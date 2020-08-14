package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"google.golang.org/api/iterator"
)

type relay struct {
	Name               string                 `firestore:"displayName"`
	Address            string                 `firestore:"publicAddress"`
	PublicKey          []byte                 `firestore:"publicKey"`
	UpdateKey          []byte                 `firestore:"updateKey"`
	NICSpeedMbps       int64                  `firestore:"nicSpeedMbps"`
	IncludedBandwithGB int64                  `firestore:"includedBandwidthGB"`
	Datacenter         *firestore.DocumentRef `firestore:"datacenter"`
	Seller             *firestore.DocumentRef `firestore:"seller"`
	ManagementAddress  string                 `firestore:"managementAddress"`
	SSHUser            string                 `firestore:"sshUser"`
	SSHPort            int64                  `firestore:"sshPort"`
	State              routing.RelayState     `firestore:"state"`
	LastUpdateTime     time.Time              `firestore:"lastUpdateTime"`
	MaxSessions        int32                  `firestore:"maxSessions"`
	MRC                int64                  `firestore:"monthlyRecurringChargeNibblins"`
	Overage            int64                  `firestore:"overage"`
	BWRule             int32                  `firestore:"bandwidthRule"`
	ContractTerm       int32                  `firestore:"contractTerm"`
	StartDate          time.Time              `firestore:"startDate"`
	EndDate            time.Time              `firestore:"endDate"`
	Type               string                 `firestore:"machineType"`
}

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
	// gcpProjectID := "network-next-v3-dev"
	gcpProjectID := "network-next-v3-prod"

	client, err := firestore.NewClient(ctx, gcpProjectID)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	relayCollection := client.Collection("Relay").Documents(ctx)
	defer relayCollection.Stop()
	for {
		rdoc, err := relayCollection.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf(" Next() err: %v\n", err)
			os.Exit(1)
		}

		var r relay
		err = rdoc.DataTo(&r)
		if err != nil {
			fmt.Printf("DataTo() err: %v\n", err)
			os.Exit(1)
		}

		rid := crypto.HashID(r.Address)
		// ridHex := fmt.Sprintf("016x", rid)
		ridSigned := int64(rid)

		fmt.Printf("Relay %s : %016x : %d\n", r.Name, rid, ridSigned)

		_, err = rdoc.Ref.Update(ctx, []firestore.Update{{Path: "signedID", Value: ridSigned}})
		_, err = rdoc.Ref.Update(ctx, []firestore.Update{{Path: "hexID", Value: fmt.Sprintf("%016x", rid)}})

		// only change one record
		// break

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

		fmt.Printf("Datacenter %s : %016x : %d\n", d.Name, did, didSigned)

		_, err = ddoc.Ref.Update(ctx, []firestore.Update{{Path: "signedID", Value: didSigned}})
		_, err = ddoc.Ref.Update(ctx, []firestore.Update{{Path: "hexID", Value: fmt.Sprintf("%016x", did)}})

		// only change one record
		// break

	}

}
