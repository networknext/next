package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/networknext/backend/modules/storage"
)

const (
	// Params for the bigtable instance and table
	// These are based on staging.env files for the portal cruncher and portal
	InstanceID = "network-next-portal-big-table-0"
	TableName  = "portal-session-history"
	CFName     = "portal-session-history"
)

var (
	// List of redis instances to resize
	ResizableRedisInstances = []string{"staging-all-redis"}
)

type StagingServiceConfig struct {
	Cores int `json:"cores"`
	Count int `json:"count"`
}

type StagingConfig struct {
	RelayGateway   StagingServiceConfig `json:"relayGateway"`
	RelayBackend   StagingServiceConfig `json:"relayBackend"`
	FakeRelays     StagingServiceConfig `json:"fakeRelays"`
	RelayFrontend  StagingServiceConfig `json:"relayFrontend"`
	RelayPusher    StagingServiceConfig `json:"relayPusher"`
	PortalCruncher StagingServiceConfig `json:"portalCruncher"`
	Vanity         StagingServiceConfig `json:"vanity"`
	Api            StagingServiceConfig `json:"api"`
	Analytics      StagingServiceConfig `json:"analytics"`
	Billing        StagingServiceConfig `json:"billing"`
	Beacon         StagingServiceConfig `json:"beacon"`
	BeaconInserter StagingServiceConfig `json:"beaconInserter"`
	Portal         StagingServiceConfig `json:"portal"`
	ServerBackend  StagingServiceConfig `json:"serverBackend"`
	FakeServer     StagingServiceConfig `json:"fakeServer"`
}

var DefaultStagingConfig = StagingConfig{
	RelayGateway: StagingServiceConfig{
		Cores: 4,
		Count: -1,
	},

	RelayBackend: StagingServiceConfig{
		Cores: 8,
		Count: 2,
	},

	FakeRelays: StagingServiceConfig{
		Cores: 16,
		Count: 1,
	},

	RelayFrontend: StagingServiceConfig{
		Cores: 1,
		Count: -1,
	},

	RelayPusher: StagingServiceConfig{
		Cores: 1,
		Count: 1,
	},

	PortalCruncher: StagingServiceConfig{
		Cores: 8,
		Count: 4,
	},

	Vanity: StagingServiceConfig{
		Cores: 4,
		Count: 4,
	},

	Api: StagingServiceConfig{
		Cores: 1,
		Count: -1,
	},

	Analytics: StagingServiceConfig{
		Cores: 1,
		Count: -1,
	},

	Billing: StagingServiceConfig{
		Cores: 1,
		Count: -1,
	},

	Beacon: StagingServiceConfig{
		Cores: 8,
		Count: -1,
	},

	BeaconInserter: StagingServiceConfig{
		Cores: 1,
		Count: -1,
	},

	Portal: StagingServiceConfig{
		Cores: 16,
		Count: -1,
	},

	ServerBackend: StagingServiceConfig{
		Cores: 16,
		Count: 2,
	},

	FakeServer: StagingServiceConfig{
		Cores: 16,
		Count: 1,
	},
}

// InstanceGroup defines the necessary functionality for a group
// of virtual machine instances to be managed by the next tool
type InstanceGroup interface {
	Name() string
	ServiceConfig() StagingServiceConfig
	Start() error
	Stop() error
	Resize(size int) error
	Instances(limit int) ([]string, error)
	CoreCount() (int, error)
}

type ManagedInstanceGroup struct {
	name          string
	serviceConfig StagingServiceConfig
	autoscale     bool
	wait          bool
}

func NewManagedInstanceGroup(name string, wait bool, serviceConfig StagingServiceConfig) *ManagedInstanceGroup {
	var autoscale bool
	if serviceConfig.Count < 0 {
		autoscale = true
	}

	return &ManagedInstanceGroup{name: name, wait: wait, autoscale: autoscale, serviceConfig: serviceConfig}
}

