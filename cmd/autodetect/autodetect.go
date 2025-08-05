package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

var mutex sync.Mutex

var cache map[string]string

var patterns []string

var re *regexp.Regexp

func main() {

	service := common.CreateService("autodetect")

	service.Router.HandleFunc("/{input_datacenter}/{server_address}", autodetectHandler).Methods("GET")

	cache = make(map[string]string)

	patternString := envvar.GetString("AUTODETECT_PATTERNS", "maxihost,latitude|latitude,latitude|i3d,i3d|gcore,gcore|g-core,gcore")
	patterns = strings.Split(patternString, "|")

	re = regexp.MustCompile(`^unity\.([a-z]+)[\.]?(.*)?$`)

	service.StartWebServer()

	service.WaitForShutdown()
}

func autodetectHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	inputDatacenter := vars["input_datacenter"]

	serverAddress := vars["server_address"]

	// extract location from the input datacenter, eg. "unity.saopaulo.1" -> "saopaulo"

	matches := re.FindStringSubmatch(inputDatacenter)
	if len(matches) < 2 {
		core.Error("invalid input datacenter: '%s'", inputDatacenter)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	location := matches[1]

	// see if we have the result in cache already

	key := "unity." + location + ":" + serverAddress

	mutex.Lock()
	value, found := cache[key]
	mutex.Unlock()

	if found {
		core.Log("%s -> %s (cached)", key, value)
		w.Write([]byte(value))
		return
	}

	// not in cache, run whois autodetect logic then add result to cache

	cmd := exec.Command("whois -I", serverAddress)
	output, err := cmd.Output()
	if err != nil {
		core.Error("error running whois command: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	outputString := strings.ToLower(string(output))

	seller := ""
	for i := range patterns {
		values := strings.Split(patterns[i], ",")
		if len(values) == 2 {
			if strings.Contains(outputString, values[0]) {
				seller = values[1]
			}
		}
	}

	if seller == "" {
		core.Error("could not find any seller for: '%s':", key)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value = fmt.Sprintf("%s.%s", seller, location)

	mutex.Lock()
	cache[key] = value
	mutex.Unlock()

	core.Log("%s -> %s", key, value)

	w.Write([]byte(value))
}
