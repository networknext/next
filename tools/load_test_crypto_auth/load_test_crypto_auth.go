package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/envvar"
)

func RunCryptoAuthThread(ctx context.Context, numMessages int, messageSize int) {

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				start := time.Now()

				numThreads := runtime.NumCPU()
				if numThreads > numMessages {
					numThreads = numMessages
				}

				numSegments := numMessages / numThreads

				messagesPerSegment := numMessages / numSegments

				waitGroup := sync.WaitGroup{}
				waitGroup.Add(numSegments)

				for segment := 0; segment < numSegments; segment++ {

					go func() {

						key := crypto.Auth_Key()

						signature := make([]byte, crypto.Auth_SignatureSize)

						for i := 0; i < messagesPerSegment; i++ {

							data := make([]byte, messageSize)
							for j := 0; j < messageSize; j++ {
								data[j] = uint8(j)
							}

							signature := crypto.Auth_Sign(data, key, signature)

							if !crypto.Auth_Verify(data, key, signature) {
								panic("auth signature did not verify")
							}

						}

						waitGroup.Done()
					}()
				}

				waitGroup.Wait()

				fmt.Printf("iteration %d: auth signed and verified %d messages of size %d (%dms)\n", iteration, numMessages, messageSize, time.Since(start).Milliseconds())

				iteration++
			}
		}
	}()
}

func main() {

	numMessages := envvar.GetInt("NUM_MESSAGES", 100000)
	messageSize := envvar.GetInt("MESSAGE_SIZE", 1024)

	RunCryptoSignThread(context.Background(), numMessages, messageSize)

	time.Sleep(time.Minute)
}
