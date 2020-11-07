package portalcruncher

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
)

type ErrReceiveMessage struct {
	err error
}

func (e *ErrReceiveMessage) Error() string {
	return fmt.Sprintf("error receiving message: %v", e.err)
}

type ErrUnknownMessage struct{}

func (*ErrUnknownMessage) Error() string {
	return "received an unknown message"
}

type ErrChannelFull struct{}

func (e *ErrChannelFull) Error() string {
	return "message channel full, dropping message"
}

type ErrUnmarshalMessage struct {
	err error
}

func (e *ErrUnmarshalMessage) Error() string {
	return fmt.Sprintf("could not unmarshal message: %v", e.err)
}

type PortalCruncher struct {
	subscriber pubsub.Subscriber
	metrics    *metrics.PortalCruncherMetrics
	btMetrics  *metrics.BigTableMetrics

	redisCountMessageChan chan *transport.SessionCountData
	redisDataMessageChan  chan *transport.SessionPortalData
	btDataMessageChan     chan *transport.SessionPortalData

	topSessions   storage.RedisClient
	sessionMap    storage.RedisClient
	sessionMeta   storage.RedisClient
	sessionSlices storage.RedisClient

	useBigtable bool
	btClient    *storage.BigTable
	btCfNames   []string
}

func NewPortalCruncher(
	ctx context.Context,
	subscriber pubsub.Subscriber,
	redisHostTopSessions string,
	redisHostSessionMap string,
	redisHostSessionMeta string,
	redisHostSessionSlices string,
	useBigtable bool,
	gcpProjectID string,
	btInstanceID string,
	btTableName string,
	btCfName string,
	btMaxAgeDays int,
	chanBufferSize int,
	logger log.Logger,
	metrics *metrics.PortalCruncherMetrics,
	btMetrics *metrics.BigTableMetrics,
) (*PortalCruncher, error) {
	topSessions, err := storage.NewRawRedisClient(redisHostTopSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostTopSessions, err)
	}

	sessionMap, err := storage.NewRawRedisClient(redisHostSessionMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostSessionMap, err)
	}

	sessionMeta, err := storage.NewRawRedisClient(redisHostSessionMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostSessionMeta, err)
	}

	sessionSlices, err := storage.NewRawRedisClient(redisHostSessionSlices)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostSessionSlices, err)
	}

	var btClient *storage.BigTable
	var btCfNames []string

	if useBigtable {
		btClient, btCfNames, err = SetupBigtable(ctx, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, logger)
		if err != nil {
			return nil, err
		}
	}

	return &PortalCruncher{
		subscriber:            subscriber,
		metrics:               metrics,
		btMetrics:             btMetrics,
		redisCountMessageChan: make(chan *transport.SessionCountData, chanBufferSize),
		redisDataMessageChan:  make(chan *transport.SessionPortalData, chanBufferSize),
		btDataMessageChan:     make(chan *transport.SessionPortalData, chanBufferSize),
		topSessions:           topSessions,
		sessionMap:            sessionMap,
		sessionMeta:           sessionMeta,
		sessionSlices:         sessionSlices,
		useBigtable:           useBigtable,
		btClient:              btClient,
		btCfNames:             btCfNames,
	}, nil
}

