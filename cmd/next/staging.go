package main

import (
	"fmt"
	"math"
	"strings"
)

const (
	ServersPerVM = 50
	ClientsPerVM = 2000

	MaxVMsPerMIG = 1000
)

func getRelayList() (string, error) {
	success, output := bashQuiet("gcloud compute instances list --project=network-next-v3-staging --filter=\"name:relay*\" --format=\"value(name)\"")
	if !success {
		return "", fmt.Errorf("could not retrieve relay list: %s", output)
	}

	// Replace endlines with spaces to allow the relays to be used as parameters in future commands
	return strings.ReplaceAll(output, "\n", " "), nil
}

func getPortalCruncherList() (string, error) {
	success, output := bashQuiet("gcloud compute instances list --project=network-next-v3-staging --filter=\"name:portal-cruncher*\" --format=\"value(name)\"")
	if !success {
		return "", fmt.Errorf("could not retrieve portal cruncher list: %s", output)
	}

	// Replace endlines with spaces to allow the portal crunchers to be used as parameters in future commands
	return strings.ReplaceAll(output, "\n", " "), nil
}

func getClientMIGCount() (int, error) {
	success, output := bashQuiet("gcloud compute instance-groups managed list --project=network-next-v3-staging --filter=\"name:load-test-clients*\" --format=\"value(name)\"")
	if !success {
		return 0, fmt.Errorf("could not retrieve client MIG list: %s", output)
	}

	// Replace endlines with spaces to allow the client MIGs to be used as parameters in future commands
	return strings.Count(output, "load-test-clients"), nil
}

func waitForMIGStable(mig string) error {
	success, output := bashQuiet(fmt.Sprintf("gcloud compute instance-groups managed wait-until %s --project=network-next-v3-staging --stable --zone=us-central1-a", mig))
	if !success {
		return fmt.Errorf("could not wait for mig to stabilize: %s", output)
	}

	return nil
}

func startInstances(instances string, wait bool) error {
	var asyncFlag string
	if !wait {
		asyncFlag = "--async"
	}

	success, output := bashQuiet(fmt.Sprintf("echo %s | xargs gcloud compute instances start --project=network-next-v3-staging %s --zone=us-central1-a", instances, asyncFlag))
	if !success {
		return fmt.Errorf("could not start instances: %s", output)
	}

	return nil
}

func stopInstances(instances string, wait bool) error {
	var asyncFlag string
	if !wait {
		asyncFlag = "--async"
	}

	success, output := bashQuiet(fmt.Sprintf("echo %s | xargs gcloud compute instances stop --project=network-next-v3-staging %s --zone=us-central1-a", instances, asyncFlag))
	if !success {
		return fmt.Errorf("could not stop instances: %s", output)
	}

	return nil
}

func resizeMIG(mig string, size int) error {
	success, output := bashQuiet(fmt.Sprintf("gcloud compute instance-groups managed resize %s --project=network-next-v3-staging --size=%d --zone=us-central1-a", mig, size))
	if !success {
		return fmt.Errorf("could not resize MIG to %d instances: %s", size, output)
	}

	return nil
}

func setAutoscaling(mig string, enabled bool) error {
	enabledString := "off"
	if enabled {
		enabledString = "on"
	}

	success, output := bashQuiet(fmt.Sprintf("gcloud compute instance-groups managed update-autoscaling %s --project=network-next-v3-staging --mode=%s --zone=us-central1-a", mig, enabledString))
	if !success {
		if enabled {
			return fmt.Errorf("could not enable autoscaling: %s", output)
		}

		return fmt.Errorf("could not disable autoscaling: %s", output)
	}

	return nil
}

func startStaging(serverBackendCount int, clientCount int) error {
	clientMIGCount, err := getClientMIGCount()
	if err != nil {
		return err
	}

	if clientCount < ClientsPerVM {
		return fmt.Errorf("must run at least %d clients", ClientsPerVM)
	}

	if clientCount > MaxVMsPerMIG*clientMIGCount*ClientsPerVM {
		return fmt.Errorf("cannot run more than %d clients", clientCount)
	}

	relays, err := getRelayList()
	if err != nil {
		return err
	}

	fmt.Println("starting relay backend and relays...")
	if err := startInstances(relays, false); err != nil {
		return err
	}

	portalCrunchers, err := getPortalCruncherList()
	if err != nil {
		return err
	}

	fmt.Println("starting portal crunchers...")
	if err := startInstances(portalCrunchers, false); err != nil {
		return err
	}

	fmt.Println("starting analytics service...")
	if err := setAutoscaling("analytics-mig", true); err != nil {
		return err
	}

	fmt.Println("starting billing service...")
	if err := setAutoscaling("billing", true); err != nil {
		return err
	}

	fmt.Println("starting portal...")
	if err := setAutoscaling("portal-mig", true); err != nil {
		return err
	}

	if err := resizeStaging(serverBackendCount, clientCount); err != nil {
		return err
	}

	fmt.Println("staging started")
	return nil
}

