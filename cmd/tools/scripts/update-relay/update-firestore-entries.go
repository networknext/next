package main

import (
	"context"
	"encoding/base64"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/api/iterator"
)

type newRelayData struct {
	Address   string
	PublicKey string
}

type firestoreRelay struct {
	Address   string `firestore:"publicAddress"`
	PublicKey []byte `firestore:"publicKey"`
}

// ADD NEW RELAY DATA HERE
// Add the relay's IP and the new public key to update it with
var relayData []newRelayData = []newRelayData{}

func main() {
	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var gcpProjectID string
	var ok bool

	if _, ok = os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !ok {
		level.Error(logger).Log("err", "GOOGLE_APPLICATION_CREDENTIALS env var not set")
		os.Exit(1)
	}
	if gcpProjectID, ok = os.LookupEnv("GOOGLE_PROJECT_ID"); !ok {
		level.Error(logger).Log("err", "GOOGLE_PROJECT_ID env var not set")
		os.Exit(1)
	}

	firestoreClient, err := firestore.NewClient(ctx, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	iter := firestoreClient.Collection("Relay").Documents(ctx)
	for {
		rdoc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		var r firestoreRelay
		err = rdoc.DataTo(&r)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		for _, data := range relayData {
			if r.Address == data.Address {
				publicKey, err := base64.StdEncoding.DecodeString(data.PublicKey)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}

				rdoc.Ref.Set(ctx, map[string]interface{}{
					"publicKey": publicKey,
				}, firestore.MergeAll)

				break
			}
		}
	}
}
