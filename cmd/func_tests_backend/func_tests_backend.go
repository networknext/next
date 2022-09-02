/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	mathRand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/common/redis_pubsub"
	"github.com/networknext/backend/modules/core"
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

func test_redis_pubsub() {

	fmt.Printf("test_redis_pubsub\n")

	ctx := context.Background()

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

	producerThreadKiller := make([]int32, producerThreads)
	consumerThreadKiller := make([]int32, consumerThreads)

	for i := 0; i < producerThreads; i++ {
		go func(threadIndex int) {
			streamProducer := redis_pubsub.NewProducer(redis_pubsub.ProducerConfig{
				RedisHostname: "127.0.0.1:6379",
				RedisPassword: "",
				ChannelName:   "test-channel",
				BatchSize:     100,
				BatchBytes:    200000,
				TimeInterval:  time.Millisecond * 100,
			})

			connectErr := streamProducer.Connect(ctx)
			if connectErr != nil {
				producerWG.Done()
				return
			}

			tickRate := time.Duration(1000000000 / 1000)

			ticker := time.NewTicker(tickRate)
			//create messages batch
			messagesBatch := make([][]byte, 0)
			start := time.Now()
			for producerThreadKiller[threadIndex] == 0 {
				select {
				case <-ticker.C:
					var err error = nil

					messageSize := mathRand.Intn(95) + 5
					messageData := make([]byte, messageSize)

					rand.Read(messageData)

					messagesBatch = append(messagesBatch, messageData)

					messagesBatch, start, err = streamProducer.SendMessages(ctx, messagesBatch, start)

					if err != nil {
						core.Error("Failed to send message: %v", err)
						continue
					}

					threadBatchesSent[threadIndex] = streamProducer.NumBatchesSent()
					threadMessagesSent[threadIndex] = streamProducer.NumMessagesSent()

				case <-ctx.Done():
					atomic.StoreInt32(&producerThreadKiller[threadIndex], 1)
				}
			}

			// If the thread is killed externally, decrement the wg counter
			producerWG.Done()
		}(i)
	}

	for i := 0; i < consumerThreads; i++ {
		go func(threadIndex int) {
			streamConsumer := redis_pubsub.NewConsumer(redis_pubsub.ConsumerConfig{
				RedisHostname: "127.0.0.1:6379",
				RedisPassword: "",
				ChannelName:   "test-channel",
			})

			connectErr := streamConsumer.Connect(ctx)
			if connectErr != nil {
				consumerWG.Done()
				return
			}

			pubsubHandler := streamConsumer.RedisDB.Subscribe(ctx, streamConsumer.Config.ChannelName)

			messageChannel := pubsubHandler.Channel()

			go func(streamConsumer *redis_pubsub.Consumer) {
				for msg := range messageChannel { // This blocks infinitely which is bad for this test...
					if err := streamConsumer.ConsumeMessage(ctx, msg); err != nil {
						core.Error("error reading redis pubsub: %v", err)
					}
				}
			}(streamConsumer)

			for consumerThreadKiller[threadIndex] == 0 {
				// Loop until this is false - used for external thread control
			}

			threadBatchesReceived[threadIndex] = streamConsumer.NumBatchesReceived()
			threadMessagesReceived[threadIndex] = streamConsumer.NumMessageReceived()

			if err := pubsubHandler.Close(); err != nil {
				core.Error("Failed to shut down pubsub handler: %v", err)
			}

			consumerWG.Done()
		}(i)
	}

	time.Sleep(time.Second * 30)

	for i := 0; i < producerThreads; i++ {
		// Loop through producer threads and shut down the message creation loops
		atomic.StoreInt32(&producerThreadKiller[i], 1)
	}

	producerWG.Wait()

	time.Sleep(time.Second * 30)

	for i := 0; i < consumerThreads; i++ {
		// Loop through consumer threads and shut down processing loops
		atomic.StoreInt32(&consumerThreadKiller[i], 1)
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

	fmt.Printf("\nTotal batches sent: %d", totalNumBatchesSent)
	fmt.Printf("\nTotal batches received: %d", totalNumBatchesReceived)
	fmt.Printf("\nTotal messages sent: %d", totalMessagesSent)
	fmt.Printf("\nTotal messages received: %d\n", totalMessagesReceived)

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
		os.Exit(1)
	}
}

type test_function func()

func main() {
	allTests := []test_function{
		test_magic_backend,
		test_redis_pubsub,
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
