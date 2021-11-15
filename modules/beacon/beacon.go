/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package beacon

import (
	"context"

	"github.com/networknext/backend/modules/transport"
)

// Beaconer is a beacon service interface that handles sending beacon packet entries through google pubsub to bigquery
type Beaconer interface {
	Submit(ctx context.Context, entry *transport.NextBeaconPacket) error
	Close()
}
