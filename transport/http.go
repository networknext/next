package transport

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"expvar"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/stats"
	"github.com/networknext/backend/storage"
)

const (
	InitRequestMagic = 0x9083708f

	MaxRelays             = 1024
	MaxRelayAddressLength = 256
)

type CommonRelayInitFuncParams struct {
	RedisClient      *redis.Client
	GeoClient        *routing.GeoClient
	IpLocator        routing.IPLocator
	Storer           storage.Storer
	Duration         metrics.Gauge
	Counter          metrics.Counter
	RouterPrivateKey []byte
}

type CommonRelayUpdateFuncParams struct {
	RedisClient           *redis.Client
	StatsDb               *routing.StatsDatabase
	Duration              metrics.Gauge
	Counter               metrics.Counter
	TrafficStatsPublisher stats.Publisher
	Storer                storage.Storer
}

// NewRouter creates a router with the specified endpoints
func NewRouter(logger log.Logger, redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, storer storage.Storer,
	statsdb *routing.StatsDatabase, initDuration metrics.Gauge, updateDuration metrics.Gauge, initCounter metrics.Counter,
	updateCounter metrics.Counter, costmatrix *routing.CostMatrix, routematrix *routing.RouteMatrix, routerPrivateKey []byte, trafficStatsPublisher stats.Publisher,
	username string, password string) *mux.Router {

	commonInitParams := CommonRelayInitFuncParams{
		RedisClient:      redisClient,
		GeoClient:        geoClient,
		IpLocator:        ipLocator,
		Storer:           storer,
		Duration:         initDuration,
		Counter:          initCounter,
		RouterPrivateKey: routerPrivateKey,
	}

	commonUpdateParams := CommonRelayUpdateFuncParams{
		RedisClient:           redisClient,
		StatsDb:               statsdb,
		Duration:              updateDuration,
		Counter:               updateCounter,
		TrafficStatsPublisher: trafficStatsPublisher,
		Storer:                storer,
	}

	router := mux.NewRouter()
	router.HandleFunc("/healthz", HealthzHandlerFunc())
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandlerFunc(logger, &commonUpdateParams)).Methods("POST")
	router.Handle("/cost_matrix", costmatrix).Methods("GET")
	router.Handle("/route_matrix", routematrix).Methods("GET")
	router.HandleFunc("/", RelayDashboardHandlerFunc(redisClient, routematrix, statsdb, username, password)).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())
	return router
}

/******************************* Init Handler *******************************/

// RelayInitHandlerFunc returns the function for the relay init endpoint
func RelayInitHandlerFunc(logger log.Logger, params *CommonRelayInitFuncParams) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "init")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Duration.Set(float64(durationSince.Milliseconds()))
			params.Counter.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(handlerLogger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		requestLogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

		contentTypes := request.Header["Content-Type"]
		if contentTypes != nil && len(contentTypes) > 0 {
			switch contentTypes[0] {
			case "application/json":
				relayInitJSON(requestLogger, writer, body, params)
			case "application/octet-stream":
				relayInitOctetStream(requestLogger, writer, body, params)
			}
		}

	}
}

