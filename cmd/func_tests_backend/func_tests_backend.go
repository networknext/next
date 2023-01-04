/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
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

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	db "github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/packets"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
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

	time.Sleep(20 * time.Second)

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

	dataset := "local"
	tableName := "test"

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	clientOptions := []option.ClientOption{
		option.WithEndpoint("http://127.0.0.1:9050"),
		option.WithoutAuthentication(),
	}

	bigquerySetupClient, err := bigquery.NewClient(cancelContext, googleProjectID, clientOptions...)
	if err != nil {
		core.Error("failed to create bigquery setup client: %v", err)
		os.Exit(1)
	}

	// Create local table under the local dataset
	tableReference := bigquerySetupClient.Dataset(dataset).Table(tableName)

	err = tableReference.Create(cancelContext, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{
				Type: bigquery.IntegerFieldType,
				Name: "timestamp",
			},
		},
	})

	if err != nil {
		core.Error("failed to setup bigquery table")
		os.Exit(1)
	}

	core.Debug("successfully set up bigquery emulator")

	bigquerySetupClient.Close()

	const NumPublishers = 1 // Emulator doesn't like multiple threads talking to the same table

	publishers := [NumPublishers]*common.GoogleBigQueryPublisher{}

	for i := 0; i < NumPublishers; i++ {
		publishers[i], err = common.CreateGoogleBigQueryPublisher(cancelContext, common.GoogleBigQueryConfig{
			ProjectId:          googleProjectID,
			Dataset:            dataset,
			TableName:          tableName,
			BatchSize:          100,
			BatchDuration:      time.Millisecond * 100,
			PublishChannelSize: 10 * 1024,
		})
		if err != nil {
			core.Error("failed to create google bigquery publisher: %v", err)
			os.Exit(1)
		}
	}

	messageChannel := make(chan bigquery.ValueSaver, 10*1024)

	const NumProducers = 10

	const NumEntriesPerProducer = 1000

	var waitGroup sync.WaitGroup

	waitGroup.Add(NumProducers)

	// Generate entries and throw them into a channel (fake pubsub functionality)
	for i := 0; i < NumProducers; i++ {
		go func() {
			for j := 0; j < NumEntriesPerProducer; j++ {
				var entry bigquery.ValueSaver = &common.TestEntry{
					Timestamp: uint32(time.Now().Nanosecond()),
				}

				messageChannel <- entry
			}

			waitGroup.Done()
		}()
	}

	waitGroup.Wait()

	waitGroup.Add(NumPublishers)

	// For each publisher, run a goroutine to publish entries off channel
	for i := 0; i < NumPublishers; i++ {

		go func(publisher *common.GoogleBigQueryPublisher) {

			for {
				select {

				case <-cancelContext.Done():
					waitGroup.Done()
					return

				case entry := <-messageChannel:
					publisher.PublishChannel <- entry
				}
			}

		}(publishers[i])
	}

	time.Sleep(time.Second * 30)

	// clean shutdown the publishing threads

	core.Debug("cancelling context")

	cancelFunc()

	core.Debug("waiting for publishers...")

	waitGroup.Wait()

	totalEntriesPublished := 0

	for i := 0; i < NumPublishers; i++ {
		totalEntriesPublished = totalEntriesPublished + int(publishers[i].NumEntriesPublished)
		if err := publishers[i].Close(); err != nil {
			core.Error("failed to shutdown publisher")
		}
	}

	if totalEntriesPublished != (NumProducers * NumEntriesPerProducer) {
		core.Error("did not receive all messages sent")
		os.Exit(1)
	}

	job, err := bigquerySetupClient.Query(fmt.Sprintf("SELECT * FROM %s.%s.%s", googleProjectID, dataset, tableName)).Run(context.Background())
	if err != nil {
		core.Error("failed to run query: %v", err)
		os.Exit(1)
	}

	rows, err := job.Read(context.Background())
	if err != nil {
		core.Error("failed to read results: %v", err)
		os.Exit(1)
	}

	if rows.TotalRows != uint64(totalEntriesPublished) {
		core.Error("not all entries were published")
		os.Exit(1)
	}

	core.Debug("done")
}

