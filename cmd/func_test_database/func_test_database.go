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

// ----------------------------------------------------------------------------------------

type test_function func()

func main() {

	allTests := []test_function{
		test_local,
		test_dev,
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
