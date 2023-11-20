package main

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net/http"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

var magicUpdateSeconds int

func main() {

	service := common.CreateService("magic_backend")

	magicUpdateSeconds = envvar.GetInt("MAGIC_UPDATE_SECONDS", 60)

	core.Debug("magic update seconds: %d", magicUpdateSeconds)

	service.Router.HandleFunc("/magic", magicHandler).Methods("GET")

	service.StartWebServer()

	service.WaitForShutdown()
}

func hashCounter(counter int64) []byte {
	hash := fnv.New64a()
	var inputValue [8]byte
	binary.LittleEndian.PutUint64(inputValue[:], uint64(counter))
	hash.Write(inputValue[:])
	hash.Write([]byte("don't worry. be happy. :)"))
	hash.Write([]byte(fmt.Sprintf("%d", counter)))
	hash.Write([]byte(fmt.Sprintf("%016x", counter)))
	hashValue := hash.Sum64()
	var result [8]byte
	binary.LittleEndian.PutUint64(result[:], uint64(hashValue))
	return result[:]
}

func magicHandler(w http.ResponseWriter, r *http.Request) {

	timestamp := time.Now().Unix()

	counter := timestamp / int64(magicUpdateSeconds)

	var counterData [8]byte
	binary.LittleEndian.PutUint64(counterData[:], uint64(counter))

	upcomingMagic := hashCounter(counter + 2)
	currentMagic := hashCounter(counter + 1)
	previousMagic := hashCounter(counter + 0)

	core.Debug("served magic values: %x -> %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
		counter,
		upcomingMagic[0],
		upcomingMagic[1],
		upcomingMagic[2],
		upcomingMagic[3],
		upcomingMagic[4],
		upcomingMagic[5],
		upcomingMagic[6],
		upcomingMagic[7],
		currentMagic[0],
		currentMagic[1],
		currentMagic[2],
		currentMagic[3],
		currentMagic[4],
		currentMagic[5],
		currentMagic[6],
		currentMagic[7],
		previousMagic[0],
		previousMagic[1],
		previousMagic[2],
		previousMagic[3],
		previousMagic[4],
		previousMagic[5],
		previousMagic[6],
		previousMagic[7])

	w.Header().Set("Content-Type", "application/octet-stream")

	w.Write(counterData[:])
	w.Write(upcomingMagic[:])
	w.Write(currentMagic[:])
	w.Write(previousMagic[:])
}