func resizeStaging(serverBackendCount int, clientCount int) error {
	// Scale down the number of servers based on how many run on a single VM and enforce a proportion of 200 clients per server
	serverCount := int(math.Ceil(float64(clientCount / 200 / ServersPerVM)))
	if serverCount == 0 && clientCount > 0 {
		serverCount = 1
	}

	clientMIGCount, err := getClientMIGCount()
	if err != nil {
		return err
	}

	if clientCount < ClientsPerVM {
		return fmt.Errorf("must run at least %d clients", ClientsPerVM)
	}

	if clientCount > MaxVMsPerMIG*clientMIGCount*ClientsPerVM {
		return fmt.Errorf("cannot run more than %d clients", clientCount)
	}

	// Scale down the number of clients based on how many run on a single VM
	clientCount = clientCount / ClientsPerVM

	// We need to stop the servers and clients first so that a change to the server backend mig
	// will keep the servers and clients evenly distributed
	fmt.Println("stopping clients...")
	for i := 1; i <= clientMIGCount; i++ {
		if err := resizeMIG(fmt.Sprintf("load-test-clients-%d", i), 0); err != nil {
			return err
		}
	}

	fmt.Println("stopping servers...")
	if err := resizeMIG("load-test-server-mig", 0); err != nil {
		return err
	}

	fmt.Printf("resizing to %d server backends...\n", serverBackendCount)
	if err := resizeMIG("server-backend4-mig", serverBackendCount); err != nil {
		return err
	}

	// Wait for the server backend mig to stabilize so that the created servers and clients will connect evenly
	if err := waitForMIGStable("server-backend4-mig"); err != nil {
		return err
	}

	fmt.Printf("resizing to %d servers (%d instances)...\n", serverCount*ServersPerVM, serverCount)
	if err := resizeMIG("load-test-server-mig", serverCount); err != nil {
		return err
	}

	// Wait for the load test server mig to stabilize so that the created clients will connect evenly
	if err := waitForMIGStable("load-test-server-mig"); err != nil {
		return err
	}

	fmt.Printf("resizing to %d clients (%d instances)...\n", clientCount*ClientsPerVM, clientCount)
	runningClientCount := clientCount
	var clientRunCount int
	for runningClientCount > 0 {
		clientRunCount++

		var overflowClientCount int
		if runningClientCount > MaxVMsPerMIG {
			overflowClientCount = runningClientCount - MaxVMsPerMIG
			runningClientCount = MaxVMsPerMIG
		}

		if err := resizeMIG(fmt.Sprintf("load-test-clients-%d", clientRunCount), runningClientCount); err != nil {
			return err
		}

		runningClientCount -= MaxVMsPerMIG
		runningClientCount += overflowClientCount
	}

	return nil
}

func stopStaging() error {
	clientMIGCount, err := getClientMIGCount()
	if err != nil {
		return err
	}

	fmt.Println("stopping clients...")
	for i := 1; i <= clientMIGCount; i++ {
		if err := resizeMIG(fmt.Sprintf("load-test-clients-%d", i), 0); err != nil {
			return err
		}
	}

	fmt.Println("stopping servers...")
	if err := resizeMIG("load-test-server-mig", 0); err != nil {
		return err
	}

	fmt.Println("stopping server backend...")
	if err := resizeMIG("server-backend4-mig", 0); err != nil {
		return err
	}

	fmt.Println("stopping portal...")
	if err := setAutoscaling("portal-mig", false); err != nil {
		return err
	}

	if err := resizeMIG("portal-mig", 0); err != nil {
		return err
	}

	fmt.Println("stopping billing service...")
	if err := setAutoscaling("billing", false); err != nil {
		return err
	}

	if err := resizeMIG("billing", 0); err != nil {
		return err
	}

	fmt.Println("stopping analytics service...")
	if err := setAutoscaling("analytics-mig", false); err != nil {
		return err
	}

	if err := resizeMIG("analytics-mig", 0); err != nil {
		return err
	}

	portalCrunchers, err := getPortalCruncherList()
	if err != nil {
		return err
	}

	fmt.Println("stopping portal crunchers...")
	if err := stopInstances(portalCrunchers, false); err != nil {
		return err
	}

	relays, err := getRelayList()
	if err != nil {
		return err
	}

	fmt.Println("stopping relay backend and relays...")
	if err := stopInstances(relays, false); err != nil {
		return err
	}

	fmt.Println("staging stopped")
	return nil
}
