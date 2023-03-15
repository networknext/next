package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"encoding/gob"
	"sort"

	// "encoding/json"
	// 
	// "os"
	// "sort"
	// "strconv"
)

// ===========================================================================================================================================

// This definition drives the set of google datacenters, eg. "google.[country/city].[number]"

var datacenterMap = map[string]*Datacenter{

	"us-east1":  {"southcarolina", 33.8361, -81.1637},

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
	Zone            string
	Region          string
	DatacenterName  string
	Latitude        float32
	Longitude       float32
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
	for _,v := range zoneMap {
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
		// for i := range unknown {
			// fmt.Printf("  %s\n", unknown[i].Zone)
		// }
	}















/*

	// generate amazon.txt

	fmt.Printf("\nGenerating amazon.txt\n")

	file, err := os.Create("config/amazon.txt")
	if err != nil {
		panic(err)
	}

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, "%s,%s\n", zones[i].AZID, zones[i].DatacenterName)
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
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].AZID, zones[i].Latitude, zones[i].Longitude)
		}
	}

	file.Close()

	// generate amazon/generated.tf

	fmt.Printf("\nGenerating amazon/generated.tf\n")

	file, err = os.Create("terraform/dev/relays/amazon/generated.tf")
	if err != nil {
		panic(err)
	}

	header := `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}
`
	fmt.Fprintf(file, header)

	format_string = "\nprovider \"aws\" { \n" +
		"  shared_config_files      = var.config\n" +
		"  shared_credentials_files = var.credentials\n" +
		"  profile                  = var.profile\n" +
		"  alias                    = \"%s\"\n" +
		"  region                   = \"%s\"\n" +
		"}\n"

	for i := range regionsResponse.Regions {
		fmt.Fprintf(file, format_string, regionsResponse.Regions[i].RegionName, regionsResponse.Regions[i].RegionName)
	}

	format_string = "\nmodule \"region_%s\" { \n" +
		"  source              = \"./region\"\n" +
		"  vpn_address         = var.vpn_address\n" +
		"  ssh_public_key_file = var.ssh_public_key_file\n" +
		"  providers = {\n" +
		"    aws = aws.%s\n" +
		"  }\n" +
		"}\n"

	for i := range regionsResponse.Regions {
		fmt.Fprintf(file, format_string, strings.ReplaceAll(regionsResponse.Regions[i].RegionName, "-", "_"), regionsResponse.Regions[i].RegionName)
	}

	fmt.Fprintf(file, "\nlocals {\n\n  datacenter_map = {\n\n")

	format_string = "    \"%s\" = {\n" +
		"      azid   = \"%s\"\n" +
		"      zone   = \"%s\"\n" +
		"      region = \"%s\"\n" +
		"    }\n" +
		"\n"

	for i := range zones {
		if zones[i].DatacenterName != "" {
			fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].AZID, zones[i].Zone, zones[i].Region)
		}
	}

	fmt.Fprintf(file, "  }\n\n  regions = [\n")

	for i := range regionsResponse.Regions {
		fmt.Fprintf(file, "    \"%s\",\n", regionsResponse.Regions[i].RegionName)
	}

	fmt.Fprintf(file, "  ]\n}\n")

	fmt.Fprintf(file, "\nlocals {\n\n  relays = {\n\n")

	for k, v := range devRelayMap {
		fmt.Fprintf(file, "    \"%s\" = { datacenter_name = \"%s\" },\n", k, v[0])
	}

	fmt.Fprintf(file, "  }\n\n}\n\n")

	relay_module := `module "relay_%s" {
  source            = "./relay"
  name              = "%s"
  zone              = local.datacenter_map["%s"].zone
  region            = local.datacenter_map["%s"].region
  type              = "%s"
  ami               = "%s"
  security_group_id = module.region_%s.security_group_id
  providers = {
    aws = aws.%s
  }
}

`
	for k, v := range devRelayMap {
		fmt.Fprintf(file, relay_module, strings.ReplaceAll(k, ".", "_"), k, v[0], v[0], v[1], v[2], strings.ReplaceAll(datacenterToRegion[v[0]], "-", "_"), datacenterToRegion[v[0]])
	}

	output_header := `output "relays" {

  description = "Data for each amazon relay setup by Terraform"

  value = {

`
	fmt.Fprintf(file, output_header)

	relay_output := `    "%s" = {
      "relay_name"       = "%s"
      "datacenter_name"  = "%s"
      "supplier_name"    = "amazon"
      "public_address"   = "${module.relay_%s.public_address}:40000"
      "internal_address" = "${module.relay_%s.internal_address}:40000"
      "internal_group"   = "%s"
      "ssh_address"      = "${module.relay_%s.public_address}:22"
      "ssh_user"         = "ubuntu"
    }

`
	for k, v := range devRelayMap {
		region := datacenterToRegion[v[0]]
		internal_group := region
		if datacenterIsLocal[v[0]] {
			internal_group = v[0]
		}
		datacenter_underscores := strings.ReplaceAll(v[0], ".", "_")
		fmt.Fprintf(file, relay_output, k, k, v[0], datacenter_underscores, datacenter_underscores, internal_group, datacenter_underscores)
	}

	fmt.Fprintf(file, "\n  }\n\n}\n")

	file.Close()
*/
}
