package portalcruncher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/ghost_army"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"

	"github.com/gomodule/redigo/redis"
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

	sessionPool *redis.Pool

	useBigtable bool
	btClient    *storage.BigTable
	btCfNames   []string
}

func NewPortalCruncher(
	ctx context.Context,
	subscriber pubsub.Subscriber,
	redisHostname string,
	redisPassword string,
	redisMaxIdleConns int,
	redisMaxActiveConns int,
	useBigtable bool,
	gcpProjectID string,
	btInstanceID string,
	btTableName string,
	btCfName string,
	btEmulatorOK bool,
	btHistoricalPath string,
	chanBufferSize int,
	metrics *metrics.PortalCruncherMetrics,
	btMetrics *metrics.BigTableMetrics,
) (*PortalCruncher, error) {
	var err error

	redisPool := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err = storage.ValidateRedisPool(redisPool); err != nil {
		return nil, fmt.Errorf("failed to validate redis pool for hostname %s: %v", redisHostname, err)
	}

	var btClient *storage.BigTable
	var btCfNames []string

	if useBigtable {
		btClient, btCfNames, err = SetupBigtable(ctx, gcpProjectID, btInstanceID, btTableName, btCfName, btEmulatorOK, btHistoricalPath)
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
		sessionPool:           redisPool,
		useBigtable:           useBigtable,
		btClient:              btClient,
		btCfNames:             btCfNames,
	}, nil
}

