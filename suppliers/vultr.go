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

// This definition drives the set of vultr datacenters, eg. "vultr.[country/city]"

var datacenterMap = map[string]*Datacenter{

	"ams": {"amsterdam", 2.3676, 4.9041},
	"atl": {"atlanta", 33.7488, -84.3877},
	"blr": {"bangalore", 12.9716, 77.5946},
	"bom": {"mumbai", 19.0760, 72.8777},
	"cdg": {"paris", 48.8566, 2.3522},
	"del": {"delhi", 28.7041, 77.1025},
	"dfw": {"dallas", 32.7767, -96.7970},
	"ewr": {"newyork", 40.7128, -74.0060},
	"fra": {"frankfurt", 50.1109, 8.6821},
	"hnl": {"honolulu", 21.3099, -157.8581},
	"icn": {"seoul", 37.5665, 126.9780},
	"itm": {"osaka", 34.6937, 135.5023},
	"jnb": {"johannesburg", -26.2041, 28.0473},
	"lax": {"losangeles", 34.0522, -118.2437},
	"lhr": {"london", 51.5072, -0.1276},
	"mad": {"madrid", 40.4168, -3.7038},
	"mel": {"melbourne", -37.8136, 144.9631},
	"mex": {"mexico", 19.4326, -99.1332},
	"mia": {"miami", 25.7617, -80.1918},
	"nrt": {"tokyo", 35.6762, 139.6503},
	"ord": {"chicago", 41.8781, -87.6298},
	"sao": {"saopaulo", -23.5558, -46.6396},
	"scl": {"santiago", -33.4489, -70.6693},
	"sea": {"seattle", 47.6062, -122.3321},
	"sgp": {"singapore", 1.3521, 103.8198},
	"sjc": {"siliconvalley", 37.3387, -121.8853},
	"sto": {"stockholm", 59.3293, 18.0686},
	"syd": {"sydney", -33.8688, 151.2093},
	"waw": {"warsaw", 52.2297, 21.0122},
	"yto": {"toronto", 43.6532, -79.3832},
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
		file, err := os.Open("cache/vultr_zones.bin")
		if err == nil {
			gob.NewDecoder(file).Decode(&zones)
			if err == nil {
				loadedZonesCache = true
			}
			file.Close()
		}
	}

	// otherwise, get vultr zones and save to cache

	if !loadedZonesCache {

		output := bash("curl -s https://api.vultr.com/v2/regions -X GET")

		type RegionData struct {
			Id   string `json:"id"`
			City string `json:"city"`
		}

		type Response struct {
			Regions []RegionData `json:"regions"`
		}

		response := Response{}

		if err := json.Unmarshal([]byte(output), &response); err != nil {
			panic(err)
		}

		for i := range response.Regions {
			zones = append(zones, &Zone{Zone: response.Regions[i].Id, Label: response.Regions[i].City})
		}

		{
			file, err := os.Create("cache/vultr_zones.bin")
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
			zones[i].DatacenterName = fmt.Sprintf("vultr.%s", datacenter.name)
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

	// generate vultr.txt

	fmt.Printf("\nGenerating vultr.txt\n")

	file, err := os.Create("config/vultr.txt")
	if err != nil {
		panic(err)
	}

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, "%s,%s\n", zones[i].Zone, zones[i].DatacenterName)
		}
	}

	file.Close()

	// generate vultr.sql

	fmt.Printf("\nGenerating vultr.sql\n")

	file, err = os.Create("schemas/sql/sellers/vultr.sql")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(file, "\n-- vultr datacenters\n")

	format_string := "\nINSERT INTO datacenters(\n" +
		"	datacenter_name,\n" +
		"	native_name,\n" +
		"	latitude,\n" +
		"	longitude,\n" +
		"	seller_id)\n" +
		"VALUES(\n" +
		"   '%s',\n" +
		"   '%s',\n" +
		"   %f,\n" +
		"   %f,\n" +
		"   (select seller_id from sellers where seller_name = 'vultr')\n" +
		");\n"

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].Zone, zones[i].Latitude, zones[i].Longitude)
		}
	}

	file.Close()

	// generate vultr/generated.tf

	file, err = os.Create("terraform/suppliers/vultr/generated.tf")
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nGenerating vultr/generated.tf\n")

	fmt.Fprintf(file, "\nlocals {\n\n  datacenter_map = {\n\n")

	format_string = "    \"%s\" = {\n" +
		"      zone        = \"%s\"\n" +
		"      native_name = \"%s\"\n" +
		"      latitude    = %.2f\n" +
		"      longitude   = %.2f\n" +
		"      seller_name = \"VULTR\"\n" +
		"      seller_code = \"vultr\"\n" +
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