func test_google_pubsub() {

	fmt.Printf("test_google_pubsub\n")

	// setup the test topic and subscription in pubsub emulator

	os.Setenv("PUBSUB_EMULATOR_HOST", "127.0.0.1:9000")

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	pubsubSetupClient, err := pubsub.NewClient(cancelContext, googleProjectID)
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

	// create the consumers first so there is no race condition

	const NumConsumers = 100

	consumers := [NumConsumers]*common.GooglePubsubConsumer{}

	for i := 0; i < NumConsumers; i++ {

		consumers[i], err = common.CreateGooglePubsubConsumer(cancelContext, common.GooglePubsubConfig{
			ProjectId:          googleProjectID,
			Topic:              "test",
			Subscription:       "test",
			MessageChannelSize: 10 * 1024,
		})

		if err != nil {
			core.Error("failed to create google pubsub consumer: %v", err)
			os.Exit(1)
		}
	}

	// send a bunch of messages via multiple producers

	var waitGroup sync.WaitGroup

	const NumProducers = 10

	producers := [NumProducers]*common.GooglePubsubProducer{}

	for i := 0; i < NumProducers; i++ {

		producers[i], err = common.CreateGooglePubsubProducer(cancelContext, common.GooglePubsubConfig{
			ProjectId:          googleProjectID,
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
				messageSize := rand.Intn(96) + 4
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

	// receive a bunch of messages via consumers

	waitGroup.Add(NumConsumers)

	var numMessagesReceived uint64

	for i := 0; i < NumConsumers; i++ {

		go func(consumer *common.GooglePubsubConsumer) {

			for {
				select {

				case <-cancelContext.Done():
					core.Debug("consumer done")
					waitGroup.Done()
					return

				case pubsubMessage := <-consumer.MessageChannel:
					msg := pubsubMessage.Data
					messageId := binary.LittleEndian.Uint32(msg[:4])
					start := int(messageId % 256)
					for j := 0; j < len(msg); j++ {
						if msg[j] != byte((start+j)%256) {
							core.Error("message validation failed. expected %d, got %d", byte((start+j)%256), msg[j])
							os.Exit(1)
						}
					}
					atomic.AddUint64(&numMessagesReceived, 1)
					pubsubMessage.Ack()
				}
			}

		}(consumers[i])
	}

	// wait until we receive all messages, or up to 30 seconds...

	receivedAllMessages := false

	for i := 0; i < 30; i++ {
		messageCount := atomic.LoadUint64(&numMessagesReceived)
		expectedCount := uint64(NumProducers * NumMessagesPerProducer)
		core.Debug("received %d/%d messages", messageCount, expectedCount)
		if messageCount > expectedCount {
			core.Error("received too many messages!")
			os.Exit(1)
		}
		if i > 10 && messageCount == expectedCount {
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

	core.Debug("waiting for consumers...")

	waitGroup.Wait()

	core.Debug("done")
}

func test_redis_pubsub() {

	fmt.Printf("test_redis_pubsub\n")

	// create the consumers first, because otherwise the consumers won't receive messages

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	const NumConsumers = 10

	consumers := [NumConsumers]*common.RedisPubsubConsumer{}

	for i := 0; i < NumConsumers; i++ {

		var err error

		consumers[i], err = common.CreateRedisPubsubConsumer(cancelContext, common.RedisPubsubConfig{
			RedisHostname:      "127.0.0.1:6379",
			RedisPassword:      "",
			PubsubChannelName:  "test-channel",
			MessageChannelSize: 10 * 1024,
		})

		if err != nil {
			core.Error("failed to create redis pubsub consumer: %v", err)
			os.Exit(1)
		}
	}

	// send a bunch of messages via multiple producers

	var waitGroup sync.WaitGroup

	const NumProducers = 3

	producers := [NumProducers]*common.RedisPubsubProducer{}

	for i := 0; i < NumProducers; i++ {

		var err error

		producers[i], err = common.CreateRedisPubsubProducer(cancelContext, common.RedisPubsubConfig{
			RedisHostname:      "127.0.0.1:6379",
			RedisPassword:      "",
			PubsubChannelName:  "test-channel",
			MessageChannelSize: 10 * 1024,
			BatchSize:          1000,
			BatchDuration:      time.Second,
		})

		if err != nil {
			core.Error("failed to create redis pubsub producer: %v", err)
			os.Exit(1)
		}
	}

	waitGroup.Add(NumProducers)

	const NumMessagesPerProducer = 100000

	for i := 0; i < NumProducers; i++ {

		go func(producer *common.RedisPubsubProducer) {

			for j := 0; j < NumMessagesPerProducer; j++ {

				messageId := j
				messageSize := rand.Intn(96) + 4
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

	// receive a bunch of messages via consumers

	waitGroup.Add(NumConsumers)

	var numMessagesReceived uint64

	for i := 0; i < NumConsumers; i++ {

		go func(consumer *common.RedisPubsubConsumer) {

			for {
				select {

				case <-cancelContext.Done():
					core.Debug("consumer done")
					waitGroup.Done()
					return

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

		}(consumers[i])
	}

	// wait until we receive all messages, or up to 30 seconds...

	// IMPORTANT: In redis pubsub, each consumer gets a full set of messages produced by all producers.
	// this is different to streams and google pubsub where messages are load balanced across consumers

	receivedAllMessages := false

	for i := 0; i < 30; i++ {
		messageCount := atomic.LoadUint64(&numMessagesReceived)
		expectedCount := uint64(NumProducers * NumMessagesPerProducer * NumConsumers)
		core.Debug("received %d/%d messages", messageCount, expectedCount)
		if messageCount > expectedCount {
			core.Error("received too many messages! %d/%d", messageCount, expectedCount)
			os.Exit(1)
		}
		if i > 10 && messageCount == expectedCount {
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

	core.Debug("waiting for consumers...")

	waitGroup.Wait()

	core.Debug("done")
}

func test_redis_streams() {

	fmt.Printf("test_redis_streams\n")

	// create the consumers first, because otherwise the consumers won't receive messages

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	const NumConsumers = 10

	consumers := [NumConsumers]*common.RedisStreamsConsumer{}

	for i := 0; i < NumConsumers; i++ {

		var err error

		consumers[i], err = common.CreateRedisStreamsConsumer(cancelContext, common.RedisStreamsConfig{
			RedisHostname:      "127.0.0.1:6379",
			RedisPassword:      "",
			StreamName:         "test-stream",
			ConsumerGroup:      "test-group",
			BatchDuration:      time.Millisecond * 100,
			BatchSize:          10,
			MessageChannelSize: 10 * 1024,
		})

		if err != nil {
			core.Error("failed to create redis streams consumer: %v", err)
			os.Exit(1)
		}
	}

	// send a bunch of messages via multiple producers

	var waitGroup sync.WaitGroup

	const NumProducers = 3

	producers := [NumProducers]*common.RedisStreamsProducer{}

	for i := 0; i < NumProducers; i++ {

		var err error

		producers[i], err = common.CreateRedisStreamsProducer(cancelContext, common.RedisStreamsConfig{
			RedisHostname:      "127.0.0.1:6379",
			RedisPassword:      "",
			StreamName:         "test-stream",
			BatchSize:          100,
			BatchDuration:      time.Millisecond * 100,
			MessageChannelSize: 10 * 1024,
		})

		if err != nil {
			core.Error("failed to create redis streams producer: %v", err)
			os.Exit(1)
		}
	}

	waitGroup.Add(NumProducers)

	const NumMessagesPerProducer = 100000

	for i := 0; i < NumProducers; i++ {

		go func(producer *common.RedisStreamsProducer) {

			for j := 0; j < NumMessagesPerProducer; j++ {

				messageId := j
				messageSize := rand.Intn(96) + 4
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

	// receive a bunch of messages via consumers

	waitGroup.Add(NumConsumers)

	var numMessagesReceived uint64

	for i := 0; i < NumConsumers; i++ {

		go func(consumer *common.RedisStreamsConsumer) {

			for {
				select {

				case <-cancelContext.Done():
					core.Debug("consumer done")
					waitGroup.Done()
					return

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

		}(consumers[i])
	}

	// wait until we receive all messages, or up to 30 seconds...

	receivedAllMessages := false

	for i := 0; i < 30; i++ {
		messageCount := atomic.LoadUint64(&numMessagesReceived)
		expectedCount := uint64(NumProducers * NumMessagesPerProducer)
		core.Debug("received %d/%d messages", messageCount, expectedCount)
		if messageCount > expectedCount {
			core.Error("received too many messages!")
			os.Exit(1)
		}
		if i > 10 && messageCount == expectedCount {
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

	core.Debug("waiting for consumers...")

	waitGroup.Wait()

	core.Debug("done")
}

func test_redis_leader_store_migration() {

	fmt.Printf("test_redis_leader_store_migration\n")

	serviceName := "store_migration"

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	redisElector, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
		Timeout:       time.Second * 5,
	})
	if err != nil {
		core.Error("failed to setup redis elector 1")
		os.Exit(1)
	}

	electorMutex1 := sync.RWMutex{}

	dataStore := []common.DataStoreConfig{
		{
			Name: "store1",
			Data: []byte("1"),
		},
		{
			Name: "store2",
			Data: []byte("2"),
		},
		{
			Name: "store3",
			Data: []byte("3"),
		},
	}

	dataStore2 := []common.DataStoreConfig{
		{
			Name: "store1",
			Data: []byte("4"),
		},
		{
			Name: "store2",
			Data: []byte("5"),
		},
		{
			Name: "store3",
			Data: []byte("6"),
		},
	}

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(3)

	go func() {

		ticker := time.NewTicker(time.Second)

		iterationNum := 0

		for {
			select {
			case <-cancelContext.Done():
				waitGroup.Done()
				return
			case <-ticker.C:

				electorMutex1.Lock()
				isLeader := redisElector.IsLeader()
				electorMutex1.Unlock()

				if isLeader {
					iterationNum++
					if iterationNum%5 == 0 {
						continue
					}
				}

				electorMutex1.Lock()
				redisElector.Store(cancelContext, dataStore...)
				newDataStore := redisElector.Load(cancelContext)
				isLeader = redisElector.IsLeader()
				electorMutex1.Unlock()

				if cancelContext.Err() != nil {
					waitGroup.Done()
					return
				}

				if isLeader {
					for i, store := range dataStore {
						if string(store.Data) != string(newDataStore[i].Data) {
							core.Error("elector 1: data loaded was not from elector 1: %s != %s", string(store.Data), string(newDataStore[i].Data))
							os.Exit(1)
						}
					}
				} else {
					for i, store := range dataStore2 {
						if string(store.Data) != string(newDataStore[i].Data) {
							core.Error("elector 1: data loaded was not from elector 2: %s != %s", string(store.Data), string(newDataStore[i].Data))
							os.Exit(1)
						}
					}
				}
			}
		}
	}()

	time.Sleep(time.Second * 2)

	electorMutex2 := sync.RWMutex{}

	redisElector2, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
		Timeout:       time.Second * 5,
	})
	if err != nil {
		core.Error("failed to setup redis elector 2")
		os.Exit(1)
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		iterationNum := 0

		for {
			select {
			case <-cancelContext.Done():
				waitGroup.Done()
				return
			case <-ticker.C:

				electorMutex2.Lock()
				isLeader := redisElector2.IsLeader()
				electorMutex2.Unlock()

				if isLeader {
					iterationNum++
					if iterationNum%5 == 0 {
						continue
					}
				}

				electorMutex2.Lock()
				redisElector2.Store(cancelContext, dataStore2...)
				newDataStore2 := redisElector2.Load(cancelContext)
				isLeader = redisElector2.IsLeader()
				electorMutex2.Unlock()

				if cancelContext.Err() != nil {
					waitGroup.Done()
					return
				}

				if isLeader {
					for i, store := range dataStore2 {
						if string(store.Data) != string(newDataStore2[i].Data) {
							core.Error("elector 2: data loaded was not from elector 2: %s != %s", string(store.Data), string(newDataStore2[i].Data))
							os.Exit(1)
						}
					}
				} else {
					for i, store := range dataStore {
						if string(store.Data) != string(newDataStore2[i].Data) {
							core.Error("elector 2: data loaded was not from elector 1: %s != %s", string(store.Data), string(newDataStore2[i].Data))
							os.Exit(1)
						}
					}

				}
			}
		}
	}()

	time.Sleep(time.Second * 2)

	redisObserver, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
		Timeout:       time.Second * 5,
	})
	if err != nil {
		core.Error("failed to setup redis observer")
		os.Exit(1)
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-cancelContext.Done():
				waitGroup.Done()
				return
			case <-ticker.C:

				observedData := redisObserver.Load(cancelContext)

				if cancelContext.Err() != nil {
					waitGroup.Done()
					return
				}

				if redisObserver.IsLeader() {
					core.Error("observer should never be leader")
					os.Exit(1)
				}

				if len(observedData) == 0 {
					core.Error("failed to successfully observe data store")
					os.Exit(1)
				}

				electorMutex1.Lock()
				isLeader1 := redisElector.IsLeader()
				electorMutex1.Unlock()

				electorMutex2.Lock()
				isLeader2 := redisElector2.IsLeader()
				electorMutex2.Unlock()

				if !isLeader1 && !isLeader2 {
					core.Error("failed to have an elected leader")
					os.Exit(1)
				}
			}
		}
	}()

	time.Sleep(time.Second * 18)

	cancelFunc()

	waitGroup.Wait()

	core.Debug("done")
}

func test_redis_leader_store_no_flap() {

	fmt.Printf("test_redis_leader_store_no_flap\n")

	serviceName := "store_flap"

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	redielector, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
		Timeout:       time.Second * 5,
	})
	if err != nil {
		core.Error("failed to setup redis elector 1")
		os.Exit(1)
	}

	dataStore := []common.DataStoreConfig{
		{
			Name: "store1",
			Data: []byte("1"),
		},
		{
			Name: "store2",
			Data: []byte("2"),
		},
		{
			Name: "store3",
			Data: []byte("3"),
		},
	}

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-cancelContext.Done():
				waitGroup.Done()
				return
			case <-ticker.C:

				redielector.Store(cancelContext, dataStore...)
				redielector.Load(cancelContext)

				if cancelContext.Err() != nil {
					waitGroup.Done()
					return
				}

				if !redielector.IsLeader() {
					core.Error("elector 1 should always be leader")
					os.Exit(1)
				}
			}
		}
	}()

	time.Sleep(time.Second * 2)

	redielector2, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
		Timeout:       time.Second * 5,
	})
	if err != nil {
		core.Error("failed to setup redis elector 2")
		os.Exit(1)
	}

	dataStore2 := []common.DataStoreConfig{
		{
			Name: "store1",
			Data: []byte("4"),
		},
		{
			Name: "store2",
			Data: []byte("5"),
		},
		{
			Name: "store3",
			Data: []byte("6"),
		},
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-cancelContext.Done():
				waitGroup.Done()
				return
			case <-ticker.C:

				redielector2.Store(cancelContext, dataStore2...)
				redielector2.Load(cancelContext)

				if cancelContext.Err() != nil {
					waitGroup.Done()
					return
				}

				if redielector2.IsLeader() {
					core.Error("elector 2 should never be leader")
					os.Exit(1)
				}
			}
		}
	}()

	time.Sleep(time.Second * 20)

	cancelFunc()

	waitGroup.Wait()

	core.Debug("done")
}

func test_redis_leader_election_migration() {

	fmt.Printf("test_redis_leader_election_migration\n")

	serviceName := "migration"

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	redisElector, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
		Timeout:       time.Second * 5,
	})
	if err != nil {
		core.Error("failed to setup redis elector 1")
		os.Exit(1)
	}

	electorMutex := sync.RWMutex{}

	go func() {

		ticker := time.NewTicker(time.Second)

		iterationNum := 0

		for {
			select {
			case <-cancelContext.Done():
				return
			case <-ticker.C:

				electorMutex.Lock()
				isLeader := redisElector.IsLeader()
				electorMutex.Unlock()

				if isLeader {
					iterationNum++
					if iterationNum%5 == 0 {
						continue
					}
				}

				electorMutex.Lock()
				redisElector.Update(cancelContext)
				isLeader = redisElector.IsLeader()
				electorMutex.Unlock()
			}
		}
	}()

	time.Sleep(time.Second * 2)

	electorMutex2 := sync.RWMutex{}

	redisElector2, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
	})
	if err != nil {
		core.Error("failed to setup redis elector 2")
		os.Exit(1)
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		iterationNum := 0

		for {
			select {
			case <-cancelContext.Done():
				return
			case <-ticker.C:

				electorMutex2.Lock()
				isLeader := redisElector2.IsLeader()
				electorMutex2.Unlock()

				if isLeader {
					iterationNum++
					if iterationNum%5 == 0 {
						continue
					}
				}

				electorMutex2.Lock()
				redisElector2.Update(cancelContext)
				isLeader = redisElector2.IsLeader()
				electorMutex2.Unlock()
			}
		}
	}()

	time.Sleep(time.Second * 2)

	redisObserver, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
	})
	if err != nil {
		core.Error("failed to setup redis observer")
		os.Exit(1)
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-cancelContext.Done():
				return
			case <-ticker.C:

				if redisObserver.IsLeader() {
					core.Error("observer should never be leader")
					os.Exit(1)
				}

				electorMutex.Lock()
				isLeader1 := redisElector.IsLeader()
				electorMutex.Unlock()

				electorMutex2.Lock()
				isLeader2 := redisElector2.IsLeader()
				electorMutex2.Unlock()

				if cancelContext.Err() != nil {
					return
				}

				if !isLeader1 && !isLeader2 {
					core.Error("failed to have an elected leader")
					os.Exit(1)
				}
			}
		}
	}()

	time.Sleep(time.Second * 18)

	cancelFunc()

	// Make sure all keys expire
	time.Sleep(time.Second * 7)

	core.Debug("done")
}

