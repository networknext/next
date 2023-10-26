package main

import (
	"fmt"
	"os"
	"bytes"
	"encoding/gob"

	"github.com/networknext/next/modules/common"
)

func main() {

	fmt.Printf("hello\n")

	data, err := os.ReadFile("data/relay_manager.bin")
	if err != nil {
		panic(err)
	}

	buffer := bytes.NewReader(data)

	relay_manager := common.RelayManager{}

	err = gob.NewDecoder(buffer).Decode(&relay_manager)
	if err != nil {
		panic(err)
	}

	fmt.Printf("loaded relay manager\n")
}
