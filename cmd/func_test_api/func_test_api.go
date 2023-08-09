/*
   Network Next Accelerate.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"

	"github.com/networknext/next/modules/admin"
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
		req.Header.Set("Authorization", "Bearer "+apiKey)
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
		req.Header.Set("Authorization", "Bearer "+apiKey)
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

	request.Header.Set("Authorization", "Bearer "+apiKey)

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

	request.Header.Set("Authorization", "Bearer "+apiKey)

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

	request.Header.Set("Authorization", "Bearer "+apiKey)

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
	Customer admin.CustomerData `json:"customer"`
	Error    string             `json:"error"`
}

type ReadCustomersResponse struct {
	Customers []admin.CustomerData `json:"customers"`
	Error     string               `json:"error"`
}

type ReadCustomerResponse struct {
	Customer admin.CustomerData `json:"customer"`
	Error    string             `json:"error"`
}

type UpdateCustomerResponse struct {
	Customer admin.CustomerData `json:"customer"`
	Error    string             `json:"error"`
}

type DeleteCustomerResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateSellerResponse struct {
	Seller admin.SellerData `json:"seller"`
	Error  string           `json:"error"`
}

type ReadSellersResponse struct {
	Sellers []admin.SellerData `json:"sellers"`
	Error   string             `json:"error"`
}

type ReadSellerResponse struct {
	Seller admin.SellerData `json:"seller"`
	Error  string           `json:"error"`
}

type UpdateSellerResponse struct {
	Seller admin.SellerData `json:"seller"`
	Error  string           `json:"error"`
}

type DeleteSellerResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateDatacenterResponse struct {
	Datacenter admin.DatacenterData `json:"datacenter"`
	Error      string               `json:"error"`
}

type ReadDatacentersResponse struct {
	Datacenters []admin.DatacenterData `json:"datacenters"`
	Error       string                 `json:"error"`
}

type ReadDatacenterResponse struct {
	Datacenter admin.DatacenterData `json:"datacenter"`
	Error      string               `json:"error"`
}

type UpdateDatacenterResponse struct {
	Datacenter admin.DatacenterData `json:"datacenter"`
	Error      string               `json:"error"`
}

type DeleteDatacenterResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateRelayResponse struct {
	Relay admin.RelayData `json:"relay"`
	Error string          `json:"error"`
}

type ReadRelaysResponse struct {
	Relays []admin.RelayData `json:"relays"`
	Error  string            `json:"error"`
}

type ReadRelayResponse struct {
	Relay admin.RelayData `json:"relay"`
	Error string          `json:"error"`
}

type UpdateRelayResponse struct {
	Relay admin.RelayData `json:"relay"`
	Error string          `json:"error"`
}

type DeleteRelayResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateRouteShaderResponse struct {
	RouteShader admin.RouteShaderData `json:"route_shader"`
	Error       string                `json:"error"`
}

type ReadRouteShadersResponse struct {
	RouteShaders []admin.RouteShaderData `json:"route_shaders"`
	Error        string                  `json:"error"`
}

type ReadRouteShaderResponse struct {
	RouteShader admin.RouteShaderData `json:"route_shader"`
	Error       string                `json:"error"`
}

type UpdateRouteShaderResponse struct {
	RouteShader admin.RouteShaderData `json:"route_shader"`
	Error       string                `json:"error"`
}

type DeleteRouteShaderResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateBuyerResponse struct {
	Buyer admin.BuyerData `json:"buyer"`
	Error string          `json:"error"`
}

type ReadBuyersResponse struct {
	Buyers []admin.BuyerData `json:"buyers"`
	Error  string            `json:"error"`
}

type ReadBuyerResponse struct {
	Buyer admin.BuyerData `json:"buyer"`
	Error string          `json:"error"`
}

type UpdateBuyerResponse struct {
	Buyer admin.BuyerData `json:"buyer"`
	Error string          `json:"error"`
}

type DeleteBuyerResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateBuyerDatacenterSettingsResponse struct {
	Settings admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                        `json:"error"`
}

type ReadBuyerDatacenterSettingsListResponse struct {
	Settings []admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                          `json:"error"`
}

type ReadBuyerDatacenterSettingsResponse struct {
	Settings admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                        `json:"error"`
}

type UpdateBuyerDatacenterSettingsResponse struct {
	Settings admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                        `json:"error"`
}

type DeleteBuyerDatacenterSettingsResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateRelayKeypairResponse struct {
	RelayKeypair admin.RelayKeypairData `json:"relay_keypair"`
	Error        string                 `json:"error"`
}

type ReadRelayKeypairsResponse struct {
	RelayKeypairs []admin.RelayKeypairData `json:"relay_keypairs"`
	Error         string                   `json:"error"`
}

type ReadRelayKeypairResponse struct {
	RelayKeypair admin.RelayKeypairData `json:"relay_keypair"`
	Error        string                 `json:"error"`
}

type UpdateRelayKeypairResponse struct {
	RelayKeypair admin.RelayKeypairData `json:"relay_keypair"`
	Error        string                 `json:"error"`
}

type DeleteRelayKeypairResponse struct {
	Error string `json:"error"`
}

// ----------------------------------------------------------------------------------------

type CreateBuyerKeypairResponse struct {
	BuyerKeypair admin.BuyerKeypairData `json:"buyer_keypair"`
	Error        string                 `json:"error"`
}

type ReadBuyerKeypairsResponse struct {
	BuyerKeypairs []admin.BuyerKeypairData `json:"buyer_keypairs"`
	Error         string                   `json:"error"`
}

type ReadBuyerKeypairResponse struct {
	BuyerKeypair admin.BuyerKeypairData `json:"buyer_keypair"`
	Error        string                 `json:"error"`
}

type UpdateBuyerKeypairResponse struct {
	BuyerKeypair admin.BuyerKeypairData `json:"buyer_keypair"`
	Error        string                 `json:"error"`
}

type DeleteBuyerKeypairResponse struct {
	Error string `json:"error"`
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
		response := DeleteCustomerResponse{}

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
		response := DeleteSellerResponse{}

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
		response := DeleteDatacenterResponse{}

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
		response := DeleteRelayResponse{}

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
		RouteShaderName:               "Test",
		ABTest:                        true,
		AcceptableLatency:             10.0,
		AcceptablePacketLossInstant:   1.0,
		AcceptablePacketLossSustained: 0.1,
		AnalysisOnly:                  true,
		BandwidthEnvelopeUpKbps:       1024,
		BandwidthEnvelopeDownKbps:     512,
		DisableNetworkNext:            true,
		LatencyReductionThreshold:     10.0,
		Multipath:                     true,
		SelectionPercent:              100,
		MaxLatencyTradeOff:            20,
		MaxNextRTT:                    200,
		RouteSwitchThreshold:          10,
		RouteSelectThreshold:          5,
		RTTVeto:                       10,
		ForceNext:                     true,
		RouteDiversity:                5,
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
		response := DeleteRouteShaderResponse{}

		err := Delete(fmt.Sprintf("admin/delete_route_shader/%x", routeShaderId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

func test_buyer() {

	fmt.Printf("\ntest_buyer\n\n")

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

	// create route shader

	routeShaderId := uint64(0)
	{
		routeShader := admin.RouteShaderData{
			RouteShaderName:               "Test",
			ABTest:                        true,
			AcceptableLatency:             10.0,
			AcceptablePacketLossInstant:   0.25,
			AcceptablePacketLossSustained: 0.01,
			AnalysisOnly:                  true,
			BandwidthEnvelopeUpKbps:       1024,
			BandwidthEnvelopeDownKbps:     512,
			DisableNetworkNext:            true,
			LatencyReductionThreshold:     10.0,
			Multipath:                     true,
			SelectionPercent:              100,
			MaxLatencyTradeOff:            20,
			MaxNextRTT:                    200,
			RouteSwitchThreshold:          10,
			RouteSelectThreshold:          5,
			RTTVeto:                       10,
			ForceNext:                     true,
			RouteDiversity:                5,
		}

		var response CreateRouteShaderResponse

		err := Create("admin/create_route_shader", routeShader, &response)

		if err != nil {
			panic(err)
		}

		routeShaderId = response.RouteShader.RouteShaderId
	}

	// create buyer

	expected := admin.BuyerData{
		BuyerName:       "Test",
		CustomerId:      customerId,
		RouteShaderId:   routeShaderId,
		PublicKeyBase64: "@@!#$@!$@#!R*$!@*R",
	}

	buyerId := uint64(0)
	{
		buyer := expected

		var response CreateBuyerResponse

		err := Create("admin/create_buyer", buyer, &response)

		if err != nil {
			panic(err)
		}

		buyerId = response.Buyer.BuyerId

		expected.BuyerId = buyerId
	}

	// read all buyers
	{
		response := ReadBuyersResponse{}

		err := GetJSON("admin/buyers", &response)

		if err != nil {
			panic(err)
		}

		if len(response.Buyers) != 1 {
			panic(fmt.Sprintf("expect one buyer in response, got %d", len(response.Buyers)))
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Buyers[0] != expected {
			panic("buyer does not match expected")
		}
	}

	// read a specific buyer
	{
		response := ReadBuyerResponse{}

		err := GetJSON(fmt.Sprintf("admin/buyer/%x", buyerId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Buyer != expected {
			panic("buyer does not match expected")
		}
	}

	// update buyer
	{
		expected.BuyerName = "Updated"

		buyer := expected

		response := UpdateBuyerResponse{}

		err := Update("admin/update_buyer", buyer, &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Buyer != expected {
			panic("buyer does not match expected")
		}
	}

	// delete route shader
	{
		response := DeleteBuyerResponse{}

		err := Delete(fmt.Sprintf("admin/delete_buyer/%x", buyerId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

func test_buyer_datacenter_settings() {

	fmt.Printf("\ntest_buyer_datacenter_settings\n\n")

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

	// create route shader

	routeShaderId := uint64(0)
	{
		routeShader := admin.RouteShaderData{
			RouteShaderName:               "Test",
			ABTest:                        true,
			AcceptableLatency:             10.0,
			AcceptablePacketLossInstant:   0.25,
			AcceptablePacketLossSustained: 0.01,
			AnalysisOnly:                  true,
			BandwidthEnvelopeUpKbps:       1024,
			BandwidthEnvelopeDownKbps:     512,
			DisableNetworkNext:            true,
			LatencyReductionThreshold:     10.0,
			Multipath:                     true,
			SelectionPercent:              100,
			MaxLatencyTradeOff:            20,
			MaxNextRTT:                    200,
			RouteSwitchThreshold:          10,
			RouteSelectThreshold:          5,
			RTTVeto:                       10,
			ForceNext:                     true,
			RouteDiversity:                5,
		}

		var response CreateRouteShaderResponse

		err := Create("admin/create_route_shader", routeShader, &response)

		if err != nil {
			panic(err)
		}

		routeShaderId = response.RouteShader.RouteShaderId
	}

	// create buyer

	buyerId := uint64(0)
	{
		buyer := admin.BuyerData{
			BuyerName:       "Test",
			CustomerId:      customerId,
			RouteShaderId:   routeShaderId,
			PublicKeyBase64: "@@!#$@!$@#!R*$!@*R",
		}

		var response CreateBuyerResponse

		err := Create("admin/create_buyer", buyer, &response)

		if err != nil {
			panic(err)
		}

		buyerId = response.Buyer.BuyerId
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

	// create buyer datacenter settings

	expected := admin.BuyerDatacenterSettings{
		BuyerId:            buyerId,
		DatacenterId:       datacenterId,
		EnableAcceleration: true,
	}

	{
		settings := expected

		var response CreateBuyerDatacenterSettingsResponse

		err := Create("admin/create_buyer_datacenter_settings", settings, &response)

		if err != nil {
			panic(err)
		}
	}

	// read all settings
	{
		response := ReadBuyerDatacenterSettingsListResponse{}

		err := GetJSON("admin/buyer_datacenter_settings", &response)

		if err != nil {
			panic(err)
		}

		if len(response.Settings) != 1 {
			panic(fmt.Sprintf("expect one setting in response, got %d", len(response.Settings)))
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Settings[0] != expected {
			panic("buyer datacenter settings do not match expected")
		}
	}

	// read specific settings
	{
		response := ReadBuyerDatacenterSettingsResponse{}

		err := GetJSON(fmt.Sprintf("admin/buyer_datacenter_settings/%x/%x", buyerId, datacenterId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Settings != expected {
			panic("settings do not match expected")
		}
	}

	// update settings
	{
		expected.EnableAcceleration = false

		settings := expected

		response := UpdateBuyerDatacenterSettingsResponse{}

		err := Update("admin/update_buyer_datacenter_settings", settings, &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		if response.Settings != expected {
			panic("settings do not match expected")
		}
	}

	// delete settings
	{
		response := DeleteBuyerDatacenterSettingsResponse{}

		err := Delete(fmt.Sprintf("admin/delete_buyer_datacenter_settings/%x/%x", buyerId, datacenterId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

func test_relay_keypair() {

	fmt.Printf("\ntest_relay_keypair\n\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create relay keypair

	relayKeypair := admin.RelayKeypairData{}

	var response CreateRelayKeypairResponse

	err := Create("admin/create_relay_keypair", relayKeypair, &response)

	if err != nil {
		panic(err)
	}

	relayKeypairId := response.RelayKeypair.RelayKeypairId

	// read all relay keypairs
	{
		response := ReadRelayKeypairsResponse{}

		err := GetJSON("admin/relay_keypairs", &response)

		if err != nil {
			panic(err)
		}

		if len(response.RelayKeypairs) != 1 {
			panic(fmt.Sprintf("expect one relay keypair in response, got %d", len(response.RelayKeypairs)))
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		data, err := base64.StdEncoding.DecodeString(response.RelayKeypairs[0].PublicKeyBase64)
		if err != nil {
			panic(err)
		}

		data, err = base64.StdEncoding.DecodeString(response.RelayKeypairs[0].PrivateKeyBase64)
		if err != nil {
			panic(err)
		}

		_ = data
	}

	// read a specific relay keypair
	{
		response := ReadRelayKeypairResponse{}

		err := GetJSON(fmt.Sprintf("admin/relay_keypair/%x", relayKeypairId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		data, err := base64.StdEncoding.DecodeString(response.RelayKeypair.PublicKeyBase64)
		if err != nil {
			panic(err)
		}

		data, err = base64.StdEncoding.DecodeString(response.RelayKeypair.PrivateKeyBase64)
		if err != nil {
			panic(err)
		}

		_ = data
	}

	// update relay keypair
	{
		relayKeypair.RelayKeypairId = relayKeypairId
		relayKeypair.PublicKeyBase64 = "aaaaaaaaaaaaaaaaaa"
		relayKeypair.PrivateKeyBase64 = "bbbbbbbbbbbbbbbbbb"

		response := UpdateRelayKeypairResponse{}

		err := Update("admin/update_relay_keypair", relayKeypair, &response)

		if err != nil {
			panic(err)
		}

		if response.Error == "" {
			panic("expect error string to be valid. we do not support updating relay keypairs")
		}
	}

	// delete relay keypair
	{
		response := DeleteRelayKeypairResponse{}

		err := Delete(fmt.Sprintf("admin/delete_relay_keypair/%x", relayKeypairId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

func test_buyer_keypair() {

	fmt.Printf("\ntest_buyer_keypair\n\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create buyer keypair

	buyerKeypair := admin.BuyerKeypairData{}

	var response CreateBuyerKeypairResponse

	err := Create("admin/create_buyer_keypair", buyerKeypair, &response)

	if err != nil {
		panic(err)
	}

	buyerKeypairId := response.BuyerKeypair.BuyerKeypairId

	// read all buyer keypairs
	{
		response := ReadBuyerKeypairsResponse{}

		err := GetJSON("admin/buyer_keypairs", &response)

		if err != nil {
			panic(err)
		}

		if len(response.BuyerKeypairs) != 1 {
			panic(fmt.Sprintf("expect one buyer keypair in response, got %d", len(response.BuyerKeypairs)))
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		data, err := base64.StdEncoding.DecodeString(response.BuyerKeypairs[0].PublicKeyBase64)
		if err != nil {
			panic(err)
		}

		data, err = base64.StdEncoding.DecodeString(response.BuyerKeypairs[0].PrivateKeyBase64)
		if err != nil {
			panic(err)
		}

		_ = data
	}

	// read a specific buyer keypair
	{
		response := ReadBuyerKeypairResponse{}

		err := GetJSON(fmt.Sprintf("admin/buyer_keypair/%x", buyerKeypairId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}

		data, err := base64.StdEncoding.DecodeString(response.BuyerKeypair.PublicKeyBase64)
		if err != nil {
			panic(err)
		}

		data, err = base64.StdEncoding.DecodeString(response.BuyerKeypair.PrivateKeyBase64)
		if err != nil {
			panic(err)
		}

		_ = data
	}

	// update buyer keypair
	{
		buyerKeypair.BuyerKeypairId = buyerKeypairId
		buyerKeypair.PublicKeyBase64 = "aaaaaaaaaaaaaaaaaa"
		buyerKeypair.PrivateKeyBase64 = "bbbbbbbbbbbbbbbbbb"

		response := UpdateBuyerKeypairResponse{}

		err := Update("admin/update_buyer_keypair", buyerKeypair, &response)

		if err != nil {
			panic(err)
		}

		if response.Error == "" {
			panic("expect error string to be valid. we do not support updating buyer keypairs")
		}
	}

	// delete buyer keypair
	{
		response := DeleteBuyerKeypairResponse{}

		err := Delete(fmt.Sprintf("admin/delete_buyer_keypair/%x", buyerKeypairId), &response)

		if err != nil {
			panic(err)
		}

		if response.Error != "" {
			panic("expect error string to be empty")
		}
	}
}

// ----------------------------------------------------------------------------------------

type test_function func()

func main() {

	allTests := []test_function{
		test_customer,
		test_seller,
		test_datacenter,
		test_relay,
		test_route_shader,
		test_buyer,
		test_buyer_datacenter_settings,
		test_relay_keypair,
		test_buyer_keypair,
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

	for i := range tests {
		tests[i]()
	}

	fmt.Printf("\n")
}
