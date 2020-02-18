package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/networknext/backend/storage"
)

func main() {
	// Sets your Google Cloud Platform project ID.
	projectID := "network-next-v3-dev"

	// Get a Firestore client.
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Close client when done.
	defer client.Close()

	db := storage.Firestore{
		Client: client,
	}

	err = db.Sync(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(db.Buyer(10149800775011964915))
}