func (mig *ManagedInstanceGroup) Name() string {
	return mig.name
}

func (mig *ManagedInstanceGroup) ServiceConfig() StagingServiceConfig {
	return mig.serviceConfig
}

func (mig *ManagedInstanceGroup) Start() error {
	if mig.autoscale {
		return mig.setAutoscaling(true)
	}

	if err := mig.Resize(mig.serviceConfig.Count); err != nil {
		return err
	}

	if mig.wait {
		waitForMIGStable(mig.name)
	}

	return nil
}

func (mig *ManagedInstanceGroup) Stop() error {
	if err := mig.setAutoscaling(false); err != nil {
		return err
	}

	return mig.Resize(0)
}

func (mig *ManagedInstanceGroup) Resize(size int) error {
	success, output := bashQuiet(fmt.Sprintf("gcloud compute instance-groups managed resize %s --project=network-next-v3-staging --size=%d --zone=us-central1-a", mig.name, size))
	if !success {
		return fmt.Errorf("could not resize MIG to %d instances: %s", size, output)
	}

	if mig.wait {
		waitForMIGStable(mig.name)
	}

	return nil
}

func (mig *ManagedInstanceGroup) Instances(limit int) ([]string, error) {
	command := fmt.Sprintf("gcloud compute instances list --project=network-next-v3-staging --filter=\"name:%s*\" --format=\"value(name)\"", mig.name)
	if limit > 0 {
		command = fmt.Sprintf("gcloud compute instances list --project=network-next-v3-staging --filter=\"name:%s*\" --format=\"value(name)\" --limit=%d", mig.name, limit)
	}

	success, output := bashQuiet(command)
	if !success {
		return nil, fmt.Errorf("could not retrieve instance list: %s", output)
	}

	return strings.Split(output, "\n"), nil
}

func (mig *ManagedInstanceGroup) CoreCount() (int, error) {
	success, output := bashQuiet(fmt.Sprintf("gcloud compute instance-groups managed describe %s --project=network-next-v3-staging --format=\"value(instanceTemplate)\" --zone=us-central1-a | xargs gcloud compute instance-templates describe --project=network-next-v3-staging --format=\"value(properties.machineType)\"", mig.name))
	if !success {
		return 0, fmt.Errorf("could not retrieve core count from MIG instance template: %s", output)
	}

	machineTypeParts := strings.Split(strings.Trim(output, " \t\r\n"), "-")
	if len(machineTypeParts) < 3 {
		return 0, fmt.Errorf("error when retrieving core count for service %s: bad machine type result", mig.name)
	}

	coreCountString := machineTypeParts[2]
	if machineTypeParts[0] == "custom" {
		coreCountString = machineTypeParts[1]
	}

	coreCount, err := strconv.Atoi(coreCountString)
	if err != nil {
		return 0, fmt.Errorf("error when retrieving core count for service %s: could not parse core count '%s'", mig.name, coreCountString)
	}

	return coreCount, nil
}

func (mig *ManagedInstanceGroup) setAutoscaling(enabled bool) error {
	enabledString := "off"
	if enabled {
		enabledString = "on"
	}

	success, output := bashQuiet(fmt.Sprintf("gcloud compute instance-groups managed update-autoscaling %s --project=network-next-v3-staging --mode=%s --zone=us-central1-a", mig.name, enabledString))
	if !success {
		if enabled {
			return fmt.Errorf("could not enable autoscaling: %s", output)
		}

		return fmt.Errorf("could not disable autoscaling: %s", output)
	}

	return nil
}

type UnmanagedInstanceGroup struct {
	name          string
	serviceConfig StagingServiceConfig
}

func NewUnmanagedInstanceGroup(name string, serviceConfig StagingServiceConfig) *UnmanagedInstanceGroup {
	return &UnmanagedInstanceGroup{name: name, serviceConfig: serviceConfig}
}

func (mig *UnmanagedInstanceGroup) Name() string {
	return mig.name
}

