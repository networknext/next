package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/modood/table"

	"github.com/networknext/backend/modules-old/routing"

	localjsonrpc "github.com/networknext/backend/modules-old/transport/jsonrpc"

	"github.com/ybbus/jsonrpc"
)

const (

	StartRelayScript = `sudo systemctl enable /app/relay.service && sudo systemctl start relay`

	StopRelayScript = `sudo systemctl stop relay && sudo systemctl disable relay`

	VersionCheckScript = `lsb_release -r | grep -Po "([0-9]{2}\.[0-9]{2})"`
)

func unitFormat(bits uint64) string {
	const (
		kilo = 1000
		mega = 1000000
		giga = 1000000000
	)

	if bits > giga {
		return fmt.Sprintf("%.02fGb/s", float64(bits)/float64(giga))
	}

	if bits > mega {
		return fmt.Sprintf("%.02fMb/s", float64(bits)/float64(mega))
	}

	if bits > kilo {
		return fmt.Sprintf("%.02fKb/s", float64(bits)/float64(kilo))
	}

	return fmt.Sprintf("%d", bits)
}

type relayInfo struct {
	id          uint64
	name        string
	user        string
	sshAddr     string
	sshPort     string
	publicAddr  string
	publicKey   string
	nicSpeed    string
	firestoreID string
	state       string
	version     string
}

func getRelayInfo(env Environment, regex string) []relayInfo {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return nil
	}

	relays := make([]relayInfo, len(reply.Relays))

	for i, r := range reply.Relays {
		relays[i] = relayInfo{
			id:         r.ID,
			name:       r.Name,
			user:       r.SSHUser,
			sshAddr:    r.ManagementAddr,
			sshPort:    fmt.Sprintf("%d", r.SSHPort),
			publicAddr: r.Addr,
			publicKey:  r.PublicKey,
			nicSpeed:   fmt.Sprintf("%d", r.NICSpeedMbps),
			state:      r.State,
			version:    r.Version,
		}
	}

	return relays
}

func getInfoForAllRelays(rpcClient jsonrpc.RPCClient, env Environment) []relayInfo {
	args := localjsonrpc.RelaysArgs{}

	var reply localjsonrpc.RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return nil
	}

	if len(reply.Relays) == 0 {
		handleRunTimeError(fmt.Sprintln("could not find a single relay"), 0)
	}

	relays := make([]relayInfo, len(reply.Relays))

	for i, relay := range reply.Relays {
		relays[i] = relayInfo{
			id:         relay.ID,
			name:       relay.Name,
			user:       relay.SSHUser,
			sshAddr:    relay.ManagementAddr,
			sshPort:    fmt.Sprintf("%d", relay.SSHPort),
			publicAddr: relay.Addr,
		}
	}

	return relays
}

func startRelays(env Environment, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			if strings.Contains(relay.name, "-removed-") || relay.state != "enabled" {
				continue
			}
			fmt.Printf("starting relay %s\n", relay.name)
			testForSSHKey(env)
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(StartRelayScript) {
				continue
			}
		}
	}
}

func stopRelays(env Environment, regexes []string) bool {
	success := true
	testForSSHKey(env)
	script := StopRelayScript
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		if len(relays) == 0 {
			fmt.Printf("no relays matched the regex '%s'\n", regex)
			continue
		}
		for _, relay := range relays {
			if strings.Contains(relay.name, "-removed-") || relay.state != "enabled" {
				continue
			}
			fmt.Printf("stopping relay %s\n", relay.name)
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			if !con.ConnectAndIssueCmd(script) {
				success = false
				continue
			}
		}
	}

	return success
}

// TODO modify to use the OpsService.UpdateRelay endpoint
func updateRelayName(env Environment, oldName string, newName string) {

	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(env, oldName); !ok {
		// error msg printed by called function
		return
	}

	reply := localjsonrpc.RelayNameUpdateReply{}
	args := localjsonrpc.RelayNameUpdateArgs{
		RelayID:   relayID,
		RelayName: newName,
	}

	if err := makeRPCCall(env, &reply, "OpsService.RelayNameUpdate", args); err != nil {
		fmt.Printf("error renaming relay: %v\n", (err))
	} else {
		fmt.Printf("Relay renamed successfully: %s -> %s\n", oldName, newName)
	}

}

