package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/storage"
)

type LiveServerService struct {
}

type LiveServerArgs struct {
	ServerBackendIPs []string
}

type LiveServerReply struct {
	ServerTrackers []storage.ServerTracker `json:"server_trackers"`
}

// LiveServers gets a json from each provided server_backend/servers and puts it on the wire
func (lss *LiveServerService) LiveServers(r *http.Request, args *LiveServerArgs, reply *LiveServerReply) error {

	authHeader := r.Header.Get("Authorization")

	var trackers []storage.ServerTracker
	for _, serverBackendIP := range args.ServerBackendIPs {
		uri := fmt.Sprintf("%s/servers", serverBackendIP)

		client := &http.Client{}
		req, _ := http.NewRequest("GET", uri, nil)
		req.Header.Set("Authorization", authHeader)

		response, err := client.Do(req)
		if err != nil {
			err = fmt.Errorf("LiveServers() error getting servers json: %v", err)
			core.Error("%v", err)
			return err
		}

		tracker := storage.NewServerTracker()

		json.NewDecoder(response.Body).Decode(&tracker.Tracker)
		response.Body.Close()

		trackers = append(trackers, *tracker)

	}

	reply.ServerTrackers = trackers

	return nil
}