func (cruncher *PortalCruncher) Start(ctx context.Context, numRedisInsertGoroutines int, numBigtableInsertGoroutines int, redisPingDuration time.Duration, redisFlushDuration time.Duration, redisFlushCount int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// Start the receive goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := cruncher.ReceiveMessage(ctx); err != nil {
					switch err.(type) {
					case *ErrChannelFull: // We don't need to stop the portal cruncher if the channel is full
						continue
					default:
						errChan <- err
						return
					}
				}
			}
		}
	}()

	// Start the redis goroutines
	for i := 0; i < numRedisInsertGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each goroutine has its own buffer to avoid syncing
			redisPortalCountBuffer := make([]*transport.SessionCountData, 0)
			redisPortalDataBuffer := make([]*transport.SessionPortalData, 0)

			flushTime := time.Now()
			pingTime := time.Now()

			for {
				// Periodically ping the redis instances and error out if we don't get a pong
				if time.Since(pingTime) >= redisPingDuration {
					pingTime = time.Now()

					if err := cruncher.PingRedis(); err != nil {
						errChan <- err
						return
					}
				}

				select {
				// Buffer up some portal count entries and only insert into redis periodically to avoid overworking redis
				case portalCount := <-cruncher.redisCountMessageChan:
					redisPortalCountBuffer = append(redisPortalCountBuffer, portalCount)

					// If it's too early to insert into redis, early out
					if time.Since(flushTime) < redisFlushDuration && len(redisPortalCountBuffer) < redisFlushCount {
						continue
					}

					flushTime = time.Now()
					minutes := flushTime.Unix() / 60

					cruncher.insertCountDataIntoRedis(redisPortalCountBuffer, minutes)
					redisPortalCountBuffer = redisPortalCountBuffer[:0]

				// Buffer up some portal data entries and only insert into redis periodically to avoid overworking redis
				case portalData := <-cruncher.redisDataMessageChan:
					redisPortalDataBuffer = append(redisPortalDataBuffer, portalData)

					// If it's too early to insert into redis, early out
					if time.Since(flushTime) < redisFlushDuration && len(redisPortalDataBuffer) < redisFlushCount {
						continue
					}

					flushTime = time.Now()
					minutes := flushTime.Unix() / 60

					cruncher.insertPortalDataIntoRedis(redisPortalDataBuffer, minutes)
					redisPortalDataBuffer = redisPortalDataBuffer[:0]

				case <-ctx.Done():
					return
				}
			}
		}()
	}

	if cruncher.useBigtable {
		// Start the bigtable goroutines
		for i := 0; i < numBigtableInsertGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				btPortalDataBuffer := make([]*transport.SessionPortalData, 0)

				for {
					select {
					// Buffer up some portal data entries and only insert into redis periodically to avoid overworking redis
					case portalData := <-cruncher.btDataMessageChan:
						btPortalDataBuffer = append(btPortalDataBuffer, portalData)

						if err := cruncher.InsertIntoBigtable(ctx, btPortalDataBuffer); err != nil {
							errChan <- err
							return
						}

					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	// Wait until either there is an error or the context is done
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		// Let the goroutines finish up
		wg.Wait()
		return ctx.Err()
	}
}

func (cruncher *PortalCruncher) ReceiveMessage(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()

	case messageInfo := <-cruncher.subscriber.ReceiveMessage():
		cruncher.metrics.ReceivedMessageCount.Add(1)

		if messageInfo.Err != nil {
			return &ErrReceiveMessage{err: messageInfo.Err}
		}

		switch messageInfo.Topic {
		case pubsub.TopicPortalCruncherSessionCounts:
			var sessionCountData transport.SessionCountData
			if err := sessionCountData.UnmarshalBinary(messageInfo.Message); err != nil {
				return &ErrUnmarshalMessage{err: err}
			}

			select {
			case cruncher.redisCountMessageChan <- &sessionCountData:
			default:
				return &ErrChannelFull{}
			}

		case pubsub.TopicPortalCruncherSessionData:
			var sessionPortalData transport.SessionPortalData
			if err := sessionPortalData.UnmarshalBinary(messageInfo.Message); err != nil {
				return &ErrUnmarshalMessage{err: err}
			}

			select {
			case cruncher.redisDataMessageChan <- &sessionPortalData:
			default:
				return &ErrChannelFull{}
			}

			select {
			case cruncher.btDataMessageChan <- &sessionPortalData:
			default:
				return &ErrChannelFull{}
			}
		default:
			return &ErrUnknownMessage{}
		}

		return nil
	}
}

func (cruncher *PortalCruncher) insertCountDataIntoRedis(redisPortalCountBuffer []*transport.SessionCountData, minutes int64) {
	for i := range redisPortalCountBuffer {
		customerID := fmt.Sprintf("%016x", redisPortalCountBuffer[i].BuyerID)
		serverID := fmt.Sprintf("%016x", redisPortalCountBuffer[i].ServerID)
		numSessions := redisPortalCountBuffer[i].NumSessions

		// Remove the old count minute bucket from 2 minutes ago if it didn't expire
		cruncher.sessionMap.Command("DEL", "c-%s-%d", customerID, minutes-2)

		// Add the new session count
		cruncher.sessionMap.Command("HSET", "c-%s-%d %s %d", customerID, minutes, serverID, numSessions)
		cruncher.sessionMap.Command("EXPIRE", "c-%s-%d %d", customerID, minutes, 30)
	}
}