func test_redis_leader_election_no_flap() {

	fmt.Printf("test_redis_leader_election_no_flap\n")

	serviceName := "falp"

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(30*time.Second))

	redielector, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
		Timeout:       time.Second * 5,
	})
	if err != nil {
		core.Error("failed to setup redis elector 1")
		os.Exit(1)
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-cancelContext.Done():
				return
			case <-ticker.C:

				redielector.Update(cancelContext)

				if cancelContext.Err() != nil {
					return
				}

				if !redielector.IsLeader() {
					core.Error("elector 1 should always be leader")
					os.Exit(1)
				}
			}
		}
	}()

	time.Sleep(time.Second * 2)

	redielector2, err := common.CreateRedisLeaderElection(cancelContext, common.RedisLeaderElectionConfig{
		RedisHostname: "127.0.0.1:6379",
		RedisPassword: "",
		ServiceName:   serviceName,
	})
	if err != nil {
		core.Error("failed to setup redis elector 2")
		os.Exit(1)
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-cancelContext.Done():
				return
			case <-ticker.C:

				redielector2.Update(cancelContext)

				if cancelContext.Err() != nil {
					return
				}

				if redielector2.IsLeader() {
					core.Error("elector 2 should never be leader")
					os.Exit(1)
				}
			}
		}
	}()

	time.Sleep(time.Second * 20)

	cancelFunc()

	// Make sure keys expire
	time.Sleep(time.Second * 7)

	core.Debug("done")
}

