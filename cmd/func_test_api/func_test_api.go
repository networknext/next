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

type CreateSellerResponse struct {
	Seller     admin.SellerData    `json:"seller"`
	Error     string               `json:"error"`
}

type ReadSellersResponse struct {
	Sellers   []admin.SellerData   `json:"sellers"`
	Error     string               `json:"error"`
}

type ReadSellerResponse struct {
	Seller     admin.SellerData    `json:"seller"`
	Error     string               `json:"error"`
}

type UpdateSellerResponse struct {
	Seller     admin.SellerData    `json:"seller"`
	Error     string               `json:"error"`
}

type DeleteSellerResponse struct {
	Error     string               `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateDatacenterResponse struct {
	Datacenter     admin.DatacenterData    `json:"datacenter"`
	Error          string               	`json:"error"`
}

type ReadDatacentersResponse struct {
	Datacenters    []admin.DatacenterData  `json:"datacenters"`
	Error          string               	`json:"error"`
}

type ReadDatacenterResponse struct {
	Datacenter     admin.DatacenterData    `json:"datacenter"`
	Error          string               	`json:"error"`
}

type UpdateDatacenterResponse struct {
	Datacenter     admin.DatacenterData    `json:"datacenter"`
	Error          string               	`json:"error"`
}

type DeleteDatacenterResponse struct {
	Error          string               	`json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateRelayResponse struct {
	Relay          admin.RelayData         `json:"relay"`
	Error          string               	`json:"error"`
}

type ReadRelaysResponse struct {
	Relays    		[]admin.RelayData  		`json:"relays"`
	Error          string               	`json:"error"`
}

type ReadRelayResponse struct {
	Relay          admin.RelayData         `json:"relay"`
	Error          string               	`json:"error"`
}

type UpdateRelayResponse struct {
	Relay          admin.RelayData         `json:"relay"`
	Error          string               	`json:"error"`
}

type DeleteRelayResponse struct {
	Error          string               	`json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateRouteShaderResponse struct {
	RouteShader    admin.RouteShaderData   `json:"route_shader"`
	Error          string               	`json:"error"`
}

type ReadRouteShadersResponse struct {
	RouteShaders   []admin.RouteShaderData `json:"route_shaders"`
	Error          string               	`json:"error"`
}

type ReadRouteShaderResponse struct {
	RouteShader    admin.RouteShaderData   `json:"route_shader"`
	Error          string               	`json:"error"`
}

type UpdateRouteShaderResponse struct {
	RouteShader    admin.RouteShaderData   `json:"route_shader"`
	Error          string               	`json:"error"`
}

type DeleteRouteShaderResponse struct {
	Error          string               	`json:"error"`
}

// ----------------------------------------------------------------------------------------

func test_customer() {

	fmt.Printf("\ntest_customer\n\n")

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
	{
		response := ReadCustomerResponse{}

		err := GetJSON(fmt.Sprintf("admin/customer/%x", customerId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Customer.CustomerId != customerId {
			panic(fmt.Sprintf("wrong customer id: got %x, expected %x", response.Customer.CustomerId, customerId))
		}

		if response.Customer.CustomerName != "Test" {
			panic("wrong customer name")
		}

		if response.Customer.CustomerCode != "test" {
			panic("wrong customer code")
		}

		if !response.Customer.Live {
			panic("customer should have live true")
		}

		if !response.Customer.Debug {
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

func test_seller() {

	fmt.Printf("\ntest_seller\n\n")

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

	// create seller

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Test"}

		var response CreateSellerResponse

		err := Create("admin/create_seller", seller, &response)

		if err != nil {
			panic(err)
		}

		sellerId = response.Seller.SellerId
	}

	// read all sellers
	{
		response := ReadSellersResponse{}

		err := GetJSON("admin/sellers", &response)

		if err != nil {
			panic(err)
		}

		if len(response.Sellers) != 1 {
			panic("expect one seller in response")
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Sellers[0].SellerId != sellerId {
			panic("wrong seller id")
		}

		if response.Sellers[0].SellerName != "Test" {
			panic("wrong seller name")
		}

		if response.Sellers[0].CustomerId != 0 {
			panic(fmt.Sprintf("wrong customer id on seller. expected %d, got %d", 0, response.Sellers[0].CustomerId))
		}
	}

	// read a specific seller
	{
		response := ReadSellerResponse{}

		err := GetJSON(fmt.Sprintf("admin/seller/%x", sellerId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Seller.SellerId != sellerId {
			panic(fmt.Sprintf("wrong seller id: got %x, expected %x", response.Seller.SellerId, sellerId))
		}

		if response.Seller.SellerName != "Test" {
			panic("wrong seller name")
		}

		if response.Seller.CustomerId != 0 {
			panic(fmt.Sprintf("wrong customer id on seller. expected %d, got %d", 0, response.Seller.CustomerId))
		}
	}

	// update seller
	{
		seller := admin.SellerData{SellerId: sellerId, SellerName: "Updated", CustomerId: customerId}

		response := UpdateSellerResponse{}

		err := Update("admin/update_seller", seller, &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Seller.SellerId != sellerId {
			panic("wrong seller id")
		}

		if response.Seller.SellerName != "Updated" {
			panic("wrong seller name")
		}

		if response.Seller.CustomerId != customerId {
			panic(fmt.Sprintf("wrong customer id on seller. expected %d, got %d", customerId, response.Seller.CustomerId))
		}
	}

	// delete seller
	{
		response := UpdateSellerResponse{}

		err := Delete(fmt.Sprintf("admin/delete_seller/%x", sellerId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

func test_datacenter() {

	fmt.Printf("\ntest_datacenter\n\n")

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

	// create seller

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Test", CustomerId: customerId}

		var response CreateSellerResponse

		err := Create("admin/create_seller", seller, &response)

		if err != nil {
			panic(err)
		}

		sellerId = response.Seller.SellerId
	}

	// create datacenter

	datacenterId := uint64(0)
	{
		datacenter := admin.DatacenterData{DatacenterName: "Test", Latitude: 100, Longitude: 200, SellerId: sellerId}

		var response CreateDatacenterResponse

		err := Create("admin/create_datacenter", datacenter, &response)

		if err != nil {
			panic(err)
		}

		datacenterId = response.Datacenter.DatacenterId
	}

	// read all datacenters
	{
		response := ReadDatacentersResponse{}

		err := GetJSON("admin/datacenters", &response)

		if err != nil {
			panic(err)
		}

		if len(response.Datacenters) != 1 {
			panic(fmt.Sprintf("expect one datacenter in response, got %d", len(response.Datacenters)))
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Datacenters[0].DatacenterId != datacenterId {
			panic("wrong datacenter id")
		}

		if response.Datacenters[0].DatacenterName != "Test" {
			panic("wrong datacenter name")
		}

		if response.Datacenters[0].Latitude != 100 {
			panic("wrong latitude")
		}

		if response.Datacenters[0].Longitude != 200 {
			panic("wrong longitude")
		}

		if response.Datacenters[0].SellerId != sellerId {
			panic("wrong seller id on datacenter")
		}

		if response.Datacenters[0].Notes != "" {
			panic("notes should be empty")
		}
	}

	// read a specific datacenter
	{
		response := ReadDatacenterResponse{}

		err := GetJSON(fmt.Sprintf("admin/datacenter/%x", datacenterId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Datacenter.DatacenterId != datacenterId {
			panic(fmt.Sprintf("wrong datacenter id: got %x, expected %x", response.Datacenter.DatacenterId, datacenterId))
		}

		if response.Datacenter.DatacenterName != "Test" {
			panic("wrong datacenter name")
		}

		if response.Datacenter.Latitude != 100.0 {
			panic("wrong latitude on datacenter")
		}

		if response.Datacenter.Longitude != 200.0 {
			panic("wrong longitude on datacenter")
		}

		if response.Datacenter.SellerId != sellerId {
			panic("wrong seller id on datacenter")
		}
	}

	// update datacenter
	{
		datacenter := admin.DatacenterData{DatacenterId: datacenterId, DatacenterName: "Updated", Latitude: 50, Longitude: 75, SellerId: sellerId}

		response := UpdateDatacenterResponse{}

		err := Update("admin/update_datacenter", datacenter, &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Datacenter.DatacenterId != datacenterId {
			panic("wrong datacenter id")
		}

		if response.Datacenter.DatacenterName != "Updated" {
			panic("wrong datacenter name after update")
		}

		if response.Datacenter.Latitude != 50.0 {
			panic("wrong latitude on datacenter")
		}

		if response.Datacenter.Longitude != 75.0 {
			panic("wrong longitude on datacenter")
		}

		if response.Datacenter.SellerId != sellerId {
			panic("wrong seller id on datacenter")
		}
	}

	// delete datacenter
	{
		response := UpdateDatacenterResponse{}

		err := Delete(fmt.Sprintf("admin/delete_datacenter/%x", datacenterId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

func test_relay() {

	fmt.Printf("\ntest_relay\n\n")

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

	// create seller

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Test", CustomerId: customerId}

		var response CreateSellerResponse

		err := Create("admin/create_seller", seller, &response)

		if err != nil {
			panic(err)
		}

		sellerId = response.Seller.SellerId
	}

	// create datacenter

	datacenterId := uint64(0)
	{
		datacenter := admin.DatacenterData{DatacenterName: "Test", Latitude: 100, Longitude: 200, SellerId: sellerId}

		var response CreateDatacenterResponse

		err := Create("admin/create_datacenter", datacenter, &response)

		if err != nil {
			panic(err)
		}

		datacenterId = response.Datacenter.DatacenterId
	}

	// create relay

	relayId := uint64(0)
	{
		relay := admin.RelayData{RelayName: "Test", DatacenterId: datacenterId, PublicIP: "127.0.0.1", InternalIP: "0.0.0.0", SSH_IP: "127.0.0.1"}

		var response CreateRelayResponse

		err := Create("admin/create_relay", relay, &response)

		if err != nil {
			panic(err)
		}

		relayId = response.Relay.RelayId
	}

	// read all relays
	{
		response := ReadRelaysResponse{}

		err := GetJSON("admin/relays", &response)

		if err != nil {
			panic(err)
		}

		if len(response.Relays) != 1 {
			panic(fmt.Sprintf("expect one relay in response, got %d", len(response.Relays)))
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Relays[0].RelayId != relayId {
			panic("wrong relay id")
		}

		if response.Relays[0].RelayName != "Test" {
			panic("wrong relay name")
		}

		if response.Relays[0].PublicIP != "127.0.0.1" {
			panic("wrong public ip")
		}

		if response.Relays[0].InternalIP != "0.0.0.0" {
			panic("wrong internal ip")
		}

		if response.Relays[0].SSH_IP != "127.0.0.1" {
			panic("wrong ssh ip")
		}

		if response.Relays[0].Notes != "" {
			panic("notes should be empty")
		}
	}

	// read a specific relay
	{
		response := ReadRelayResponse{}

		err := GetJSON(fmt.Sprintf("admin/relay/%x", relayId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Relay.RelayId != relayId {
			panic(fmt.Sprintf("wrong relay id: got %x, expected %x", response.Relay.RelayId, relayId))
		}

		if response.Relay.RelayName != "Test" {
			panic("wrong relay name")
		}

		if response.Relay.PublicIP != "127.0.0.1" {
			panic("wrong public ip")
		}

		if response.Relay.InternalIP != "0.0.0.0" {
			panic("wrong internal ip")
		}

		if response.Relay.SSH_IP != "127.0.0.1" {
			panic("wrong ssh ip")
		}

		if response.Relay.Notes != "" {
			panic("notes should be empty")
		}
	}

	// update relay
	{
		relay := admin.RelayData{RelayId: relayId, RelayName: "Updated", DatacenterId: datacenterId, PublicIP: "127.0.0.1", InternalIP: "0.0.0.0", SSH_IP: "127.0.0.1"}

		response := UpdateRelayResponse{}

		err := Update("admin/update_relay", relay, &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Relay.RelayId != relayId {
			panic("wrong relay id")
		}

		if response.Relay.RelayName != "Updated" {
			panic("wrong relay name after update")
		}

		if response.Relay.PublicIP != "127.0.0.1" {
			panic("wrong public ip")
		}

		if response.Relay.InternalIP != "0.0.0.0" {
			panic("wrong internal ip")
		}

		if response.Relay.SSH_IP != "127.0.0.1" {
			panic("wrong ssh ip")
		}

		if response.Relay.Notes != "" {
			panic("notes should be empty")
		}
	}

	// delete relay
	{
		response := UpdateRelayResponse{}

		err := Delete(fmt.Sprintf("admin/delete_relay/%x", relayId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

func test_route_shader() {

	fmt.Printf("\ntest_route_shader\n\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create route shader

	expected := admin.RouteShaderData{
		RouteShaderName: "Test",
		ABTest: true,
		AcceptableLatency: 10.0,
		AcceptablePacketLoss: 0.1,
		PacketLossSustained: 0.25,
		AnalysisOnly: true,
		BandwidthEnvelopeUpKbps: 1024,
		BandwidthEnvelopeDownKbps: 512,
		DisableNetworkNext: true,
		LatencyThreshold: 10.0,
		Multipath: true,
		ReduceLatency: true,
		ReducePacketLoss: true,
		SelectionPercent: 100,
		MaxLatencyTradeOff: 20,
		MaxNextRTT: 200,
		RouteSwitchThreshold: 10,
		RouteSelectThreshold: 5,
		RTTVeto_Default: 10,
		RTTVeto_Multipath: 20,
		RTTVeto_PacketLoss: 30,
		ForceNext: true,
		RouteDiversity: 5,
	}

	routeShaderId := uint64(0)
	{
		routeShader := expected

		var response CreateRouteShaderResponse

		err := Create("admin/create_route_shader", routeShader, &response)

		if err != nil {
			panic(err)
		}

		routeShaderId = response.RouteShader.RouteShaderId

		expected.RouteShaderId = routeShaderId
	}

	// read all route shaders
	{
		response := ReadRouteShadersResponse{}

		err := GetJSON("admin/route_shaders", &response)

		if err != nil {
			panic(err)
		}

		if len(response.RouteShaders) != 1 {
			panic(fmt.Sprintf("expect one route shader in response, got %d", len(response.RouteShaders)))
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.RouteShaders[0] != expected {
			panic("route shader does not match expected")
		}
	}

	// read a specific route shader
	{
		response := ReadRouteShaderResponse{}

		err := GetJSON(fmt.Sprintf("admin/route_shader/%x", routeShaderId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.RouteShader != expected {
			panic("route shader does not match expected")
		}
	}

	// update route shader
	{
		expected.RouteShaderName = "Updated"

		routeShader := expected

		response := UpdateRouteShaderResponse{}

		err := Update("admin/update_route_shader", routeShader, &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.RouteShader != expected {
			panic("route shader does not match expected")
		}
	}

	// delete route shader
	{
		response := UpdateRouteShaderResponse{}

		err := Delete(fmt.Sprintf("admin/delete_route_shader/%x", routeShaderId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

/*
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
		routeShader := admin.RouteShaderData{
			RouteShaderName: "Route Shader"
		}

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
*/

// ----------------------------------------------------------------------------------------

// todo: come back once route shader is tested

/*
func test_buyer() {

	fmt.Printf("test_buyer\n")

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

	// create buyer

	dummyBase64 := "oaneuthoanuthath"

	buyerId := uint64(0)
	{
		buyer := admin.BuyerData{BuyerName: "Buyer", PublicKeyBase64: dummyBase64, CustomerId: customerId, RouteShaderId: routeShaderId}

		var response CreateBuyerResponse

		err := Create("admin/create_buyer", buyer, &response)

		if err != nil {
			panic(err)
		}

		buyerId = response.Buyer.BuyerId
	}

	// todo
	_ = buyerId

	/*
	// read all buyers
	{
		buyersResponse := ReadCustomersResponse{}

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
		test_customer,
		test_seller,
		test_datacenter,
		test_relay,
		test_route_shader,
		/*
		test_buyer,
		test_buyer_datacenter_setting,
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


/*/
type CreateBuyerResponse struct {
	Buyer     admin.BuyerData      `json:"buyer"`
	Error     string               `json:"error"`
}

type ReadBuyersResponse struct {
	Buyers    []admin.BuyerData    `json:"buyers"`
	Error     string               `json:"error"`
}

type ReadBuyerResponse struct {
	Buyer     admin.BuyerData      `json:"buyer"`
	Error     string               `json:"error"`
}

type UpdateBuyerResponse struct {
	Buyer     admin.BuyerData      `json:"buyer"`
	Error     string               `json:"error"`
}

type DeleteBuyerResponse struct {
	Error     string               `json:"error"`
}
*/

