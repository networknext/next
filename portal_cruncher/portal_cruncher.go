package portalcruncher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/metrics"
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

	// portalCountMessageChan chan *transport.SessionCountData
	redisDataMessageChan 	chan *transport.SessionPortalData
	btDataMessageChan 		chan *transport.SessionPortalData

	topSessions   storage.RedisClient
	sessionMap    storage.RedisClient
	sessionMeta   storage.RedisClient
	sessionSlices storage.RedisClient

	// portalCountBuffer []*transport.SessionCountData
	redisPortalDataBuffer 	[]*transport.SessionPortalData
	btPortalDataBuffer		[]*transport.SessionPortalData

	useBigtable 	bool
	btClient 		storage.BigTable
	btCfNames		[]string

	redisFlushCount int
	flushTime       time.Time
	pingTime        time.Time
}

func NewPortalCruncher(
	ctx context.Context,
	subscriber pubsub.Subscriber,
	redisHostTopSessions string,
	redisHostSessionMap string,
	redisHostSessionMeta string,
	redisHostSessionSlices string,
	gcpProjectID string,
	useBigtable bool,
	chanBufferSize int,
	redisFlushCount int,
	logger log.Logger,
	metrics *metrics.PortalCruncherMetrics,
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
		btClient, btCfNames, err := SetupBigtable(ctx, gcpProjectID, logger)
		if err != nil {
			return nil, err
		}		
	}

	return &PortalCruncher{
		subscriber: subscriber,
		metrics:    metrics,
		// portalCountMessageChan: make(chan *transport.SessionCountData, chanBufferSize),
		redisPortalDataMessageChan: make(chan *transport.SessionPortalData, chanBufferSize),
		btPortalDataMessageChan: make(chan *transport.SessionPortalData, chanBufferSize),
		topSessions:           topSessions,
		sessionMap:            sessionMap,
		sessionMeta:           sessionMeta,
		sessionSlices:         sessionSlices,
		redisFlushCount:       redisFlushCount,
		useBigtable 		   useBigtable,
		btClient			   btClient,
		btCfNames			   btCfnames,
		flushTime:             time.Now(),
		pingTime:              time.Now(),

	}, nil
}

func (cruncher *PortalCruncher) Start(ctx context.Context, numReceiveGoroutines int, numRedisInsertGoroutines int, numBigtableInsertGoroutines int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// Start the receive goroutines
	for i := 0; i < numReceiveGoroutines; i++ {
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
	}

	// Start the redis goroutines
	for i := 0; i < numRedisInsertGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				// Buffer up some portal data entries and only insert into redis periodically to avoid overworking redis
				case portalData := <-cruncher.redisPortalDataMessageChan:
					cruncher.redisPortalDataBuffer = append(cruncher.redisPortalDataBuffer, portalData)

					if time.Since(cruncher.flushTime) < time.Second && len(cruncher.redisPortalDataBuffer) < cruncher.redisFlushCount {
						continue
					}

					// Periodically ping the redis instances and error out if we don't get a pong
					if time.Since(cruncher.pingTime) >= time.Second*10 {
						if err := cruncher.PingRedis(); err != nil {
							errChan <- err
							return
						}

						cruncher.pingTime = time.Now()
					}

					cruncher.flushTime = time.Now()
					minutes := cruncher.flushTime.Unix() / 60

					cruncher.InsertIntoRedis(minutes)

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

				for {
					select {
					// Buffer up some portal data entries and only insert into redis periodically to avoid overworking redis
					case portalData := <-cruncher.btPortalDataMessageChan:
						cruncher.btPortalDataBuffer = append(cruncher.btPortalDataBuffer, portalData)

						if err := cruncher.InsertIntoBigtable(ctx); err != nil {
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
	topic, messageChan, err := cruncher.subscriber.ReceiveMessage(ctx)
	if err != nil {
		return &ErrReceiveMessage{err: err}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()

	case message := <-messageChan:
		cruncher.metrics.ReceivedMessageCount.Add(1)

		switch topic {
		case pubsub.TopicPortalCruncherSessionCounts:
			// var sessionCountData transport.SessionCountData
			// if err := sessionCountData.UnmarshalBinary(message); err != nil {
			// 	return &ErrUnmarshalMessage{err: err}
			// }

			// select {
			// case cruncher.portalCountMessageChan <- &sessionCountData:
			// default:
			// 	return &ErrChannelFull{}
			// }

		case pubsub.TopicPortalCruncherSessionData:
			var sessionPortalData transport.SessionPortalData
			if err := sessionPortalData.UnmarshalBinary(message); err != nil {
				return &ErrUnmarshalMessage{err: err}
			}

			select {
			case cruncher.portalDataMessageChan <- &sessionPortalData:
			default:
				return &ErrChannelFull{}
			}
		default:
			return &ErrUnknownMessage{}
		}

		return nil
	}
}

func (cruncher *PortalCruncher) InsertIntoRedis(minutes int64) {
	// Remove the old global top sessions minute bucket from 2 minutes ago if it didn't expire
	cruncher.topSessions.Command("DEL", "s-%d", minutes-2)

	// Update the current global top sessions minute bucket
	var format string
	args := make([]interface{}, 0)

	format += "s-%d"
	args = append(args, minutes)

	for j := range cruncher.redisPortalDataBuffer {
		meta := cruncher.redisPortalDataBuffer[j].Meta
		largeCustomer := cruncher.redisPortalDataBuffer[j].LargeCustomer
		everOnNext := cruncher.redisPortalDataBuffer[j].EverOnNext

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

	for j := range cruncher.redisPortalDataBuffer {
		meta := &cruncher.redisPortalDataBuffer[j].Meta
		largeCustomer := cruncher.redisPortalDataBuffer[j].LargeCustomer
		everOnNext := cruncher.redisPortalDataBuffer[j].EverOnNext

		// For large customers, only insert the session if they have ever taken network next
		if largeCustomer && !meta.OnNetworkNext && !everOnNext {
			continue // Early out if we shouldn't add this session
		}

		slice := &cruncher.redisPortalDataBuffer[j].Slice
		point := &cruncher.redisPortalDataBuffer[j].Point
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

		// Remove the old per-buyer top sessions minute bucket from 2 minutes ago if it didnt expire
		// and update the current per-buyer top sessions list
		cruncher.topSessions.Command("DEL", "sc-%s-%d", customerID, minutes-2)
		cruncher.topSessions.Command("ZADD", "sc-%s-%d %.2f %s", customerID, minutes, score, sessionID)
		cruncher.topSessions.Command("EXPIRE", "sc-%s-%d %d", customerID, minutes, 30)

		// Remove the old map points minute buckets from 2 minutes ago if it didn't expire
		cruncher.sessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes-2, sessionID)
		cruncher.sessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes-2, sessionID)

		// Update the map points for this minute bucket
		// Make sure to remove the session ID from the opposite bucket in case the session
		// has switched from direct -> next or next -> direct
		pointString := point.RedisString()
		if next {
			cruncher.sessionMap.Command("HSET", "n-%s-%d %s %s", customerID, minutes, sessionID, pointString)
			cruncher.sessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes-1, sessionID)
			cruncher.sessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes, sessionID)
		} else {
			cruncher.sessionMap.Command("HSET", "d-%s-%d %s %s", customerID, minutes, sessionID, pointString)
			cruncher.sessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes-1, sessionID)
			cruncher.sessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes, sessionID)
		}

		// Expire map points
		cruncher.sessionMap.Command("EXPIRE", "n-%s-%d %d", customerID, minutes, 30)
		cruncher.sessionMap.Command("EXPIRE", "d-%s-%d %d", customerID, minutes, 30)

		// Update session meta
		cruncher.sessionMeta.Command("SET", "sm-%s \"%s\" EX %d", sessionID, meta.RedisString(), 120)

		// Update session slices
		cruncher.sessionSlices.Command("RPUSH", "ss-%s %s", sessionID, slice.RedisString())
		cruncher.sessionSlices.Command("EXPIRE", "ss-%s %d", sessionID, 120)
	}

	cruncher.redisPortalDataBuffer = cruncher.redisPortalDataBuffer[:0]
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

