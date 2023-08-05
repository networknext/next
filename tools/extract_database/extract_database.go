package main

import (
	"fmt"
	"os"

	db "github.com/networknext/next/modules/database"
)

func main() {

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

	database.Save("database.bin")

	loaded, err := db.LoadDatabase("database.bin")
	if err != nil {
		fmt.Printf("error: could not load database.bin: %v\n", err)
		os.Exit(1)
	}

	err = loaded.Validate()
	if err != nil {
		fmt.Printf("error: loaded database did not validate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(loaded.String())
}