func (mig *UnmanagedInstanceGroup) ServiceConfig() StagingServiceConfig {
	return mig.serviceConfig
}

func (ig *UnmanagedInstanceGroup) Start() error {
	instances, err := ig.Instances(0)
	if err != nil {
		return err
	}

	instancesString := strings.Join(instances, " ")
	success, output := bashQuiet(fmt.Sprintf("echo %s | xargs gcloud compute instances start --project=network-next-v3-staging --async --zone=us-central1-a", instancesString))
	if !success {
		return fmt.Errorf("could not start instances: %s", output)
	}

	return nil
}

func (ig *UnmanagedInstanceGroup) Stop() error {
	instances, err := ig.Instances(0)
	if err != nil {
		return err
	}

	instancesString := strings.Join(instances, " ")
	success, output := bashQuiet(fmt.Sprintf("echo %s | xargs gcloud compute instances stop --project=network-next-v3-staging --async --zone=us-central1-a", instancesString))
	if !success {
		return fmt.Errorf("could not stop instances: %s", output)
	}

	return nil
}

func (ig *UnmanagedInstanceGroup) Resize(size int) error {
	return fmt.Errorf("unimplemented")
}

func (ig *UnmanagedInstanceGroup) Instances(limit int) ([]string, error) {
	command := fmt.Sprintf("gcloud compute instances list --project=network-next-v3-staging --filter=\"name:%s*\" --format=\"value(name)\"", ig.name)
	if limit > 0 {
		command = fmt.Sprintf("gcloud compute instances list --project=network-next-v3-staging --filter=\"name:%s*\" --format=\"value(name)\" --limit=%d", ig.name, limit)
	}

	success, output := bashQuiet(command)
	if !success {
		return nil, fmt.Errorf("could not retrieve instance list: %s", output)
	}

	return strings.Split(output, "\n"), nil
}

func (ig *UnmanagedInstanceGroup) CoreCount() (int, error) {
	instance, err := ig.Instances(1)
	if err != nil {
		return 0, err
	}

	if len(instance) == 0 {
		return 0, fmt.Errorf("could not retrieve core count: no instances in %s", ig.name)
	}

	success, output := bashQuiet(fmt.Sprintf("gcloud compute instances describe %s --project=network-next-v3-staging --format=\"value(machineType)\" --zone=us-central1-a", instance[0]))
	if !success {
		return 0, fmt.Errorf("could not retrieve instance machine type: %s", output)
	}

	lastSlashIndex := strings.LastIndex(output, "/")
	if lastSlashIndex+1 < 0 || lastSlashIndex+1 >= len(output) {
		return 0, fmt.Errorf("error when retrieving core count for service %s: bad machine type result", ig.name)
	}

	machineTypeString := strings.Trim(output[lastSlashIndex+1:], " \t\r\n")

	machineTypeParts := strings.Split(machineTypeString, "-")
	if len(machineTypeParts) < 3 {
		return 0, fmt.Errorf("error when retrieving core count for service %s: bad machine type result", ig.name)
	}

	coreCountString := machineTypeParts[2]
	if machineTypeParts[0] == "custom" {
		coreCountString = machineTypeParts[1]
	}

	coreCount, err := strconv.Atoi(coreCountString)
	if err != nil {
		return 0, fmt.Errorf("error when retrieving core count for service %s: could not parse core count '%s'", ig.name, coreCountString)
	}

	return coreCount, nil
}

// Waits for the given MIG to stabilize before continuing
func waitForMIGStable(mig string) error {
	success, output := bashQuiet(fmt.Sprintf("gcloud compute instance-groups managed wait-until %s --project=network-next-v3-staging --stable --zone=us-central1-a", mig))
	if !success {
		return fmt.Errorf("could not wait for mig to stabilize: %s", output)
	}

	return nil
}

