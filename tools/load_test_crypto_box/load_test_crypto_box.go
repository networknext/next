package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/crypto"
	"github.com/networknext/accelerate/modules/envvar"
)

func RunCryptoBoxThread(ctx context.Context, numMessages int, messageSize int) {

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

						senderPublicKey, senderPrivateKey := crypto.Box_KeyPair()

						receiverPublicKey, receiverPrivateKey := crypto.Box_KeyPair()

						nonce := make([]byte, crypto.Box_NonceSize)
						common.RandomBytes(nonce)

						for i := 0; i < messagesPerSegment; i++ {

							data := make([]byte, messageSize)
							for i := 0; i < messageSize; i++ {
								data[i] = uint8(i)
							}

							encryptedData := make([]byte, messageSize+crypto.Box_MacSize)

							encryptedBytes := crypto.Box_Encrypt(senderPrivateKey[:], receiverPublicKey[:], nonce, encryptedData, len(data))

							if encryptedBytes != messageSize+crypto.Box_MacSize {
								panic("bad encrypted bytes")
							}

							err := crypto.Box_Decrypt(senderPublicKey[:], receiverPrivateKey[:], nonce, encryptedData, encryptedBytes)
							if err != nil {
								panic("failed to decrypt")
							}

						}

						waitGroup.Done()
					}()
				}

				waitGroup.Wait()

				fmt.Printf("iteration %d: encrypted and decrypted %d messages of size %d (%dms)\n", iteration, numMessages, messageSize, time.Since(start).Milliseconds())

				iteration++
			}
		}
	}()
}

func main() {

	numMessages := envvar.GetInt("NUM_MESSAGES", 100000)
	messageSize := envvar.GetInt("MESSAGE_SIZE", 1024)

	RunCryptoBoxThread(context.Background(), numMessages, messageSize)

	time.Sleep(time.Minute)
}
