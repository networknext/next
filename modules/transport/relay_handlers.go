package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/networknext/backend/modules/envvar"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
)

const (
	InitRequestMagic = 0x9083708f

	MaxRelays = 1024
)

var (
	MaxJitter float64
)

type RelayInitHandlerConfig struct {
	RelayMap         *routing.RelayMap
	Storer           storage.Storer
	Metrics          *metrics.RelayInitMetrics
	RouterPrivateKey []byte
}

type RelayUpdateHandlerConfig struct {
	RelayMap   *routing.RelayMap
	StatsDB    *routing.StatsDatabase
	Metrics    *metrics.RelayUpdateMetrics
	Storer     storage.Storer
	RB15NoInit bool
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
func RelayInitHandlerFunc(logger log.Logger, params *RelayInitHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "init")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.Metrics.Invocations.Add(1)
		}()

		locallogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(locallogger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		defer request.Body.Close()

		var relayInitRequest RelayInitRequest
		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			err = relayInitRequest.UnmarshalBinary(body)
		default:
			err = errors.New("unsupported content type")
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		locallogger = log.With(locallogger, "relay_addr", relayInitRequest.Address.String())

		if relayInitRequest.Magic != InitRequestMagic {
			level.Error(locallogger).Log("msg", "magic number mismatch", "magic_number", relayInitRequest.Magic)
			http.Error(writer, "magic number mismatch", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.InvalidMagic.Add(1)
			return
		}

		id := crypto.HashID(relayInitRequest.Address.String())
		relay, err := params.Storer.Relay(id)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to get relay from storage", "err", err)
			http.Error(writer, "failed to get relay from storage", http.StatusNotFound)
			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		if _, ok := crypto.Open(relayInitRequest.EncryptedToken, relayInitRequest.Nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
			level.Error(locallogger).Log("msg", "crypto open failed")
			http.Error(writer, "crypto open failed", http.StatusUnauthorized)
			params.Metrics.ErrorMetrics.DecryptionFailure.Add(1)
			return
		}

		params.RelayMap.Lock()
		relayData := params.RelayMap.GetRelayData(relayInitRequest.Address.String())
		params.RelayMap.Unlock()
		if relayData != nil {
			level.Error(locallogger).Log("msg", "relay already initialized")
			fmt.Printf("relay %v %v tried to reinitialized", relayData.ID, relayData.Name)
			http.Error(writer, "relay already initialized", http.StatusConflict)
			params.Metrics.ErrorMetrics.RelayAlreadyExists.Add(1)
			return
		}

		relayData, err, errCode := initRelayOnBackend(&relay, locallogger, relayInitRequest.RelayVersion, &params.Metrics.ErrorMetrics, &RelayMapAndStorer{params.RelayMap, params.Storer})
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		var responseData []byte
		response := RelayInitResponse{
			Version:   VersionNumberInitResponse,
			Timestamp: uint64(relayData.LastUpdateTime.Unix()),
			PublicKey: relayData.PublicKey,
		}

		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			responseData, err = response.MarshalBinary()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		writer.Write(responseData)
	}
}

type RelayMapAndStorer struct {
	RelayMap *routing.RelayMap
	Storer   storage.Storer
}

func initRelayOnBackend(relay *routing.Relay, logger log.Logger, relayVersion string, errorMetrics *metrics.RelayInitErrorMetrics, mapStore *RelayMapAndStorer) (*routing.RelayData, error, int) {
	relayData := routing.NewRelayData()
	// Don't allow quarantined relays back in

	if relay.State == routing.RelayStateQuarantine {
		level.Error(logger).Log("msg", "quaratined relay attempted to reconnect", "relay", relay.Name)
		errorMetrics.RelayQuarantined.Add(1)
		return relayData, fmt.Errorf("cannot permit quarantined relay"), http.StatusUnauthorized
	}

	// Set the relay's state to enabled
	relay.State = routing.RelayStateEnabled

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := mapStore.Storer.SetRelay(ctx, *relay); err != nil {
		level.Error(logger).Log("msg", "failed to set relay state in storage", "err", err)
		return relayData, fmt.Errorf("failed to set relay state in storage"), http.StatusInternalServerError

	}

	relayData.ID = relay.ID
	relayData.Name = relay.Name
	relayData.Addr = relay.Addr
	relayData.PublicKey = relay.PublicKey
	relayData.Seller = relay.Seller
	relayData.Datacenter = relay.Datacenter
	relayData.LastUpdateTime = time.Now()
	relayData.MaxSessions = relay.MaxSessions
	relayData.Version = relayVersion

	mapStore.RelayMap.Lock()
	mapStore.RelayMap.AddRelayDataEntry(relayData.Addr.String(), relayData)
	mapStore.RelayMap.Unlock()

	level.Debug(logger).Log("msg", "relay initialized")

	return relayData, nil, 0
}

