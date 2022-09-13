/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"sync"
	"sync/atomic"

	// "context"
	// "encoding/binary"
	// "sync"

	// "context"
	// "crypto/rand"
	"fmt"
	"io/ioutil"

	mathRand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	// "sync"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	// "github.com/networknext/backend/modules/common"
	// "github.com/networknext/backend/modules/core"
	// "github.com/go-redis/redis/v9"
	// "github.com/networknext/backend/modules/core"
	// "github.com/networknext/backend/modules/common"
)

func check_output(substring string, cmd *exec.Cmd, stdout bytes.Buffer, stderr bytes.Buffer) {
	if !strings.Contains(stdout.String(), substring) {
		fmt.Printf("\nerror: missing output '%s'\n\n", substring)
		fmt.Printf("--------------------------------------------------\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("--------------------------------------------------\n")
		if len(stderr.String()) > 0 {
			fmt.Printf("%s", stderr.String())
			fmt.Printf("--------------------------------------------------\n")
		}
		fmt.Printf("\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}
}

func test_magic_backend() {

	fmt.Printf("test_magic_backend\n")

	// run the magic backend and make sure it runs and does things it's expected to do

	cmd := exec.Command("./magic_backend")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = make([]string, 0)
	cmd.Env = append(cmd.Env, "ENV=local")
	cmd.Env = append(cmd.Env, "HTTP_PORT=40000")
	cmd.Env = append(cmd.Env, "NEXT_DEBUG_LOGS=1")
	cmd.Env = append(cmd.Env, "MAGIC_UPDATE_SECONDS=5")

	err := cmd.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic backend!\n\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("%s", stderr.String())
		os.Exit(1)
	}

	time.Sleep(10 * time.Second)

	check_output("magic_backend", cmd, stdout, stderr)
	check_output("starting http server on port 40000", cmd, stdout, stderr)

	// test the health check

	response, err := http.Get("http://127.0.0.1:40000/health")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: health check failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the version endpoint

	_, err = http.Get("http://127.0.0.1:40000/version")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: version endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the magic values endpoint

	response, err = http.Get("http://127.0.0.1:40000/magic")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: magic endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	magicData, error := ioutil.ReadAll(response.Body)
	if error != nil {
		fmt.Printf("error: failed to read magic response data\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	time.Sleep(time.Second)

	check_output("served magic values", cmd, stdout, stderr)

	// test that the magic values shuffle from upcoming -> current -> previous over time

	var upcomingMagic [8]byte
	var currentMagic [8]byte
	var previousMagic [8]byte

	copy(upcomingMagic[:], magicData[0:8])
	copy(currentMagic[:], magicData[8:16])
	copy(previousMagic[:], magicData[16:24])

	magicUpdates := 0

	for i := 0; i < 30; i++ {

		response, err = http.Get("http://127.0.0.1:40000/magic")
		if err != nil || response.StatusCode != 200 {
			fmt.Printf("error: magic endpoint failed\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		magicData, error := ioutil.ReadAll(response.Body)
		if error != nil {
			fmt.Printf("error: failed to read magic response data\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		if bytes.Compare(magicData[0:8], upcomingMagic[:]) != 0 {

			magicUpdates++

			if bytes.Compare(magicData[8:16], upcomingMagic[:]) != 0 {
				fmt.Printf("error: did not see upcoming magic shuffle to current magic\n")
				cmd.Process.Signal(syscall.SIGTERM)
				os.Exit(1)
			}

			if bytes.Compare(magicData[16:24], currentMagic[:]) != 0 {
				fmt.Printf("error: did not see current magic shuffle to previous magic\n")
				cmd.Process.Signal(syscall.SIGTERM)
				os.Exit(1)
			}

			copy(upcomingMagic[:], magicData[0:8])
			copy(currentMagic[:], magicData[8:16])
			copy(previousMagic[:], magicData[16:24])
		}

		time.Sleep(time.Second)

	}

	// we should see 5,6 or 7 magic updates (30 seconds with updates once every 5 seconds...)

	if magicUpdates != 5 && magicUpdates != 6 && magicUpdates != 7 {
		fmt.Printf("error: did not see magic values update every ~5 seconds (%d magic updates)", magicUpdates)
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// run a second magic backend. it should match the same magic values

	cmd2 := exec.Command("./magic_backend")

	var stdout2 bytes.Buffer
	var stderr2 bytes.Buffer
	cmd2.Stdout = &stdout2
	cmd2.Stderr = &stderr2

	cmd2.Env = make([]string, 0)
	cmd2.Env = append(cmd2.Env, "ENV=local")
	cmd2.Env = append(cmd2.Env, "HTTP_PORT=40001")
	cmd2.Env = append(cmd2.Env, "NEXT_DEBUG_LOGS=1")
	cmd2.Env = append(cmd2.Env, "MAGIC_UPDATE_SECONDS=5")

	err = cmd2.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic backend #2!\n\n")
		fmt.Printf("%s", stdout2.String())
		fmt.Printf("%s", stderr2.String())
		os.Exit(1)
	}

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {

		response1, err := http.Get("http://127.0.0.1:40000/magic")
		if err != nil || response1.StatusCode != 200 {
			fmt.Printf("error: magic endpoint failed (1)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		magicData1, error := ioutil.ReadAll(response.Body)
		if error != nil {
			fmt.Printf("error: failed to read magic response data (1)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		response2, err := http.Get("http://127.0.0.1:40001/magic")
		if err != nil || response2.StatusCode != 200 {
			fmt.Printf("error: magic endpoint failed (2)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		magicData2, error := ioutil.ReadAll(response.Body)
		if error != nil {
			fmt.Printf("error: failed to read magic response data (2)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		if bytes.Compare(magicData1, magicData2) != 0 && !(bytes.Compare(magicData1[0:16], magicData2[8:24]) == 0 || bytes.Compare(magicData2[0:16], magicData1[8:24]) == 0) {
			fmt.Printf("error: magic data mismatch between two magic backends\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		time.Sleep(time.Second)

	}

	// test that the service shuts down cleanly

	cmd.Process.Signal(os.Interrupt)
	cmd2.Process.Signal(os.Interrupt)

	cmd.Wait()
	cmd2.Wait()

	check_output("received shutdown signal", cmd, stdout, stderr)
	check_output("successfully shutdown", cmd, stdout, stderr)

	check_output("received shutdown signal", cmd2, stdout, stderr)
	check_output("successfully shutdown", cmd2, stdout, stderr)
}

func test_google_bigquery() {

	fmt.Printf("test_google_bigquery\n")

	action := "dev-bigquery-emulator"

	fmt.Printf("make %s\n", action)

	cmd := exec.Command("make", action)
	if cmd == nil {
		core.Error("could not run make!\n")
		os.Exit(1)
	}

	stdout_pipe, err := cmd.StdoutPipe()
	if err != nil {
		core.Error("could not create stdout pipe for make")
		os.Exit(1)
	}

	defer stdout_pipe.Close()
}

func test_google_pubsub() {

	fmt.Printf("test_google_pubsub\n")

	// setup the test topic and subscription in pubsub emulator

	os.Setenv("PUBSUB_EMULATOR_HOST", "127.0.0.1:9000")

	cancelContext, cancelFunc := context.WithCancel(context.Background())

	pubsubSetupClient, err := pubsub.NewClient(cancelContext, "local")
	if err != nil {
		core.Error("failed to create pubsub setup client: %v", err)
		os.Exit(1)
	}

	topic := "test"

	pubsubSetupClient.CreateTopic(cancelContext, topic)
	pubsubSetupClient.CreateSubscription(cancelContext, topic, pubsub.SubscriptionConfig{
		Topic: pubsubSetupClient.Topic(topic),
	})

	pubsubSetupClient.Close()

	// send a bunch of messages via multiple producers

	var waitGroup sync.WaitGroup

	const NumProducers = 10

	producers := [NumProducers]*common.GooglePubsubProducer{}

	for i := 0; i < NumProducers; i++ {

		producers[i], err = common.CreateGooglePubsubProducer(cancelContext, common.GooglePubsubConfig{
			ProjectId:          "local",
			Topic:              "test",
			MessageChannelSize: 10 * 1024,
			BatchSize:          100,
			BatchDuration:      time.Millisecond * 100,
		})

		if err != nil {
			core.Error("failed to create google pubsub producer: %v", err)
			os.Exit(1)
		}
	}

	waitGroup.Add(NumProducers)

	const NumMessagesPerProducer = 10000

	for i := 0; i < NumProducers; i++ {

		go func(producer *common.GooglePubsubProducer) {
			
			for j := 0; j < NumMessagesPerProducer; j++ {

				messageId := j
				messageSize := mathRand.Intn(96) + 4
				messageData := make([]byte, messageSize)

				binary.LittleEndian.PutUint32(messageData[:4], uint32(messageId))

				start := messageId % 256
				for k := 0; k < messageSize; k++ {
					messageData[k] = byte((start + k) % 256)
				}

				producer.MessageChannel <- messageData
			}

			waitGroup.Done()

		}(producers[i])
	}

	fmt.Printf("waiting for producers...\n")

	waitGroup.Wait()

	// receive a bunch of messages via the consumer

	const NumConsumers = 100

	consumers := [NumConsumers]*common.GooglePubsubConsumer{}

	for i := 0; i < NumConsumers; i++ {

		consumers[i], err = common.CreateGooglePubsubConsumer(cancelContext, common.GooglePubsubConfig{
			ProjectId:          "local",
			Topic:              "test",
			Subscription:       "test",
			MessageChannelSize: 10 * 1024,
		})

		if err != nil {
			core.Error("failed to create google pubsub consumer: %v", err)
			os.Exit(1)
		}
	}

	waitGroup.Add(NumConsumers)

	var numMessagesReceived uint64

	for i := 0; i < NumConsumers; i++ {

		go func(consumer *common.GooglePubsubConsumer) {
			
			for {
				select {

				case <-cancelContext.Done():
					break

				case msg := <-consumer.MessageChannel:
					messageId := binary.LittleEndian.Uint32(msg[:4])
					start := int(messageId % 256)
					for j := 0; j < len(msg); j++ {
						if msg[j] != byte((start+j)%256) {
							core.Error("message validation failed. expected %d, got %d", byte((start+j)%256), msg[j])
							os.Exit(1)
						}
					}
					atomic.AddUint64(&numMessagesReceived, 1)
				}
			}

			core.Debug("consumer done")

			waitGroup.Done()

		}(consumers[i])
	}

	// wait until we receive all messages, or up to 60 seconds...

	receivedAllMessages := false

	for i := 0; i < 60; i++ {
		messageCount := atomic.LoadUint64(&numMessagesReceived)
		expectedCount := uint64(NumProducers * NumMessagesPerProducer)
		core.Debug("received %d/%d messages", messageCount, expectedCount)
		if messageCount == expectedCount {
			core.Debug("received all")
			receivedAllMessages = true
			break
		}
		time.Sleep(time.Second)
	}

	if !receivedAllMessages {
		core.Error("did not receive all messages sent")
		os.Exit(1)
	}

	// clean shutdown the consumer threads

	core.Debug("cancelling context")

	cancelFunc()

	waitGroup.Wait()
}

func test_redis_pubsub() {

	fmt.Printf("test_redis_pubsub\n")

	parentContext := context.Background()

	producerThreads := 2
	consumerThreads := 10

	var producerWG sync.WaitGroup
	var consumerWG sync.WaitGroup

	producerWG.Add(producerThreads)
	consumerWG.Add(consumerThreads)

	threadMessagesSent := make([]int64, producerThreads)
	threadMessagesReceived := make([]int64, consumerThreads)

	threadBatchesSent := make([]int64, producerThreads)
	threadBatchesReceived := make([]int64, consumerThreads)

	producerThreadQuit := make([]context.CancelFunc, producerThreads)
	consumerThreadQuit := make([]context.CancelFunc, consumerThreads)

	for i := 0; i < producerThreads; i++ {
		ctx, cancel := context.WithCancel(parentContext)

		producerThreadQuit[i] = cancel

		go func(threadIndex int, ctx context.Context) {
			streamProducer, err := common.CreateRedisPubsubProducer(ctx, common.RedisPubsubConfig{
				RedisHostname:      "127.0.0.1:6379",
				RedisPassword:      "",
				PubsubChannelName:  "test-channel",
				MessageChannelSize: 10 * 1024,
				BatchSize:          100,
				BatchDuration:      time.Millisecond * 100,
			})

			if err != nil {
				producerWG.Done()
				return
			}

			tickRate := time.Duration(1000000000 / 1000)

			ticker := time.NewTicker(tickRate)

			numMessagesSent := 0

		producerLoop:
			for {
				select {
				case <-ticker.C:
					messageID := numMessagesSent
					messageSize := mathRand.Intn(96) + 4
					messageData := make([]byte, messageSize)

					binary.LittleEndian.PutUint32(messageData[:4], uint32(messageID))

					start := messageID % 256
					for i := 0; i < messageSize; i++ {
						messageData[i] = byte((start + i) % 256)
					}

					streamProducer.MessageChannel <- messageData

					numMessagesSent++

				case <-ctx.Done():
					break producerLoop
				}
			}

			threadBatchesSent[threadIndex] = int64(streamProducer.NumBatchesSent())
			threadMessagesSent[threadIndex] = int64(streamProducer.NumMessagesSent())

			// If the thread is killed externally, decrement the wg counter
			producerWG.Done()
		}(i, ctx)
	}

	for i := 0; i < consumerThreads; i++ {
		ctx, cancel := context.WithCancel(parentContext)

		consumerThreadQuit[i] = cancel

		go func(threadIndex int, ctx context.Context) {
			streamConsumer, err := common.CreateRedisPubsubConsumer(ctx, common.RedisPubsubConfig{
				RedisHostname:      "127.0.0.1:6379",
				RedisPassword:      "",
				PubsubChannelName:  "test-channel",
				MessageChannelSize: 10 * 1024,
			})

			if err != nil {
				consumerWG.Done()
				return
			}

		consumerLoop:
			for {
				select {
				case msg := <-streamConsumer.MessageChannel:
					messageID := binary.LittleEndian.Uint32(msg[:4])

					start := int(messageID % 256)
					for i := 0; i < len(msg); i++ {
						if msg[i] != byte((start+i)%256) {
							core.Error("Message validation failed!")
						}
					}
				case <-ctx.Done():
					break consumerLoop
				}
			}

			threadBatchesReceived[threadIndex] = int64(streamConsumer.NumBatchesReceived())
			threadMessagesReceived[threadIndex] = int64(streamConsumer.NumMessageReceived())

			consumerWG.Done()
		}(i, ctx)
	}

	time.Sleep(time.Second * 30)

	for i := 0; i < producerThreads; i++ {
		// Loop through producer threads and shut down the message creation loops
		producerThreadQuit[i]()
	}

	producerWG.Wait()

	time.Sleep(time.Second * 30)

	for i := 0; i < consumerThreads; i++ {
		// Loop through consumer threads and shut down processing loops
		consumerThreadQuit[i]()
	}

	consumerWG.Wait()

	totalMessagesSent := 0
	totalMessagesReceived := 0

	for _, numMessages := range threadMessagesSent {
		totalMessagesSent = totalMessagesSent + int(numMessages)
	}

	for _, numMessages := range threadMessagesReceived {
		totalMessagesReceived = totalMessagesReceived + int(numMessages)
	}

	totalNumBatchesSent := 0
	for i := 0; i < producerThreads; i++ {
		totalNumBatchesSent = totalNumBatchesSent + int(threadBatchesSent[i])
	}

	totalNumBatchesReceived := 0
	for i := 0; i < consumerThreads; i++ {
		totalNumBatchesReceived = totalNumBatchesReceived + int(threadBatchesReceived[i])
	}

	// Divide num batches received across all threads by num consumers to make sure everyone got the same num batches
	totalNumBatchesReceived = (totalNumBatchesReceived / consumerThreads)

	// Divide num messages received across all threads by num consumers to make sure everyone got the same num messages
	totalMessagesReceived = (totalMessagesReceived / consumerThreads)

	failed := false
	if totalNumBatchesReceived == totalNumBatchesSent {
		fmt.Printf("\nTest Results - Batches Sent: Passed\n")
	} else {
		fmt.Printf("\nTest Results - Batches Sent: Failed\n")
		failed = true
	}

	if totalMessagesReceived == totalMessagesSent {
		fmt.Println("Test Results - Messages Sent: Passed")
	} else {
		fmt.Println("Test Results - Messages Sent: Failed")
		failed = true
	}

	if failed {
		fmt.Printf("Total number of batches sent: %d\n", totalNumBatchesSent)
		fmt.Printf("Total number of messages sent: %d\n", totalMessagesSent)

		fmt.Printf("Total number of batches received: %d\n", totalNumBatchesReceived)
		fmt.Printf("Total number of messages received: %d\n", totalMessagesReceived)
		os.Exit(1)
	}
}

func test_redis_streams() {

	fmt.Printf("test_redis_streams\n")

	parentContext := context.Background()

	producerThreads := 2
	consumerThreads := 10

	var producerWG sync.WaitGroup
	var consumerWG sync.WaitGroup

	producerWG.Add(producerThreads)
	consumerWG.Add(consumerThreads)

	threadMessagesSent := make([]int64, producerThreads)
	threadMessagesReceived := make([]int64, consumerThreads)

	threadBatchesSent := make([]int64, producerThreads)
	threadBatchesReceived := make([]int64, consumerThreads)

	producerThreadQuit := make([]context.CancelFunc, producerThreads)
	consumerThreadQuit := make([]context.CancelFunc, consumerThreads)

	for i := 0; i < producerThreads; i++ {
		ctx, cancel := context.WithCancel(parentContext)

		producerThreadQuit[i] = cancel

		go func(threadIndex int, ctx context.Context) {
			streamProducer, err := common.CreateRedisStreamsProducer(ctx, common.RedisStreamsConfig{
				RedisHostname:      "127.0.0.1:6379",
				RedisPassword:      "",
				StreamName:         "test-stream",
				BatchSize:          100,
				BatchDuration:      time.Millisecond * 100,
				MessageChannelSize: 10 * 1024,
			})

			if err != nil {
				producerWG.Done()
				return
			}

			tickRate := time.Duration(1000000000 / 1000)

			ticker := time.NewTicker(tickRate)
			//create messages batch

			numMessagesSent := 0
		producerLoop:
			for {
				select {
				case <-ticker.C:
					messageID := numMessagesSent
					messageSize := mathRand.Intn(96) + 4
					messageData := make([]byte, messageSize)

					binary.LittleEndian.PutUint32(messageData[:4], uint32(messageID))

					start := messageID % 256
					for i := 0; i < messageSize; i++ {
						messageData[i] = byte((start + i) % 256)
					}

					streamProducer.MessageChannel <- messageData

					numMessagesSent++

				case <-ctx.Done():
					break producerLoop
				}
			}

			threadBatchesSent[threadIndex] = int64(streamProducer.NumBatchesSent())
			threadMessagesSent[threadIndex] = int64(streamProducer.NumMessagesSent())

			// If the thread is killed externally, decrement the wg counter
			producerWG.Done()
		}(i, ctx)
	}

	for i := 0; i < consumerThreads; i++ {
		ctx, cancel := context.WithCancel(parentContext)

		consumerThreadQuit[i] = cancel

		go func(threadIndex int, ctx context.Context) {
			streamConsumer, err := common.CreateRedisStreamsConsumer(ctx, common.RedisStreamsConfig{
				RedisHostname:      "127.0.0.1:6379",
				RedisPassword:      "",
				StreamName:         "test-stream",
				ConsumerGroup:      "test-group",
				BatchDuration:      time.Millisecond * 100,
				BatchSize:          10,
				MessageChannelSize: 10 * 1024,
			})

			if err != nil {
				consumerWG.Done()
				return
			}

		consumerLoop:
			for {
				select {
				case msg := <-streamConsumer.MessageChannel:
					messageID := binary.LittleEndian.Uint32(msg[:4])

					start := int(messageID % 256)
					for i := 0; i < len(msg); i++ {
						if msg[i] != byte((start+i)%256) {
							core.Error("Message validation failed!")
						}
					}
				case <-ctx.Done():
					break consumerLoop
				}
			}

			threadBatchesReceived[threadIndex] = int64(streamConsumer.NumBatchesReceived())
			threadMessagesReceived[threadIndex] = int64(streamConsumer.NumMessageReceived())

			consumerWG.Done()
		}(i, ctx)
	}

	time.Sleep(time.Second * 30)

	for i := 0; i < producerThreads; i++ {
		// Loop through producer threads and shut down the message creation loops
		producerThreadQuit[i]()
	}

	producerWG.Wait()

	time.Sleep(time.Second * 30)

	for i := 0; i < consumerThreads; i++ {
		// Loop through consumer threads and shut down processing loops
		consumerThreadQuit[i]()
	}

	consumerWG.Wait()

	totalMessagesSent := 0
	totalMessagesReceived := 0

	for _, numMessages := range threadMessagesSent {
		totalMessagesSent = totalMessagesSent + int(numMessages)
	}

	for _, numMessages := range threadMessagesReceived {
		totalMessagesReceived = totalMessagesReceived + int(numMessages)
	}

	totalNumBatchesSent := 0
	for i := 0; i < producerThreads; i++ {
		totalNumBatchesSent = totalNumBatchesSent + int(threadBatchesSent[i])
	}

	totalNumBatchesReceived := 0
	for i := 0; i < consumerThreads; i++ {
		totalNumBatchesReceived = totalNumBatchesReceived + int(threadBatchesReceived[i])
	}

	failed := false
	if totalNumBatchesReceived == totalNumBatchesSent {
		fmt.Println("Test Results - Batches Sent: Passed")
	} else {
		fmt.Println("Test Results - Batches Sent: Failed")
		failed = true
	}

	if totalMessagesReceived == totalMessagesSent {
		fmt.Println("Test Results - Messages Sent: Passed")
	} else {
		fmt.Println("Test Results - Messages Sent: Failed")
		failed = true
	}

	if failed {
		os.Exit(1)
	}
}

type test_function func()

func main() {
	allTests := []test_function{
		// test_magic_backend,
		// test_redis_pubsub,
		// test_redis_streams,
		test_google_pubsub,
		// test_google_bigquery,
	}

	var tests []test_function

	if len(os.Args) > 1 {
		funcName := os.Args[1]
		for _, test := range allTests {
			name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
			name = name[len("main."):]
			if funcName == name {
				tests = append(tests, test)
				break
			}
		}
		if len(tests) == 0 {
			panic(fmt.Sprintf("could not find any test: '%s'", funcName))
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	go func() {
		time.Sleep(time.Duration(len(tests)*120) * time.Second)
		panic("tests took too long!")
	}()

	for i := range tests {
		tests[i]()
	}
}
