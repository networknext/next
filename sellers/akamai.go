package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// ===========================================================================================================================================

// This definition drives the set of akamai datacenters, eg. "akamai.[country/city]"

var datacenterMap = map[string]*Datacenter{

	"ap-west":      {"mumbai.1", 19.0760, 72.8777},
	"in-bom-2":     {"mumbai.2", 19.0760, 72.8777},
	"ca-central":   {"toronto", 43.6532, -79.3832},
	"ap-southeast": {"sydney", -33.8688, 151.2093},
	"us-central":   {"dallas", 32.7767, -96.7970},
	"us-west":      {"fremont", 37.5485, -121.9886},
	"us-southeast": {"atlanta", 33.7488, -84.3877},
	"us-east":      {"newyork", 40.7128, -74.0060},
	"eu-west":      {"london", 51.5072, -0.1276},
	"ap-south":     {"singapore.1", 1.3521, 103.8198},
	"sg-sin-2":     {"singapore.2", 1.3521, 103.8198},
	"eu-central":   {"frankfurt", 50.1109, 8.6821},
	"ap-northeast": {"tokyo", 35.6762, 139.6503},
	"jp-osa":       {"osaka", 34.6937, 135.5023},
	"us-iad":       {"washington", 47.751076, -120.740135},
	"us-ord":       {"chicago", 41.8781, -87.6298},
	"us-sea":       {"seattle", 47.6061, -122.3328},
	"us-mia":       {"miami", 25.7617, -80.1918},
	"us-lax":       {"losangeles", 34.0549, -118.2426},
	"nl-ams":       {"amsterdam", 52.3676, 4.9041},
	"fr-par":       {"paris", 48.8575, 2.3514},
	"br-gru":       {"saopaulo", -23.5558, -46.6396},
	"se-sto":       {"stockholm", 59.3327, 18.0656}
	"es-mad":       {"madrid", 40.416775, -3.703790},
	"it-mil":       {"milan", 45.4685, 9.1824},
	"gb-lon":       {"london", 51.5072, -0.1276},
	"au-mel":       {"melbourne", -37.8136, 144.9631},
	"id-cgk":       {"jakarta", -6.1944, 106.8229},
	"in-maa":       {"chennai", 13.0843, 80.2705},
}

type Datacenter struct {
	name      string
	latitude  float32
	longitude float32
}

// ===========================================================================================================================================

func bash(command string) string {
	var output bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	return output.String()
}

type Zone struct {
	Zone           string
	Label          string
	DatacenterName string
	Latitude       float32
	Longitude      float32
}

func main() {

	// create cache dir if needed

	bash("mkdir -p cache")

	// load zones cache, if possible

	zones := make([]*Zone, 0)

	loadedZonesCache := false

	{
		file, err := os.Open("cache/akamai_zones.bin")
		if err == nil {
			gob.NewDecoder(file).Decode(&zones)
			if err == nil {
				loadedZonesCache = true
			}
			file.Close()
		}
	}

	// otherwise, get akamai zones and save to cache

	if !loadedZonesCache {

		output := bash("curl -s https://api.linode.com/v4/regions")

		type ResponseData struct {
			Id    string `json:"id"`
			Label string `json:"label"`
		}

		type Response struct {
			Data []ResponseData `json:"data"`
		}

		response := Response{}

		if err := json.Unmarshal([]byte(output), &response); err != nil {
			panic(err)
		}

		for i := range response.Data {
			zones = append(zones, &Zone{Zone: response.Data[i].Id, Label: response.Data[i].Label})
		}

		{
			file, err := os.Create("cache/akamai_zones.bin")
			if err != nil {
				panic(err)
			}
			gob.NewEncoder(file).Encode(&zones)
			file.Close()
		}

		fmt.Printf("\nZones:\n\n")
		for i := range zones {
			fmt.Printf("  %s -> %s\n", zones[i].Zone, zones[i].Label)
		}
	}

	// print out the known datacenters

	fmt.Printf("\nKnown datacenters:\n\n")

	unknown := make([]*Zone, 0)

	for i := range zones {
		datacenter := datacenterMap[zones[i].Zone]
		if datacenter != nil {
			zones[i].DatacenterName = fmt.Sprintf("akamai.%s", datacenter.name)
			zones[i].Latitude = datacenter.latitude
			zones[i].Longitude = datacenter.longitude
			fmt.Printf("  %s\n", zones[i].DatacenterName)
		} else {
			unknown = append(unknown, zones[i])
		}
	}

	// print out the unknown datacenters

	if len(unknown) > 0 {
		fmt.Printf("\nUnknown datacenters:\n\n")
		for i := range unknown {
			fmt.Printf("  %s\n", unknown[i].Zone)
		}
	}

	// generate akamai.txt

	fmt.Printf("\nGenerating akamai.txt\n")

	file, err := os.Create("config/akamai.txt")
	if err != nil {
		panic(err)
	}

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, "%s,%s\n", zones[i].Zone, zones[i].DatacenterName)
		}
	}

	file.Close()

	// generate akamai/generated.tf

	file, err = os.Create("terraform/sellers/akamai/generated.tf")
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nGenerating akamai/generated.tf\n")

	fmt.Fprintf(file, "\nlocals {\n\n  datacenter_map = {\n\n")

	format_string := "    \"%s\" = {\n" +
		"      zone        = \"%s\"\n" +
		"      native_name = \"%s\"\n" +
		"      latitude    = %.2f\n" +
		"      longitude   = %.2f\n" +
		"      seller_name = \"Akamai\"\n" +
		"      seller_code = \"akamai\"\n" +
		"    }\n" +
		"\n"

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].Zone, zones[i].Zone, zones[i].Latitude, zones[i].Longitude)
		}
	}

	fmt.Fprintf(file, "  }\n\n")

	fmt.Fprintf(file, "}\n")

	file.Close()

	fmt.Printf("\n")
}