func (cruncher *PortalCruncher) insertPortalDataIntoRedis(redisPortalDataBuffer []*transport.SessionPortalData, minutes int64) {
	// Remove the old global top sessions minute bucket from 2 minutes ago if it didn't expire
	cruncher.topSessions.Command("DEL", "s-%d", minutes-2)

	// Update the current global top sessions minute bucket
	var format string
	args := make([]interface{}, 0)

	format += "s-%d"
	args = append(args, minutes)

	for i := range redisPortalDataBuffer {
		meta := redisPortalDataBuffer[i].Meta
		largeCustomer := redisPortalDataBuffer[i].LargeCustomer
		everOnNext := redisPortalDataBuffer[i].EverOnNext

		// For large customers, only insert the session if they have ever taken network next
		if largeCustomer && !meta.OnNetworkNext && !everOnNext {
			continue // Early out if we shouldn't add this session
		}

		sessionID := fmt.Sprintf("%016x", meta.ID)
		score := meta.DeltaRTT
		if score < 0 {
			score = 0
		}
		if !meta.OnNetworkNext {
			score = -meta.DirectRTT
		}

		format += " %.2f %s"
		args = append(args, score, sessionID)
	}

	cruncher.topSessions.Command("ZADD", format, args...)
	cruncher.topSessions.Command("EXPIRE", "s-%d %d", minutes, 30)

	for i := range redisPortalDataBuffer {
		meta := &redisPortalDataBuffer[i].Meta
		slice := &redisPortalDataBuffer[i].Slice
		point := &redisPortalDataBuffer[i].Point
		sessionID := fmt.Sprintf("%016x", meta.ID)
		customerID := fmt.Sprintf("%016x", meta.BuyerID)
		next := meta.OnNetworkNext
		score := meta.DeltaRTT
		if score < 0 {
			score = 0
		}
		if !next {
			score = -100000 + meta.DirectRTT
		}

		largeCustomer := redisPortalDataBuffer[i].LargeCustomer
		everOnNext := redisPortalDataBuffer[i].EverOnNext

		// For large customers, only insert the session if they have ever taken network next
		if largeCustomer && !meta.OnNetworkNext && !everOnNext {
			continue // Early out if we shouldn't add this session
		}

		// Update the map points for this minute bucket
		// Make sure to remove the session ID from the opposite bucket in case the session
		// has switched from direct -> next or next -> direct
		pointString := point.RedisString()
		if next {
			cruncher.sessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes-1, sessionID)
			cruncher.sessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes, sessionID)
			cruncher.sessionMap.Command("HSET", "n-%s-%d %s %s", customerID, minutes, sessionID, pointString)

		} else {
			cruncher.sessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes-1, sessionID)
			cruncher.sessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes, sessionID)
			cruncher.sessionMap.Command("HSET", "d-%s-%d %s %s", customerID, minutes, sessionID, pointString)
		}

		// Remove the old per-buyer top sessions minute bucket from 2 minutes ago if it didnt expire
		// and update the current per-buyer top sessions list
		cruncher.topSessions.Command("DEL", "sc-%s-%d", customerID, minutes-2)
		cruncher.topSessions.Command("ZADD", "sc-%s-%d %.2f %s", customerID, minutes, score, sessionID)
		cruncher.topSessions.Command("EXPIRE", "sc-%s-%d %d", customerID, minutes, 30)

		// Remove the old map points minute buckets from 2 minutes ago if it didn't expire
		cruncher.sessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes-2, sessionID)
		cruncher.sessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes-2, sessionID)

		// Expire map points
		cruncher.sessionMap.Command("EXPIRE", "n-%s-%d %d", customerID, minutes, 30)
		cruncher.sessionMap.Command("EXPIRE", "d-%s-%d %d", customerID, minutes, 30)

		// Update session meta
		cruncher.sessionMeta.Command("SET", "sm-%s \"%s\" EX %d", sessionID, meta.RedisString(), 120)

		// Update session slices
		cruncher.sessionSlices.Command("RPUSH", "ss-%s %s", sessionID, slice.RedisString())
		cruncher.sessionSlices.Command("EXPIRE", "ss-%s %d", sessionID, 120)
	}
}