func (cruncher *PortalCruncher) Start(ctx context.Context, numRedisInsertGoroutines int, numBigtableInsertGoroutines int, redisPingDuration time.Duration, redisFlushDuration time.Duration, redisFlushCount int, env string, errChan chan error, wg *sync.WaitGroup) {
	// Start the receive goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case err := <- cruncher.ReceiveMessage(ctx):
				if err != nil {
					switch err.(type) {
					case *ErrChannelFull: // We don't need to stop the portal cruncher if the channel is full
						continue
					default:
						core.Error("failed to receive message: %v", err)
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
						core.Error("failed to ping redis: %v", err)
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
		// Get the Ghost Army buyerID
		ghostArmyBuyerID := ghostarmy.GhostArmyBuyerID(env)

		// Start the bigtable goroutines
		for i := 0; i < numBigtableInsertGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				btPortalDataBuffer := make([]*transport.SessionPortalData, 0)

				for {
					select {
					// Buffer up some portal data entries to insert into bigtable
					case portalData := <-cruncher.btDataMessageChan:
						btPortalDataBuffer = append(btPortalDataBuffer, portalData)

						if err := cruncher.InsertIntoBigtable(ctx, btPortalDataBuffer, env, ghostArmyBuyerID); err != nil {
							core.Error("failed to insert data into bigtable: %v", err)
							errChan <- err
							return
						}
						btPortalDataBuffer = btPortalDataBuffer[:0]

					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}
}

func (cruncher *PortalCruncher) ReceiveMessage(ctx context.Context) <-chan error {
	errChan := make(chan error)

	go func() {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()

		case messageInfo := <-cruncher.subscriber.ReceiveMessage():
			cruncher.metrics.ReceivedMessageCount.Add(1)

			if messageInfo.Err != nil {
				errChan <- &ErrReceiveMessage{err: messageInfo.Err}
			}

			switch messageInfo.Topic {
			case pubsub.TopicPortalCruncherSessionCounts:
				var sessionCountData transport.SessionCountData
				if err := transport.ReadSessionCountData(&sessionCountData, messageInfo.Message); err != nil {
					errChan <- &ErrUnmarshalMessage{err: err}
				}

				select {
				case cruncher.redisCountMessageChan <- &sessionCountData:
				default:
					errChan <- &ErrChannelFull{}
				}

			case pubsub.TopicPortalCruncherSessionData:
				var sessionPortalData transport.SessionPortalData
				if err := transport.ReadSessionPortalData(&sessionPortalData, messageInfo.Message); err != nil {
					errChan <- &ErrUnmarshalMessage{err: err}
				}

				select {
				case cruncher.redisDataMessageChan <- &sessionPortalData:
				default:
					errChan <- &ErrChannelFull{}
				}

				select {
				case cruncher.btDataMessageChan <- &sessionPortalData:
				default:
					errChan <- &ErrChannelFull{}
				}
			default:
				errChan <- &ErrUnknownMessage{}
			}

			errChan <- nil
		}
	}()

	return errChan
}

func (cruncher *PortalCruncher) insertCountDataIntoRedis(redisPortalCountBuffer []*transport.SessionCountData, minutes int64) {
	sessionMapConn := cruncher.sessionPool.Get()
	defer sessionMapConn.Close()

	var cmd string

	for i := range redisPortalCountBuffer {
		customerID := fmt.Sprintf("%016x", redisPortalCountBuffer[i].BuyerID)
		serverID := fmt.Sprintf("%016x", redisPortalCountBuffer[i].ServerID)
		numSessions := redisPortalCountBuffer[i].NumSessions

		// Remove the old count minute bucket from 2 minutes ago if it didn't expire
		cmd = fmt.Sprintf("c-%s-%d", customerID, minutes-2)
		sessionMapConn.Do("DEL", cmd)

		// Add the new session count
		cmd = fmt.Sprintf("c-%s-%d %s %d", customerID, minutes, serverID, numSessions)
		sessionMapConn.Do("HSET", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		cmd = fmt.Sprintf("c-%s-%d %d", customerID, minutes, 30)
		sessionMapConn.Do("EXPIRE", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)
	}
}

func (cruncher *PortalCruncher) insertPortalDataIntoRedis(redisPortalDataBuffer []*transport.SessionPortalData, minutes int64) {
	topSessionsConn := cruncher.sessionPool.Get()
	sessionMapConn := cruncher.sessionPool.Get()
	sessionMetaConn := cruncher.sessionPool.Get()
	sessionSlicesConn := cruncher.sessionPool.Get()
	defer func() {
		topSessionsConn.Close()
		sessionMapConn.Close()
		sessionMetaConn.Close()
		sessionSlicesConn.Close()
	}()
	var cmd string

	// Remove the old global top sessions minute bucket from 2 minutes ago if it didn't expire
	cmd = fmt.Sprintf("s-%d", minutes-2)
	topSessionsConn.Do("DEL", cmd)

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

	cmd = fmt.Sprintf(format, args...)
	topSessionsConn.Do("ZADD", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

	cmd = fmt.Sprintf("s-%d %d", minutes, 30)
	topSessionsConn.Do("EXPIRE", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

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
			cmd = fmt.Sprintf("d-%s-%d %s", customerID, minutes-1, sessionID)
			sessionMapConn.Do("HDEL", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

			cmd = fmt.Sprintf("d-%s-%d %s", customerID, minutes, sessionID)
			sessionMapConn.Do("HDEL", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

			cmd = fmt.Sprintf("n-%s-%d %s %s", customerID, minutes, sessionID, pointString)
			sessionMapConn.Do("HSET", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)
		} else {
			cmd = fmt.Sprintf("n-%s-%d %s", customerID, minutes-1, sessionID)
			sessionMapConn.Do("HDEL", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

			cmd = fmt.Sprintf("n-%s-%d %s", customerID, minutes, sessionID)
			sessionMapConn.Do("HDEL", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

			cmd = fmt.Sprintf("d-%s-%d %s %s", customerID, minutes, sessionID, pointString)
			sessionMapConn.Do("HSET", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)
		}

		// Remove the old per-buyer top sessions minute bucket from 2 minutes ago if it didnt expire
		// and update the current per-buyer top sessions list
		cmd = fmt.Sprintf("sc-%s-%d", customerID, minutes-2)
		topSessionsConn.Do("DEL", cmd)

		cmd = fmt.Sprintf("sc-%s-%d %.2f %s", customerID, minutes, score, sessionID)
		topSessionsConn.Do("ZADD", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		cmd = fmt.Sprintf("sc-%s-%d %d", customerID, minutes, 30)
		topSessionsConn.Do("EXPIRE", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		// Remove the old map points minute buckets from 2 minutes ago if it didn't expire
		cmd = fmt.Sprintf("d-%s-%d %s", customerID, minutes-2, sessionID)
		sessionMapConn.Do("HDEL", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		cmd = fmt.Sprintf("n-%s-%d %s", customerID, minutes-2, sessionID)
		sessionMapConn.Do("HDEL", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		// Expire map points
		cmd = fmt.Sprintf("n-%s-%d %d", customerID, minutes, 30)
		sessionMapConn.Do("EXPIRE", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		cmd = fmt.Sprintf("d-%s-%d %d", customerID, minutes, 30)
		sessionMapConn.Do("EXPIRE", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		// Update session meta
		// There can be spaces in the ISP string, so manually create the string slice instead of splitting on white space
		cmdSlice := []string{fmt.Sprintf("sm-%s", sessionID), meta.RedisString(), "EX", fmt.Sprintf("%d", 120)}
		sessionMetaConn.Do("SET", redis.Args{}.AddFlat(cmdSlice)...)

		// Update session slices
		cmd = fmt.Sprintf("ss-%s %s", sessionID, slice.RedisString())
		sessionSlicesConn.Do("RPUSH", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)

		cmd = fmt.Sprintf("ss-%s %d", sessionID, 120)
		sessionSlicesConn.Do("EXPIRE", redis.Args{}.AddFlat(strings.Split(cmd, " "))...)
	}
}

func (cruncher *PortalCruncher) PingRedis() error {
	if err := storage.ValidateRedisPool(cruncher.sessionPool); err != nil {
		return err
	}

	return nil
}

func (cruncher *PortalCruncher) CloseRedisPool() {
	cruncher.sessionPool.Close()
}

func SetupBigtable(ctx context.Context,
	gcpProjectID string,
	btInstanceID string,
	btTableName string,
	btCfName string,
	btEmulatorOK bool,
	btHistoricalPath string) (*storage.BigTable, []string, error) {
	// Setup Bigtable
	if btEmulatorOK {
		// Emulator is used for local testing
		// Requires that emulator has been started in another terminal to work as intended
		gcpProjectID = "local"
		core.Debug("Detected Bigtable emulator host.\n\tTable Name: %s\n\tProject ID: %s\n\tInstance ID: %s", btTableName, gcpProjectID, btInstanceID)
	}

	if gcpProjectID == "" && !btEmulatorOK {
		return nil, nil, fmt.Errorf("SetupBigtable() No GCP Project ID found. Could not find $BIGTABLE_EMULATOR_HOST for local testing.")
	}

	// Put the column family names in a slice
	btCfNames := []string{btCfName}
	// Create a bigtable admin for setup
	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
	if err != nil {
		return nil, nil, err
	}

	// Check if the table exists in the instance
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	if err != nil {
		return nil, nil, fmt.Errorf("SetupBigtable() Failed to verify if table exists: %v", err)
	}

	if !tableExists {
		return nil, nil, fmt.Errorf("SetupBigtable() Could not find table %s in bigtable instance. Create the table before starting the portal cruncher", btTableName)
	} else {
		core.Debug("Found the table in the bigtable instance")
	}

	// Close the admin client
	if err = btAdmin.Close(); err != nil {
		return nil, nil, err
	}

	// Create a standard client for writing to the table
	btClient, err := storage.NewBigTable(ctx, gcpProjectID, btInstanceID, btTableName)
	if err != nil {
		return nil, nil, err
	}

	if btEmulatorOK {
		if err = btClient.SeedBigtable(ctx, btCfNames, btHistoricalPath); err != nil {
			core.Error("failed to seed BigTable: %v", err)
		}
	}

	return btClient, btCfNames, nil
}

func (cruncher *PortalCruncher) InsertIntoBigtable(ctx context.Context, btPortalDataBuffer []*transport.SessionPortalData, env string, ghostArmyBuyerID uint64) error {
	for j := range btPortalDataBuffer {
		meta := &btPortalDataBuffer[j].Meta
		slice := &btPortalDataBuffer[j].Slice

		// Do not insert ghost army in prod
		if env == "prod" && meta.BuyerID == ghostArmyBuyerID {
			continue
		}

		// This seems redundant, try to figure out a better prefix to limit the number of keys
		// Key for session ID
		sessionRowKey := fmt.Sprintf("%016x", meta.ID)
		sliceRowKey := fmt.Sprintf("%016x#%v", meta.ID, slice.Timestamp)

		// Key for all user specific
		userRowKey := fmt.Sprintf("%016x#%016x", meta.UserHash, meta.ID)

		metaRowKeys := []string{sessionRowKey, userRowKey}
		sliceRowKeys := []string{sliceRowKey}

		// Have 2 columns under 1 column family
		// 1) Meta
		// 2) Slice

		// Create byte slices of the session data
		metaBinary, err := transport.WriteSessionMeta(meta)
		if err != nil {
			return err
		}
		sliceBinary, err := transport.WriteSessionSlice(slice)
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

	return nil
}

func (cruncher *PortalCruncher) CloseBigTable() {
	if cruncher.useBigtable {
		cruncher.btClient.Close()
	}
}
