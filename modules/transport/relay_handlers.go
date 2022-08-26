package transport

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
)

var (
	MaxJitter float64
)

type RelayUpdateHandlerConfig struct {
	RelayMap     *routing.RelayMap
	StatsDB      *routing.StatsDatabase
	Metrics      *metrics.RelayUpdateMetrics
	GetRelayData func() ([]routing.Relay, map[uint64]routing.Relay)
}

/*
RelayUpdateHandlerFunc() receives batched HTTP relay update requests from the Relay Gateway.
It unbatches the requests and processes them individually, recording the relay ping stats
and relay data in the StatsDB and relay map.
*/
func RelayUpdateHandlerFunc(params *RelayUpdateHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		core.Debug("%s - relay update", request.RemoteAddr)

		// Get the batched updates
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			core.Error("%s: relay update could not read request body: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		if request.Header.Get("Content-Type") != "application/octet-stream" {
			core.Error("%s: relay update unsupported content type", request.RemoteAddr)
			params.Metrics.ErrorMetrics.ContentTypeFailure.Add(1)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// Unbatch the updates
		updates, err := unbatchRelayUpdates(body)
		if err != nil {
			core.Error("%s: relay update could not unbatch relay updates: %v", request.RemoteAddr, err)
			params.Metrics.ErrorMetrics.UnbatchFailure.Add(1)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		for i := range updates {
			// Use anonymous function to allow for defers to complete
			func() {
				durationStart := time.Now()
				defer func() {
					durationSince := time.Since(durationStart)
					params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
					params.Metrics.Invocations.Add(1)
				}()

				var relayUpdateRequest RelayUpdateRequest
				if err = relayUpdateRequest.UnmarshalBinary(updates[i]); err != nil {
					core.Error("%s: relay update could not read request packet", request.RemoteAddr)
					params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
					return
				}

				// Get the relay's ID
				_, relayHash := params.GetRelayData()
				id := crypto.HashID(relayUpdateRequest.Address.String())

				relay, ok := relayHash[id]
				if !ok {
					// If we could not find the relay, skip it and move on
					core.Error("%s: could not find relay: %s [%x]", request.RemoteAddr, relayUpdateRequest.Address.String(), id)
					params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
					return
				}

				// Update relay data
				relayData := routing.RelayData{}

				relayData.ID = id
				relayData.Addr = relayUpdateRequest.Address
				relayData.LastUpdateTime = time.Now()
				relayData.Name = relay.Name
				relayData.PublicKey = relay.PublicKey
				relayData.MaxSessions = relay.MaxSessions
				relayData.SessionCount = int(relayUpdateRequest.SessionCount)
				relayData.ShuttingDown = relayUpdateRequest.ShuttingDown
				relayData.Version = relayUpdateRequest.RelayVersion
				relayData.CPU = relayUpdateRequest.CPU
				relayData.NICSpeedMbps = relay.NICSpeedMbps
				relayData.MaxBandwidthMbps = relay.MaxBandwidthMbps

				// Envelope Up/Down and Bandwidth Sent/Recv are sent by the Relay in kbps
				relayData.EnvelopeUpMbps = float32(float64(relayUpdateRequest.EnvelopeUpKbps) / 1000.0)
				relayData.EnvelopeDownMbps = float32(float64(relayUpdateRequest.EnvelopeDownKbps) / 1000.0)

				relayData.BandwidthSentMbps = float32(float64(relayUpdateRequest.BandwidthSentKbps) / 1000.0)
				relayData.BandwidthRecvMbps = float32(float64(relayUpdateRequest.BandwidthRecvKbps) / 1000.0)

				params.RelayMap.Lock()
				params.RelayMap.UpdateRelayData(relayData)
				params.RelayMap.Unlock()

				// Update relay ping stats
				statsUpdate := routing.RelayStatsUpdate{}

				statsUpdate.ID = relayData.ID
				statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)
				params.StatsDB.ProcessStats(&statsUpdate)

				// core.Debug("%s - processed relay update", relayData.Addr)
			}()
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		writer.WriteHeader(http.StatusOK) // 200
	}
}

func unbatchRelayUpdates(updateChain []byte) ([][]byte, error) {
	updates := make([][]byte, 0)

	var offset int
	for {
		if offset >= len(updateChain) {
			break
		}

		var updateLength uint32
		var updateRequest []byte
		if !encoding.ReadUint32(updateChain, &offset, &updateLength) {
			return nil, fmt.Errorf("failed to read batched message length at offset %d (length %d)", offset, len(updateChain))
		}

		if !encoding.ReadBytes(updateChain, &offset, &updateRequest, updateLength) {
			return nil, fmt.Errorf("failed to read batched message length at offset %d (length %d)", offset, len(updateChain))
		}

		updates = append(updates, updateRequest)
	}

	return updates, nil
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

func RelayDashboardDataHandlerFunc(
	relayMap *routing.RelayMap,
	GetRouteMatrix func() *routing.RouteMatrix,
	statsdb *routing.StatsDatabase,
	maxJitter float64,
) func(writer http.ResponseWriter, request *http.Request) {

	type displayRelay struct {
		ID   uint64
		Name string
		Addr string
	}

	type response struct {
		Analysis string
		Relays   []displayRelay
		Stats    map[string]map[string]routing.Stats
	}

	MaxJitter = maxJitter

	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		var res response

		routeMatrix := GetRouteMatrix()
		res.Analysis = string(routeMatrix.GetAnalysisData())

		allRelayData := relayMap.GetAllRelayData()

		for _, relayData := range allRelayData {
			display := displayRelay{
				ID:   relayData.ID,
				Name: relayData.Name,
				Addr: relayData.Addr.String(),
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

		// write json response here
		type jsonRelay struct {
			Name string
			Addr string
		}

		type jsonResponse struct {
			Analysis routing.JsonMatrixAnalysis
			Relays   []jsonRelay
			Stats    map[string]map[string]routing.Stats
		}

		var jResponse jsonResponse

		jResponse.Analysis = routeMatrix.GetJsonAnalysis()
		jResponse.Stats = res.Stats

		for _, b := range res.Relays {
			var jRelay jsonRelay
			jRelay.Addr = b.Addr
			jRelay.Name = b.Name
			jResponse.Relays = append(jResponse.Relays, jRelay)
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(jResponse)

	}

}

func RelayDashboardAnalysisHandlerFunc(
	GetRouteMatrix func() *routing.RouteMatrix,
) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		routeMatrix := GetRouteMatrix()

		type jsonResponse struct {
			Analysis routing.JsonMatrixAnalysis `json:"analysis"`
		}

		var jResponse jsonResponse

		jResponse.Analysis = routeMatrix.GetJsonAnalysis()

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(jResponse)
	}
}

func DatabaseBinVersionFunc(creator *string, creationTime *string, env *string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		binInfo := map[string]string{
			"creator":      *creator,
			"creationTime": *creationTime,
			"env":          *env,
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(binInfo); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
