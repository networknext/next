package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/storage"
)

func main() {
	ctx := context.Background()
	logger := log.NewNopLogger()

	// Set these variables depending on the environment
	// Remember to also export GOOGLE_APPLICATION_CREDENTIALS env var
	gcpProjectID := "local"
	btInstanceID := "localhost:8086"
	btTableName := "portal-session-history"
	prefix := "prefix_of_rows_to_delete_goes_here"

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
