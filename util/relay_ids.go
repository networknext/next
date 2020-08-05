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
		ridSigned := int64(rid)

		fmt.Printf("%s : %016x : %d\n", r.Name, rid, ridSigned)

		_, err = rdoc.Ref.Update(ctx, []firestore.Update{{Path: "signedID", Value: ridSigned}})

		break

	}

}
