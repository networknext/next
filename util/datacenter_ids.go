package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/networknext/backend/crypto"
	"google.golang.org/api/iterator"
)

type datacenter struct {
	Name         string  `firestore:"name"`
	Enabled      bool    `firestore:"enabled"`
	Latitude     float64 `firestore:"latitude"`
	Longitude    float64 `firestore:"longitude"`
	SupplierName string  `firestore:"supplierName"`
}

func main() {

	// don't forget to e.g.:
	// export GOOGLE_APPLICATION_CREDENTIALS=../../../dev-credentials.json

	ctx := context.Background()
	gcpProjectID := "network-next-v3-dev"

	client, err := firestore.NewClient(ctx, gcpProjectID)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	datacenterCollection := client.Collection("Datacenter").Documents(ctx)
	defer datacenterCollection.Stop()
	for {
		ddoc, err := datacenterCollection.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf(" Next() err: %v\n", err)
			os.Exit(1)
		}

		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			fmt.Printf("DataTo() err: %v\n", err)
			os.Exit(1)
		}

		did := crypto.HashID(d.Name)
		didSigned := int64(did)

		fmt.Printf("%s : %016x : %d\n", d.Name, did, didSigned)

		_, err = ddoc.Ref.Update(ctx, []firestore.Update{{Path: "signedID", Value: didSigned}})

		break

	}

}
