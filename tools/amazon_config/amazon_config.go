package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"encoding/json"
	"strings"
	"sort"
)

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
	ZoneName string
	ZoneId string
	ZoneType string
}

type Zone struct {
	zone string
	azid string
	region string
	local bool
	datacenter_name string
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

	    	zones = append(zones, Zone{zone, azid, region, local, ""})
	    }

    }

    // sort by azid and print out zones

    fmt.Printf("\nAll zones:\n\n")

    sort.SliceStable(zones, func(i, j int) bool { return zones[i].azid < zones[j].azid })

    for i := range zones {
		values := strings.Split(zones[i].azid, "-")
		a := values[len(values)-2]
		b := values[len(values)-1]
    	fmt.Printf("  %s|%s -> %s\n", a, b, zones[i].region)
    }
}
