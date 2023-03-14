package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------------

/*

	IMPORTANT: This is the definition of the set of datacenters for Amazon!!!

	All data for amazon relays are all generated from here:

		1. config/amazon.txt
		2. terraform/amazon.tfvars
		3. schemas/sql/sellers/amazon.sql

	Do not edit the generated data directly, edit the map below!
*/

var datacenterMap = map[string]*Datacenter{

	"afs1": {"johannesburg", -33.9249, 18.4241},
	"ape1": {"hongkong", 22.3193, 114.1694},
	"apne1": {"tokyo", 35.6762, 139.6503},
	"apne2": {"seoul", 37.5665, 126.9780},
	"apne3": {"osaka", 34.6937, 135.5023},
	"aps1": {"mumbai", 19.0760, 72.8777},
	"aps2": {"hyderabad", 17.3850, 78.4867},
	"apse1": {"singapore", 1.3521, 103.8198},
	"apse2": {"sydney", -33.8688, 151.2093},
	"apse3": {"jakarta", -6.2088, 106.8456},
	"apse4": {"melbourne", -37.8136, 144.9631},
	"cac1": {"montreal", 45.5019, -73.5674},
	"euc1": {"frankfurt", 50.1109, 8.6821},
	"euc2": {"zurich", 	47.3769, 8.5417},
	"eun1": {"stockholm", 59.3293, 18.0686},
	"eus1": {"milan", 45.4642, 9.1900},
	"eus2": {"spain", 41.5976, -0.9057},
	"euw1": {"ireland", 53.7798, -7.3055},
	"euw2": {"london", 51.5072, -0.1276},
	"euw3": {"paris", 48.8566, 2.3522},
	"mec1": {"uae", 23.4241, 53.8478},
	"mes1": {"bahrain", 26.0667, 50.5577},
	"sae1": {"saopaulo", -23.5558, -46.6396},
	"use1": {"virginia", 39.0438, -77.4874},
	"use2": {"ohio", 40.4173, -82.9071},
	"usw1": {"sanjose", 37.3387, -121.8853},
	"usw2": {"oregon", 45.8399, -119.7006},
}

type Datacenter struct {
	name      string
	latitude  float32
	longitude float32
}

// ----------------------------------------------------------------------------------

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

type RegionsResponse struct {
	Regions []RegionData
}

type RegionData struct {
	RegionName string
}

type AvailabilityZonesResponse struct {
	AvailabilityZones []AvailabilityZoneData
}

type AvailabilityZoneData struct {
	RegionName string
	ZoneName   string
	ZoneId     string
	ZoneType   string
}

type Zone struct {
	zone            string
	azid            string
	region          string
	local           bool
	datacenter_name string
	latitude        float32
	longitude       float32
}

