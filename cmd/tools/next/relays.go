package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

const (
	PortCheckScript = `echo "$(sudo lsof -i -P -n | grep '*:40000' | tr -s ' ' | cut -d ' ' -f 1 | head -1)"`
)

func relays(rpcClient jsonrpc.RPCClient, env Environment, filter string) {
	args := localjsonrpc.RelaysArgs{
		Name: filter,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].SessionCount > reply.Relays[j].SessionCount
	})

	relays := []struct {
		Name        string
		Address     string
		State       string
		Sessions    string
		Tx          string
		Rx          string
		Version     string
		LastUpdated string
	}{}

	for _, relay := range reply.Relays {
		tx := fmt.Sprintf("%.02fGB", float64(relay.BytesSent)/float64(1000000000))
		if relay.BytesSent < 1000000000 {
			tx = fmt.Sprintf("%.02fMB", float64(relay.BytesSent)/float64(1000000))
		}
		rx := fmt.Sprintf("%.02fGB", float64(relay.BytesReceived)/float64(1000000000))
		if relay.BytesReceived < 1000000000 {
			rx = fmt.Sprintf("%.02fMB", float64(relay.BytesReceived)/float64(1000000))
		}
		lastUpdated := "n/a"
		if relay.State == "enabled" {
			lastUpdated = time.Since(relay.LastUpdateTime).Truncate(time.Second).String()
		}

		address := relay.Addr

		relays = append(relays, struct {
			Name        string
			Address     string
			State       string
			Sessions    string
			Tx          string
			Rx          string
			Version     string
			LastUpdated string
		}{
			Name:        relay.Name,
			Address:     address,
			State:       relay.State,
			Sessions:    fmt.Sprintf("%d", relay.SessionCount),
			Tx:          tx,
			Rx:          rx,
			Version:     relay.Version,
			LastUpdated: lastUpdated,
		})
	}

	table.Output(relays)
}

func checkRelays(rpcClient jsonrpc.RPCClient, env Environment, filter string) {
	args := localjsonrpc.RelaysArgs{
		Name: filter,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Name < reply.Relays[j].Name
	})

	type checkInfo struct {
		Name           string
		CanSSH         string `table:"SSH Success"`
		UbuntuVersion  string `table:"Ubuntu"`
		CPUCores       string `table:"Cores"`
		CanPingBackend string `table:"Ping Backend"`
		ServiceRunning string `table:"Running"`
		PortBound      string `table:"Bound"`
	}

	info := make([]checkInfo, len(reply.Relays))

	var wg sync.WaitGroup
	wg.Add(len(reply.Relays))

	fmt.Printf("acquiring check info for %d relays\n", len(info))
	for i, relay := range reply.Relays {
		r := relay
		go func(indx int, wg *sync.WaitGroup) {
			defer wg.Done()

			infoIndx := &info[indx]
			infoIndx.Name = r.Name

			con := NewSSHConn(r.SSHUser, r.ManagementAddr, strconv.FormatInt(r.SSHPort, 10), env.SSHKeyFilePath)

			// test ssh capability, if not success return
			if con.TestConnect() {
				infoIndx.CanSSH = "yes"
			} else {
				infoIndx.CanSSH = "no"
				return
			}

			// get ubuntu version
			{
				if out, err := con.IssueCmdAndGetOutput(`lsb_release -r | grep -Po "([0-9]{2}\.[0-9]{2})"`); err == nil {
					infoIndx.UbuntuVersion = out
				} else {
					log.Printf("error when acquiring ubuntu version for relay %s: %v\n", r.Name, err)
					infoIndx.UbuntuVersion = "SSH Error"
				}
			}

			// get logical core count
			{
				if out, err := con.IssueCmdAndGetOutput("nproc"); err == nil {
					// test if the output of nproc is a number
					if _, err := strconv.ParseUint(out, 10, 64); err == nil {
						infoIndx.CPUCores = out
					} else {
						log.Printf("could not parse value of nproc (%s) to uint for relay %s: %v\n", out, r.Name, err)
						infoIndx.CPUCores = "Unknown"
					}
				} else {
					log.Printf("error when acquiring number of logical cpu cores for relay %s: %v\n", r.Name, err)
					infoIndx.CPUCores = "SSH Error"
				}
			}

			// test ping ability
			{
				if backend, err := env.RelayBackendHostname(); err == nil {
					if out, err := con.IssueCmdAndGetOutput("ping -c 1 " + backend + " > /dev/null; echo $?"); err == nil {
						if out == "0" {
							infoIndx.CanPingBackend = "yes"
						} else {
							infoIndx.CanPingBackend = "no"
						}
					} else {
						log.Printf("error when checking relay %s can ping the backend: %v\n", r.Name, err)
					}
				} else {
					log.Printf("%v\n", err)
				}
			}

			// check if the service is running
			{
				if out, err := con.IssueCmdAndGetOutput("sudo systemctl status relay > /dev/null; echo $?"); err == nil {
					if out == "0" {
						infoIndx.ServiceRunning = "yes"
					} else {
						infoIndx.ServiceRunning = "no"
					}
				} else {
					log.Printf("error when checking if relay %s has the service running: %v\n", r.Name, err)
				}
			}

			// check if the port is bound
			{
				if out, err := con.IssueCmdAndGetOutput(PortCheckScript); err == nil {
					if out == "relay" {
						infoIndx.PortBound = "yes"
					} else {
						infoIndx.PortBound = "no"
					}
				} else {
					log.Printf("error when checking if relay %s has the right port bound: %v\n", r.Name, err)
				}
			}

			log.Printf("gathered info for relay %s\n", r.Name)
		}(i, &wg)
	}

	log.Println("waiting for relays to complete")
	wg.Wait()
	log.Println("done, outputting results...")

	table.Output(info)
}

func addRelay(rpcClient jsonrpc.RPCClient, env Environment, relay routing.Relay) {
	args := localjsonrpc.AddRelayArgs{
		Relay: relay,
	}

	var reply localjsonrpc.AddRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" added to storage.\n", relay.Name)
}

func removeRelay(rpcClient jsonrpc.RPCClient, env Environment, name string) {
	info := getRelayInfo(rpcClient, name)

	args := localjsonrpc.RemoveRelayArgs{
		RelayID: info.id,
	}

	var reply localjsonrpc.RemoveRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" removed from storage.\n", name)
}
