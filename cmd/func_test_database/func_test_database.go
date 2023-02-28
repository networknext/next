/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"
	"bytes"
	"net/http"
	"encoding/json"
	"io/ioutil"

	db "github.com/networknext/backend/modules/database"
)

// ----------------------------------------------------------------------------------------

func bash(command string) {

	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: failed to run command: %v\n", err)
		os.Exit(1)
	}

	cmd.Wait()
}

func api() (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command("./api")
	if cmd == nil {
		panic("could not create api!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ENABLE_PORTAL=false")
	cmd.Env = append(cmd.Env, "ENABLE_ADMIN=false")
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

func ValidateDatabase() {

	database, err := db.ExtractDatabase("host=127.0.0.1 port=5432 user=developer dbname=postgres sslmode=disable")
	if err != nil {
		fmt.Printf("error: failed to extract database: %v\n", err)
		os.Exit(1)
	}

	err = database.Validate()
	if err != nil {
		fmt.Printf("error: database did not validate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(database.String())
}

func test_local() {

	fmt.Printf("test_local\n")

	bash("psql -U developer -h localhost postgres -f ../schemas/sql/destroy.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/create.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/local.sql")

	ValidateDatabase()
}

func test_dev() {

	fmt.Printf("test_dev\n")

	bash("psql -U developer -h localhost postgres -f ../schemas/sql/destroy.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/create.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/dev.sql")

	ValidateDatabase()
}

func Get(url string, object interface{}) {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	err = json.Unmarshal([]byte(body), &object)
	if err != nil {
		panic(fmt.Sprintf("could not parse json response for %s: %v", url, err))
	}
}

func test_api() {

	fmt.Printf("test_api\n")

	// create a dummy database

	database := db.CreateDatabase()

	database.CreationTime = "now"
	database.Creator = "test"
	database.BuyerMap[1] = &db.Buyer{Id: 1, Name: "buyer", Live: true, Debug: true}
	database.SellerMap[1] = &db.Seller{Id: 1, Name: "seller"}
	database.DatacenterMap[1] = &db.Datacenter{Id: 1, Name: "datacenter", Latitude: 100, Longitude: 200}
	database.Relays = append(database.Relays, db.Relay{Id: 1, Name: "relay", Datacenter: *database.DatacenterMap[1]})
	datacenterRelays := [1]uint64{1}
	database.DatacenterRelays[1] = datacenterRelays[:]
	database.BuyerDatacenterSettings[1] = make(map[uint64]*db.BuyerDatacenterSettings)
	database.BuyerDatacenterSettings[1][1] = &db.BuyerDatacenterSettings{BuyerId: 1, DatacenterId: 1, EnableAcceleration: true}

	err := database.Validate()
	if err != nil {
		fmt.Printf("error: database did not validate: %v\n", err)
		os.Exit(1)
	}

	// save it to database.bin

	database.Save("database.bin")

	// run API service and it will load in database.bin

	api_cmd, _ := api()

	// query the database REST API

	// ...

	// shut down API service

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()

	// check results of all queries

	// ...
}

// ----------------------------------------------------------------------------------------

type test_function func()

func main() {

	allTests := []test_function{
		test_local,
		test_dev,
		test_api,
	}

	var tests []test_function

	if len(os.Args) > 1 {
		funcName := os.Args[1]
		for _, test := range allTests {
			name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
			name = name[len("main."):]
			if funcName == name {
				tests = append(tests, test)
				break
			}
		}
		if len(tests) == 0 {
			panic(fmt.Sprintf("could not find any test: '%s'", funcName))
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	go func() {
		time.Sleep(time.Duration(len(tests)*120) * time.Second)
		panic("tests took too long!")
	}()

	fmt.Printf("\n")

	for i := range tests {
		tests[i]()
	}
}