func (cruncher *PortalCruncher) SetupBigtable(ctx context.Context, gcpProjectID string, logger log.Logger) (*storage.Bigtable, []string, error) {
	// Setup Bigtable
	btEmulatorOK := envvar.Exists("BIGTABLE_EMULATOR_HOST")
	if btEmulatorOK {
		// Emulator is used for local testing
		// Requires that emulator has been started in another terminal to work as intended
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected Bigtable emulator")
	}

	if gcpProjectID == "" && !btEmulatorOK {
		return nil, nil, fmt.Errorf("No GCP Project ID found. Could not find $BIGTABLE_EMULATOR_HOST for local testing.")
	}

	// Get Bigtable instance ID
	btInstanceID := envvar.Get("GOOGLE_BIGTABLE_INSTANCE_ID", "")

	// Get the table name
	btTableName := envvar.Get("GOOGLE_BIGTABLE_TABLE_NAME", "")

	// Get the column family names and put them in a slice
	btCfName := envvar.Get("GOOGLE_BIGTABLE_CF_NAME", "")
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

	// Verify if the table needed exists
	if !tableExists {
		// Create a table with the given name and column families
		if err = btAdmin.CreateTable(ctx, btTableName, btCfNames); err != nil {
			return nil, nil, err
		}

		// Get the max number of days the data should be kept in Bigtable
		maxDays, err := envvar.GetInt("GOOGLE_BIGTABLE_MAX_AGE_DAYS", 90)
		if err != nil {
			return nil, nil, err
		}

		// Set a garbage collection policy of 90 days
		maxAge := time.Hour * time.Duration(24*maxDays)
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

func (cruncher *PortalCruncher) InsertIntoBigtable(ctx context.Context) error {
	for j := range cruncher.btPortalDataBuffer {
		meta := &cruncher.redisPortalDataBuffer[j].Meta
		slice := &cruncher.redisPortalDataBuffer[j].Slice
		
		sessionRowKey := fmt.Sprintf("%016x#%016x", meta.BuyerID, meta.ID)
		userRowKey := fmt.Sprintf("%016x#%016x", meta.UserHash, meta.ID)

		rowKeys := []string{sessionRowKey, userRowKey}

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

		// Insert session data into Bigtable
		if err := bt.InsertSessionData(ctx, cruncher.btCfNames, metaBinary, sliceBinary, rowKeys); err != nil {
			return err
		}
	}

	cruncher.btPortalDataBuffer = cruncher.btPortalDataBuffer[:0]
	return nil
}