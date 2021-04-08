package transport

// update for merge

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/crypto"
)

const InitRequestMagic = 0x9083708f

const MaxRelays = 1024

var (
	MaxJitter float64
)

type RelayUpdateHandlerConfig struct {
	RelayMap     *routing.RelayMap
	StatsDB      *routing.StatsDatabase
	Metrics      *metrics.RelayUpdateMetrics
	GetRelayData func() ([]routing.Relay, map[uint64]routing.Relay)
}

func RelayUpdateHandlerFunc(params *RelayUpdateHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.Metrics.Invocations.Add(1)
		}()

		core.Debug("%s - relay update", request.RemoteAddr)

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			core.Debug("%s - error: relay update could not read request body: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		if request.Header.Get("Content-Type") != "application/octet-stream" {
			core.Debug("%s - error: relay update unsupported content type", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		var relayUpdateRequest RelayUpdateRequest
		err = relayUpdateRequest.UnmarshalBinary(body)
		if err != nil {
			core.Debug("%s - error: relay update could not read request packet", request.RemoteAddr)
			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if relayUpdateRequest.Version > VersionNumberUpdateRequest {
			core.Debug("%s - error: relay update version mismatch: %d > %d", request.RemoteAddr, relayUpdateRequest.Version, VersionNumberUpdateRequest)
			params.Metrics.ErrorMetrics.InvalidVersion.Add(1)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if len(relayUpdateRequest.PingStats) > MaxRelays {
			core.Debug("%s - error: relay update too many relays in ping stats: %d > %d", request.RemoteAddr, relayUpdateRequest.PingStats, MaxRelays)
			params.Metrics.ErrorMetrics.ExceedMaxRelays.Add(1)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// check if relay exists

		relayArray, relayHash := params.GetRelayData()

		id := crypto.HashID(relayUpdateRequest.Address.String())

		relay, ok := relayHash[id]

		if !ok {
			core.Debug("%s - error: could not find relay: %s [%x]", request.RemoteAddr, relayUpdateRequest.Address.String(), id)
			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
			writer.WriteHeader(http.StatusNotFound) // 404
			return
		}		

		// todo: bring back crypto check

		// update relay data

		relayData := routing.RelayData{}

		relayData.ID = id
		relayData.Addr = relayUpdateRequest.Address
		relayData.LastUpdateTime = time.Now()
		relayData.Name = relay.Name
		relayData.PublicKey = relay.PublicKey
		relayData.MaxSessions = relay.MaxSessions
		relayData.SessionCount = int(relayUpdateRequest.TrafficStats.SessionCount)
		relayData.Version = relayUpdateRequest.RelayVersion

		params.RelayMap.Lock()
		params.RelayMap.UpdateRelayData(relayData)
		params.RelayMap.Unlock()

		// update relay ping stats

		statsUpdate := routing.RelayStatsUpdate{}

		statsUpdate.ID = relayData.ID

		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)

		params.StatsDB.ProcessStats(&statsUpdate)

		// get relays to ping

		relaysToPing := make([]routing.RelayPingData, 0)

		sellerName := relayHash[relayData.ID].Seller.Name

		for i := range relayArray {

			if relayArray[i].ID == relayData.ID {
				continue
			}

			var address string
			if sellerName == relayArray[i].Seller.Name && relayArray[i].InternalAddr.String() != ":0" {
				address = relayArray[i].InternalAddr.String()
			} else {
				address = relayArray[i].Addr.String()
			}

			relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(relayArray[i].ID), Address: address})
		}

		// build and write the response

		var responseData []byte

		response := RelayUpdateResponse{}

		for i := range relaysToPing {
			response.RelaysToPing = append(response.RelaysToPing, routing.RelayPingData{
				ID:      relaysToPing[i].ID,
				Address: relaysToPing[i].Address,
			})
		}

		response.Timestamp = time.Now().Unix()

		response.TargetVersion = "2.0.0"

		responseData, err = response.MarshalBinary()
		if err != nil {
			core.Debug("%s - error: failed to write relay update response: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		writer.Write(responseData)

		// core.Debug("%s - wrote relay update response", request.RemoteAddr)
	}
}

// func RelayUpdatePubSubFunc(requestBody []byte, logger log.Logger, params *RelayUpdateHandlerConfig) {
// 	durationStart := time.Now()
// 	defer func() {
// 		durationSince := time.Since(durationStart)
// 		params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
// 		params.Metrics.Invocations.Add(1)
// 	}()

// 	var relayUpdateRequest RelayUpdateRequest
// 	err := relayUpdateRequest.UnmarshalBinary(requestBody)
// 	if err != nil {
// 		_ = level.Error(logger).Log("msg", "error unmarshaling relay update request", "err", err)
// 		params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
// 		return
// 	}

// 	// If the relay does not exist in Firestore it's a ghost, ignore it
// 	id := crypto.HashID(relayUpdateRequest.Address.String())

// 	var relay routing.Relay
// 	if !params.LoadTest {
// 		// If the relay does not exist in Firestore it's a ghost, ignore it
// 		relay, err = params.Storer.Relay(id)
// 		if err != nil {
// 			level.Error(logger).Log("msg", "relay does not exist in Firestore (ghost)", "err", err)
// 			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
// 			return
// 		}
// 	} else {
// 		relay = loadTestRelay(relayUpdateRequest.Address.String(), routing.RelayStateEnabled)
// 	}

// 	if relayUpdateRequest.ShuttingDown {
// 		params.RelayMap.Lock()
// 		params.RelayMap.RemoveRelayData(relayUpdateRequest.Address.String())
// 		params.RelayMap.Unlock()
// 		return
// 	}

// 	params.RelayMap.RLock()
// 	relayDataReadOnly := params.RelayMap.GetRelayData(relayUpdateRequest.Address.String())
// 	params.RelayMap.RUnlock()

// 	if relayDataReadOnly == nil {
// 		relayDataReadOnly = initRelayOnBackend(&relay, relayUpdateRequest, logger, relayUpdateRequest.RelayVersion, &params.Metrics.InitErrorMetrics, params.RelayMap)
// 	}

// 	statsUpdate := &routing.RelayStatsUpdate{}
// 	statsUpdate.ID = id
// 	statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)
// 	params.StatsDB.ProcessStats(statsUpdate)

// 	// Update the relay data
// 	params.RelayMap.Lock()
// 	params.RelayMap.UpdateRelayDataEntry(relayUpdateRequest.Address.String(), relayUpdateRequest.TrafficStats, float32(relayUpdateRequest.CPUUsage)*100.0, float32(relayUpdateRequest.MemUsage)*100.0)
// 	params.RelayMap.Unlock()
// }

// func initRelayOnBackend(relay *routing.Relay, relayUpdateRequest RelayUpdateRequest, logger log.Logger, relayVersion string, errorMetrics *metrics.RelayInitErrorMetrics, relayMap *routing.RelayMap) *routing.RelayData {
// 	relayData := routing.NewRelayData()
// 	relayData.ID = crypto.HashID(relayUpdateRequest.Address.String())
// 	relayData.Name = relay.Name
// 	relayData.Addr = relayUpdateRequest.Address
// 	relayData.PublicKey = relay.PublicKey
// 	relayData.Seller = relay.Seller
// 	relayData.Datacenter = relay.Datacenter
// 	relayData.LastUpdateTime = time.Now()
// 	relayData.MaxSessions = relay.MaxSessions
// 	relayData.Version = relayVersion

// 	relayMap.Lock()
// 	relayMap.AddRelayDataEntry(relayData.Addr.String(), relayData)
// 	relayMap.Unlock()

// 	level.Debug(logger).Log("msg", "relay initialized")

// 	return relayData
// }

// NewRelayUpdateHandlerFunc returns the function for the new relay backend update endpoint
// func NewRelayUpdateHandlerFunc(logger log.Logger, relayslogger log.Logger, params *RelayUpdateHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
// 	handlerLogger := log.With(logger, "handler", "update")

// 	return func(writer http.ResponseWriter, request *http.Request) {

// 		durationStart := time.Now()
// 		defer func() {
// 			durationSince := time.Since(durationStart)
// 			params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
// 			params.Metrics.Invocations.Add(1)
// 		}()

// 		body, err := ioutil.ReadAll(request.Body)
// 		if err != nil {
// 			level.Error(handlerLogger).Log("msg", "could not read packet", "err", err)
// 			writer.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		request.Body.Close()

// 		locallogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

// 		var relayUpdateRequest RelayUpdateRequest
// 		switch request.Header.Get("Content-Type") {
// 		case "application/octet-stream":
// 			err = relayUpdateRequest.UnmarshalBinary(body)
// 		default:
// 			err = errors.New("unsupported content type")
// 		}
// 		if err != nil {
// 			level.Error(locallogger).Log("msg", "error unmarshaling relay update request", "err", err)
// 			http.Error(writer, err.Error(), http.StatusBadRequest)
// 			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
// 			return
// 		}

// 		// If the relay does not exist in Firestore it's a ghost, ignore it
// 		id := crypto.HashID(relayUpdateRequest.Address.String())

// 		var relay routing.Relay
// 		if !params.LoadTest {
// 			relay, err = params.Storer.Relay(id)
// 			if err != nil {
// 				level.Error(locallogger).Log("msg", "relay does not exist in Firestore (ghost)", "err", err)
// 				http.Error(writer, "relay does not exist in Firestore (ghost)", http.StatusNotFound)
// 				params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
// 				return
// 			}
// 		} else {
// 			relay = loadTestRelay(relayUpdateRequest.Address.String(), routing.RelayStateEnabled)
// 		}

// 		if !bytes.Equal(relayUpdateRequest.Token, relay.PublicKey) {
// 			level.Error(locallogger).Log("msg", "relay public key doesn't match")
// 			http.Error(writer, "relay public key doesn't match", http.StatusBadRequest)
// 			params.Metrics.ErrorMetrics.InvalidToken.Add(1)
// 			return
// 		}

// 		// If the relay is shutting down, set the state to maintenance if it was previously operating correctly
// 		if relayUpdateRequest.ShuttingDown {
// 			params.RelayMap.Lock()
// 			params.RelayMap.RemoveRelayData(relayUpdateRequest.Address.String())
// 			params.RelayMap.Unlock()
// 			return
// 		}

// 		params.RelayMap.RLock()
// 		relayDataReadOnly := params.RelayMap.GetRelayData(relayUpdateRequest.Address.String())
// 		params.RelayMap.RUnlock()

// 		if relayDataReadOnly == nil {
// 			relayDataReadOnly = initRelayOnBackend(&relay, relayUpdateRequest, locallogger, relayUpdateRequest.RelayVersion, &params.Metrics.InitErrorMetrics, params.RelayMap)
// 		}

// 		statsUpdate := &routing.RelayStatsUpdate{}
// 		statsUpdate.ID = relayDataReadOnly.ID
// 		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)
// 		params.StatsDB.ProcessStats(statsUpdate)

// 		// Update the relay data
// 		params.RelayMap.Lock()
// 		params.RelayMap.UpdateRelayDataEntry(relayUpdateRequest.Address.String(), relayUpdateRequest.TrafficStats, float32(relayUpdateRequest.CPUUsage)*100.0, float32(relayUpdateRequest.MemUsage)*100.0)
// 		params.RelayMap.Unlock()

// 		level.Debug(relayslogger).Log(
// 			"id", relayDataReadOnly.ID,
// 			"name", relayDataReadOnly.Name,
// 			"addr", relayDataReadOnly.Addr.String(),
// 			"datacenter", relayDataReadOnly.Datacenter.Name,
// 			"session_count", relayUpdateRequest.TrafficStats.SessionCount,
// 			"bytes_received", relayUpdateRequest.TrafficStats.AllRx(),
// 			"bytes_send", relayUpdateRequest.TrafficStats.AllTx(),
// 		)

// 		level.Debug(locallogger).Log("msg", "relay updated")

// 		var responseData []byte
// 		response := RelayUpdateResponse{}
// 		response.RelaysToPing = make([]routing.RelayPingData, 0)
// 		response.Timestamp = time.Now().Unix()

// 		switch request.Header.Get("Content-Type") {
// 		case "application/octet-stream":
// 			responseData, err = response.MarshalBinary()
// 			if err != nil {
// 				writer.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}
// 		}

// 		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
// 		writer.Write(responseData)

// 		return
// 	}
// }

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
