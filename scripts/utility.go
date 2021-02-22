package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/storage"
)

// Chose the utility function to use
func main() {
	DatacenterReverseLookup()

	// Set these variables depending on the environment
	// Remember to also export GOOGLE_APPLICATION_CREDENTIALS env var
	gcpProjectID := "local"
	btInstanceID := "localhost:8086"
	btTableName := "portal-session-history"
	prefix := "prefix_of_rows_to_delete_goes_here"
	DeleteBigtableRows(gcpProjectID, btInstanceID, btTableName, prefix)
}

// Fill in the "matches" string slice with datacenter hashes you want to brute force search.
// The program will pair up all known suppliers with a list of common cities to try and find a match.
func DatacenterReverseLookup() {
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

// Delete rows from a Bigtable instance based on a prefix
func DeleteBigtableRows(gcpProjectID, btInstanceID, btTableName, prefix string) {
	ctx := context.Background()
	logger := log.NewNopLogger()

	if os.Getenv("BIGTABLE_EMULATOR_HOST") != "" {
		fmt.Println("Detected Bigtable emulator")
	}

	// Get a bigtable admin client
	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID, logger)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	defer func() {
		// Close the admin client
		err = btAdmin.Close()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}()

	// Verify table exists
	exists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if !exists {
		fmt.Printf("Table %s does not exist in instance %s. Aborting.\n", btTableName, btInstanceID)
		return
	}

	// Delete rows with prefix from table
	err = btAdmin.DropRowsByPrefix(ctx, btTableName, prefix)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully deleted rows with prefix %s from table %s\n", prefix, btTableName)
}