func (cruncher *PortalCruncher) PingRedis() error {
	if err := cruncher.topSessions.Ping(); err != nil {
		return err
	}

	if err := cruncher.sessionMap.Ping(); err != nil {
		return err
	}

	if err := cruncher.sessionMeta.Ping(); err != nil {
		return err
	}

	if err := cruncher.sessionSlices.Ping(); err != nil {
		return err
	}

	return nil
}

func SetupBigtable(ctx context.Context,
	gcpProjectID string,
	btInstanceID string,
	btTableName string,
	btCfName string,
	btMaxAgeDays int,
	logger log.Logger) (*storage.BigTable, []string, error) {
	// Setup Bigtable
	_, btEmulatorOK := os.LookupEnv("BIGTABLE_EMULATOR_HOST")
	if btEmulatorOK {
		// Emulator is used for local testing
		// Requires that emulator has been started in another terminal to work as intended
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected Bigtable emulator")
	}

	if gcpProjectID == "" && !btEmulatorOK {
		return nil, nil, fmt.Errorf("No GCP Project ID found. Could not find $BIGTABLE_EMULATOR_HOST for local testing.")
	}

	// Put the column family names in a slice
	btCfNames := []string{btCfName}

	// Create a bigtable admin for setup
	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID, logger)
	if err != nil {
		return nil, nil, err
	}

	// Check if the table exists in the instance
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	if err != nil {
		return nil, nil, err
	}

	if !tableExists {
		// Create a table with the given name and column families
		if err = btAdmin.CreateTable(ctx, btTableName, btCfNames); err != nil {
			return nil, nil, err
		}

		// Set a garbage collection policy of maxAgeDays
		maxAge := time.Hour * time.Duration(24*btMaxAgeDays)
		if err = btAdmin.SetMaxAgePolicy(ctx, btTableName, btCfNames, maxAge); err != nil {
			return nil, nil, err
		}
	}

	// Close the admin client
	if err = btAdmin.Close(); err != nil {
		return nil, nil, err
	}

	// Create a standard client for writing to the table
	btClient, err := storage.NewBigTable(ctx, gcpProjectID, btInstanceID, logger)
	if err != nil {
		return nil, nil, err
	}

	return btClient, btCfNames, nil
}

func (cruncher *PortalCruncher) InsertIntoBigtable(ctx context.Context, btPortalDataBuffer []*transport.SessionPortalData) error {
	for j := range btPortalDataBuffer {
		meta := &btPortalDataBuffer[j].Meta
		slice := &btPortalDataBuffer[j].Slice

		// This seems redundant, try to figure out a better prefix to limit the number of keys
		// Key for session ID
		sessionRowKey := fmt.Sprintf("%016x", meta.ID)
		sliceRowKey := fmt.Sprintf("%016x#%v", meta.ID, slice.Timestamp)

		// Key for all buyer specific sessions
		buyerRowKey := fmt.Sprintf("%016x#%016x", meta.BuyerID, meta.ID)

		// Key for all user specific
		userRowKey := fmt.Sprintf("%016x#%016x", meta.UserHash, meta.ID)

		metaRowKeys := []string{sessionRowKey, buyerRowKey, userRowKey}
		sliceRowKeys := []string{sliceRowKey}

		// Have 2 columns under 1 column family
		// 1) Meta
		// 2) Slice

		// Create byte slices of the session data
		metaBinary, err := meta.MarshalBinary()
		if err != nil {
			return err
		}
		sliceBinary, err := slice.MarshalBinary()
		if err != nil {
			return err
		}

		// Insert session meta data into Bigtable
		if err := cruncher.btClient.InsertSessionMetaData(ctx, cruncher.btCfNames, metaBinary, metaRowKeys); err != nil {
			cruncher.btMetrics.WriteMetaFailureCount.Add(1)
			return err
		}
		cruncher.btMetrics.WriteMetaSuccessCount.Add(1)

		// Insert session slice data into Bigtable
		if err := cruncher.btClient.InsertSessionSliceData(ctx, cruncher.btCfNames, sliceBinary, sliceRowKeys); err != nil {
			cruncher.btMetrics.WriteSliceFailureCount.Add(1)
			return err
		}
		cruncher.btMetrics.WriteSliceSuccessCount.Add(1)
	}

	btPortalDataBuffer = btPortalDataBuffer[:0]
	return nil
}
