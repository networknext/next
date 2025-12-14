package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// ===========================================================================================================================================

// DEV RELAYS

var devRelayMap = map[string][]string{
	"amazon.ohio.1": {"amazon.ohio.1", "m5a.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.ohio.2": {"amazon.ohio.2", "m5a.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
}

// PROD RELAYS

var prodRelayMap = map[string][]string{

	"amazon.calgary.1": {"amazon.calgary.1", "t3.medium", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.portland.1": {"amazon.portland.1", "t3.medium", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.seattle.1": {"amazon.seattle.1", "t3.medium", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.denver.1": {"amazon.denver.1", "t3.medium", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.montreal.1": {"amazon.montreal.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.oregon.1": {"amazon.oregon.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.ohio.1": {"amazon.ohio.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.saopaulo.1": {"amazon.saopaulo.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.saopaulo.2": {"amazon.saopaulo.2", "m5a.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.saopaulo.3": {"amazon.saopaulo.3", "m5a.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	
	"amazon.buenosaires.1": {"amazon.buenosaires.1", "t3.medium" /*"r5.large"*/, "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.lima.1":        {"amazon.lima.1", "t3.medium", "ubuntu-minimal/images/hvm-ssd/ubuntu-jammy-22.04-amd64-minimal-*"},
	
	"amazon.santiago.1":    {"amazon.santiago.1", "t3.medium" /*"t3.large"*/, "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	
	"amazon.frankfurt.1":   {"amazon.frankfurt.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	/*
	"amazon.frankfurt.2": {"amazon.frankfurt.2", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.frankfurt.3": {"amazon.frankfurt.3", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	*/

	"amazon.virginia.1": {"amazon.virginia.1", "m5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.virginia.2": {"amazon.virginia.2", "m5a.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.virginia.3": {"amazon.virginia.3", "c3.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.virginia.4": {"amazon.virginia.4", "m5a.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.virginia.5": {"amazon.virginia.5", "m5a.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.virginia.6": {"amazon.virginia.6", "r5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.dallas.1":     {"amazon.dallas.1", "c6i.large" /*"c6i.xlarge"*/, "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.miami.1":      {"amazon.miami.1", "c6i.large" /*"c6i.xlarge"*/, "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.queretaro.1":  {"amazon.queretaro.1", "t3.medium" /*"c5.2xlarge"*/, "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.losangeles.1": {"amazon.losangeles.1", "c5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.losangeles.2": {"amazon.losangeles.2", "c5.2xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.bahrain.1": {"amazon.bahrain.1", "c5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	/*
	"amazon.bahrain.2": {"amazon.bahrain.2", "c5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	"amazon.bahrain.3": {"amazon.bahrain.3", "c5a.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	*/

	"amazon.uae.1": {"amazon.uae.1", "c5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.uae.2": {"amazon.uae.2", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.uae.3": {"amazon.uae.3", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.telaviv.1": {"amazon.telaviv.1", "c5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.telaviv.2": {"amazon.telaviv.2", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.telaviv.3": {"amazon.telaviv.3", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	// "amazon.mexico.1": {"amazon.mexico.1", "m6i.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.mexico.2": {"amazon.mexico.2", "m6i.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.mexico.3": {"amazon.mexico.3", "m6i.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.london.1": {"amazon.london.1", "c5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.london.2": {"amazon.london.2", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.london.3": {"amazon.london.3", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	// "amazon.ireland.1": {"amazon.ireland.1", "c5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.ireland.2": {"amazon.ireland.2", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.ireland.3": {"amazon.ireland.3", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.oman.1":  {"amazon.oman.1", "t3.medium", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.spain.1": {"amazon.spain.1", "c5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.spain.2": {"amazon.spain.2", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.spain.3": {"amazon.spain.3", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},

	"amazon.paris.1": {"amazon.paris.1", "c5.large", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.paris.2": {"amazon.paris.2", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
	// "amazon.paris.3": {"amazon.paris.3", "c5.xlarge", "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
}

// Exclude regions

var excludedRegions = map[string]bool{
	"ap-southeast-6": true,
}

// ===========================================================================================================================================

// Amazon datacenters, eg. "amazon.[country/city].[number]"

var datacenterMap = map[string]*Datacenter{

	// regions (AZID)

	"afs1":  {"johannesburg", -33.9249, 18.4241},
	"ape1":  {"hongkong", 22.3193, 114.1694},
	"apne1": {"tokyo", 35.6762, 139.6503},
	"apne2": {"seoul", 37.5665, 126.9780},
	"apne3": {"osaka", 34.6937, 135.5023},
	"aps1":  {"mumbai", 19.0760, 72.8777},
	"aps2":  {"hyderabad", 17.3850, 78.4867},
	"apse1": {"singapore", 1.3521, 103.8198},
	"apse2": {"sydney", -33.8688, 151.2093},
	"apse3": {"jakarta", -6.2088, 106.8456},
	"apse4": {"melbourne", -37.8136, 144.9631},
	"apse5": {"malaysia", 4.2105, 101.9758},
	"apse6": {"newzealand", -36.8509, 174.7645},
	"apse7": {"thailand", 15.8700, 100.9925},
	"cac1":  {"montreal", 45.5019, -73.5674},
	"euc1":  {"frankfurt", 50.1109, 8.6821},
	"euc2":  {"zurich", 47.3769, 8.5417},
	"eun1":  {"stockholm", 59.3293, 18.0686},
	"eus1":  {"milan", 45.4642, 9.1900},
	"eus2":  {"spain", 41.5976, -0.9057},
	"euw1":  {"ireland", 53.7798, -7.3055},
	"euw2":  {"london", 51.5072, -0.1276},
	"euw3":  {"paris", 48.8566, 2.3522},
	"mec1":  {"uae", 23.4241, 53.8478},
	"mes1":  {"bahrain", 26.0667, 50.5577},
	"sae1":  {"saopaulo", -23.5558, -46.6396},
	"use1":  {"virginia", 39.0438, -77.4874},
	"use2":  {"ohio", 40.4173, -82.9071},
	"usw1":  {"sanjose", 37.3387, -121.8853},
	"usw2":  {"oregon", 45.8399, -119.7006},
	"mxc1":  {"mexico", 23.6345, -102.5528},
	"ilc1":  {"telaviv", 32.0853, 34.7818},
	"caw1":  {"calgary", 51.0447, -114.0719},

	// local zones (AZID)

	"los1": {"nigeria", 6.5244, 3.3792},
	"tpe1": {"taipei", 25.0330, 121.5654},
	"ccu1": {"kolkata", 22.5726, 88.3639},
	"del1": {"delhi", 28.7041, 77.1025},
	"bkk1": {"bangkok", 13.7563, 100.5018},
	"per1": {"perth", -31.9523, 115.8613},
	"ham1": {"hamburg", 53.5488, 9.9872},
	"waw1": {"warsaw", 52.2297, 21.0122},
	"cph1": {"copenhagen", 55.6761, 12.5683},
	"hel1": {"finland", 60.1699, 24.9384},
	"mct1": {"oman", 23.5880, 58.3829},
	"atl1": {"atlanta", 33.7488, -84.3877},
	"bos1": {"boston", 42.3601, -71.0589},
	"bue1": {"buenosaires", -34.6037, -58.3816},
//	"chi1": {"chicago", 41.8781, -87.6298},
	"dfw2": {"dallas", 32.7767, -96.7970},
	"iah1": {"houston", 29.7604, -95.3698},
	"lim1": {"lima", -12.0464, -77.0428},
	"mci1": {"kansas", 39.0997, -94.5786},
	"mia2": {"miami", 25.7617, -80.1918},
	"msp1": {"minneapolis", 44.9778, -93.2650},
// "nyc1": {"newyork", 40.7128, -74.0060},
	"phl1": {"philadelphia", 39.9526, -75.1652},
	"qro1": {"queretaro", 23.6345, -102.5528},
	"scl1": {"santiago", -33.4489, -70.6693},
	"den1": {"denver", 39.7392, -104.9903},
	"las1": {"lasvegas", 36.1716, -115.1391},
	"lax1": {"losangeles", 34.0522, -118.2437},
	"pdx1": {"portland", 45.5152, -122.6784},
	"phx1": {"phoenix", 33.4484, -112.0740},
	"sea1": {"seattle", 47.6062, -122.3321},
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
		return ""
	}
	return output.String()
}

type RegionsResponse struct {
	Regions []RegionData
}

type RegionData struct {
	RegionName string
	Excluded   bool
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
	Zone           string
	AZID           string
	Region         string
	Local          bool
	DatacenterName string
	Latitude       float32
	Longitude      float32
}

func main() {

	// create cache dir if needed

	bash("mkdir -p cache")

	// load regions cache, if possible

	regionsResponse := RegionsResponse{}

	loadedRegionsCache := false

	{
		file, err := os.Open("cache/amazon_regions.bin")
		if err == nil {
			gob.NewDecoder(file).Decode(&regionsResponse)
			if err == nil {
				loadedRegionsCache = true
			}
			file.Close()
		}
	}

	// otherwise, get all regions and save to cache

	needToSaveRegionsCache := false

	if !loadedRegionsCache {

		fmt.Printf("\n")

		output := bash("aws ec2 describe-regions --all-regions")

		if err := json.Unmarshal([]byte(output), &regionsResponse); err != nil {
			panic(err)
		}

		for i := range regionsResponse.Regions {
			fmt.Printf("  %s\n", regionsResponse.Regions[i].RegionName)
		}

		needToSaveRegionsCache = true
	}

	// load zones cache, if possible

	zones := make([]Zone, 0)

	loadedZonesCache := false

	{
		file, err := os.Open("cache/amazon_zones.bin")
		if err == nil {
			gob.NewDecoder(file).Decode(&zones)
			if err == nil {
				loadedZonesCache = true
			}
			file.Close()
		}
	}

	// otherwise, iterate across each region and get zones then save to cache

	if !loadedZonesCache {

		for i := range regionsResponse.Regions {

			fmt.Printf("\n%s zones:\n\n", regionsResponse.Regions[i].RegionName)

			output := bash(fmt.Sprintf("aws ec2 describe-availability-zones --region=%s --all-availability-zones", regionsResponse.Regions[i].RegionName))

			if output == "" {
				fmt.Printf("  Excluding region '%s' because it's not enabled in your AWS account.\n", regionsResponse.Regions[i].RegionName)
				regionsResponse.Regions[i].Excluded = true
				continue
			}

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

		// sort by azid then cache results

		sort.SliceStable(zones, func(i, j int) bool { return zones[i].AZID < zones[j].AZID })

		{
			file, err := os.Create("cache/amazon_zones.bin")
			if err != nil {
				panic(err)
			}
			gob.NewEncoder(file).Encode(zones)
			file.Close()
		}
	}

	// write regions cache here, because we need to save the excluded flag above, post-iterating across all regions and getting zones

	if needToSaveRegionsCache {
		file, err := os.Create("cache/amazon_regions.bin")
		if err != nil {
			panic(err)
		}
		gob.NewEncoder(file).Encode(&regionsResponse)
		file.Close()
	}

	// print all zones

	fmt.Printf("\nAll zones:\n\n")

	for i := range zones {
		values := strings.Split(zones[i].AZID, "-")
		a := values[len(values)-2]
		b := values[len(values)-1]
		fmt.Printf("  %s|%s\n", a, b)
	}

	// print out the known datacenters

	fmt.Printf("\nKnown datacenters:\n\n")

	unknown := make([]*Zone, 0)

	datacenterToRegion := make(map[string]string)
	datacenterIsLocal := make(map[string]bool)

	for i := range zones {
		values := strings.Split(zones[i].AZID, "-")
		a := values[len(values)-2]
		b := values[len(values)-1]
		datacenter := datacenterMap[a]
		number, _ := strconv.Atoi(b[2:])
		if datacenter != nil {
			zones[i].DatacenterName = fmt.Sprintf("amazon.%s.%d", datacenter.name, number)
			zones[i].Latitude = datacenter.latitude
			zones[i].Longitude = datacenter.longitude
			fmt.Printf("  %s\n", zones[i].DatacenterName)
			datacenterToRegion[zones[i].DatacenterName] = zones[i].Region
			datacenterIsLocal[zones[i].DatacenterName] = zones[i].Local
		} else {
			unknown = append(unknown, &zones[i])
		}
	}

	// print out the unknown datacenters

	if len(unknown) > 0 {
		fmt.Printf("\nUnknown datacenters:\n\n")
		for i := range unknown {
			fmt.Printf("  %s -> %s\n", unknown[i].Zone, unknown[i].AZID)
		}
	}

	// print excluded regions (not enabled in AWS account)

	fmt.Printf("\nExcluded regions:\n\n")

	for i := range regionsResponse.Regions {
		if regionsResponse.Regions[i].Excluded || excludedRegions[regionsResponse.Regions[i].RegionName] {
			fmt.Printf("  %s\n", regionsResponse.Regions[i].RegionName)
		}
	}

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

	// generate dev amazon/generated.tf
	{
		fmt.Printf("\nGenerating dev amazon/generated.tf\n")

		file, err = os.Create("terraform/dev/relays/amazon/generated.tf")
		if err != nil {
			panic(err)
		}

		header := `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
`
		fmt.Fprintf(file, header)

		format_string := "\nprovider \"aws\" { \n" +
			"  shared_config_files      = var.config\n" +
			"  shared_credentials_files = var.credentials\n" +
			"  profile                  = var.profile\n" +
			"  alias                    = \"%s\"\n" +
			"  region                   = \"%s\"\n" +
			"}\n"

		for i := range regionsResponse.Regions {
			if !regionsResponse.Regions[i].Excluded && !excludedRegions[regionsResponse.Regions[i].RegionName] {
				fmt.Fprintf(file, format_string, regionsResponse.Regions[i].RegionName, regionsResponse.Regions[i].RegionName)
			}
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
			if !regionsResponse.Regions[i].Excluded && !excludedRegions[regionsResponse.Regions[i].RegionName] {
				fmt.Fprintf(file, format_string, strings.ReplaceAll(regionsResponse.Regions[i].RegionName, "-", "_"), regionsResponse.Regions[i].RegionName)
			}
		}

		fmt.Fprintf(file, "\nlocals {\n\n  datacenter_map = {\n\n")

		format_string = "    \"%s\" = {\n" +
			"      azid        = \"%s\"\n" +
			"      zone        = \"%s\"\n" +
			"      region      = \"%s\"\n" +
			"      native_name = \"%s\"\n" +
			"      latitude    = %.2f\n" +
			"      longitude   = %.2f\n" +
			"      seller_name = \"Amazon\"\n" +
			"      seller_code = \"amazon\"\n" +
			"    }\n" +
			"\n"

		for i := range zones {
			if zones[i].DatacenterName != "" {
				fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].AZID, zones[i].Zone, zones[i].Region, fmt.Sprintf("%s (%s)", zones[i].AZID, zones[i].Zone), zones[i].Latitude, zones[i].Longitude)
			}
		}

		fmt.Fprintf(file, "  }\n\n  regions = [\n")

		for i := range regionsResponse.Regions {
			if !regionsResponse.Regions[i].Excluded && !excludedRegions[regionsResponse.Regions[i].RegionName] {
				fmt.Fprintf(file, "    \"%s\",\n", regionsResponse.Regions[i].RegionName)
			}
		}

		fmt.Fprintf(file, "  ]\n}\n")

		fmt.Fprintf(file, "\nlocals {\n\n  relays = {\n\n")

		devRelayNames := make([]string, len(devRelayMap))
		index := 0
		for k := range devRelayMap {
			devRelayNames[index] = k
			index++
		}

		sort.SliceStable(devRelayNames, func(i, j int) bool { return devRelayNames[i] < devRelayNames[j] })

		for i := range devRelayNames {
			k := devRelayNames[i]
			v := devRelayMap[k]
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
	  vpn_address       = var.vpn_address
	  providers = {
	    aws = aws.%s
	  }
	}
	`

		for i := range devRelayNames {
			k := devRelayNames[i]
			v := devRelayMap[k]
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
	      "seller_name"      = "Amazon"
	      "seller_code"      = "amazon"
	      "public_ip"        = module.relay_%s.public_address
	      "public_port"      = 40000
	      "internal_ip"      = module.relay_%s.internal_address
	      "internal_port"    = 40000
	      "internal_group"   = "%s"
	      "ssh_ip"           = module.relay_%s.public_address
	      "ssh_port"         = 22
	      "ssh_user"         = "ubuntu"
	      "bandwidth_price"  = 2
	    }

	`
		for i := range devRelayNames {
			k := devRelayNames[i]
			v := devRelayMap[k]
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
	}

	// generate prod amazon/generated.tf
	{
		fmt.Printf("\nGenerating prod amazon/generated.tf\n")

		file, err = os.Create("terraform/prod/relays/amazon/generated.tf")
		if err != nil {
			panic(err)
		}

		header := `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
`
		fmt.Fprintf(file, header)

		format_string := "\nprovider \"aws\" { \n" +
			"  shared_config_files      = var.config\n" +
			"  shared_credentials_files = var.credentials\n" +
			"  profile                  = var.profile\n" +
			"  alias                    = \"%s\"\n" +
			"  region                   = \"%s\"\n" +
			"}\n"

		for i := range regionsResponse.Regions {
			if !regionsResponse.Regions[i].Excluded && !excludedRegions[regionsResponse.Regions[i].RegionName] {
				fmt.Fprintf(file, format_string, regionsResponse.Regions[i].RegionName, regionsResponse.Regions[i].RegionName)
			}
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
			if !regionsResponse.Regions[i].Excluded && !excludedRegions[regionsResponse.Regions[i].RegionName] {
				fmt.Fprintf(file, format_string, strings.ReplaceAll(regionsResponse.Regions[i].RegionName, "-", "_"), regionsResponse.Regions[i].RegionName)
			}
		}

		fmt.Fprintf(file, "\nlocals {\n\n  datacenter_map = {\n\n")

		format_string = "    \"%s\" = {\n" +
			"      azid        = \"%s\"\n" +
			"      zone        = \"%s\"\n" +
			"      region      = \"%s\"\n" +
			"      native_name = \"%s\"\n" +
			"      latitude    = %.2f\n" +
			"      longitude   = %.2f\n" +
			"      seller_name = \"Amazon\"\n" +
			"      seller_code = \"amazon\"\n" +
			"    }\n" +
			"\n"

		for i := range zones {
			if zones[i].DatacenterName != "" {
				fmt.Fprintf(file, format_string, zones[i].DatacenterName, zones[i].AZID, zones[i].Zone, zones[i].Region, fmt.Sprintf("%s (%s)", zones[i].AZID, zones[i].Zone), zones[i].Latitude, zones[i].Longitude)
			}
		}

		fmt.Fprintf(file, "  }\n\n  regions = [\n")

		for i := range regionsResponse.Regions {
			if !regionsResponse.Regions[i].Excluded && !excludedRegions[regionsResponse.Regions[i].RegionName] {
				fmt.Fprintf(file, "    \"%s\",\n", regionsResponse.Regions[i].RegionName)
			}
		}

		fmt.Fprintf(file, "  ]\n}\n")

		fmt.Fprintf(file, "\nlocals {\n\n  relays = {\n\n")

		prodRelayNames := make([]string, len(prodRelayMap))
		index := 0
		for k := range prodRelayMap {
			prodRelayNames[index] = k
			index++
		}

		sort.SliceStable(prodRelayNames, func(i, j int) bool { return prodRelayNames[i] < prodRelayNames[j] })

		for i := range prodRelayNames {
			k := prodRelayNames[i]
			v := prodRelayMap[k]
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
	  vpn_address       = var.vpn_address
	  providers = {
	    aws = aws.%s
	  }
	}
	`

		for i := range prodRelayNames {
			k := prodRelayNames[i]
			v := prodRelayMap[k]

			region, ok := datacenterToRegion[v[0]]
			if !ok {
				fmt.Printf("missing datacenter to region for '%s'\n", v[0])
				fmt.Printf("========================\n")
				for k := range datacenterToRegion {
					fmt.Printf("%s\n", k)
				}
				fmt.Printf("========================\n")
				os.Exit(1)
			}
			_ = region

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
	      "seller_name"      = "Amazon"
	      "seller_code"      = "amazon"
	      "public_ip"        = module.relay_%s.public_address
	      "public_port"      = 40000
	      "internal_ip"      = module.relay_%s.internal_address
	      "internal_port"    = 40000
	      "internal_group"   = "%s"
	      "ssh_ip"           = module.relay_%s.public_address
	      "ssh_port"         = 22
	      "ssh_user"         = "ubuntu"
	      "bandwidth_price"  = 2
	    }

	`
		for i := range prodRelayNames {
			k := prodRelayNames[i]
			v := prodRelayMap[k]
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
	}

	fmt.Printf("\n")
}
