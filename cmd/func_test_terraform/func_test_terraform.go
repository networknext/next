/*
   Network Next. Copyright 2017 - 2026 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"
)

const APIPrivateKey = "uKUsmTySUVEssqBmVNciJWWolchcGGhFzRWMpydwOtVExvqYpHMotnkanNTaGHHh"

// ----------------------------------------------------------------------------------------

func bash(command string) {

	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: failed to run command: %v\n", err)
		os.Exit(1)
	}

	cmd.Wait()
}

func bash_quiet(command string) {

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
	bash_quiet("psql -U developer -h localhost postgres -f ../schemas/sql/destroy.sql")
	bash_quiet("psql -U developer -h localhost postgres -f ../schemas/sql/create.sql")
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
	cmd.Env = append(cmd.Env, fmt.Sprintf("API_PRIVATE_KEY=%s", APIPrivateKey))

	var output bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Start()

	return cmd, &output
}

// ----------------------------------------------------------------------------------------

var terraform_create = `

terraform {
  required_providers {
    networknext = {
      source = "networknext/networknext"
      version = "~> 5.0"
    }
  }
}

provider "networknext" {
  hostname = "http://localhost:50000"
  api_key  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTc0OTczODE4OX0.I89NXJCRMU_pIjnSleAnbux5HNsHhymzQ_SVatFo3b4"
}

# ---------------------------------------------------------

resource "networknext_seller" "test" {
  name = "Test"
  code = "test"
}

data "networknext_sellers" "test" {
  depends_on = [
    networknext_seller.test,
  ]
}

output "sellers" {
  value = data.networknext_sellers.test
}

# ---------------------------------------------------------

resource "networknext_datacenter" "test" {
  name = "test"
  seller_id = networknext_seller.test.id
  latitude = 100
  longitude = 50
}

data "networknext_datacenters" "test" {
  depends_on = [
    networknext_datacenter.test,
  ]
}

output "datacenters" {
  value = data.networknext_datacenters.test
}

# ---------------------------------------------------------

resource "networknext_relay_keypair" "test" {}

data "networknext_relay_keypairs" "test" {
  depends_on = [
    networknext_relay_keypair.test,
  ]
}

output "relay_keypairs" {
  value = data.networknext_relay_keypairs.test
}

# ---------------------------------------------------------

resource "networknext_relay" "test" {
  name = "test.relay"
  datacenter_id = networknext_datacenter.test.id
  public_ip = "127.0.0.1"
  public_key_base64=networknext_relay_keypair.test.public_key_base64
  private_key_base64=networknext_relay_keypair.test.private_key_base64
}

data "networknext_relays" "test" {
  depends_on = [
    networknext_relay.test,
  ]
}

output "relays" {
  value = data.networknext_relays.test
}

# ---------------------------------------------------------

resource "networknext_route_shader" test {
  name = "test"
}

data "networknext_route_shaders" "test" {
  depends_on = [
    networknext_route_shader.test,
  ]
}

output "route_shaders" {
  value = data.networknext_route_shaders.test
}

# ---------------------------------------------------------

resource "networknext_buyer" "test" {
  name = "Test Buyer"
  code = "test"
  route_shader_id = networknext_route_shader.test.id
  public_key_base64 = "C5Pyu+73R2XZvdxDZW4v8qfLs4K2vGb6q9lXgHnaxf/lInW1HWJMHQ=="
  live = true
  debug = true
}

data "networknext_buyers" "test" {
  depends_on = [
    networknext_buyer.test,
  ]
}

output "buyers" {
  value = data.networknext_buyers.test
}

# ---------------------------------------------------------

resource "networknext_buyer_datacenter_settings" "test" {
  buyer_id = networknext_buyer.test.id
  datacenter_id = networknext_datacenter.test.id
  enable_acceleration = true
}

data "networknext_buyer_datacenter_settings" "test" {
  depends_on = [
    networknext_buyer_datacenter_settings.test,
  ]
}

output "buyer_datacenter_settings" {
  value = data.networknext_buyer_datacenter_settings.test
}

# ---------------------------------------------------------
`

// ----------------------------------------------------------------------------------------

var terraform_update = `

terraform {
  required_providers {
    networknext = {
      source = "networknext/networknext"
      version = "~> 5.0"
    }
  }
}

provider "networknext" {
  hostname = "http://localhost:50000"
  api_key  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTc0OTczODE4OX0.I89NXJCRMU_pIjnSleAnbux5HNsHhymzQ_SVatFo3b4"
}

# ---------------------------------------------------------

resource "networknext_seller" "test" {
  name = "test (update)"
  code = "test"
}

data "networknext_sellers" "test" {
  depends_on = [
    networknext_seller.test,
  ]
}

output "sellers" {
  value = data.networknext_sellers.test
}

# ---------------------------------------------------------

resource "networknext_datacenter" "test" {
  name = "test (update)"
  seller_id = networknext_seller.test.id
  latitude = 100
  longitude = 50
}

data "networknext_datacenters" "test" {
  depends_on = [
    networknext_datacenter.test,
  ]
}

output "datacenters" {
  value = data.networknext_datacenters.test
}

# ---------------------------------------------------------

resource "networknext_relay_keypair" "test" {}

data "networknext_relay_keypairs" "test" {
  depends_on = [
    networknext_relay_keypair.test,
  ]
}

output "relay_keypairs" {
  value = data.networknext_relay_keypairs.test
}

# ---------------------------------------------------------

resource "networknext_relay" "test" {
  name = "test.relay.update"
  datacenter_id = networknext_datacenter.test.id
  public_ip = "127.0.0.1"
  public_key_base64=networknext_relay_keypair.test.public_key_base64
  private_key_base64=networknext_relay_keypair.test.private_key_base64
}

data "networknext_relays" "test" {
  depends_on = [
    networknext_relay.test,
  ]
}

output "relays" {
  value = data.networknext_relays.test
}

# ---------------------------------------------------------

resource "networknext_route_shader" test {
  name = "test (update)"
}

data "networknext_route_shaders" "test" {
  depends_on = [
    networknext_route_shader.test,
  ]
}

output "route_shaders" {
  value = data.networknext_route_shaders.test
}

# ---------------------------------------------------------

resource "networknext_buyer" "test" {
  name = "Test Buyer (update)"
  code = "test"
  route_shader_id = networknext_route_shader.test.id
  public_key_base64 = "C5Pyu+73R2XZvdxDZW4v8qfLs4K2vGb6q9lXgHnaxf/lInW1HWJMHQ=="
}

data "networknext_buyers" "test" {
  depends_on = [
    networknext_buyer.test,
  ]
}

output "buyers" {
  value = data.networknext_buyers.test
}

# ---------------------------------------------------------

resource "networknext_buyer_datacenter_settings" "test" {
  buyer_id = networknext_buyer.test.id
  datacenter_id = networknext_datacenter.test.id
  enable_acceleration = false
}

data "networknext_buyer_datacenter_settings" "test" {
  depends_on = [
    networknext_buyer_datacenter_settings.test,
  ]
}

output "buyer_datacenter_settings" {
  value = data.networknext_buyer_datacenter_settings.test
}

# ---------------------------------------------------------
`

// ----------------------------------------------------------------------------------------

func test_terraform() {

	fmt.Printf("\ntest_terraform\n\n")

	clearDatabase()

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

	// create a temporary directory to work in

	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		panic("could not create temp dir")
	}

	// clean up the temp dir when we are finished

	defer func() {
		os.RemoveAll(tmp)
	}()

	// write the main.tf file to the temp dir

	os.WriteFile(fmt.Sprintf("%s/main.tf", tmp), []byte(terraform_create), 0666)

	// run terraform init in the temp dir

	bash(fmt.Sprintf("cd %s && terraform init", tmp))

	// apply terraform

	bash(fmt.Sprintf("cd %s && terraform apply -auto-approve", tmp))

	// apply terraform a second time, there should be no changes

	bash(fmt.Sprintf("cd %s && terraform apply -auto-approve", tmp))

	// apply an update to all properties that accept them

	os.WriteFile(fmt.Sprintf("%s/main.tf", tmp), []byte(terraform_update), 0666)

	bash(fmt.Sprintf("cd %s && terraform apply -auto-approve", tmp))

	// destroy everything

	bash(fmt.Sprintf("cd %s && terraform destroy -auto-approve", tmp))
}

// ----------------------------------------------------------------------------------------

type test_function func()

func main() {

	allTests := []test_function{
		test_terraform,
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
