package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// ===========================================================================================================================================

// This definition drives the set of google datacenters, eg. "google.[country/city].[number]"

var datacenterMap = map[string]*Datacenter{

	"us-east1":                {"southcarolina", 33.8361, -81.1637},
	"asia-east1":              {"taiwan", 25.105497, 121.597366},
	"asia-east2":              {"hongkong", 22.3193, 114.1694},
	"asia-northeast1":         {"tokyo", 35.6762, 139.6503},
	"asia-northeast2":         {"osaka", 34.6937, 135.5023},
	"asia-northeast3":         {"seoul", 37.5665, 126.9780},
	"asia-south1":             {"mumbai", 19.0760, 72.8777},
	"asia-south2":             {"delhi", 28.7041, 77.1025},
	"asia-southeast1":         {"singapore", 1.3521, 103.8198},
	"asia-southeast2":         {"jakarta", 6.2088, 106.8456},
	"australia-southeast1":    {"sydney", -33.8688, 151.2093},
	"australia-southeast2":    {"melbourne", -37.8136, 144.9631},
	"europe-central2":         {"warsaw", 52.2297, 21.0122},
	"europe-north1":           {"finland", 60.5693, 27.1878},
	"europe-southwest1":       {"madrid", 40.4168, 3.7038},
	"europe-west1":            {"belgium", 50.4706, 3.8170},
	"europe-west2":            {"london", 51.5072, -0.1276},
	"europe-west3":            {"frankfurt", 50.1109, 8.6821},
	"europe-west4":            {"netherlands", 53.4386, 6.8355},
	"europe-west6":            {"zurich", 47.3769, 8.5417},
	"europe-west8":            {"milan", 45.4642, 9.1900},
	"europe-west9":            {"paris", 48.8566, 2.3522},
	"me-west1":                {"telaviv", 32.0853, 34.7818},
	"northamerica-northeast1": {"montreal", 45.5019, -73.5674},
	"northamerica-northeast2": {"toronto", 43.6532, -79.3832},
	"southamerica-east1":      {"saopaulo", -23.5558, -46.6396},
	"southamerica-west1":      {"santiago", -33.4489, -70.6693},
	"us-central1":             {"iowa", 41.8780, -93.0977},
	"us-east4":                {"virginia", 37.4316, -78.6569},
	"us-east5":                {"ohio", 39.9612, -82.9988},
	"us-south1":               {"dallas", 32.7767, -96.7970},
	"us-west1":                {"oregon", 45.5946, -121.1787},
	"us-west2":                {"losangeles", 34.0522, -118.2437},
	"us-west3":                {"saltlakecity", 40.7608, -111.8910},
	"us-west4":                {"lasvegas", 36.1716, -115.1391},
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
	Region         string
	DatacenterName string
	Latitude       float32
	Longitude      float32
}

func main() {

	// create cache dir if needed

	bash("mkdir -p cache")

	// load regions cache, if possible

	regions := make([]string, 0)

	loadedRegionsCache := false

	{
		file, err := os.Open("cache/google_regions.bin")
		if err == nil {
			gob.NewDecoder(file).Decode(&regions)
			if err == nil {
				loadedRegionsCache = true
			}
			file.Close()
		}
	}

	// otherwise, get google regions and save to cache

	if !loadedRegionsCache {

		output := bash("gcloud compute regions list")

		lines := strings.Split(output, "\n")

		for i := 1; i < len(lines); i++ {
			re := regexp.MustCompile(`^([a-zA-Z0-9-]+)\w+`)
			match := re.FindStringSubmatch(lines[i])
			if len(match) > 0 {
				regions = append(regions, match[0])
			}
		}

		{
			file, err := os.Create("cache/google_regions.bin")
			if err != nil {
				panic(err)
			}
			gob.NewEncoder(file).Encode(&regions)
			file.Close()
		}

		fmt.Printf("\nRegions:\n\n")
		for i := range regions {
			fmt.Printf("  %s\n", regions[i])
		}
	}

	// load zones cache, if possible

	zones := make([]*Zone, 0)

	loadedZonesCache := false

	{
		file, err := os.Open("cache/google_zones.bin")
		if err == nil {
			gob.NewDecoder(file).Decode(&zones)
			if err == nil {
				loadedZonesCache = true
			}
			file.Close()
		}
	}

	// otherwise, get google zones and save to cache

	if !loadedZonesCache {

		output := bash("gcloud compute zones list")

		lines := strings.Split(output, "\n")

		for i := 1; i < len(lines); i++ {
			re := regexp.MustCompile(`^([a-zA-Z0-9-]+)\w+`)
			match := re.FindStringSubmatch(lines[i])
			if len(match) >= 1 {
				zones = append(zones, &Zone{match[0], "", "", 0, 0})
			}
		}

		{
			file, err := os.Create("cache/google_zones.bin")
			if err != nil {
				panic(err)
			}
			gob.NewEncoder(file).Encode(&zones)
			file.Close()
		}

		fmt.Printf("\nZones:\n\n")
		for i := range zones {
			fmt.Printf("  %s\n", zones[i].Zone)
		}
	}

	// unique the zones (not sure why I need to do this...) then sort by zone name

	zoneMap := make(map[string]*Zone)

	for i := range zones {
		zoneMap[zones[i].Zone] = zones[i]
	}

	index := 0
	zones = make([]*Zone, len(zoneMap))
	for _, v := range zoneMap {
		zones[index] = v
		index++
	}

	sort.SliceStable(zones, func(i, j int) bool { return zones[i].Zone < zones[j].Zone })

	// print out the known datacenters

	fmt.Printf("\nKnown datacenters:\n\n")

	unknown := make([]*Zone, 0)

	datacenterToRegion := make(map[string]string)

	for i := range zones {
		values := strings.Split(zones[i].Zone, "-")
		a := values[0]
		b := values[1]
		c := values[2]
		region := fmt.Sprintf("%s-%s", a, b)
		datacenter := datacenterMap[region]
		number := c[0] - 'a' + 1
		if datacenter != nil {
			zones[i].Region = region
			zones[i].DatacenterName = fmt.Sprintf("google.%s.%d", datacenter.name, number)
			zones[i].Latitude = datacenter.latitude
			zones[i].Longitude = datacenter.longitude
			fmt.Printf("  %s\n", zones[i].DatacenterName)
			datacenterToRegion[zones[i].DatacenterName] = region
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

	// generate google.txt

	fmt.Printf("\nGenerating google.txt\n")

	file, err := os.Create("config/google.txt")
	if err != nil {
		panic(err)
	}

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, "%s,%s\n", zones[i].Zone, zones[i].DatacenterName)
		}
	}

	file.Close()

	// generate google.sql

	fmt.Printf("\nGenerating google.sql\n")

	file, err = os.Create("schemas/sql/sellers/google.sql")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(file, "\n-- google datacenters\n")

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
		"   (select seller_id from sellers where seller_name = 'google')\n" +
		");\n"

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].Zone, zones[i].Latitude, zones[i].Longitude)
		}
	}

	file.Close()

	// generate google/generated.tf

	file, err = os.Create("terraform/suppliers/google/generated.tf")
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nGenerating google/generated.tf\n")

	fmt.Fprintf(file, "\nlocals {\n\n  datacenter_map = {\n\n")

	format_string = "    \"%s\" = {\n" +
		"      zone   = \"%s\"\n" +
		"      region = \"%s\"\n" +
		"    }\n" +
		"\n"

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].Zone, zones[i].Region)
		}
	}

	fmt.Fprintf(file, "  }\n\n")

	fmt.Fprintf(file, "}\n")

	file.Close()
}