func test_cost_matrix_read_write() {

	fmt.Printf("test_cost_matrix_read_write\n")

	startTime := time.Now()

	for {

		if time.Since(startTime) > 60*time.Second {
			break
		}

		writeMessage := common.GenerateRandomCostMatrix()

		readMessage := common.CostMatrix{}

		const BufferSize = 100 * 1024

		buffer, err := writeMessage.Write(BufferSize)
		if err != nil {
			panic(err)
		}

		err = readMessage.Read(buffer)

		if !reflect.DeepEqual(writeMessage, readMessage) {
			panic("cost matrix read write failure")
		}
	}
}

func test_route_matrix_read_write() {

	fmt.Printf("test_route_matrix_read_write\n")

	startTime := time.Now()

	for {

		if time.Since(startTime) > 60*time.Second {
			break
		}

		writeMessage := common.GenerateRandomRouteMatrix()

		readMessage := common.RouteMatrix{}

		const BufferSize = 1024 * 1024

		buffer, err := writeMessage.Write(BufferSize)
		if err != nil {
			panic(err)
		}

		err = readMessage.Read(buffer)

		if !reflect.DeepEqual(writeMessage, readMessage) {
			panic("route matrix read write failure")
		}
	}
}

func test_session_data_serialize() {

	fmt.Printf("test_session_data_serialize\n")

	startTime := time.Now()

	for {

		if time.Since(startTime) > 60*time.Second {
			break
		}

		writePacket := packets.GenerateRandomSessionData()

		readPacket := packets.SDK5_SessionData{}

		const BufferSize = 10 * 1024

		buffer := [BufferSize]byte{}

		writeStream := encoding.CreateWriteStream(buffer[:])

		err := writePacket.Serialize(writeStream)
		if err != nil {
			panic(err)
		}
		writeStream.Flush()
		packetBytes := writeStream.GetBytesProcessed()

		readStream := encoding.CreateReadStream(buffer[:packetBytes])
		err = readPacket.Serialize(readStream)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(writePacket, readPacket) {
			panic("session data serialize failure")
		}
	}
}

