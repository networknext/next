/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"

	"github.com/networknext/backend/modules/admin"
)

var hostname = "http://127.0.0.1:50000"
var apiKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZGF0YWJhc2UiOnRydWUsInBvcnRhbCI6dHJ1ZX0.QFPdb-RcP8wyoaOIBYeB_X6uA7jefGPVxm2VevJvpwU"
var apiPrivateKey = "this is the private key that generates API keys. make sure you change this value in production"

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

func clearDatabase() {
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/destroy.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/create.sql")
}

func api() (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command("./api")
	if cmd == nil {
		panic("could not create api!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ENABLE_PORTAL=false")
	cmd.Env = append(cmd.Env, "ENABLE_DATABASE=false")
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")
	cmd.Env = append(cmd.Env, fmt.Sprintf("API_PRIVATE_KEY=%s", apiPrivateKey))

	var output bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Start()

	return cmd, &output
}

// ----------------------------------------------------------------------------------------

func GetText(path string) (string, error) {

    url := hostname + "/" + path

    var err error
    var response *http.Response
    for i := 0; i < 5; i++ {
        req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
        req.Header.Set("Authorization", "Bearer " + apiKey)
        client := &http.Client{}
        response, err = client.Do(req)
        if err == nil {
            break
        }
        time.Sleep(time.Second)
    }

    if err != nil {
        return "", fmt.Errorf("failed to read %s: %v", url, err)
    }

    if response == nil {
        return "", fmt.Errorf("no response from %s", url)
    }

    if response.StatusCode != 200 {
        return "", fmt.Errorf("got %d response for %s", response.StatusCode, url)
    }

    body, error := ioutil.ReadAll(response.Body)
    if error != nil {
        return "", fmt.Errorf("could not read response body for %s: %v", url, err)
    }

    response.Body.Close()

    return string(body), nil
}

func GetJSON(path string, object interface{}) error {

    url := hostname + "/" + path

    var err error
    var response *http.Response
    for i := 0; i < 5; i++ {
        req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
        req.Header.Set("Authorization", "Bearer " + apiKey)
        client := &http.Client{}
        response, err = client.Do(req)
        if err == nil {
            break
        }
        time.Sleep(time.Second)
    }

    if err != nil {
        return fmt.Errorf("failed to read %s: %v", url, err)
    }

    if response == nil {
        return fmt.Errorf("no response from %s", url)
    }

    body, error := ioutil.ReadAll(response.Body)
    if error != nil {
        return fmt.Errorf("could not read response body for %s: %v", url, err)
    }

    response.Body.Close()

    err = json.Unmarshal([]byte(body), &object)
    if err != nil {
        return fmt.Errorf("could not parse json response for %s: %v", url, err)
    }

    return nil
}

func Create(path string, requestData interface{}, responseData interface{}) error {

    url := hostname + "/" + path

    buffer := new(bytes.Buffer)

    json.NewEncoder(buffer).Encode(requestData)

    request, err := http.NewRequest("POST", url, buffer)
    if err != nil {
        return fmt.Errorf("could not create HTTP POST request for %s: %v", url, err)
    }

    request.Header.Set("Authorization", "Bearer " + apiKey)

    httpClient := &http.Client{}

    var response *http.Response
    for i := 0; i < 5; i++ {
        response, err = httpClient.Do(request)
        if err == nil {
            break
        }
        time.Sleep(time.Second)
    }

    if err != nil {
        return fmt.Errorf("create failed on %s: %v", url, err)
    }

    if response == nil {
        return fmt.Errorf("no response from %s", url)
    }

    body, error := ioutil.ReadAll(response.Body)
    if error != nil {
        return fmt.Errorf("could not read response body for %s: %v", url, err)
    }

    response.Body.Close()

    err = json.Unmarshal([]byte(body), &responseData)
    if err != nil {
        return fmt.Errorf("could not parse json response for %s: %v", url, err)
    }

    return nil
}

func Update(path string, requestData interface{}, responseData interface{}) error {

    url := hostname + "/" + path

    buffer := new(bytes.Buffer)

    json.NewEncoder(buffer).Encode(requestData)

    request, _ := http.NewRequest("PUT", url, buffer)

    request.Header.Set("Authorization", "Bearer " + apiKey)

    httpClient := &http.Client{}

    var err error
    var response *http.Response
    for i := 0; i < 5; i++ {
        response, err = httpClient.Do(request)
        if err == nil {
            break
        }
        time.Sleep(time.Second)
    }

    if err != nil {
        return fmt.Errorf("failed to read %s: %v", url, err)
    }

    if response == nil {
        return fmt.Errorf("no response from %s", url)
    }

    body, error := ioutil.ReadAll(response.Body)
    if error != nil {
        return fmt.Errorf("could not read response body for %s: %v", url, err)
    }

    response.Body.Close()

    err = json.Unmarshal([]byte(body), &responseData)
    if err != nil {
        return fmt.Errorf("could not parse json response for %s: %v", url, err)
    }

    return nil
}

func Delete(path string, responseData interface{}) error {

    url := hostname + "/" + path

    request, _ := http.NewRequest("DELETE", url, nil)

    request.Header.Set("Authorization", "Bearer " + apiKey)

    httpClient := &http.Client{}

    var err error
    var response *http.Response
    for i := 0; i < 5; i++ {
        response, err = httpClient.Do(request)
        if err == nil {
            break
        }
        time.Sleep(time.Second)
    }

    if err != nil {
        return fmt.Errorf("failed to read %s: %v", url, err)
    }

    if response == nil {
        return fmt.Errorf("no response from %s", url)
    }

    body, error := ioutil.ReadAll(response.Body)
    if error != nil {
        return fmt.Errorf("could not read response body for %s: %v", url, err)
    }

    response.Body.Close()

    err = json.Unmarshal([]byte(body), &responseData)
    if err != nil {
        return fmt.Errorf("could not parse json response for %s: %v", url, err)
    }

    return err
}

// ----------------------------------------------------------------------------------------

type CreateCustomerResponse struct {
	Customer  admin.CustomerData   `json:"customer"`
	Error     string               `json:"error"`
}

type ReadCustomersResponse struct {
	Customers []admin.CustomerData `json:"customers"`
	Error     string               `json:"error"`
}

type ReadCustomerResponse struct {
	Customer  admin.CustomerData   `json:"customer"`
	Error     string               `json:"error"`
}

type UpdateCustomerResponse struct {
	Customer  admin.CustomerData   `json:"customer"`
	Error     string               `json:"error"`
}

type DeleteCustomerResponse struct {
	Error     string               `json:"error"`
}

// ----------------------------------------------------------------------------------------

func test_customers() {

	fmt.Printf("test_customers\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create customer

	customerId := uint64(0)
	{
		customer := admin.CustomerData{CustomerName: "Test", CustomerCode: "test", Live: true, Debug: true}

		var response CreateCustomerResponse

		err := Create("admin/create_customer", customer, &response)

		if err != nil {
			panic(err)
		}

		customerId = response.Customer.CustomerId
	}

	// read all customers
	{
		customersResponse := ReadCustomersResponse{}

		err := GetJSON("admin/customers", &customersResponse)

		if err != nil {
			panic(err)
		}

		if len(customersResponse.Customers) != 1 {
			panic("expect one customer in response")
		}

		if customersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if customersResponse.Customers[0].CustomerId != customerId {
			panic("wrong customer id")
		}

		if customersResponse.Customers[0].CustomerName != "Test" {
			panic("wrong customer name")
		}

		if customersResponse.Customers[0].CustomerCode != "test" {
			panic("wrong customer code")
		}

		if !customersResponse.Customers[0].Live {
			panic("customer should have live true")
		}

		if !customersResponse.Customers[0].Debug {
			panic("customer should have debug true")
		}
	}

	// read a specific customer

	// read all customers
	{
		customerResponse := ReadCustomerResponse{}

		err := GetJSON(fmt.Sprintf("admin/customer/%x", customerId), &customerResponse)

		if err != nil {
			panic(err)
		}

		if customerResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if customerResponse.Customer.CustomerId != customerId {
			panic(fmt.Sprintf("wrong customer id: got %x, expected %x", customerResponse.Customer.CustomerId, customerId))
		}

		if customerResponse.Customer.CustomerName != "Test" {
			panic("wrong customer name")
		}

		if customerResponse.Customer.CustomerCode != "test" {
			panic("wrong customer code")
		}

		if !customerResponse.Customer.Live {
			panic("customer should have live true")
		}

		if !customerResponse.Customer.Debug {
			panic("customer should have debug true")
		}
	}

	// update customer
	{
		customer := admin.CustomerData{CustomerId: customerId, CustomerName: "Updated", CustomerCode: "updated", Live: false, Debug: false}

		response := UpdateCustomerResponse{}

		err := Update("admin/update_customer", customer, &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Customer.CustomerId != customerId {
			panic("wrong customer id")
		}

		if response.Customer.CustomerName != "Updated" {
			panic("wrong customer name")
		}

		if response.Customer.CustomerCode != "updated" {
			panic("wrong customer code")
		}

		if response.Customer.Live {
			panic("customer should have live false")
		}

		if response.Customer.Debug {
			panic("customer should have debug false")
		}
	}

	// delete customer
	{
		response := UpdateCustomerResponse{}

		err := Delete(fmt.Sprintf("admin/delete_customer/%x", customerId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

/*
type BuyersResponse struct {
	Buyers []admin.BuyerData `json:"buyers"`
	Error  string            `json:"error"`
}

func test_buyers() {

	fmt.Printf("test_buyers\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create customer (needed by buyer)

	customerId := uint64(0)
	{
		customer := admin.CustomerData{CustomerName: "Test", CustomerCode: "test", Live: true, Debug: true}

		customerId = Create("http://127.0.0.1:50000/admin/create_customer", customer)
	}

	// create route shader (needed by buyer)

	routeShaderId := uint64(0)
	{
		routeShader := admin.RouteShaderData{}

		routeShaderId = Create("http://127.0.0.1:50000/admin/create_route_shader", routeShader)
	}

	// create buyer

	dummyBase64 := "oaneuthoanuthath"

	buyerId := uint64(0)
	{
		buyer := admin.BuyerData{BuyerName: "Buyer", PublicKeyBase64: dummyBase64, CustomerId: customerId, RouteShaderId: routeShaderId}

		buyerId = Create("http://127.0.0.1:50000/admin/create_buyer", buyer)
	}

	// read buyers
	{
		buyersResponse := BuyersResponse{}

		Read("http://127.0.0.1:50000/admin/buyers", &buyersResponse)

		if len(buyersResponse.Buyers) != 1 {
			panic("expect one buyer in response")
		}

		if buyersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if buyersResponse.Buyers[0].BuyerId != buyerId {
			panic("wrong buyer id")
		}

		if buyersResponse.Buyers[0].BuyerName != "Buyer" {
			panic("wrong buyer name")
		}

		if buyersResponse.Buyers[0].PublicKeyBase64 != dummyBase64 {
			panic("wrong public key base64")
		}

		if buyersResponse.Buyers[0].CustomerId != customerId {
			panic("wrong customer id")
		}

		if buyersResponse.Buyers[0].RouteShaderId != routeShaderId {
			panic("wrong route shader id")
		}
	}

	// update buyer
	{
		buyer := admin.BuyerData{BuyerId: buyerId, BuyerName: "Updated", PublicKeyBase64: dummyBase64, CustomerId: customerId, RouteShaderId: routeShaderId}

		Update("http://127.0.0.1:50000/admin/update_buyer", buyer)

		buyersResponse := BuyersResponse{}

		Read("http://127.0.0.1:50000/admin/buyers", &buyersResponse)

		if len(buyersResponse.Buyers) != 1 {
			panic("expect one buyer in response")
		}

		if buyersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if buyersResponse.Buyers[0].BuyerId != buyerId {
			panic("wrong buyer id")
		}

		if buyersResponse.Buyers[0].BuyerName != "Updated" {
			panic("wrong buyer name")
		}

		if buyersResponse.Buyers[0].PublicKeyBase64 != dummyBase64 {
			panic("wrong public key base64")
		}

		if buyersResponse.Buyers[0].CustomerId != customerId {
			panic("wrong customer id")
		}

		if buyersResponse.Buyers[0].RouteShaderId != routeShaderId {
			panic("wrong route shader id")
		}
	}

	// delete buyer
	{
		Delete("http://127.0.0.1:50000/admin/delete_buyer", buyerId)

		buyersResponse := BuyersResponse{}

		Read("http://127.0.0.1:50000/admin/buyers", &buyersResponse)

		if len(buyersResponse.Buyers) != 0 {
			panic("should be no buyers after delete")
		}

		if buyersResponse.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

type SellersResponse struct {
	Sellers []admin.SellerData `json:"sellers"`
	Error   string             `json:"error"`
}

func test_sellers() {

	fmt.Printf("test_sellers\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create seller

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Seller"}

		sellerId = Create("http://127.0.0.1:50000/admin/create_seller", seller)
	}

	// read sellers
	{
		sellersResponse := SellersResponse{}

		Read("http://127.0.0.1:50000/admin/sellers", &sellersResponse)

		if len(sellersResponse.Sellers) != 1 {
			panic("expect one seller in response")
		}

		if sellersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if sellersResponse.Sellers[0].SellerId != sellerId {
			panic("wrong seller id")
		}

		if sellersResponse.Sellers[0].SellerName != "Seller" {
			panic("wrong seller name")
		}

		if sellersResponse.Sellers[0].CustomerId != 0 {
			panic("wrong customer id")
		}
	}

	// update seller
	{
		seller := admin.SellerData{SellerId: sellerId, SellerName: "Updated"}

		Update("http://127.0.0.1:50000/admin/update_seller", seller)

		sellersResponse := SellersResponse{}

		Read("http://127.0.0.1:50000/admin/sellers", &sellersResponse)

		if len(sellersResponse.Sellers) != 1 {
			panic("expect one seller in response")
		}

		if sellersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if sellersResponse.Sellers[0].SellerId != sellerId {
			panic("wrong seller id")
		}

		if sellersResponse.Sellers[0].SellerName != "Updated" {
			panic("wrong seller name")
		}

		if sellersResponse.Sellers[0].CustomerId != 0 {
			panic("wrong customer id")
		}
	}

	// delete seller
	{
		Delete("http://127.0.0.1:50000/admin/delete_seller", sellerId)

		sellersResponse := SellersResponse{}

		Read("http://127.0.0.1:50000/admin/sellers", &sellersResponse)

		if len(sellersResponse.Sellers) != 0 {
			panic("should be no sellers after delete")
		}

		if sellersResponse.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

type DatacentersResponse struct {
	Datacenters []admin.DatacenterData `json:"datacenters"`
	Error       string                 `json:"error"`
}

func test_datacenters() {

	fmt.Printf("test_datacenters\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create seller (needed by datacenter)

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Seller"}

		sellerId = Create("http://127.0.0.1:50000/admin/create_seller", seller)
	}

	// create datacenter

	datacenterId := uint64(0)
	{
		datacenter := admin.DatacenterData{DatacenterName: "Datacenter", Latitude: 100, Longitude: 200, SellerId: sellerId}

		datacenterId = Create("http://127.0.0.1:50000/admin/create_datacenter", datacenter)
	}

	// read datacenters
	{
		datacentersResponse := DatacentersResponse{}

		Read("http://127.0.0.1:50000/admin/datacenters", &datacentersResponse)

		if len(datacentersResponse.Datacenters) != 1 {
			panic("expect one datacenter in response")
		}

		if datacentersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if datacentersResponse.Datacenters[0].DatacenterId != datacenterId {
			panic("wrong datacenter id")
		}

		if datacentersResponse.Datacenters[0].DatacenterName != "Datacenter" {
			panic("wrong datacenter name")
		}

		if datacentersResponse.Datacenters[0].Latitude != 100 {
			panic("wrong latitude")
		}

		if datacentersResponse.Datacenters[0].Longitude != 200 {
			panic("wrong longitude")
		}

		if datacentersResponse.Datacenters[0].Notes != "" {
			panic("notes should be empty")
		}
	}

	// update datacenter
	{
		datacenter := admin.DatacenterData{DatacenterId: datacenterId, DatacenterName: "Updated", Latitude: 110, Longitude: 220, Notes: "notes"}

		Update("http://127.0.0.1:50000/admin/update_datacenter", datacenter)

		datacentersResponse := DatacentersResponse{}

		Read("http://127.0.0.1:50000/admin/datacenters", &datacentersResponse)

		if len(datacentersResponse.Datacenters) != 1 {
			panic("expect one datacenter in response")
		}

		if datacentersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if datacentersResponse.Datacenters[0].DatacenterId != datacenterId {
			panic("wrong datacenter id")
		}

		if datacentersResponse.Datacenters[0].DatacenterName != "Updated" {
			panic("wrong datacenter name")
		}

		if datacentersResponse.Datacenters[0].Latitude != 110 {
			panic("wrong latitude")
		}

		if datacentersResponse.Datacenters[0].Longitude != 220 {
			panic("wrong longitude")
		}

		if datacentersResponse.Datacenters[0].Notes != "notes" {
			panic("wrong notes")
		}
	}

	// delete datacenter
	{
		Delete("http://127.0.0.1:50000/admin/delete_datacenter", datacenterId)

		datacentersResponse := DatacentersResponse{}

		Read("http://127.0.0.1:50000/admin/datacenters", &datacentersResponse)

		if len(datacentersResponse.Datacenters) != 0 {
			panic("should be no datacenters after delete")
		}

		if datacentersResponse.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

type RelaysResponse struct {
	Relays []admin.RelayData `json:"relays"`
	Error  string            `json:"error"`
}

func test_relays() {

	fmt.Printf("test_relays\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create seller (needed by datacenter)

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Seller"}

		sellerId = Create("http://127.0.0.1:50000/admin/create_seller", seller)
	}

	// create datacenter (needed by relay)

	datacenterId := uint64(0)
	{
		datacenter := admin.DatacenterData{DatacenterName: "Datacenter", Latitude: 100, Longitude: 200, SellerId: sellerId}

		datacenterId = Create("http://127.0.0.1:50000/admin/create_datacenter", datacenter)
	}

	// create relay

	relayId := uint64(0)
	{
		relay := admin.RelayData{RelayName: "Relay", DatacenterId: datacenterId, PublicIP: "127.0.0.1", InternalIP: "0.0.0.0", SSH_IP: "127.0.0.1"}

		relayId = Create("http://127.0.0.1:50000/admin/create_relay", relay)
	}

	// read relays
	{
		relaysResponse := RelaysResponse{}

		Read("http://127.0.0.1:50000/admin/relays", &relaysResponse)

		if len(relaysResponse.Relays) != 1 {
			panic("expect one relay in response")
		}

		if relaysResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if relaysResponse.Relays[0].RelayId != relayId {
			panic("wrong relay id")
		}

		if relaysResponse.Relays[0].RelayName != "Relay" {
			panic("wrong relay name")
		}

		if relaysResponse.Relays[0].PublicIP != "127.0.0.1" {
			panic("wrong public ip")
		}

		if relaysResponse.Relays[0].InternalIP != "0.0.0.0" {
			panic("wrong internal ip")
		}

		if relaysResponse.Relays[0].SSH_IP != "127.0.0.1" {
			panic("wrong ssh ip")
		}

		if relaysResponse.Relays[0].Notes != "" {
			panic("notes should be empty")
		}
	}

	// update relay
	{
		relay := admin.RelayData{RelayId: relayId, RelayName: "Updated", DatacenterId: datacenterId, Notes: "notes", PublicIP: "127.0.0.10", InternalIP: "5.5.5.5", SSH_IP: "127.0.0.10"}

		Update("http://127.0.0.1:50000/admin/update_relay", relay)

		relaysResponse := RelaysResponse{}

		Read("http://127.0.0.1:50000/admin/relays", &relaysResponse)

		if relaysResponse.Relays[0].RelayName != "Updated" {
			panic("wrong relay name")
		}

		if relaysResponse.Relays[0].PublicIP != "127.0.0.10" {
			panic("wrong public ip")
		}

		if relaysResponse.Relays[0].InternalIP != "5.5.5.5" {
			panic("wrong internal ip")
		}

		if relaysResponse.Relays[0].SSH_IP != "127.0.0.10" {
			panic("wrong ssh ip")
		}

		if relaysResponse.Relays[0].Notes != "notes" {
			panic("wrong notes")
		}
	}

	// delete relay
	{
		Delete("http://127.0.0.1:50000/admin/delete_relay", relayId)

		relaysResponse := RelaysResponse{}

		Read("http://127.0.0.1:50000/admin/relays", &relaysResponse)

		if len(relaysResponse.Relays) != 0 {
			panic("should be no relays after delete")
		}

		if relaysResponse.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

type RouteShadersResponse struct {
	RouteShaders []admin.RouteShaderData `json:"route_shaders"`
	Error        string                  `json:"error"`
}

func test_route_shaders() {

	fmt.Printf("test_route_shaders\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create route shader

	routeShaderId := uint64(0)
	{
		routeShader := admin.RouteShaderData{RouteShaderName: "Route Shader"}

		routeShaderId = Create("http://127.0.0.1:50000/admin/create_route_shader", routeShader)
	}

	// read route shaders
	{
		routeShadersResponse := RouteShadersResponse{}

		Read("http://127.0.0.1:50000/admin/route_shaders", &routeShadersResponse)

		if routeShadersResponse.RouteShaders[0].RouteShaderName != "Route Shader" {
			panic("wrong route shader name")
		}

		if len(routeShadersResponse.RouteShaders) != 1 {
			panic("expect one route shader in response")
		}

		if routeShadersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if routeShadersResponse.RouteShaders[0].RouteShaderId != routeShaderId {
			panic("wrong route shader id")
		}
	}

	// update route shader
	{
		routeShader := admin.RouteShaderData{RouteShaderId: routeShaderId, RouteShaderName: "Updated"}

		Update("http://127.0.0.1:50000/admin/update_route_shader", routeShader)

		routeShadersResponse := RouteShadersResponse{}

		Read("http://127.0.0.1:50000/admin/route_shaders", &routeShadersResponse)

		if routeShadersResponse.RouteShaders[0].RouteShaderName != "Updated" {
			panic("wrong route shader name")
		}

		if len(routeShadersResponse.RouteShaders) != 1 {
			panic("expect one route shader in response")
		}

		if routeShadersResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if routeShadersResponse.RouteShaders[0].RouteShaderId != routeShaderId {
			panic("wrong route shader id")
		}
	}

	// delete route shader
	{
		Delete("http://127.0.0.1:50000/admin/delete_route_shader", routeShaderId)

		routeShadersResponse := RouteShadersResponse{}

		Read("http://127.0.0.1:50000/admin/route_shaders", &routeShadersResponse)

		if len(routeShadersResponse.RouteShaders) != 0 {
			panic("should be no route shaders after delete")
		}

		if routeShadersResponse.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

type BuyerDatacenterSettingsResponse struct {
	Settings []admin.BuyerDatacenterSettings `json:"buyer_datacenter_settings"`
	Error    string                          `json:"error"`
}

func test_buyer_datacenter_settings() {

	fmt.Printf("test_buyer_datacenter_settings\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create customer

	customerId := uint64(0)
	{
		customer := admin.CustomerData{CustomerName: "Test", CustomerCode: "test", Live: true, Debug: true}

		customerId = Create("http://127.0.0.1:50000/admin/create_customer", customer)
	}

	// create route shader

	routeShaderId := uint64(0)
	{
		routeShader := admin.RouteShaderData{}

		routeShaderId = Create("http://127.0.0.1:50000/admin/create_route_shader", routeShader)
	}

	// create buyer

	dummyBase64 := "oaneuthoanuthath"

	buyerId := uint64(0)
	{
		buyer := admin.BuyerData{BuyerName: "Buyer", PublicKeyBase64: dummyBase64, CustomerId: customerId, RouteShaderId: routeShaderId}

		buyerId = Create("http://127.0.0.1:50000/admin/create_buyer", buyer)
	}

	// create seller

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Seller"}

		sellerId = Create("http://127.0.0.1:50000/admin/create_seller", seller)
	}

	// create datacenter

	datacenterId := uint64(0)
	{
		datacenter := admin.DatacenterData{DatacenterName: "Datacenter", Latitude: 100, Longitude: 200, SellerId: sellerId}

		datacenterId = Create("http://127.0.0.1:50000/admin/create_datacenter", datacenter)
	}

	// create buyer datacenter settings
	{
		settings := admin.BuyerDatacenterSettings{BuyerId: buyerId, DatacenterId: datacenterId, EnableAcceleration: true}

		Create("http://127.0.0.1:50000/admin/create_buyer_datacenter_settings", settings)
	}

	// read buyer datacenter settings
	{
		buyerDatacenterSettingsResponse := BuyerDatacenterSettingsResponse{}

		Read("http://127.0.0.1:50000/admin/buyer_datacenter_settings", &buyerDatacenterSettingsResponse)

		if len(buyerDatacenterSettingsResponse.Settings) != 1 {
			panic(fmt.Sprintf("expect one settings in response, got %d", len(buyerDatacenterSettingsResponse.Settings)))
		}

		if buyerDatacenterSettingsResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if buyerDatacenterSettingsResponse.Settings[0].BuyerId != buyerId {
			panic("wrong buyer id")
		}

		if buyerDatacenterSettingsResponse.Settings[0].DatacenterId != datacenterId {
			panic("wrong datacenter id")
		}

		if buyerDatacenterSettingsResponse.Settings[0].EnableAcceleration != true {
			panic("wrong enable acceleration")
		}
	}

	// update buyer datacenter settings
	{
		settings := admin.BuyerDatacenterSettings{BuyerId: buyerId, DatacenterId: datacenterId, EnableAcceleration: false}

		Update("http://127.0.0.1:50000/admin/update_buyer_datacenter_settings", settings)

		buyerDatacenterSettingsResponse := BuyerDatacenterSettingsResponse{}

		Read("http://127.0.0.1:50000/admin/buyer_datacenter_settings", &buyerDatacenterSettingsResponse)

		if len(buyerDatacenterSettingsResponse.Settings) != 1 {
			panic(fmt.Sprintf("expect one settings in response, got %d", len(buyerDatacenterSettingsResponse.Settings)))
		}

		if buyerDatacenterSettingsResponse.Error != "" {
			panic("expect error string to be empty")
		}

		if buyerDatacenterSettingsResponse.Settings[0].BuyerId != buyerId {
			panic("wrong buyer id")
		}

		if buyerDatacenterSettingsResponse.Settings[0].DatacenterId != datacenterId {
			panic("wrong datacenter id")
		}

		if buyerDatacenterSettingsResponse.Settings[0].EnableAcceleration != false {
			panic("wrong enable acceleration")
		}
	}

	// delete buyer datacenter settings
	{
		Delete(fmt.Sprintf("http://127.0.0.1:50000/admin/delete_buyer_datacenter_settings/%d/%d", buyerId, datacenterId), 1)

		buyerDatacenterSettingsResponse := BuyerDatacenterSettingsResponse{}

		Read("http://127.0.0.1:50000/admin/buyer_datacenter_settings", &buyerDatacenterSettingsResponse)

		if len(buyerDatacenterSettingsResponse.Settings) != 0 {
			panic("should be no settings after delete")
		}

		if buyerDatacenterSettingsResponse.Error != "" {
			panic("expect error string to be empty")
		}
	}

	// get route shader defaults
	{
		routeShader := admin.RouteShaderData{}

		Read("http://127.0.0.1:50000/admin/route_shader_defaults", &routeShader)
	}
}
*/

// ----------------------------------------------------------------------------------------

type test_function func()

func main() {

	allTests := []test_function{
		test_customers,
		/*
		test_buyers,
		test_sellers,
		test_datacenters,
		test_relays,
		test_route_shaders,
		test_buyer_datacenter_settings,
		*/
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