func StartStaging(config StagingConfig) error {
	instanceGroups := createInstanceGroups(config)

	// Create the Bigtable instance and table if needed
	fmt.Printf("Setting up Bigtable...")
	err := createBigTable()
	if err != nil {
		return err
	}
	fmt.Println("done")

	// Copy the latest SQLite template for the portal to use
	fmt.Printf("Copying ./testdata/sqlite3-empty.sql to GCP cloud storage...")
	if err = pushSQLiteTemplate(); err != nil {
		return err
	}
	fmt.Println("done")

	var wg sync.WaitGroup
	// Resize Redis instances to 100 GB
	fmt.Printf("Resizing redis instances...\n")
	for _, instanceName := range ResizableRedisInstances {
		wg.Add(1)
		go func(redisInstanceName string) {
			defer wg.Done()
			err := resizeRedisInstance(redisInstanceName, 100)
			if err != nil {
				fmt.Printf("failed to resize redis instance %s to 100 GB: %v\n", redisInstanceName, err)
			} else {
				fmt.Printf("resized redis instance %s to 100 GB\n", redisInstanceName)
			}
		}(instanceName)
	}
	wg.Wait()
	fmt.Println("done")

	for _, instanceGroup := range instanceGroups {
		serviceConfig := instanceGroup.ServiceConfig()

		if serviceConfig.Count < 0 {
			fmt.Printf("configuring %s with autoscaling and with %d cores each...", instanceGroup.Name(), serviceConfig.Cores)
		} else {
			fmt.Printf("configuring %s with %d instances with %d cores each...", instanceGroup.Name(), serviceConfig.Count, serviceConfig.Cores)
		}

		coreCount, err := instanceGroup.CoreCount()
		if err != nil {
			return err
		}

		if coreCount != serviceConfig.Cores {
			// Update with desired core count
		}

		instances, err := instanceGroup.Instances(0)
		if err != nil {
			return err
		}

		if len(instances) != serviceConfig.Count {
			// Update with desired instance count
		}

		fmt.Print("starting...")
		if err := instanceGroup.Start(); err != nil {
			return err
		}

		fmt.Println()
	}

	fmt.Println("\nstaging started")
	return nil
}

func StopStaging() []error {
	instanceGroups := createInstanceGroups(DefaultStagingConfig)

	var wg sync.WaitGroup
	errChan := make(chan error, len(instanceGroups)+len(ResizableRedisInstances)+1)

	fmt.Println("stopping staging...")

	for i := len(instanceGroups) - 1; i >= 0; i-- {
		wg.Add(1)
		go func(i int) {
			if err := instanceGroups[i].Stop(); err != nil {
				errChan <- err
			}

			fmt.Printf("stopped %s\n", instanceGroups[i].Name())

			wg.Done()
		}(i)

	}

	// Delete bigtable
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := deleteBigTable(); err != nil {
			errChan <- err
		}

		fmt.Println("deleted Bigtable")
	}()

	// Resize Redis instances to 1 GB
	for _, instanceName := range ResizableRedisInstances {
		wg.Add(1)
		go func(redisInstanceName string) {
			defer wg.Done()
			err := resizeRedisInstance(redisInstanceName, 1)
			if err != nil {
				errChan <- err
			}
			fmt.Printf("resized redis instance %s to 1 GB\n", redisInstanceName)
		}(instanceName)
	}

	wg.Wait()

	errs := make([]error, 0)
	select {
	case err := <-errChan:
		errs = append(errs, err)
	default:
		if len(errs) == 0 {
			fmt.Println("\nstaging stopped")
			return nil
		}

		fmt.Println()
		return errs
	}

	return nil
}

