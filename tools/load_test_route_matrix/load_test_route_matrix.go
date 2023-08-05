package main

import (
	"context"
	"fmt"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
)

func RunRouteMatrixThread(ctx context.Context) {

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				numRelays := constants.MaxRelays

				routeMatrix := common.GenerateRandomRouteMatrix(numRelays)

				start := time.Now()

				routeMatrixData, err := routeMatrix.Write()
				if err != nil {
					panic(fmt.Sprintf("could not write route matrix: %v", err))
				}

				routeMatrixRead := common.RouteMatrix{}
				err = routeMatrixRead.Read(routeMatrixData)
				if err != nil {
					panic(fmt.Sprintf("could not read route matrix: %v", err))
				}

				fmt.Printf("iteration %d: read/write route matrix - %d relays, %d bytes (%dms)\n", iteration, numRelays, len(routeMatrixData), time.Since(start).Milliseconds())

				iteration++
			}
		}
	}()
}

func main() {

	RunRouteMatrixThread(context.Background())

	time.Sleep(time.Minute)
}