// RelayUpdateHandlerFunc returns the function for the relay update endpoint
func RelayUpdateHandlerFunc(logger log.Logger, relayslogger log.Logger, params *RelayUpdateHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.Metrics.Invocations.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(handlerLogger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer request.Body.Close()

		locallogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

		var relayUpdateRequest RelayUpdateRequest
		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			err = relayUpdateRequest.UnmarshalBinary(body)
		default:
			err = errors.New("unsupported content type")
		}
		if err != nil {
			level.Error(locallogger).Log("msg", "error unmarshaling relay update request", "err", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)

			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		if len(relayUpdateRequest.PingStats) > MaxRelays {
			level.Error(locallogger).Log("msg", "max relays exceeded", "relay count", len(relayUpdateRequest.PingStats))
			http.Error(writer, "max relays exceeded", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.ExceedMaxRelays.Add(1)
			return
		}

		// If the relay does not exist in Firestore it's a ghost, ignore it
		id := crypto.HashID(relayUpdateRequest.Address.String())
		relay, err := params.Storer.Relay(id)
		if err != nil {
			level.Error(locallogger).Log("msg", "relay does not exist in Firestore (ghost)", "err", err)
			http.Error(writer, "relay does not exist in Firestore (ghost)", http.StatusNotFound)
			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		// Check if the relay state isn't set to enabled, and as a failsafe quarantine the relay
		if relay.State != routing.RelayStateEnabled && !params.RB15NoInit {
			level.Error(locallogger).Log("msg", "non-enabled relay attempting to update", "relay_name", relay.Name, "relay_address", relay.Addr.String(), "relay_state", relay.State)
			http.Error(writer, "cannot allow non-enabled relay to update", http.StatusUnauthorized)
			params.Metrics.ErrorMetrics.RelayNotEnabled.Add(1)
			return
		}

		params.RelayMap.RLock()
		relayDataReadOnly := params.RelayMap.GetRelayData(relayUpdateRequest.Address.String())
		params.RelayMap.RUnlock()

		if relayDataReadOnly == nil {
			if params.RB15NoInit {
				relayData, err, errCode := initRelayOnBackend(&relay, locallogger, relayUpdateRequest.RelayVersion, &params.Metrics.InitErrorMetrics, &RelayMapAndStorer{params.RelayMap, params.Storer})
				if err != nil {
					http.Error(writer, err.Error(), errCode)
					return
				}
				relayDataReadOnly = relayData

			} else {
				level.Warn(locallogger).Log("msg", "relay not initialized")
				http.Error(writer, "relay not initialized", http.StatusNotFound)
				params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
				return
			}
		}

		id = relayDataReadOnly.ID

		// If the relay is shutting down, set the state to maintenance if it was previously operating correctly
		if relayUpdateRequest.ShuttingDown {
			if relay.State == routing.RelayStateEnabled {
				relay.State = routing.RelayStateMaintenance
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := params.Storer.SetRelay(ctx, relay); err != nil {
				level.Error(locallogger).Log("msg", "failed to set relay state in storage while shutting down", "err", err)
				http.Error(writer, "failed to set relay state in storage while shutting down", http.StatusInternalServerError)
				return
			}

			params.RelayMap.Lock()
			params.RelayMap.RemoveRelayData(relayUpdateRequest.Address.String())
			params.RelayMap.Unlock()
			return
		}

		if !bytes.Equal(relayUpdateRequest.Token, relayDataReadOnly.PublicKey) {
			level.Error(locallogger).Log("msg", "relay public key doesn't match")
			http.Error(writer, "relay public key doesn't match", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.InvalidToken.Add(1)
			return
		}

		statsUpdate := &routing.RelayStatsUpdate{}
		statsUpdate.ID = relayDataReadOnly.ID
		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)
		params.StatsDB.ProcessStats(statsUpdate)

		relaysToPing := make([]routing.RelayPingData, 0)
		allRelayData := params.RelayMap.GetAllRelayData()

		enableInternalIPs, err := envvar.GetBool("FEATURE_ENABLE_INTERNAL_IPS", false)
		if err != nil {
			level.Error(logger).Log("msg", "unable to parse value of 'ENABLE_INTERNAL_IPS'", "err", err)
		}

		for _, v := range allRelayData {
			if v.ID != relay.ID {
				otherRelay, err := params.Storer.Relay(v.ID)
				if err != nil {
					level.Error(locallogger).Log("msg", "failed to get other relay from storage", "err", err)
					continue
				}

				if otherRelay.State == routing.RelayStateEnabled {
					var address string
					if enableInternalIPs && relay.Seller.Name == otherRelay.Seller.Name && relay.InternalAddr.String() != ":0" && otherRelay.InternalAddr.String() != ":0" {
						address = otherRelay.InternalAddr.String()
					} else {
						address = v.Addr.String()
					}
					relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(v.ID), Address: address})
				}
			}
		}

		// Update the relay data
		params.RelayMap.Lock()
		params.RelayMap.UpdateRelayDataEntry(relayUpdateRequest.Address.String(), relayUpdateRequest.TrafficStats, float32(relayUpdateRequest.CPUUsage)*100.0, float32(relayUpdateRequest.MemUsage)*100.0)
		params.RelayMap.Unlock()

		level.Debug(relayslogger).Log(
			"id", relayDataReadOnly.ID,
			"name", relayDataReadOnly.Name,
			"addr", relayDataReadOnly.Addr.String(),
			"datacenter", relayDataReadOnly.Datacenter.Name,
			"session_count", relayUpdateRequest.TrafficStats.SessionCount,
			"bytes_received", relayUpdateRequest.TrafficStats.AllRx(),
			"bytes_send", relayUpdateRequest.TrafficStats.AllTx(),
		)

		level.Debug(locallogger).Log("msg", "relay updated")

		var responseData []byte
		response := RelayUpdateResponse{}
		for _, pingData := range relaysToPing {
			response.RelaysToPing = append(response.RelaysToPing, routing.RelayPingData{
				ID:      pingData.ID,
				Address: pingData.Address,
			})
		}
		response.Timestamp = time.Now().Unix()

		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			responseData, err = response.MarshalBinary()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		writer.Write(responseData)

	}
}

func statsTable(stats map[string]map[string]routing.Stats) template.HTML {
	html := strings.Builder{}
	html.WriteString("<table>")

	names := make([]string, 0)
	for name := range stats {
		names = append(names, name)
	}

	sort.Strings(names)

	html.WriteString("<tr>")
	html.WriteString("<th>Name</th>")
	for _, name := range names {
		html.WriteString("<th>" + name + "</th>")
	}
	html.WriteString("</tr>")

	for x, a := range names {
		html.WriteString("<tr>")
		html.WriteString("<th>" + a + "</th>")

		for y, b := range names {
			if a == b || y > x {
				html.WriteString("<td>&nbsp;</td>")
				continue
			}

			RTT := stats[a][b].RTT
			Jitter := stats[a][b].Jitter
			PacketLoss := stats[a][b].PacketLoss

			rttStyle := "<div>"
			jitterStyle := "</div><div>"
			packetLossStyle := "</div><div>"

			if RTT >= 10000 {
				rttStyle = "<div style='color: red;'>"
			}
			if Jitter > MaxJitter {
				jitterStyle = "</div><div style='color: red;'>"
			}
			if PacketLoss > .001 {
				packetLossStyle = "</div><div style='color: red;'>"
			}

			html.WriteString("<td>" + rttStyle +
				fmt.Sprintf("RTT(%.0f)", RTT) + jitterStyle +
				fmt.Sprintf("Jitter(%.2f)", Jitter) + packetLossStyle +
				fmt.Sprintf("PacketLoss(%.2f)", PacketLoss) + "</div></td>")

		}

		html.WriteString("</tr>")
	}

	html.WriteString("</table>")

	return template.HTML(html.String())
}

func RelayDashboardHandlerFunc(relayMap *routing.RelayMap, GetRouteMatrix func() *routing.RouteMatrix, statsdb *routing.StatsDatabase, username string, password string, maxJitter float64) func(writer http.ResponseWriter, request *http.Request) {
	type displayRelay struct {
		ID         uint64
		Name       string
		Addr       string
		Datacenter routing.Datacenter
	}

	type response struct {
		Analysis string
		Relays   []displayRelay
		Stats    map[string]map[string]routing.Stats
	}

	MaxJitter = maxJitter

	funcmap := template.FuncMap{
		"statsTable": statsTable,
	}

	tmpl := template.Must(template.New("dashboard").Funcs(funcmap).Parse(`
		<html>
			<head>
				<title>Relay Dashboard</title>
				<style>
					body { font-family: monospace; }
					table { width: 100%; border-collapse: collapse; }
					table, th, td { padding: 3px; border: 1px solid black; }
					td { text-align: center; }
				</style>
			</head>
			<body>
				<h1>Relay Dashboard</h1>

				<h2>Route Matrix Analysis</h2>
				<pre>{{ .Analysis }}</pre>

				<h2>{{ len .Relays }} Relays</h2>
				<table>
					<tr>
						<th>Name</th>
						<th>Address</th>
						<th>Datacenter</th>
						<th>Lat / Long</th>
					</tr>
					{{ range .Relays }}
					<tr>
						<td>{{ .Name }}</td>
						<td>{{ .Addr }}</td>
						<td>{{ .Datacenter.Name }}</td>
						<td>{{ printf "%.2f" .Datacenter.Location.Latitude }} / {{ printf "%.2f" .Datacenter.Location.Longitude }}</td>
					</tr>
					{{ end }}
				</table>

				<h2>Stats</h2>
				{{ .Stats | statsTable }}
			</body>
		</html>
	`))

	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		u, p, _ := request.BasicAuth()
		if u != username && p != password {
			writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		var res response

		routeMatrix := GetRouteMatrix()
		res.Analysis = string(routeMatrix.GetAnalysisData())

		allRelayData := relayMap.GetAllRelayData()

		for _, relayData := range allRelayData {
			display := displayRelay{
				ID:   relayData.ID,
				Name: relayData.Name,
				// needs to be stringified before html,
				//otherwise braces are displayed surrounding the ip
				Addr:       relayData.Addr.String(),
				Datacenter: relayData.Datacenter,
			}
			if display.Name == "" {
				display.Name = display.Addr
			}
			res.Relays = append(res.Relays, display)
		}

		sort.Slice(res.Relays, func(i int, j int) bool {
			return res.Relays[i].Name < res.Relays[j].Name
		})

		res.Stats = make(map[string]map[string]routing.Stats)
		for _, a := range res.Relays {
			aKey := a.Name
			if aKey == "" {
				aKey = a.Addr
			}

			res.Stats[aKey] = make(map[string]routing.Stats)

			for _, b := range res.Relays {
				bKey := b.Name
				if bKey == "" {
					bKey = b.Addr
				}

				rtt, jitter, packetloss := statsdb.GetSample(a.ID, b.ID)
				res.Stats[aKey][bKey] = routing.Stats{RTT: float64(rtt), Jitter: float64(jitter), PacketLoss: float64(packetloss)}
			}
		}

		if err := tmpl.Execute(writer, res); err != nil {
			fmt.Println(err)
		}
	}
}

func RelayStatsFunc(logger log.Logger, rmap *routing.RelayMap) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, request *http.Request) {
		if bin, err := rmap.MarshalBinary(); err == nil {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(bin)))
			w.WriteHeader(http.StatusOK)
			w.Write(bin)
		} else {
			level.Error(logger).Log("msg", "could not marshal relay map", "err", err)
			http.Error(w, "could not marshal relay map", http.StatusInternalServerError)
		}
	}
}