// Creates a Bigtable instance and table, if needed
// Bigtable is required for the portal crunchers and portal to function
func createBigTable() error {
	ctx := context.Background()
	gcpProjectID := "network-next-v3-staging"

	// Create a bigtable instance admin
	btInstanceAdmin, err := storage.NewBigTableInstanceAdmin(ctx, gcpProjectID)
	if err != nil {
		return err
	}
	defer func() {
		_ = btInstanceAdmin.Close()
	}()

	// Verify if the instance already exists
	instanceExists, err := btInstanceAdmin.VerifyInstanceExists(ctx, InstanceID)
	if err != nil {
		return err
	}
	if !instanceExists {
		// Create the instance with 2 clusters and 5 zones per cluster
		displayName := "bigtable-staging"
		zones := []string{"us-central1-a", "us-central1-b"}
		numClusters := 2
		numNodesPerCluster := 5

		err = btInstanceAdmin.CreateInstance(ctx, InstanceID, displayName, zones, numClusters, numNodesPerCluster)
		if err != nil {
			return err
		}
	}

	// Create a bigtable admin
	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, InstanceID)
	if err != nil {
		return err
	}
	defer func() {
		_ = btAdmin.Close()
	}()

	// Verify if the table already exists
	tableExists, err := btAdmin.VerifyTableExists(ctx, TableName)
	if err != nil {
		return err
	}
	if !tableExists {
		// Create the table with a max age policy of 90 days
		err = btAdmin.CreateTable(ctx, TableName, []string{CFName})
		if err != nil {
			return err
		}

		maxAge := time.Hour * time.Duration(1) * 24 * 90
		err = btAdmin.SetMaxAgePolicy(ctx, TableName, []string{CFName}, maxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

// Deletes a Bigtable instance, if needed
// Deleting an instance automatically deletes any tables for that instance
func deleteBigTable() error {
	ctx := context.Background()
	gcpProjectID := "network-next-v3-staging"

	// Create a bigtable instance admin
	btInstanceAdmin, err := storage.NewBigTableInstanceAdmin(ctx, gcpProjectID)
	if err != nil {
		return err
	}
	defer func() {
		_ = btInstanceAdmin.Close()
	}()

	// Verify if the instance already exists
	instanceExists, err := btInstanceAdmin.VerifyInstanceExists(ctx, InstanceID)
	if err != nil {
		return err
	}
	if instanceExists {
		// Delete the instance
		err = btInstanceAdmin.DeleteInstance(ctx, InstanceID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Copies ./testdata/sqlite3-empty.sql to the staging artifacts bucket to be used by the portal
func pushSQLiteTemplate() error {

	bucketName := "gs://staging_artifacts"

	if _, err := os.Stat("./testdata/sqlite3-empty.sql"); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Local file ./testdata/sqlite3-empty.sql does not exist: %v", err)
	}

	gsutilCpCommand := exec.Command("gsutil", "cp", "./testdata/sqlite3-empty.sql", bucketName)

	err := gsutilCpCommand.Run()
	if err != nil {
		return fmt.Errorf("Error copying ./testdata/sqlite3-empty.sql to %s: %v\n", bucketName, err)
	}

	return nil
}

func resizeRedisInstance(instanceName string, size int) error {
	success, output := bashQuiet(fmt.Sprintf("gcloud redis instances update %s --project=network-next-v3-staging --size=%d --region=us-central1", instanceName, size))
	if !success {
		return fmt.Errorf("could not resize redis instance %s to %d GB: %s", instanceName, size, output)
	}

	return nil
}

func createInstanceGroups(config StagingConfig) []InstanceGroup {
	instanceGroups := make([]InstanceGroup, 0)

	instanceGroups = append(instanceGroups, NewUnmanagedInstanceGroup("relay-backend", config.RelayBackend))
	instanceGroups = append(instanceGroups, NewUnmanagedInstanceGroup("portal-cruncher", config.PortalCruncher))
	instanceGroups = append(instanceGroups, NewUnmanagedInstanceGroup("vanity", config.Vanity))
	instanceGroups = append(instanceGroups, NewUnmanagedInstanceGroup("relay-pusher", config.RelayPusher))
	instanceGroups = append(instanceGroups, NewUnmanagedInstanceGroup("fake-relays", config.FakeRelays))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("relay-gateway-mig", false, config.RelayGateway))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("relay-frontend-mig", false, config.RelayFrontend))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("api-mig", false, config.Api))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("analytics-mig", false, config.Analytics))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("billing", false, config.Billing))
	// instanceGroups = append(instanceGroups, NewManagedInstanceGroup("beacon-mig", false, config.Beacon))
	// instanceGroups = append(instanceGroups, NewManagedInstanceGroup("beacon-inserter-mig", false, config.BeaconInserter))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("portal-mig", false, config.Portal))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("server-backend-mig", true, config.ServerBackend))
	instanceGroups = append(instanceGroups, NewManagedInstanceGroup("fake-server-mig", true, config.FakeServer))

	return instanceGroups
}