func test_relay_manager() {

	fmt.Printf("test_relay_manager\n")

	relayManager := common.CreateRelayManager()

	ctx, contextCancelFunc := context.WithCancel(context.Background())

	// setup a lot of relays

	const NumRelays = 1500

	relayNames := make([]string, NumRelays)
	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)

	for i := range relayIds {
		relayIds[i] = common.RelayId(relayNames[i])
		relayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
	}

	// get costs once per-second

	go func() {
		counter := 0
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				if counter < 30 {
					panic("get costs deadlocked!")
				}
				return
			case <-ticker.C:
				const MaxRTT = 255
				const MaxJitter = 100
				const MaxPacketLoss = 1
				currentTime := time.Now().Unix()
				costs := relayManager.GetCosts(currentTime, relayIds, MaxRTT, MaxJitter, MaxPacketLoss, false)
				_ = costs
				fmt.Printf("costs %d\n", counter)
				counter++
			}
		}
	}()

	// really slam in the relay updates once per-second, randomly for 1000 relays

	numSamples := NumRelays
	sampleRelayId := make([]uint64, numSamples)
	sampleRTT := make([]float32, numSamples)
	sampleJitter := make([]float32, numSamples)
	samplePacketLoss := make([]float32, numSamples)

	for i := 0; i < numSamples; i++ {
		sampleRelayId[i] = uint64(i)
		sampleRTT[i] = 10
		sampleJitter[i] = 5
		samplePacketLoss[i] = 0
	}

	for i := 0; i < NumRelays; i++ {

		go func(index int) {

			ticker := time.NewTicker(time.Second)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					currentTime := time.Now().Unix()
					// fmt.Printf("relay update %d\n", index)
					relayManager.ProcessRelayUpdate(currentTime, relayIds[index], relayNames[index], relayAddresses[index], 0, "test", false, numSamples, sampleRelayId, sampleRTT, sampleJitter, samplePacketLoss)
				}
			}

		}(i)

	}

	time.Sleep(60 * time.Second)

	contextCancelFunc()
}