func relayInitJSON(logger log.Logger, writer http.ResponseWriter, body []byte, params *CommonRelayInitFuncParams) {
	var jsonPacket RelayInitRequestJSON
	if err := json.Unmarshal(body, &jsonPacket); err != nil {
		level.Error(logger).Log("msg", "could not parse init json", "err", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var packet RelayInitPacket
	if err := jsonPacket.ToInitPacket(&packet); err != nil {
		level.Error(logger).Log("msg", "could not convert json to init packet", "err", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	relay, statusCode := relayInitHandler(logger, &packet, params)

	writer.WriteHeader(statusCode)

	if relay == nil || statusCode != http.StatusOK {
		level.Error(logger).Log("msg", "could not process relay init in json handler")
		return
	}

	var response RelayInitResponseJSON
	response.Timestamp = relay.LastUpdateTime * 1000 // convert to millis, this is what the curr prod relay expects, the new relay uses seconds so it convertes back

	if responseData, err := json.Marshal(response); err != nil {
		level.Error(logger).Log("msg", "could not marshal init json response", "err", err)
		writer.WriteHeader(http.StatusBadRequest)
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(responseData)
	}
}

func relayInitOctetStream(logger log.Logger, writer http.ResponseWriter, body []byte, params *CommonRelayInitFuncParams) {
	relayInitPacket := RelayInitPacket{}
	if err := relayInitPacket.UnmarshalBinary(body); err != nil {
		level.Error(logger).Log("msg", "could not unmarshal init packet", "err", err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	relay, statusCode := relayInitHandler(logger, &relayInitPacket, params)

	writer.WriteHeader(statusCode)

	if relay == nil || statusCode != http.StatusOK {
		level.Error(logger).Log("msg", "could not process relay init in octet-stream handler")
		return
	}

	index := 0
	var responseData [PacketSizeRelayInitResponse]byte
	{
		encoding.WriteUint32(responseData[:], &index, VersionNumberInitResponse)
		encoding.WriteUint64(responseData[:], &index, relay.LastUpdateTime)
		encoding.WriteBytes(responseData[:], &index, relay.PublicKey, crypto.KeySize)
	}

	writer.Header().Set("Content-Type", "application/octet-stream")

	writer.Write(responseData[:])
}

func relayInitHandler(logger log.Logger, relayInitPacket *RelayInitPacket, params *CommonRelayInitFuncParams) (*routing.Relay, int) {
	locallogger := log.With(logger, "relay_addr", relayInitPacket.Address.String())

	if relayInitPacket.Magic != InitRequestMagic {
		level.Error(locallogger).Log("msg", "magic number mismatch", "magic_number", relayInitPacket.Magic)
		return nil, http.StatusBadRequest
	}

	if relayInitPacket.Version != VersionNumberInitRequest {
		level.Error(locallogger).Log("msg", "version mismatch", "version", relayInitPacket.Version)
		return nil, http.StatusBadRequest
	}

	id := crypto.HashID(relayInitPacket.Address.String())

	relayEntry, ok := params.Storer.Relay(id)
	if !ok {
		level.Error(locallogger).Log("msg", "relay not in firestore")
		return nil, http.StatusInternalServerError
	}

	relay := routing.Relay{
		ID:             id,
		Addr:           relayInitPacket.Address,
		PublicKey:      relayEntry.PublicKey,
		Datacenter:     relayEntry.Datacenter,
		Seller:         relayEntry.Seller,
		LastUpdateTime: uint64(time.Now().Unix()),
		Latitude:       relayEntry.Latitude,
		Longitude:      relayEntry.Longitude,
	}

	if _, ok := crypto.Open(relayInitPacket.EncryptedToken, relayInitPacket.Nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
		level.Error(locallogger).Log("msg", "crypto open failed")
		return nil, http.StatusUnauthorized
	}

	exists := params.RedisClient.HExists(routing.HashKeyAllRelays, relay.Key())

	if exists.Err() != nil && exists.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
		return nil, http.StatusNotFound
	}

	if exists.Val() {
		level.Warn(locallogger).Log("msg", "relay already initialized")
		return nil, http.StatusConflict
	}

	if loc, err := params.IpLocator.LocateIP(relay.Addr.IP); err == nil {
		relay.Latitude = loc.Latitude
		relay.Longitude = loc.Longitude
	} else {
		level.Warn(locallogger).Log("msg", "using default geolocation from storage for relay")
	}

	// Regular set for expiry
	if res := params.RedisClient.Set(relay.Key(), relay.ID, routing.RelayTimeout); res.Err() != nil && res.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to initialize relay", "err", res.Err())
		return nil, http.StatusInternalServerError
	}

	// HSet for full relay data
	if res := params.RedisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil && res.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to initialize relay", "err", res.Err())
		return nil, http.StatusInternalServerError
	}

	if err := params.GeoClient.Add(relay); err != nil {
		level.Error(locallogger).Log("msg", "failed to initialize relay", "err", err)
		return nil, http.StatusInternalServerError
	}

	level.Debug(locallogger).Log("msg", "relay initialized")

	return &relay, http.StatusOK
}

/******************************* Update Handler *******************************/

// RelayUpdateHandlerFunc returns the function for the relay update endpoint
func RelayUpdateHandlerFunc(logger log.Logger, params *CommonRelayUpdateFuncParams) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Duration.Set(float64(durationSince.Milliseconds()))
			params.Counter.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		requestLogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

		contentTypes := request.Header["Content-Type"]
		if contentTypes != nil && len(contentTypes) > 0 {
			switch contentTypes[0] {
			case "application/json":
				relayUpdateJSON(requestLogger, writer, body, params)
			case "application/octet-stream":
				relayUpdateOctetStream(requestLogger, writer, body, params)
			}
		}
	}
}

func relayUpdateJSON(logger log.Logger, writer http.ResponseWriter, body []byte, params *CommonRelayUpdateFuncParams) {
	var jsonPacket RelayUpdateRequestJSON
	if err := json.Unmarshal(body, &jsonPacket); err != nil {
		level.Error(logger).Log("msg", "could not parse update json", "err", err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var packet RelayUpdatePacket
	if err := jsonPacket.ToUpdatePacket(&packet); err != nil {
		level.Error(logger).Log("msg", "could not convert json to update packet", "err", err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	relaysToPing, statusCode := relayUpdateHandler(logger, &packet, params)

	writer.WriteHeader(statusCode)

	if relaysToPing == nil || statusCode != http.StatusOK {
		level.Error(logger).Log("msg", "could not process relay update in json handler")
		return
	}

	var response RelayUpdateResponseJSON
	for _, pingData := range relaysToPing {
		var token routing.LegacyPingToken
		token.Timeout = uint64(time.Now().Unix() * 100000) // some arbitrarily large number just to make things compatable for the moment
		token.RelayID = crypto.HashID(jsonPacket.StringAddr)
		bin, _ := token.MarshalBinary()
		var legacy routing.LegacyPingData
		legacy.ID = pingData.ID
		legacy.Address = pingData.Address
		legacy.PingToken = base64.StdEncoding.EncodeToString(bin)
		response.RelaysToPing = append(response.RelaysToPing, legacy)
	}

	if responseData, err := json.Marshal(response); err != nil {
		level.Error(logger).Log("msg", "could not marshal json update response", "err", err)
		writer.WriteHeader(http.StatusBadRequest)
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(responseData)
	}

	if ts, err := ptypes.TimestampProto(time.Unix(int64(jsonPacket.Timestamp), 0)); err == nil {
		// can find the relay based on its address hash in firestore
		if relay, ok := params.Storer.Relay(crypto.HashID(jsonPacket.StringAddr)); ok {
			stats := &stats.RelayTrafficStats{
				RelayId:            stats.NewEntityID("Relay", jsonPacket.RelayName), // use its name here
				Usage:              jsonPacket.Usage,
				Timestamp:          ts,
				BytesPaidTx:        jsonPacket.TrafficStats.BytesPaidTx,
				BytesPaidRx:        jsonPacket.TrafficStats.BytesPaidRx,
				BytesManagementTx:  jsonPacket.TrafficStats.BytesManagementTx,
				BytesManagementRx:  jsonPacket.TrafficStats.BytesManagementRx,
				BytesMeasurementTx: jsonPacket.TrafficStats.BytesMeasurementTx,
				BytesMeasurementRx: jsonPacket.TrafficStats.BytesMeasurementRx,
				BytesInvalidRx:     jsonPacket.TrafficStats.BytesInvalidRx,
				SessionCount:       jsonPacket.TrafficStats.SessionCount,
			}

			str, _ := json.Marshal(stats)
			level.Debug(logger).Log("msg", fmt.Sprintf("Publishing: %s", str))
			if err := params.TrafficStatsPublisher.Publish(context.Background(), relay.ID, stats); err != nil {
				level.Error(logger).Log("msg", fmt.Sprintf("Publish error: %v", err))
			}
		} else {
			level.Error(logger).Log("msg", fmt.Sprintf("TrafficStats, cannot lookup relay in firestore, %d", jsonPacket.Metadata.ID))
		}
	} else {
		level.Error(logger).Log("msg", fmt.Sprintf("Can't publish to pubsub. Timestamp error: %v", err))
	}
}

func relayUpdateOctetStream(logger log.Logger, writer http.ResponseWriter, body []byte, params *CommonRelayUpdateFuncParams) {
	relayUpdatePacket := RelayUpdatePacket{}
	if err := relayUpdatePacket.UnmarshalBinary(body); err != nil {
		level.Error(logger).Log("msg", "could not unmarshal update packet", "err", err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	relaysToPing, statusCode := relayUpdateHandler(logger, &relayUpdatePacket, params)

	writer.WriteHeader(statusCode)

	if relaysToPing == nil || statusCode != http.StatusOK {
		level.Error(logger).Log("msg", "could not process relay update in octet-stream handler")
		return
	}

	index := 0
	responseData := make([]byte, 8+(8+MaxRelayAddressLength)*len(relaysToPing))
	{
		encoding.WriteUint32(responseData, &index, VersionNumberUpdateResponse)
		encoding.WriteUint32(responseData, &index, uint32(len(relaysToPing)))

		for i := range relaysToPing {
			encoding.WriteUint64(responseData, &index, relaysToPing[i].ID)
			encoding.WriteString(responseData, &index, relaysToPing[i].Address, MaxRelayAddressLength)
		}
	}

	writer.Header().Set("Content-Type", "application/octet-stream")

	writer.Write(responseData[:index])

	relayID := crypto.HashID(relayUpdatePacket.Address.String())
	if relay, ok := params.Storer.Relay(relayID); ok {
		stats := &stats.RelayTrafficStats{
			RelayId:            stats.NewEntityID("Relay", relay.Addr.String()), // TODO Until the db is fixed up, this needs to be the relay's firestore id, not its address
			BytesMeasurementRx: relayUpdatePacket.BytesReceived,
		}

		if err := params.TrafficStatsPublisher.Publish(context.Background(), relay.ID, stats); err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("Publish error: %v", err))
		}
	} else {
		level.Error(logger).Log("msg", fmt.Sprintf("TrafficStats, cannot lookup relay in firestore, %d", relayID))
		return
	}
}

func relayUpdateHandler(logger log.Logger, relayUpdatePacket *RelayUpdatePacket, params *CommonRelayUpdateFuncParams) ([]routing.RelayPingData, int) {
	locallogger := log.With(logger, "relay_addr", relayUpdatePacket.Address.String())

	if relayUpdatePacket.Version != VersionNumberUpdateRequest || relayUpdatePacket.NumRelays > MaxRelays {
		level.Error(locallogger).Log("msg", "version mismatch", "version", relayUpdatePacket.Version)
		return nil, http.StatusBadRequest
	}

	relay := routing.Relay{
		ID: crypto.HashID(relayUpdatePacket.Address.String()),
	}

	exists := params.RedisClient.HExists(routing.HashKeyAllRelays, relay.Key())

	if exists.Err() != nil && exists.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
		return nil, http.StatusInternalServerError
	}

	if !exists.Val() {
		level.Warn(locallogger).Log("msg", "relay not initialized")
		return nil, http.StatusNotFound
	}

	hgetResult := params.RedisClient.HGet(routing.HashKeyAllRelays, relay.Key())
	if hgetResult.Err() != nil && hgetResult.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to get relays", "err", exists.Err())
		return nil, http.StatusNotFound
	}

	data, err := hgetResult.Bytes()
	if err != nil && err != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to get relay data", "err", err)
		return nil, http.StatusInternalServerError
	}

	if err = relay.UnmarshalBinary(data); err != nil {
		level.Error(locallogger).Log("msg", "failed to unmarshal relay data", "err", err)
		return nil, http.StatusBadRequest
	}

	if !bytes.Equal(relayUpdatePacket.Token, relay.PublicKey) {
		level.Error(locallogger).Log("msg", "relay public key doesn't match")
		return nil, http.StatusBadRequest
	}

	statsUpdate := &routing.RelayStatsUpdate{}
	statsUpdate.ID = relay.ID
	statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdatePacket.PingStats...)

	params.StatsDb.ProcessStats(statsUpdate)

	relay.LastUpdateTime = uint64(time.Now().Unix())

	relaysToPing := make([]routing.RelayPingData, 0)

	// Regular set for expiry
	if res := params.RedisClient.Set(relay.Key(), 0, routing.RelayTimeout); res.Err() != nil {
		level.Error(locallogger).Log("msg", "failed to store relay update expiry", "err", res.Err())
		return nil, http.StatusInternalServerError
	}

	// HSet for full relay data
	if res := params.RedisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil {
		level.Error(locallogger).Log("msg", "failed to store relay update", "err", res.Err())
		return nil, http.StatusInternalServerError
	}

	hgetallResult := params.RedisClient.HGetAll(routing.HashKeyAllRelays)
	if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to get other relays", "err", hgetallResult.Err())
		return nil, http.StatusNotFound
	}

	for k, v := range hgetallResult.Val() {
		if k != relay.Key() {
			var unmarshaledValue routing.Relay
			if err := unmarshaledValue.UnmarshalBinary([]byte(v)); err != nil {
				level.Error(locallogger).Log("msg", "failed to get other relay", "err", err)
				continue
			}
			relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(unmarshaledValue.ID), Address: unmarshaledValue.Addr.String()})
		}
	}

	level.Debug(locallogger).Log("msg", "relay updated")

	return relaysToPing, http.StatusOK
}

func HealthzHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func statsTable(stats map[string]map[string]routing.Stats) template.HTML {
	html := strings.Builder{}
	html.WriteString("<table>")

	addrs := make([]string, 0)
	for addr := range stats {
		addrs = append(addrs, addr)
	}

	html.WriteString("<tr>")
	html.WriteString("<th>Address</th>")
	for _, addr := range addrs {
		html.WriteString("<th>" + addr + "</th>")
	}
	html.WriteString("</tr>")

	for _, a := range addrs {
		html.WriteString("<tr>")
		html.WriteString("<th>" + a + "</th>")

		for _, b := range addrs {
			if a == b {
				html.WriteString("<td>&nbsp;</td>")
				continue
			}

			html.WriteString("<td>" + stats[a][b].String() + "</td>")
		}

		html.WriteString("</tr>")
	}

	html.WriteString("</table>")

	return template.HTML(html.String())
}

// NearHandlerFunc returns the function for the near endpoint
func RelayDashboardHandlerFunc(redisClient *redis.Client, routeMatrix *routing.RouteMatrix, statsdb *routing.StatsDatabase, username string, password string) func(writer http.ResponseWriter, request *http.Request) {
	type response struct {
		Analysis string
		Relays   []routing.Relay
		Stats    map[string]map[string]routing.Stats
	}

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

				<h2>Relays</h2>
				<table>
					<tr>
						<th>Address</th>
						<th>Datacenter</th>
						<th>Lat / Long</th>
						<th>Seller</th>
						<th>Ingress / Egress</th>
					</tr>
					{{ range .Relays }}
					<tr>
						<td>{{ .Addr }}</td>
						<td>{{ .Datacenter.Name }}</td>
						<td>{{ .Latitude }} / {{ .Longitude }}</td>
						<td>{{ .Seller.Name }}</td>
						<td>{{ .Seller.IngressPriceCents }} / {{ .Seller.EgressPriceCents }}</td>
					</tr>
					{{ end }}
				</table>

				<h2>Stats</h2>
				{{ .Stats | statsTable }}
			</body>
		</html>
	`))

	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		u, p, _ := request.BasicAuth()
		if u != username && p != password {
			writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		var res response

		buf := bytes.Buffer{}

		routeMatrix.WriteAnalysisTo(&buf)
		res.Analysis = buf.String()

		hgetallResult := redisClient.HGetAll(routing.HashKeyAllRelays)
		if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
			fmt.Println(hgetallResult.Err())
			return
		}

		for _, rawRelay := range hgetallResult.Val() {
			var relay routing.Relay
			if err := relay.UnmarshalBinary([]byte(rawRelay)); err != nil {
				fmt.Println(err)
				return
			}
			res.Relays = append(res.Relays, relay)
		}

		res.Stats = make(map[string]map[string]routing.Stats)
		for _, a := range res.Relays {
			res.Stats[a.Addr.String()] = make(map[string]routing.Stats)

			for _, b := range res.Relays {
				rtt, jitter, packetloss := statsdb.GetSample(a.ID, b.ID)
				res.Stats[a.Addr.String()][b.Addr.String()] = routing.Stats{RTT: float64(rtt), Jitter: float64(jitter), PacketLoss: float64(packetloss)}
			}
		}

		if err := tmpl.Execute(writer, res); err != nil {
			fmt.Println(err)
		}
	}
}
