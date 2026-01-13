/*
   Network Next. Copyright 2017 - 2026 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"

	"github.com/networknext/next/modules/core"
	db "github.com/networknext/next/modules/database"
)

const TestAPIKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTc0OTczODE4OX0.I89NXJCRMU_pIjnSleAnbux5HNsHhymzQ_SVatFo3b4"
const TestAPIPrivateKey = "uKUsmTySUVEssqBmVNciJWWolchcGGhFzRWMpydwOtVExvqYpHMotnkanNTaGHHh"

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
	cmd.Env = append(cmd.Env, fmt.Sprintf("API_PRIVATE_KEY=%s", TestAPIPrivateKey))

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

	fmt.Printf("%s\n\n", database.String())
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
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/test.sql")

	ValidateDatabase()
}

func GetJSON(url string, object interface{}) {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
		req.Header.Set("Authorization", "Bearer "+TestAPIKey)
		client := &http.Client{}
		response, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("got %d response for %s", response.StatusCode, url))
	}

	body, error := io.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	err = json.Unmarshal([]byte(body), &object)
	if err != nil {
		panic(fmt.Sprintf("could not parse json response for %s: %v", url, err))
	}
}

func GetBinary(url string) []byte {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
		req.Header.Set("Authorization", "Bearer "+TestAPIKey)
		client := &http.Client{}
		response, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("got %d response for %s", response.StatusCode, url))
	}

	body, error := io.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	return body
}

func test_api() {

	fmt.Printf("test_api\n")

	// create a dummy database

	database := db.CreateDatabase()

	database.CreationTime = "now"
	database.Creator = "test"
	database.BuyerMap[1] = &db.Buyer{Id: 1, Name: "buyer", Live: true, Debug: true}
	database.SellerMap[1] = &db.Seller{Id: 1, Name: "seller"}
	database.DatacenterMap[1] = &db.Datacenter{Id: 1, Name: "local", Latitude: 100, Longitude: 200}
	for i := 0; i < 1; i++ {
		relayId := uint64(1 + i)
		relay := db.Relay{
			Id:            relayId,
			Name:          fmt.Sprintf("local-%d", i+1),
			PublicAddress: core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i)),
			SSHAddress:    core.ParseAddress("127.0.0.1:22"),
			Datacenter:    database.DatacenterMap[1],
			Seller:        database.SellerMap[1],
		}
		database.Relays = append(database.Relays, relay)
		database.DatacenterRelays[1] = append(database.DatacenterRelays[1], uint64(1+i))
	}
	datacenterRelays := [1]uint64{1}
	database.DatacenterRelays[1] = datacenterRelays[:]
	database.BuyerDatacenterSettings[1] = make(map[uint64]*db.BuyerDatacenterSettings)
	database.BuyerDatacenterSettings[1][1] = &db.BuyerDatacenterSettings{BuyerId: 1, DatacenterId: 1, EnableAcceleration: true}

	database.Fixup()

	err := database.Validate()
	if err != nil {
		fmt.Printf("error: database did not validate: %v\n", err)
		os.Exit(1)
	}

	// save it to database.bin

	database.Save("database.bin")

	// run API service and it will load in database.bin

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// run API queries

	database_binary := GetBinary("http://127.0.0.1:50000/database/binary")

	database_json := db.Database{}
	GetJSON("http://127.0.0.1:50000//database/json", &database_json)

	database_header := db.HeaderResponse{}
	GetJSON("http://127.0.0.1:50000//database/header", &database_header)

	database_buyers := db.BuyersResponse{}
	GetJSON("http://127.0.0.1:50000//database/buyers", &database_buyers)

	database_sellers := db.SellersResponse{}
	GetJSON("http://127.0.0.1:50000//database/sellers", &database_sellers)

	database_datacenters := db.DatacentersResponse{}
	GetJSON("http://127.0.0.1:50000//database/datacenters", &database_datacenters)

	database_relays := db.RelaysResponse{}
	GetJSON("http://127.0.0.1:50000//database/relays", &database_relays)

	database_buyer_datacenter_settings := db.BuyerDatacenterSettingsResponse{}
	GetJSON("http://127.0.0.1:50000//database/buyer_datacenter_settings", &database_buyer_datacenter_settings)

	// verify database binary

	expected_binary := database.GetBinary()

	if !bytes.Equal(database_binary, expected_binary) {
		panic("database binary does not match expected")
	}

	json_binary := database_json.GetBinary()

	if !bytes.Equal(json_binary, expected_binary) {
		panic("json database does not match expected")
	}

	if err := database_json.Validate(); err != nil {
		fmt.Printf("error: json database did not validate: %v\n", err)
		os.Exit(1)
	}

	if database_header.CreationTime != "now" {
		panic("wrong database creation time")
	}

	if database_header.Creator != "test" {
		panic("wrong database creator")
	}

	if database_header.NumBuyers != 1 {
		panic("wrong number of database buyers")
	}

	if database_header.NumSellers != 1 {
		panic("wrong number of database sellers")
	}

	if database_header.NumDatacenters != 1 {
		panic("wrong number of database datacenters")
	}

	if database_header.NumRelays != 1 {
		panic("wrong number of database relays")
	}

	if len(database_buyers.Buyers) != 1 {
		panic("wrong number of buyers")
	}

	if len(database_sellers.Sellers) != 1 {
		panic("wrong number of sellers")
	}

	if len(database_datacenters.Datacenters) != 1 {
		panic("wrong number of datacenters")
	}

	if len(database_relays.Relays) != 1 {
		panic("wrong number of relays")
	}

	if len(database_buyer_datacenter_settings.BuyerDatacenterSettings) != 1 {
		panic("wrong number of buyer datacenter settings")
	}

	if database_buyers.Buyers[0].Id != 1 || database_buyers.Buyers[0].Name != "buyer" || database_buyers.Buyers[0].Live != true || database_buyers.Buyers[0].Debug != true {
		panic("buyer is invalid")
	}

	if database_sellers.Sellers[0].Id != 1 || database_sellers.Sellers[0].Name != "seller" {
		panic("seller is invalid")
	}

	if database_datacenters.Datacenters[0].Id != 1 || database_datacenters.Datacenters[0].Name != "local" || database_datacenters.Datacenters[0].Latitude != 100 || database_datacenters.Datacenters[0].Longitude != 200 {
		panic("datacenter is invalid")
	}

	if database_relays.Relays[0].Id != 1 || database_relays.Relays[0].Name != "local-1" {
		panic("relay is invalid")
	}

	if database_buyer_datacenter_settings.BuyerDatacenterSettings[0].BuyerId != 1 || database_buyer_datacenter_settings.BuyerDatacenterSettings[0].DatacenterId != 1 || database_buyer_datacenter_settings.BuyerDatacenterSettings[0].EnableAcceleration == false {
		panic("buyer datacenter settings are invalid")
	}
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