// func resizeStaging(config StagingConfig) error {
// 	// Scale down the number of servers based on how many run on a single VM and enforce a proportion of 200 clients per server
// 	serverCount := int(math.Ceil(float64(config.Client.Count / 200 / config.Server.ServersPerVM)))
// 	if serverCount == 0 && config.Client.Count > 0 {
// 		serverCount = 1
// 	}

// 	clientMIGCount, err := getClientMIGCount()
// 	if err != nil {
// 		return err
// 	}

// 	if config.Client.Count < config.Client.ClientsPerVM {
// 		return fmt.Errorf("must run at least %d clients", config.Client.ClientsPerVM)
// 	}

// 	if config.Client.Count > MaxVMsPerMIG*clientMIGCount*config.Client.ClientsPerVM {
// 		return fmt.Errorf("cannot run more than %d clients", config.Client.Count)
// 	}

// 	// We need to stop the servers and clients first so that a change to the server backend mig
// 	// will keep the servers and clients evenly distributed
// 	fmt.Println("stopping clients...")
// 	for i := 1; i <= clientMIGCount; i++ {
// 		if err := resizeMIG(fmt.Sprintf("load-test-clients-%d", i), 0); err != nil {
// 			return err
// 		}
// 	}

// 	fmt.Println("stopping servers...")
// 	if err := resizeMIG("load-test-server-mig", 0); err != nil {
// 		return err
// 	}

// 	fmt.Printf("resizing to %d server backends...\n", config.ServerBackend.Count)
// 	if err := resizeMIG("server-backend4-mig", config.ServerBackend.Count); err != nil {
// 		return err
// 	}

// 	// Wait for the server backend mig to stabilize so that the created servers and clients will connect evenly
// 	if err := waitForMIGStable("server-backend4-mig"); err != nil {
// 		return err
// 	}

// 	fmt.Printf("resizing to %d servers (%d instances)...\n", serverCount*config.Server.ServersPerVM, serverCount)
// 	if err := resizeMIG("load-test-server-mig", serverCount); err != nil {
// 		return err
// 	}

// 	// Wait for the load test server mig to stabilize so that the created clients will connect evenly
// 	if err := waitForMIGStable("load-test-server-mig"); err != nil {
// 		return err
// 	}

// 	fmt.Printf("resizing to %d clients (%d instances)...\n", config.Client.Count*config.Client.ClientsPerVM, config.Client.Count)

// 	// Scale down the number of clients based on how many run on a single VM
// 	runningClientCount := config.Client.Count / config.Client.ClientsPerVM

// 	var clientRunCount int
// 	for runningClientCount > 0 {
// 		clientRunCount++

// 		var overflowClientCount int
// 		if runningClientCount > MaxVMsPerMIG {
// 			overflowClientCount = runningClientCount - MaxVMsPerMIG
// 			runningClientCount = MaxVMsPerMIG
// 		}

// 		if err := resizeMIG(fmt.Sprintf("load-test-clients-%d", clientRunCount), runningClientCount); err != nil {
// 			return err
// 		}

// 		runningClientCount -= MaxVMsPerMIG
// 		runningClientCount += overflowClientCount
// 	}

// 	return nil
// }