func test_optimize() {

	fmt.Printf("test_optimize\n")

	relayManager := common.CreateRelayManager()

	ctx, contextCancelFunc := context.WithCancel(context.Background())

	// setup a lot of relays

	const NumRelays = 100

	relayNames := make([]string, NumRelays)
	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayLatitudes := make([]float32, NumRelays)
	relayLongitudes := make([]float32, NumRelays)
	relayDatacenterIds := make([]uint64, NumRelays)
	destRelays := make([]bool, NumRelays)

	for i := range relayIds {
		relayNames[i] = fmt.Sprintf("relay%d", i)
		relayIds[i] = common.RelayId(relayNames[i])
		relayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		relayLatitudes[i] = float32(common.RandomInt(-90, +90))
		relayLongitudes[i] = float32(common.RandomInt(-90, +90))
		relayDatacenterIds[i] = uint64(common.RandomInt(0, 5))
		destRelays[i] = true
	}

	// get costs once per-second

	go func() {
		counter := 0
		ticker := time.NewTicker(time.Second)
		for {
			select {

			case <-ctx.Done():
				if counter < 30 {
					panic("optimize deadlocked!")
				}
				return

			case <-ticker.C:

				const MaxRTT = 255
				const MaxJitter = 100
				const MaxPacketLoss = 1

				currentTime := time.Now().Unix()

				costs := relayManager.GetCosts(currentTime, relayIds, MaxRTT, MaxJitter, MaxPacketLoss, false)

				costMatrix := &common.CostMatrix{
					Version:            common.CostMatrixVersion_Write,
					RelayIds:           relayIds,
					RelayAddresses:     relayAddresses,
					RelayNames:         relayNames,
					RelayLatitudes:     relayLatitudes,
					RelayLongitudes:    relayLongitudes,
					RelayDatacenterIds: relayDatacenterIds,
					DestRelays:         destRelays,
					Costs:              costs,
				}

				costMatrixData, err := costMatrix.Write(10 * 1024 * 1024)
				if err != nil {
					panic("could not write cost matrix")
				}
				_ = costMatrixData

				numCPUs := runtime.NumCPU()
				numSegments := NumRelays
				if numCPUs < NumRelays {
					numSegments = NumRelays / 5
					if numSegments == 0 {
						numSegments = 1
					}
				}

				costThreshold := int32(1)

				binFileData := make([]byte, 100*1024)

				routeMatrix := &common.RouteMatrix{
					CreatedAt:          uint64(time.Now().Unix()),
					Version:            common.RouteMatrixVersion_Write,
					RelayIds:           relayIds,
					RelayAddresses:     relayAddresses,
					RelayNames:         relayNames,
					RelayLatitudes:     relayLatitudes,
					RelayLongitudes:    relayLongitudes,
					RelayDatacenterIds: relayDatacenterIds,
					DestRelays:         destRelays,
					RouteEntries:       core.Optimize2(NumRelays, numSegments, costs, costThreshold, relayDatacenterIds, destRelays),
					BinFileBytes:       int32(len(binFileData)),
					BinFileData:        binFileData,
				}

				routeMatrixData, err := routeMatrix.Write(100 * 1024 * 1024)
				if err != nil {
					panic("could not write route matrix")
					continue
				}
				_ = routeMatrixData

				fmt.Printf("optimize %d\n", counter)

				counter++
			}
		}
	}()

	// relay updates once per-second for each relay

	numSamples := NumRelays
	sampleRelayId := make([]uint64, numSamples)
	sampleRTT := make([]float32, numSamples)
	sampleJitter := make([]float32, numSamples)
	samplePacketLoss := make([]float32, numSamples)

	for i := 0; i < numSamples; i++ {
		sampleRelayId[i] = uint64(i)
		sampleRTT[i] = float32(common.RandomInt(1, 100))
		sampleJitter[i] = float32(common.RandomInt(0, 50))
		samplePacketLoss[i] = float32(common.RandomInt(0, 2))
	}

	for i := 0; i < NumRelays; i++ {

		go func(index int) {

			ticker := time.NewTicker(time.Second)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					currentTime := time.Now().Unix()
					relayManager.ProcessRelayUpdate(currentTime, relayIds[index], relayNames[index], relayAddresses[index], 0, "test", false, numSamples, sampleRelayId, sampleRTT, sampleJitter, samplePacketLoss)
				}
			}

		}(i)

	}

	time.Sleep(60 * time.Second)

	contextCancelFunc()
}

