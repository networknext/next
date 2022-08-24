package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"hash/fnv"
	"encoding/binary"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
)

var magicUpdateSeconds int

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

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

	upcomingMagic := hashCounter(counter+2)
	currentMagic := hashCounter(counter+1)
	previousMagic := hashCounter(counter+0)

	core.Debug("served magic values: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
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

	w.Write(upcomingMagic[:])
	w.Write(currentMagic[:])
	w.Write(previousMagic[:])
}

func main() {

	serviceName := "magic_backend";

	fmt.Printf("%s\n", serviceName)

	fmt.Printf("git hash: %s\n", sha)

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("error getting env: %v", err)
		os.Exit(1)
	}

	fmt.Printf("env: %s\n", env)

	magicUpdateSeconds, _ = envvar.GetInt("MAGIC_UPDATE_SECONDS", 60)

	fmt.Printf("magic update seconds: %d\n", magicUpdateSeconds)

	// Start HTTP server
	{
		port := envvar.Get("PORT", "41007")
		if port == "" {
			core.Error("PORT not set")
			os.Exit(1)
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/magic", magicHandler).Methods("GET")

		go func() {
			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				os.Exit(1)
			}
		}()
	}

	// Wait for shutdown signal

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		break
	}

	core.Debug("successfully shutdown")
}