func checkRelays(
	env Environment,
	regex string,
	relaysStateShowFlags [6]bool,
	relaysStateHideFlags [6]bool,
	relaysDownFlag bool,
	csvOutputFlag bool,
) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Name < reply.Relays[j].Name
	})

	type checkInfo struct {
		Name           string
		Address        string `table:"Address"`
		CanSSH         string `table:"SSH"`
		UbuntuVersion  string `table:"Ubuntu"`
		CPUCores       string `table:"Cores"`
		CanPingBackend string `table:"Ping Backend"`
		ServiceRunning string `table:"Running"`
		PortBound      string `table:"Bound"`
	}

	// filter the relays
	includedRelays := make([]relayInfo, 0)

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			fmt.Printf("could not parse relay state %s for relay %s", relay.State, relay.Name)
			continue
		}

		includeRelay := true

		for i, flag := range relaysStateShowFlags {
			if flag {
				if relayState != routing.RelayState(i) {
					// An "only show" flag is set and this relay doesn't match that state, so don't include it in the final output
					includeRelay = false
				} else {
					// One of the flags should include the relay, so set to true and break out, since combining the flags is an OR operation
					includeRelay = true
					break
				}
			}
		}

		if relaysStateHideFlags[relayState] {
			// Relay should be hidden, so don't include in final output
			includeRelay = false
		}

		if !includeRelay {
			continue
		}

		includedRelays = append(includedRelays, relayInfo{
			name:    relay.Name,
			user:    relay.SSHUser,
			sshAddr: relay.ManagementAddr,
			sshPort: strconv.FormatInt(relay.SSHPort, 10),
		})
	}

	var wg sync.WaitGroup
	wg.Add(len(includedRelays))

	fmt.Printf("Checking %d relays\n\n", len(includedRelays))

	info := make([]checkInfo, len(includedRelays))

	for i, relay := range includedRelays {
		r := relay
		go func(indx int, wg *sync.WaitGroup) {
			defer wg.Done()

			infoIndx := &info[indx]
			infoIndx.Name = r.name
			infoIndx.Address = r.sshAddr

			con := NewSSHConn(r.user, r.sshAddr, r.sshPort, env.SSHKeyFilePath)

			// test ssh capability, if not success return
			if con.TestConnect() {
				infoIndx.CanSSH = "yes"
			} else {
				infoIndx.CanSSH = "no"
				return
			}

			// get ubuntu version
			{
				if out, err := con.IssueCmdAndGetOutput(VersionCheckScript); err == nil {
					infoIndx.UbuntuVersion = out
				} else {
					fmt.Printf("error when acquiring ubuntu version for relay %s: %v\n", r.name, err)
					infoIndx.UbuntuVersion = "Error"
				}
			}

			// test ping ability
			{
				if backend, err := env.RelayBackendHostname(); err == nil {
					if out, err := con.IssueCmdAndGetOutput("ping -c 20 " + backend + " > /dev/null; echo $?"); err == nil {
						if out == "0" {
							infoIndx.CanPingBackend = "yes"
						} else {
							infoIndx.CanPingBackend = "no"
						}
					} else {
						fmt.Printf("error when checking relay %s can ping the backend: %v\n", r.name, err)
						infoIndx.CanPingBackend = "Error"
					}
				} else {
					fmt.Printf("%v\n", err)
				}
			}

			// check if the service is running
			{
				if out, err := con.IssueCmdAndGetOutput("sudo systemctl status relay > /dev/null 2>&1; echo $?"); err == nil {
					if out == "0" {
						infoIndx.ServiceRunning = "yes"
					} else {
						infoIndx.ServiceRunning = "no"
					}
				} else {
					fmt.Printf("error when checking if relay %s has the service running: %v\n", r.name, err)
					infoIndx.ServiceRunning = "Error"
				}
			}

			fmt.Printf("gathered info for relay %s\n", r.name)
		}(i, &wg)
	}

	wg.Wait()

	if len(info) > 0 {
		fmt.Printf("\n")
	}

	if csvOutputFlag {
		var csvInfo [][]string
		csvInfo = append(csvInfo, []string{
			"Name", "Address", "SSH", "Ubuntu", "Cores", "Ping Backend", "Running", "Bound"})

		for _, relayInfo := range info {
			csvInfo = append(csvInfo, []string{
				relayInfo.Name,
				relayInfo.Address,
				relayInfo.CanSSH,
				relayInfo.UbuntuVersion,
				relayInfo.CPUCores,
				relayInfo.CanPingBackend,
				relayInfo.ServiceRunning,
				relayInfo.PortBound,
			})

			fileName := "./relay-check.csv"
			f, err := os.Create(fileName)
			if err != nil {
				fmt.Printf("Error creating local CSV file %s: %v\n", fileName, err)
				return
			}

			writer := csv.NewWriter(f)
			err = writer.WriteAll(csvInfo)
			if err != nil {
				fmt.Printf("Error writing local CSV file %s: %v\n", fileName, err)
			}
		}
		fmt.Println("CSV file written: relay-check.csv")

	} else {
		table.Output(info)
	}

}

func relayLogs(env Environment, lines uint, regexes []string) {
	for _, regex := range regexes {
		relays := getRelayInfo(env, regex)
		for _, relay := range relays {
			con := NewSSHConn(relay.user, relay.sshAddr, relay.sshPort, env.SSHKeyFilePath)
			con.ConnectAndIssueCmd("journalctl -fu relay -n 1000")
			break
		}
	}
}