const (
	magicBackendBin = "./magic_backend"
	relayGatewayBin = "./relay_gateway"
	relayBackendBin = "./relay_backend"
)

func test_relay_backend() {

	fmt.Printf("test_relay_backend\n")

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(60*time.Second))

	// setup datacenters

	const NumDatacenters = 10

	datacenterIds := make([]uint64, NumDatacenters)
	datacenterNames := make([]string, NumDatacenters)
	datacenterLatitudes := make([]float32, NumDatacenters)
	datacenterLongitudes := make([]float32, NumDatacenters)

	for i := 0; i < NumDatacenters; i++ {
		datacenterIds[i] = uint64(i)
		datacenterNames[i] = fmt.Sprintf("datacenter%d", i)
		datacenterLatitudes[i] = float32(common.RandomInt(-90, +90))
		datacenterLongitudes[i] = float32(common.RandomInt(-90, +90))
	}

	// setup relays

	const NumRelays = 100

	relayIds := make([]uint64, NumRelays)
	relayNames := make([]string, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayDatacenterIds := make([]uint64, NumRelays)
	destRelays := make([]bool, NumRelays)

	for i := range relayIds {
		relayNames[i] = fmt.Sprintf("relay%d", i)
		relayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		relayIds[i] = common.RelayId(relayAddresses[i].String())
		relayDatacenterIds[i] = uint64(common.RandomInt(0, NumDatacenters-1))
		destRelays[i] = true
	}

	// setup a database containing the relays

	database := db.CreateDatabase()

	database.CreationTime = time.Now().String()
	database.Creator = "test"

	for i := 0; i < NumRelays; i++ {

		relay := db.Relay{}

		relay.ID = relayIds[i]
		relay.Name = relayNames[i]
		relay.Addr = relayAddresses[i]
		relay.Version = "test"
		relay.Datacenter.ID = relayDatacenterIds[i]
		relay.Datacenter.Name = datacenterNames[relay.Datacenter.ID]
		relay.Datacenter.Latitude = datacenterLatitudes[relay.Datacenter.ID]
		relay.Datacenter.Longitude = datacenterLongitudes[relay.Datacenter.ID]

		database.Relays = append(database.Relays, relay)

		database.RelayMap[relay.ID] = relay
	}

	// write the database out to a temporary file

	file, err := ioutil.TempFile(".", "temp-database-")
	if err != nil {
		panic("could not create temporary database file")
	}

	databaseFilename := file.Name()

	defer os.Remove(databaseFilename)

	fmt.Println(databaseFilename)

	database.Save(databaseFilename)

	// start the magic backend

	magic_backend_cmd := exec.Command(magicBackendBin)
	if magic_backend_cmd == nil {
		panic("could not create magic backend!\n")
	}

	magic_backend_cmd.Env = os.Environ()
	magic_backend_cmd.Env = append(magic_backend_cmd.Env, "HTTP_PORT=41007")

	var magic_backend_output bytes.Buffer
	magic_backend_cmd.Stdout = &magic_backend_output
	magic_backend_cmd.Stderr = &magic_backend_output
	magic_backend_cmd.Start()

	// run the relay gateway, such that it loads the temporary database file

	relay_gateway_cmd := exec.Command(relayGatewayBin)
	if relay_gateway_cmd == nil {
		panic("could not create relay gateway!\n")
	}

	relay_gateway_cmd.Env = os.Environ()
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, fmt.Sprintf("DATABASE_PATH=%s", databaseFilename))
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, "OVERLAY_PATH=nopenopenope")
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, "HTTP_PORT=30000")

	var relay_gateway_output bytes.Buffer
	relay_gateway_cmd.Stdout = &relay_gateway_output
	relay_gateway_cmd.Stderr = &relay_gateway_output
	relay_gateway_cmd.Start()

	// run the relay backend, such that it loads the temporary database file

	relay_backend_cmd := exec.Command(relayBackendBin)
	if relay_backend_cmd == nil {
		panic("could not create relay backend!\n")
	}

	relay_backend_cmd.Env = os.Environ()
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, fmt.Sprintf("DATABASE_PATH=%s", databaseFilename))
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, "OVERLAY_PATH=nopenopenope")
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, "HTTP_PORT=30001")
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, "READY_DELAY=1s")
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, "DISABLE_GOOGLE_PUBSUB=1")

	var relay_backend_output bytes.Buffer
	relay_backend_cmd.Stdout = &relay_backend_output
	relay_backend_cmd.Stderr = &relay_backend_output
	relay_backend_cmd.Start()

	// wait until the relay gateway and relay backend are ready

	for {
		if strings.Contains(relay_gateway_output.String(), "starting http server on port 30000") &&
			strings.Contains(relay_backend_output.String(), "starting http server on port 30001") {
			break
		}

		time.Sleep(time.Second)
	}

	// hammer the relay backend with relay updates

	var waitGroup sync.WaitGroup

	waitGroup.Add(NumRelays)

	var errorCount uint64

	for i := 0; i < NumRelays; i++ {

		go func(index int) {

			// create http client

			transport := &http.Transport{
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
			}

			client := &http.Client{Transport: transport}

			ticker := time.NewTicker(1 * time.Second)

			for {
				select {

				case <-cancelContext.Done():
					waitGroup.Done()
					return

				case <-ticker.C:

					requestPacket := packets.RelayUpdateRequestPacket{}

					requestPacket.Version = packets.VersionNumberRelayUpdateRequest
					requestPacket.Address = relayAddresses[index]
					requestPacket.Token = make([]byte, packets.RelayTokenSize)
					requestPacket.NumSamples = NumRelays

					for i := 0; i < NumRelays; i++ {
						requestPacket.SampleRelayId[i] = relayIds[i]
						requestPacket.SampleRTT[i] = float32(common.RandomInt(1, 100))
						requestPacket.SampleJitter[i] = float32(common.RandomInt(1, 50))
						requestPacket.SamplePacketLoss[i] = float32(common.RandomInt(0, 2))
					}

					body := requestPacket.Write(make([]byte, 100*1024))

					request, err := http.NewRequest("POST", "http://127.0.0.1:30000/relay_update", bytes.NewBuffer(body))
					if err != nil {
						fmt.Printf("error creating http request: %v\n", err)
						atomic.AddUint64(&errorCount, 1)
						break
					}

					request.Header.Set("Content-Type", "application/octet-stream")

					response, err := client.Do(request)
					if err != nil {
						fmt.Printf("error running http request: %v\n", err)
						atomic.AddUint64(&errorCount, 1)
						break
					}

					if response.StatusCode != 200 {
						fmt.Printf("bad http response %d\n", response.StatusCode)
						atomic.AddUint64(&errorCount, 1)
					}

					response.Body.Close()
				}
			}

		}(i)
	}

	// run a goroutine to pull down the route matrix once per-second from the relay backend

	waitGroup.Add(1)

	routeMatrixCounter := 0

	go func() {

		transport := &http.Transport{
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
		}

		client := &http.Client{Transport: transport}

		// wait until the relay backend is ready

		for {
			response, err := client.Get("http://127.0.0.1:30001/health")
			if err == nil && response.StatusCode == 200 {
				break
			}
		}

		// request route matrix once per-second

		ticker := time.NewTicker(1 * time.Second)

		for {
			select {

			case <-cancelContext.Done():
				waitGroup.Done()
				return

			case <-ticker.C:

				response, err := client.Get("http://127.0.0.1:30001/route_matrix")
				if err != nil {
					core.Error("failed to http get route matrix: %v", err)
					atomic.AddUint64(&errorCount, 1)
					break
				}

				buffer, err := ioutil.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					atomic.AddUint64(&errorCount, 1)
					break
				}

				response.Body.Close()

				routeMatrix := common.RouteMatrix{}

				err = routeMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					atomic.AddUint64(&errorCount, 1)
					break
				}

				if len(routeMatrix.RelayIds) != NumRelays {
					core.Error("wrong num relays in route matrix: %d", len(routeMatrix.RelayIds))
					atomic.AddUint64(&errorCount, 1)
					break
				}

				fmt.Printf("route matrix %d\n", routeMatrixCounter)

				routeMatrixCounter++
			}
		}
	}()

	// wait for 60 seconds

	time.Sleep(60 * time.Second)

	// wait for all goroutines to finish

	cancelFunc()

	waitGroup.Wait()

	// print output from services

	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("%s", magic_backend_output.String())
	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("%s", relay_gateway_output.String())
	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("%s", relay_backend_output.String())
	fmt.Printf("-----------------------------------------------\n")

	magic_backend_cmd.Process.Signal(os.Interrupt)
	magic_backend_cmd.Wait()

	relay_gateway_cmd.Process.Signal(os.Interrupt)
	relay_gateway_cmd.Wait()

	relay_backend_cmd.Process.Signal(os.Interrupt)
	relay_backend_cmd.Wait()

	if errorCount != 0 {
		panic("error count is not zero")
	}

	if routeMatrixCounter < 30 {
		panic("not enough valid route matrices")
	}
}

type test_function func()

var googleProjectID string

func main() {

	googleProjectID = "local"

	allTests := []test_function{
		test_magic_backend,
		test_redis_pubsub,
		test_redis_streams,
		test_redis_leader_store_no_flap,
		test_redis_leader_store_migration,
		test_redis_leader_election_no_flap,
		test_redis_leader_election_migration,
		test_google_pubsub,
		test_google_bigquery,
		test_cost_matrix_read_write,
		test_route_matrix_read_write,
		test_session_data_serialize,
		test_relay_manager,
		test_optimize,
		test_relay_backend,
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

	fmt.Printf("\n")

	for i := range tests {
		tests[i]()
	}
}
