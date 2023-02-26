/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"reflect"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"

	"github.com/networknext/backend/modules/admin"
)

func bash(command string) {

	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=.")

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: failed to run command: %v\n", err)
		os.Exit(1)
	}

	cmd.Wait()
}

func clearDatabase() {
	bash("psql postgres -f ../schemas/sql/destroy.sql")
	bash("psql postgres -f ../schemas/sql/create.sql")
}

func api() (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command("./api")
	if cmd == nil {
		panic("could not create api!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

func test_customers() {

	fmt.Printf("test_customers\n")

	clearDatabase()

	api_cmd, _ := api()

	// create customer
	{
		customer := admin.CustomerData{CustomerName: "Test", CustomerCode: "test", Live: true, Debug: true}
		
		buffer := new(bytes.Buffer)
		
		json.NewEncoder(buffer).Encode(customer)
		
		request, _ := http.NewRequest("POST", "http://127.0.0.1:50000/admin/create_customer", buffer)
		
		client := &http.Client{}

		var err error		
		var response *http.Response
		for i := 0; i < 30; i++ {
			response, err = client.Do(request)
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}

		if err != nil {
		    panic(fmt.Sprintf("could not post create customer: %v\n", err))
		}

		if response.StatusCode != 200 {
	      panic(fmt.Sprintf("bad http response on create customer: %d", response.StatusCode))
		}

	   body, error := ioutil.ReadAll(response.Body)
	   if error != nil {
	      panic(fmt.Sprintf("could not read response: %v", err))
	   }

	   customerId, err := strconv.ParseUint(string(body), 10, 64)
	   if err != nil {
	   	panic(fmt.Sprintf("could not parse customer id in response: %v\n", err))
	   }

	   _ = customerId

	   response.Body.Close()
	}

	// read customers
	{
		var err error
		var response *http.Response
		for i := 0; i < 30; i++ {
			response, err = http.Get("http://127.0.0.1:50000/admin/customers")
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}

		if err != nil {
			panic(fmt.Sprintf("failed to get customers: %v", err))
		}

	   body, error := ioutil.ReadAll(response.Body)
	   if error != nil {
	      panic(fmt.Sprintf("could not read response: %v", err))
	   }

	   // todo: parse customers json

	   response.Body.Close()

		fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)
	}

	// update customer

	// (read customers again to make sure updated)

	// delete customer

	// (read customers again to make sure deleted)

	api_cmd.Process.Signal(os.Interrupt)

	api_cmd.Wait()
}

/*
func test_buyers() {

	fmt.Printf("test_buyers\n")

	api_cmd, _ := api()

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get("http://127.0.0.1:50000/admin/buyers")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to get buyers: %v", err))
	}

   body, error := ioutil.ReadAll(response.Body)
   if error != nil {
      panic(fmt.Sprintf("could not read response: %v", err))
   }

   response.Body.Close()

	fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()
}

func test_sellers() {

	fmt.Printf("test_sellers\n")

	api_cmd, _ := api()

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get("http://127.0.0.1:50000/admin/sellers")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to get sellers: %v", err))
	}

   body, error := ioutil.ReadAll(response.Body)
   if error != nil {
      panic(fmt.Sprintf("could not read response: %v", err))
   }

   response.Body.Close()

	fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()
}

func test_datacenters() {

	fmt.Printf("test_datacenters\n")

	api_cmd, _ := api()

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get("http://127.0.0.1:50000/admin/datacenters")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to get datacenters: %v", err))
	}

   body, error := ioutil.ReadAll(response.Body)
   if error != nil {
      panic(fmt.Sprintf("could not read response: %v", err))
   }

   response.Body.Close()

	fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()
}

func test_relays() {

	fmt.Printf("test_relays\n")

	api_cmd, _ := api()

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get("http://127.0.0.1:50000/admin/relays")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to get relays: %v", err))
	}

   body, error := ioutil.ReadAll(response.Body)
   if error != nil {
      panic(fmt.Sprintf("could not read response: %v", err))
   }

   response.Body.Close()

	fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()
}

func test_route_shaders() {

	fmt.Printf("test_route_shaders\n")

	api_cmd, _ := api()

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get("http://127.0.0.1:50000/admin/route_shaders")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to get route shaders: %v", err))
	}

   body, error := ioutil.ReadAll(response.Body)
   if error != nil {
      panic(fmt.Sprintf("could not read response: %v", err))
   }

   response.Body.Close()

	fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()
}

func test_buyer_datacenter_settings() {

	fmt.Printf("test_buyer_datacenter_settings\n")

	api_cmd, _ := api()

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get("http://127.0.0.1:50000/admin/buyer_datacenter_settings")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to get buyer datacenter settings: %v", err))
	}

   body, error := ioutil.ReadAll(response.Body)
   if error != nil {
      panic(fmt.Sprintf("could not read response: %v", err))
   }

   response.Body.Close()

	fmt.Printf("--------------------------------------------------\n%s--------------------------------------------------\n", body)

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()
}
*/

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
