/*
   Network Next. Copyright 2017 - 2025 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"

	"github.com/networknext/next/modules/admin"
	"github.com/networknext/next/modules/common"
	db "github.com/networknext/next/modules/database"
)

const Hostname = "http://127.0.0.1:50000"
const TestAPIKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTc0OTczODE4OX0.I89NXJCRMU_pIjnSleAnbux5HNsHhymzQ_SVatFo3b4"
const TestAPIPrivateKey = "uKUsmTySUVEssqBmVNciJWWolchcGGhFzRWMpydwOtVExvqYpHMotnkanNTaGHHh"
const TestBuyerPublicKey = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ=="

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
	cmd.Env = append(cmd.Env, fmt.Sprintf("API_PRIVATE_KEY=%s", TestAPIPrivateKey))

	var output bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Start()

	return cmd, &output
}

// ----------------------------------------------------------------------------------------

func GetText(path string) (string, error) {

	url := Hostname + "/" + path

	var err error
	var response *http.Response
	for i := 0; i < 5; i++ {
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
		return "", fmt.Errorf("failed to read %s: %v", url, err)
	}

	if response == nil {
		return "", fmt.Errorf("no response from %s", url)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("got %d response for %s", response.StatusCode, url)
	}

	body, error := io.ReadAll(response.Body)
	if error != nil {
		return "", fmt.Errorf("could not read response body for %s: %v", url, err)
	}

	response.Body.Close()

	return string(body), nil
}

func GetJSON(path string, object interface{}) error {

	url := Hostname + "/" + path

	var err error
	var response *http.Response
	for i := 0; i < 5; i++ {
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
		return fmt.Errorf("failed to read %s: %v", url, err)
	}

	if response == nil {
		return fmt.Errorf("no response from %s", url)
	}

	body, error := io.ReadAll(response.Body)
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

	url := Hostname + "/" + path

	buffer := new(bytes.Buffer)

	json.NewEncoder(buffer).Encode(requestData)

	request, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return fmt.Errorf("could not create HTTP POST request for %s: %v", url, err)
	}

	request.Header.Set("Authorization", "Bearer "+TestAPIKey)

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

	body, error := io.ReadAll(response.Body)
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

	url := Hostname + "/" + path

	buffer := new(bytes.Buffer)

	json.NewEncoder(buffer).Encode(requestData)

	request, _ := http.NewRequest("PUT", url, buffer)

	request.Header.Set("Authorization", "Bearer "+TestAPIKey)

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

	body, error := io.ReadAll(response.Body)
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

	url := Hostname + "/" + path

	request, _ := http.NewRequest("DELETE", url, nil)

	request.Header.Set("Authorization", "Bearer "+TestAPIKey)

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

	body, error := io.ReadAll(response.Body)
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

func test_seller() {

	fmt.Printf("\ntest_seller\n\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create seller

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Test", SellerCode: "test"}

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

		if response.Sellers[0].SellerCode != "test" {
			panic("wrong seller code")
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
			panic(fmt.Sprintf("expect error string to be empty: %s", response.Error))
		}

		if response.Seller.SellerId != sellerId {
			panic(fmt.Sprintf("wrong seller id: got %x, expected %x", response.Seller.SellerId, sellerId))
		}

		if response.Seller.SellerName != "Test" {
			panic("wrong seller name")
		}

		if response.Seller.SellerCode != "test" {
			panic("wrong seller code")
		}
	}

	// update seller
	{
		seller := admin.SellerData{SellerId: sellerId, SellerName: "Updated"}

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

	// create datacenter

	datacenterId := uint64(0)
	{
		datacenter := admin.DatacenterData{DatacenterName: "test", Latitude: 100, Longitude: 200, SellerId: sellerId}

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

		if response.Datacenters[0].DatacenterName != "test" {
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

		if response.Datacenter.DatacenterName != "test" {
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

	// create seller

	sellerId := uint64(0)
	{
		seller := admin.SellerData{SellerName: "Test", SellerCode: "test"}

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
		datacenter := admin.DatacenterData{DatacenterName: "test", Latitude: 100, Longitude: 200, SellerId: sellerId}

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
		relay := admin.RelayData{RelayName: "test", DatacenterId: datacenterId, PublicIP: "127.0.0.1", InternalIP: "0.0.0.0", SSH_IP: "127.0.0.1", PublicKeyBase64: "+ONHHci1bizkWzi4MTt1E5b0p0M5Xe0PhUay5H5KIl4=", PrivateKeyBase64: "+ONHHci1bizkWzi4MTt1E5b0p0M5Xe0PhUay5H5KIl4="}

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

		if response.Relays[0].RelayName != "test" {
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

		if response.Relay.RelayName != "test" {
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
		relay := admin.RelayData{RelayId: relayId, RelayName: "Updated", DatacenterId: datacenterId, PublicIP: "127.0.0.1", InternalIP: "0.0.0.0", SSH_IP: "127.0.0.1", PublicKeyBase64: "+ONHHci1bizkWzi4MTt1E5b0p0M5Xe0PhUay5H5KIl4=", PrivateKeyBase64: "+ONHHci1bizkWzi4MTt1E5b0p0M5Xe0PhUay5H5KIl4="}

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
		BandwidthEnvelopeUpKbps:       1024,
		BandwidthEnvelopeDownKbps:     512,
		DisableNetworkNext:            true,
		LatencyReductionThreshold:     10.0,
		SelectionPercent:              100,
		MaxLatencyTradeOff:            20,
		RouteSwitchThreshold:          10,
		RouteSelectThreshold:          5,
		RTTVeto:                       10,
		ForceNext:                     true,
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

	// create route shader

	routeShaderId := uint64(0)
	{
		routeShader := admin.RouteShaderData{
			RouteShaderName:               "Test",
			ABTest:                        true,
			AcceptableLatency:             10.0,
			AcceptablePacketLossInstant:   0.25,
			AcceptablePacketLossSustained: 0.01,
			BandwidthEnvelopeUpKbps:       1024,
			BandwidthEnvelopeDownKbps:     512,
			DisableNetworkNext:            true,
			LatencyReductionThreshold:     10.0,
			SelectionPercent:              100,
			MaxLatencyTradeOff:            20,
			RouteSwitchThreshold:          10,
			RouteSelectThreshold:          5,
			RTTVeto:                       10,
			ForceNext:                     true,
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
		BuyerCode:       "test",
		RouteShaderId:   routeShaderId,
		PublicKeyBase64: TestBuyerPublicKey,
		Live:            true,
		Debug:           true,
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
		expected.BuyerCode = "updated"

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

	// create route shader

	routeShaderId := uint64(0)
	{
		routeShader := admin.RouteShaderData{
			RouteShaderName:               "Test",
			ABTest:                        true,
			AcceptableLatency:             10.0,
			AcceptablePacketLossInstant:   0.25,
			AcceptablePacketLossSustained: 0.01,
			BandwidthEnvelopeUpKbps:       1024,
			BandwidthEnvelopeDownKbps:     512,
			DisableNetworkNext:            true,
			LatencyReductionThreshold:     10.0,
			SelectionPercent:              100,
			MaxLatencyTradeOff:            20,
			RouteSwitchThreshold:          10,
			RouteSelectThreshold:          5,
			RTTVeto:                       10,
			ForceNext:                     true,
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
			BuyerCode:       "test",
			RouteShaderId:   routeShaderId,
			PublicKeyBase64: TestBuyerPublicKey,
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
		seller := admin.SellerData{SellerName: "Test", SellerCode: "test"}

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
		datacenter := admin.DatacenterData{DatacenterName: "test", Latitude: 100, Longitude: 200, SellerId: sellerId}

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

type AdminDatabaseResponse struct {
	Database string `json:"database_base64"`
	Error    string `json:"error"`
}

func test_database() {

	fmt.Printf("\ntest_database\n\n")

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
			RouteShaderName:               "Test",
			ABTest:                        true,
			AcceptableLatency:             10.0,
			AcceptablePacketLossInstant:   0.25,
			AcceptablePacketLossSustained: 0.01,
			BandwidthEnvelopeUpKbps:       1024,
			BandwidthEnvelopeDownKbps:     512,
			DisableNetworkNext:            true,
			LatencyReductionThreshold:     10.0,
			SelectionPercent:              100,
			MaxLatencyTradeOff:            20,
			RouteSwitchThreshold:          10,
			RouteSelectThreshold:          5,
			RTTVeto:                       10,
			ForceNext:                     true,
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
			BuyerCode:       "test",
			RouteShaderId:   routeShaderId,
			PublicKeyBase64: TestBuyerPublicKey,
			Live:            true,
			Debug:           true,
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
		seller := admin.SellerData{SellerName: "Test", SellerCode: "test"}

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
		datacenter := admin.DatacenterData{DatacenterName: "test", Latitude: 100, Longitude: 200, SellerId: sellerId}

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

	// get database and verify it

	var response AdminDatabaseResponse

	err := GetJSON("admin/database", &response)

	if err != nil {
		panic(err)
	}

	database_binary, err := base64.StdEncoding.DecodeString(response.Database)
	if err != nil {
		panic(err)
	}

	tempFile := fmt.Sprintf("/tmp/database-%s.bin", common.RandomString(64))
	err = os.WriteFile(tempFile, database_binary, 0666)
	if err != nil {
		panic(err)
	}
	database, err := db.LoadDatabase(tempFile)
	if err != nil {
		panic(err)
	}
	err = database.Validate()
	if err != nil {
		panic(err)
	}
}

// ----------------------------------------------------------------------------------------

type test_function func()

func main() {

	allTests := []test_function{
		test_seller,
		test_datacenter,
		test_relay,
		test_route_shader,
		test_buyer,
		test_buyer_datacenter_settings,
		test_relay_keypair,
		test_database,
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
