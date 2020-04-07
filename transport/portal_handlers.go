package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

func PortalHandlerFunc(redisClient redis.Cmdable, routeMatrix *routing.RouteMatrix, username string, password string) func(writer http.ResponseWriter, request *http.Request) {
	type response struct {
		Analysis string
		Relays   []routing.Relay
	}

	tmpl := template.Must(template.New("portal").Parse(`
		<html>
			<head>
				<title>Portal</title>
				<style>
					body { font-family: monospace; }
					table { width: 100%; border-collapse: collapse; }
					table, th, td { padding: 3px; border: 1px solid black; }
					td { text-align: center; }
				</style>
			</head>
			<body>
				<h1>Portal</h1>

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
			</body>
		</html>
	`))

	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		u, p, _ := request.BasicAuth()
		if u != username || p != password {
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

		if err := tmpl.Execute(writer, res); err != nil {
			fmt.Println(err)
		}
	}
}

func RelaySSHInfoHandlerFunc(logger log.Logger, storer storage.Storer) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		relayNames := request.URL.Query()["relay_name"]
		if len(relayNames) == 0 {
			level.Error(logger).Log("msg", "No relay name given")
		}

		found := false
		var r routing.Relay
		relayName := relayNames[0]
		for _, relay := range storer.Relays() {
			if relay.Name == relayName {
				r = relay
				found = true
			}
		}

		if !found {
			level.Error(logger).Log("msg", fmt.Sprintf("failed to lookup relay with name '%s'", relayName))
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		type sshInfo struct {
			User    string `json:"user"`
			Address string `json:"address"`
			Port    int64  `json:"port"`
		}

		info := sshInfo{
			User:    r.SSHUser,
			Address: r.ManagementAddress,
			Port:    r.SSHPort,
		}

		respData, err := json.Marshal(&info)
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("failed to marshal json response: %v", err))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(respData)
	}
}