func main() {

	// get all regions

	fmt.Printf("\nRegions:\n\n")

	output := bash("aws ec2 describe-regions --all-regions")

	regionsResponse := RegionsResponse{}
	if err := json.Unmarshal([]byte(output), &regionsResponse); err != nil {
		panic(err)
	}

	for i := range regionsResponse.Regions {
		fmt.Printf("  %s\n", regionsResponse.Regions[i].RegionName)
	}

	// iterate across each region and get zones

	zones := make([]Zone, 0)

	for i := range regionsResponse.Regions {

		fmt.Printf("\n%s zones:\n\n", regionsResponse.Regions[i].RegionName)

		output = bash(fmt.Sprintf("aws ec2 describe-availability-zones --region=%s --all-availability-zones", regionsResponse.Regions[i].RegionName))

		availabilityZonesResponse := AvailabilityZonesResponse{}
		if err := json.Unmarshal([]byte(output), &availabilityZonesResponse); err != nil {
			panic(err)
		}

		for j := range availabilityZonesResponse.AvailabilityZones {
			zoneType := availabilityZonesResponse.AvailabilityZones[j].ZoneType
			if zoneType != "availability-zone" && zoneType != "local-zone" {
				continue
			}
			local := zoneType == "local-zone"
			zone := availabilityZonesResponse.AvailabilityZones[j].ZoneName
			azid := availabilityZonesResponse.AvailabilityZones[j].ZoneId
			region := availabilityZonesResponse.AvailabilityZones[j].RegionName
			if local {
				fmt.Printf("  %s [%s] (local)\n", zone, azid)
			} else {
				fmt.Printf("  %s [%s]\n", zone, azid)
			}
			zones = append(zones, Zone{zone, azid, region, local, "", 0, 0})
		}

	}

	// sort by azid and print out zones

	fmt.Printf("\nAll zones:\n\n")

	sort.SliceStable(zones, func(i, j int) bool { return zones[i].azid < zones[j].azid })

	for i := range zones {
		values := strings.Split(zones[i].azid, "-")
		a := values[len(values)-2]
		b := values[len(values)-1]
		fmt.Printf("  %s|%s\n", a, b)
	}

	// print out the known datacenters

	fmt.Printf("\nKnown datacenters:\n\n")

	unknown := make([]*Zone, 0)

	for i := range zones {
		values := strings.Split(zones[i].azid, "-")
		a := values[len(values)-2]
		b := values[len(values)-1]
		datacenter := datacenterMap[a]
		number, _ := strconv.Atoi(b[2:])
		if datacenter != nil {
			zones[i].datacenter_name = fmt.Sprintf("amazon.%s.%d", datacenter.name, number)
			zones[i].latitude = datacenter.latitude
			zones[i].longitude = datacenter.longitude
			fmt.Printf("  %s\n", zones[i].datacenter_name)
		} else {
			unknown = append(unknown, &zones[i])
		}
	}

	// print out the unknown datacenters

	fmt.Printf("\nUnknown datacenters:\n\n")

	for i := range unknown {
		fmt.Printf("  %s -> %s\n", unknown[i].zone, unknown[i].azid)
	}

	// generate amazon.txt

	fmt.Printf("\nGenerating amazon.txt\n")

    file, err := os.Create("config/amazon.txt")
    if err != nil {
        panic(err)
    }

    for i := range zones {
    	if zones[i].datacenter_name != "" {
    		fmt.Fprintf(file, "%s,%s\n", zones[i].azid, zones[i].datacenter_name)
    	}
    }

    file.Close()

    // generate amazon.sql

	fmt.Printf("\nGenerating amazon.sql\n")

    file, err = os.Create("schemas/sql/sellers/amazon.sql")
    if err != nil {
        panic(err)
    }

    fmt.Fprintf(file, "\n-- amazon datacenters\n")

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
		"   (select seller_id from sellers where seller_name = 'amazon')\n" +
		");\n"

    for i := range zones {
    	if zones[i].datacenter_name != "" {
    		fmt.Fprintf(file, format_string, zones[i].datacenter_name, zones[i].azid, zones[i].latitude, zones[i].longitude)
    	}
    }

    file.Close()

    // generate amazon.tfvars

	fmt.Printf("\nGenerating amazon.tfvars\n")

    file, err = os.Create("terraform/amazon.tfvars")
    if err != nil {
        panic(err)
    }

    fmt.Fprintf(file, "\namazon_datacenter_map = {\n\n")

    format_string = "  \"%s\" = {\n" +
    	"    azid   = \"%s\"\n" + 
    	"    zone   = \"%s\"\n" + 
    	"    region = \"%s\"\n" + 
    	"  }\n" +
    	"\n"

    for i := range zones {
    	if zones[i].datacenter_name != "" {
    		fmt.Fprintf(file, format_string, zones[i].datacenter_name, zones[i].azid, zones[i].zone, zones[i].region)
    	}
    }

    fmt.Fprintf(file, "}\n")

    fmt.Fprintf(file, "\namazon_regions = [\n")

	for i := range regionsResponse.Regions {
		fmt.Fprintf(file, "  \"%s\",\n", regionsResponse.Regions[i].RegionName)
	}

    fmt.Fprintf(file, "]\n")

    file.Close()
}
