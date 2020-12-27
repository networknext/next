package main

import (
	"fmt"

	"github.com/networknext/backend/modules/crypto"
)

// Fill in the "matches" string slice with datacenter hashes you want to brute force search.
// The program will pair up all known suppliers with a list of common cities to try and find a match.
func main() {
	cityNames := []string{
		"",
		"amsterdam",
		"atlanta",
		"beijing",
		"bangalore",
		"chennai",
		"chicago",
		"copenhagen",
		"dallas",
		"daressalaam",
		"delhi",
		"denpasar",
		"dubai",
		"eindhoven",
		"florida",
		"frankfurt",
		"fujairah",
		"heerlen",
		"hongkong",
		"hyderabad",
		"jakarta",
		"johannesburg",
		"kolkata",
		"kuala",
		"kyiv",
		"london",
		"losangeles",
		"luanda",
		"luxembourg",
		"madrid",
		"manila",
		"mexico",
		"miami",
		"montreal",
		"moscow",
		"mumbai",
		"newdelhi",
		"newjersey",
		"newyork",
		"osaka",
		"paris",
		"perth",
		"phoenix",
		"rotterdam",
		"saintlouis",
		"saintpetersburg",
		"saltlakecity",
		"sanjose",
		"santaclara",
		"sao",
		"saopaolo",
		"saopaulo",
		"seattle",
		"semarang",
		"seoul",
		"singapore",
		"shanghai",
		"stlouis",
		"stpetersburg",
		"stockholm",
		"strasbourg",
		"surabaya",
		"sydney",
		"tokyo",
		"toronto",
		"vancouver",
		"warsaw",
		"washingtondc",
	}

	supplierNames := []string{
		"",
		"multiplay",
		"100tb",
		"amazon",
		"atman",
		"azure",
		"buzinessware",
		"colocrossing",
		"datapacket",
		"dedicatedsolutions",
		"dinahosting",
		"gcore",
		"glesys",
		"google",
		"handynetworks",
		"hostdime",
		"hosthink",
		"hostirian",
		"hostroyale",
		"i3d",
		"ibm",
		"inap",
		"intergrid",
		"kamatera",
		"leaseweb",
		"limelight",
		"maxihost",
		"netdedi",
		"nforce",
		"opencolo",
		"ovh",
		"packet",
		"pallada",
		"phoenixnap",
		"psychz",
		"radore",
		"riot",
		"selectel",
		"serversaustralia",
		"serversdotcom",
		"serverzoo",
		"singlehop",
		"springshosting",
		"stackpath",
		"totalserversolutions",
		"vultr",
		"zenlayer",
	}

	matches := []string{
		"6d7c3a02b218131d",
		"83462735ba7ef971",
		"c38c805d821cd0d3",
		"f45cb29b3b74e986",
	}

	oneMatchFound := false
	for _, supplierName := range supplierNames {
		for _, cityName := range cityNames {
			datacenterName := supplierName + "." + cityName

			for _, match := range matches {
				if match == fmt.Sprintf("%016x", crypto.HashID(datacenterName)) {
					fmt.Printf("match found! datacenter name for %s is %s\n", match, datacenterName)
					oneMatchFound = true
				}
			}
		}
	}

	if !oneMatchFound {
		fmt.Println("no matches found")
	}
}
