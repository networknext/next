package vanity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gomodule/redigo/redis"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/pubsub"
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

type VanityMetricHandler struct {
	handler              metrics.Handler
	metrics              *metrics.VanityServiceMetrics
	vanityMetricDataChan chan *VanityMetrics
	buyerMetricMap       map[string]*metrics.VanityMetric
	mapMutex             sync.RWMutex
	redisSessionsMap     *redis.Pool
	redisSetName         string
	maxUserIdleTime      time.Duration
	subscriber           pubsub.Subscriber
	hourMetricsMap       map[string]bool
	displayMap           map[string]string
	logger               log.Logger
}

// VanityMetrics is the struct for all desired vanity metrics
// passed internally before being written by the metrics handler.
// The metric counterpart is located at modules/metrics/vanity_metric.go
// Metrics are derived from the billingEntry in ExtractVanityMetrics() in modules/transport/post_session_handler.go
type VanityMetrics struct {
	BuyerID                 uint64
	UserHash                uint64
	SessionID               uint64
	Timestamp               uint64
	SlicesAccelerated       uint64
	SlicesLatencyReduced    uint64
	SlicesPacketLossReduced uint64
	SlicesJitterReduced     uint64
	SessionsAccelerated     uint64
}

func NewVanityMetricHandler(vanityHandler metrics.Handler, vanityServiceMetrics *metrics.VanityServiceMetrics, chanBufferSize int,
	vanitySubscriber pubsub.Subscriber, redisSessions string, redisMaxIdleConnections int, redisMaxActiveConnections int,
	vanityMaxUserIdleTime time.Duration, vanitySetName string, logger log.Logger) *VanityMetricHandler {

	// Create Redis client for userHash -> sessionID, timestamp map
	vanitySessionsMap := storage.NewRedisPool(redisSessions, redisMaxIdleConnections, redisMaxActiveConnections)

	// List of metrics that need the number of hours calculated (i.e. Hours of Latency Reduced)
	vanityHourMetricsMap := map[string]bool{
		"Slices Accelerated":         true,
		"Slices Latency Reduced":     true,
		"Slices Packet Loss Reduced": true,
		"Slices Jitter Reduced":      true,
	}
	// Map of internal vanity metric Display Name to actual name shown to customers
	vanityDisplayMap := map[string]string{
		"Slices Accelerated":         "Hours Accelerated",
		"Slices Latency Reduced":     "Hours Latency Reduced",
		"Slices Packet Loss Reduced": "Hours Packet Loss Reduced",
		"Slices Jitter Reduced":      "Hours Jitter Reduced",
		"Sessions Accelerated":       "Sessions Accelerated",
	}

	return &VanityMetricHandler{
		handler:              vanityHandler,
		metrics:              vanityServiceMetrics,
		vanityMetricDataChan: make(chan *VanityMetrics, chanBufferSize),
		buyerMetricMap:       make(map[string]*metrics.VanityMetric),
		mapMutex:             sync.RWMutex{},
		redisSessionsMap:     vanitySessionsMap,
		redisSetName:         vanitySetName,
		maxUserIdleTime:      vanityMaxUserIdleTime,
		subscriber:           vanitySubscriber,
		hourMetricsMap:       vanityHourMetricsMap,
		displayMap:           vanityDisplayMap,
		logger:               logger,
	}
}

func (v VanityMetrics) Size() uint64 {
	sum := 8 + // BuyerID
		8 + // UserHash
		8 + // SessionID
		8 + // Timestamp
		8 + // SlicesAccelerated
		8 + // SlicesLatencyReduced
		8 + // SlicesPacketLossReduced
		8 + // SlicesJitterReduced
		8 // SessionsAccelerated

	return uint64(sum)
}

func (v VanityMetrics) MarshalBinary() ([]byte, error) {
	data := make([]byte, v.Size())
	index := 0

	encoding.WriteUint64(data, &index, v.BuyerID)
	encoding.WriteUint64(data, &index, v.UserHash)
	encoding.WriteUint64(data, &index, v.SessionID)
	encoding.WriteUint64(data, &index, v.Timestamp)

	encoding.WriteUint64(data, &index, v.SlicesAccelerated)
	encoding.WriteUint64(data, &index, v.SlicesLatencyReduced)
	encoding.WriteUint64(data, &index, v.SlicesPacketLossReduced)
	encoding.WriteUint64(data, &index, v.SlicesJitterReduced)
	encoding.WriteUint64(data, &index, v.SessionsAccelerated)

	return data, nil
}

func (v *VanityMetrics) UnmarshalBinary(data []byte) error {
	index := 0

	if !encoding.ReadUint64(data, &index, &v.BuyerID) {
		return errors.New("[VanityMetrics] invalid read at buyer ID")
	}

	if !encoding.ReadUint64(data, &index, &v.UserHash) {
		return errors.New("[VanityMetrics] invalid read at user hash")
	}

	if !encoding.ReadUint64(data, &index, &v.SessionID) {
		return errors.New("[VanityMetrics] invalid read at session ID")
	}

	if !encoding.ReadUint64(data, &index, &v.Timestamp) {
		return errors.New("[VanityMetrics] invalid read at timestamp")
	}
	if !encoding.ReadUint64(data, &index, &v.SlicesAccelerated) {
		return errors.New("[VanityMetrics] invalid read at slices accelerated")
	}

	if !encoding.ReadUint64(data, &index, &v.SlicesLatencyReduced) {
		return errors.New("[VanityMetrics] invalid read at slices latency reduced")
	}

	if !encoding.ReadUint64(data, &index, &v.SlicesPacketLossReduced) {
		return errors.New("[VanityMetrics] invalid read at slices packet loss reduced")
	}

	if !encoding.ReadUint64(data, &index, &v.SlicesJitterReduced) {
		return errors.New("[VanityMetrics] invalid read at slices jitter reduced")
	}

	if !encoding.ReadUint64(data, &index, &v.SessionsAccelerated) {
		return errors.New("[VanityMetrics] invalid read at sessions accelerated")
	}

	return nil
}

func (vm *VanityMetricHandler) Start(ctx context.Context, numVanityUpdateGoroutines int) error {
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
				if err := vm.ReceiveMessage(ctx); err != nil {
					switch err.(type) {
					case *ErrChannelFull: // We don't need to stop the vanity metric handler if the channel is full
						continue
					default:
						errChan <- err
						return
					}
				}
			}
		}
	}()

	// Start the goroutines for preparing and updating the metrics for the write loop
	for i := 0; i < numVanityUpdateGoroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// Each goroutine has its own buffer to avoid syncing
			vanityMetricDataBuffer := make([]*VanityMetrics, 0)

			for {
				select {
				// Buffer up some vanity metric entries to insert into StackDriver
				case vanityData := <-vm.vanityMetricDataChan:
					vanityMetricDataBuffer = append(vanityMetricDataBuffer, vanityData)
					fmt.Printf("got data from channel %v\n", vanityData)
					if err := vm.UpdateMetrics(ctx, vanityMetricDataBuffer); err != nil {
						fmt.Printf("Error updating metrics: %v\n", err)
						vm.metrics.UpdateVanityFailureCount.Add(1)
						errChan <- err
						return
					}
					vanityMetricDataBuffer = vanityMetricDataBuffer[:0]
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Wait until either there is an error or the context is done
	select {
	case err := <-errChan:
		fmt.Printf("The exit error is: %v\n", err)
		return err
	case <-ctx.Done():
		// Let the goroutines finish up
		wg.Wait()
		fmt.Printf("Waitgroups are done, return ctx error: %v", ctx.Err())
		return ctx.Err()
	}
}

// Receive messages from ZeroMQ and insert them into the VanityMetricHandler's data channel
func (vm *VanityMetricHandler) ReceiveMessage(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()

	case messageInfo := <-vm.subscriber.ReceiveMessage():
		vm.metrics.ReceivedVanityCount.Add(1)
		fmt.Printf("received message from ZMQ\n")
		if messageInfo.Err != nil {
			fmt.Printf("Error from message info: %v\n", messageInfo.Err)
			level.Error(vm.logger).Log("err", messageInfo.Err)
			return &ErrReceiveMessage{err: messageInfo.Err}
		}

		switch messageInfo.Topic {
		case pubsub.TopicVanityMetricData:
			var vanityData VanityMetrics
			if err := vanityData.UnmarshalBinary(messageInfo.Message); err != nil {
				fmt.Printf("Error from unmarshaling binary: %v\n", err)
				level.Error(vm.logger).Log("msg", "Could not unmarshal binary", "err", err)
				return &ErrUnmarshalMessage{err: err}
			}

			// Add any new buyers to map
			buyerID := fmt.Sprintf("%016x", vanityData.BuyerID)
			isNewBuyerID, err := vm.AddNewBuyerID(ctx, buyerID)
			if err != nil {
				fmt.Printf("Error from AddNewBuyerID(): %v\n", err)
				return err
			}
			if isNewBuyerID {
				level.Debug(vm.logger).Log("msg", "BuyerID is new", "BuyerID", buyerID)
			} else {
				level.Debug(vm.logger).Log("msg", "BuyerID already exists in map", "BuyerID", buyerID)
			}

			select {
			case vm.vanityMetricDataChan <- &vanityData:
				fmt.Printf("Inserted vanity data into data chan\n")
				level.Debug(vm.logger).Log("msg", "Successfully received vanity data from ZeroMQ")
			default:
				return &ErrChannelFull{}
			}
		default:
			return &ErrUnknownMessage{}
		}

		return nil
	}
}

// Adds a new buyerID to the buyer map if it doesn't already exist
// Returns true if the buyerID is new
func (vm *VanityMetricHandler) AddNewBuyerID(ctx context.Context, buyerID string) (bool, error) {
	vm.mapMutex.RLock()
	_, exists := vm.buyerMetricMap[buyerID]
	vm.mapMutex.RUnlock()

	if !exists {
		// Creates counters / gauges / histograms per vanity metric for this buyer ID,
		// or provides existing ones from a previous run
		vanityMetricPerBuyer, err := metrics.NewVanityMetric(ctx, vm.handler, buyerID)
		if err != nil {
			fmt.Printf("Error getting new vanity metric: %v\n", err)
			level.Error(vm.logger).Log("err", err)
			return true, err
		}

		// Store this vanity metric in the map for future look up
		vm.mapMutex.Lock()
		vm.buyerMetricMap[buyerID] = vanityMetricPerBuyer
		vm.mapMutex.Unlock()
		fmt.Printf("Found new buyerID %s inserted into map\n", buyerID)
		level.Debug(vm.logger).Log("msg", "Found new buyer ID, inserted into map for quick lookup", "buyerID", buyerID)
		return true, nil
	}

	return false, nil
}

// Updates the metrics per buyer
func (vm *VanityMetricHandler) UpdateMetrics(ctx context.Context, vanityMetricDataBuffer []*VanityMetrics) error {
	fmt.Printf("inside UpdateMetrics()\n")
	for j := range vanityMetricDataBuffer {
		fmt.Printf("Inside update metric loop\n")
		buyerID := fmt.Sprintf("%016x", vanityMetricDataBuffer[j].BuyerID)
		level.Debug(vm.logger).Log("msg", "Buyer ID obtained from data", "buyerID", buyerID)
		fmt.Printf("Got buyerID from data %s\n", buyerID)
		// Get the counters / gauges / histograms per vanity metric for this buyer ID
		fmt.Printf("before RLock\n")
		vm.mapMutex.RLock()
		fmt.Printf("After RLock()\n")
		vanityMetricPerBuyer, exists := vm.buyerMetricMap[buyerID]
		fmt.Printf("Accessed buyerMetricMap: %v, %v\n", vanityMetricPerBuyer, exists)
		vm.mapMutex.RUnlock()
		fmt.Printf("After RUnlock()\n")

		if !exists {
			fmt.Printf("Error could not find buyerID %s in map\n", buyerID)
			return fmt.Errorf("Could not find buyerID %s in map", buyerID)
		}
		fmt.Printf("After !exists\n")
		// Calculate sessionsAccelerated
		newSession, err := vm.IsNewSession(vanityMetricDataBuffer[j].SessionID)
		if err != nil {
			fmt.Printf("Error from IsNewSession(): %v\n", err)
			level.Error(vm.logger).Log("err", err)
			return err
		}
		if newSession {
			fmt.Printf("newSession is True\n")
			level.Debug(vm.logger).Log("msg", "Found new accelerated session for user", "userHash", fmt.Sprintf("%016x", vanityMetricDataBuffer[j].UserHash), "sessionID", fmt.Sprintf("%016x", vanityMetricDataBuffer[j].SessionID))
			vanityMetricDataBuffer[j].SessionsAccelerated = 1
		}
		fmt.Printf("After IsNewSession() check\n")

		currentSlicesAccelerated := vanityMetricPerBuyer.SlicesAccelerated.Value()
		currentSlicesLatencyReduced := vanityMetricPerBuyer.SlicesLatencyReduced.Value()
		currentSlicesPacketLossReduced := vanityMetricPerBuyer.SlicesPacketLossReduced.Value()
		currentSlicesJitterReduced := vanityMetricPerBuyer.SlicesJitterReduced.Value()
		currentSessionsAccelerated := vanityMetricPerBuyer.SessionsAccelerated.Value()

		level.Debug(vm.logger).Log("msg", "Before updating metric values",
			"buyerID", buyerID,
			"userHash", vanityMetricDataBuffer[j].UserHash,
			"sessionID", vanityMetricDataBuffer[j].SessionID,
			"timestamp", vanityMetricDataBuffer[j].Timestamp,
			"SlicesAccelerated", currentSlicesAccelerated,
			"SlicesLatencyReduced", currentSlicesLatencyReduced,
			"SlicesPacketLossReduced", currentSlicesPacketLossReduced,
			"SlicesJitterReduced", currentSlicesJitterReduced,
			"SessionsAccelerated", currentSessionsAccelerated,
		)
		fmt.Printf("Before updating metric values\n")
		// Update each metric's value
		// Writing to stack driver is taken care of by the tsMetricsHandler's WriteLoop() in cmd/vanity/vanity.go
		vanityMetricPerBuyer.SlicesAccelerated.Add(float64(vanityMetricDataBuffer[j].SlicesAccelerated))
		vanityMetricPerBuyer.SlicesLatencyReduced.Add(float64(vanityMetricDataBuffer[j].SlicesLatencyReduced))
		vanityMetricPerBuyer.SlicesPacketLossReduced.Add(float64(vanityMetricDataBuffer[j].SlicesPacketLossReduced))
		vanityMetricPerBuyer.SlicesJitterReduced.Add(float64(vanityMetricDataBuffer[j].SlicesJitterReduced))
		vanityMetricPerBuyer.SessionsAccelerated.Add(float64(vanityMetricDataBuffer[j].SessionsAccelerated))
		fmt.Printf("After updating metric values")
		level.Debug(vm.logger).Log("msg", "Updating metric values",
			"buyerID", buyerID,
			"userHash", vanityMetricDataBuffer[j].UserHash,
			"sessionID", vanityMetricDataBuffer[j].SessionID,
			"timestamp", vanityMetricDataBuffer[j].Timestamp,
			"SlicesAccelerated", vanityMetricPerBuyer.SlicesAccelerated.Value(),
			"SlicesLatencyReduced", vanityMetricPerBuyer.SlicesLatencyReduced.Value(),
			"SlicesPacketLossReduced", vanityMetricPerBuyer.SlicesPacketLossReduced.Value(),
			"SlicesJitterReduced", vanityMetricPerBuyer.SlicesJitterReduced.Value(),
			"SessionsAccelerated", vanityMetricPerBuyer.SessionsAccelerated.Value(),
		)

		vm.metrics.UpdateVanitySuccessCount.Add(1)
	}

	return nil
}

// Checks if a userHash has moved onto a new sessionID at a later point in time
func (vm *VanityMetricHandler) IsNewSession(sessionID uint64) (bool, error) {
	sessionIDStr := fmt.Sprintf("%016x", sessionID)
	fmt.Printf("Before SessionIDExists()")
	exists, err := vm.SessionIDExists(sessionIDStr)
	if err != nil {
		fmt.Printf("Error from SessionIDExists(): %v\n", err)
		return exists, nil
	}
	fmt.Printf("After SessionIDExists()\n")
	if exists {
		// Not a new sessionID
		fmt.Printf("Not a new sessionID\n")
		return false, nil
	}
	fmt.Printf("Before AddSessionID\n")
	// Found a new sessionID, add it to the set
	err = vm.AddSessionID(sessionIDStr)
	if err != nil {
		fmt.Printf("Error from AddSessionID(): %v\n", err)
		return exists, err
	}
	fmt.Printf("After AddSessionID()")

	return true, nil
}

// Returns a marshaled JSON of an empty VanityMetrics struct
func (vm *VanityMetricHandler) GetEmptyMetrics() ([]byte, error) {
	ret_val, err := json.Marshal(&VanityMetrics{})
	if err != nil {
		return nil, err
	}

	return ret_val, nil
}

// Deletes a metric descriptor from StackDriver, given a gcpProject ID, service name, and metric name
// For vanity metrics, the service name would be the buyerID, and the metric name could be `vanity_metric.slices_accelerated`
func (vm *VanityMetricHandler) DeleteMetricDescriptor(ctx context.Context, sd *metrics.StackDriverHandler, gcpProjectID string, serviceName string, metricName string) error {
	name := fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", gcpProjectID, serviceName, metricName)
	req := &monitoringpb.DeleteMetricDescriptorRequest{Name: name}
	err := sd.Client.DeleteMetricDescriptor(ctx, req)

	return err
}

// Returns a marshaled JSON of all custom metrics tracked through Stackdriver
func (vm *VanityMetricHandler) ListCustomMetrics(ctx context.Context, sd *metrics.StackDriverHandler, gcpProjectID string, serviceName string) ([]byte, error) {
	descFilter := fmt.Sprintf(`metric.type = starts_with("custom.googleapis.com/%s")`, serviceName)
	descReq := &monitoringpb.ListMetricDescriptorsRequest{
		Name:   fmt.Sprintf("projects/%s", gcpProjectID),
		Filter: descFilter,
	}
	descIt := sd.Client.ListMetricDescriptors(ctx, descReq)

	customMetrics := make(map[string]string)
	for {
		resp, err := descIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		customMetrics[resp.DisplayName] = resp.Type
	}

	// Encode the map of custom metric names to descriptions
	ret_val, err := json.Marshal(customMetrics)
	if err != nil {
		return nil, err
	}

	return ret_val, nil
}

// Returns a map of display name to metric type for a custom service tracked through Stackdriver
// Example: {"Route Matrix Bytes": "custom.googleapis.com/server_backend/route_matrix_update.bytes"}
func (vm *VanityMetricHandler) GetCustomMetricTypes(ctx context.Context, sd *metrics.StackDriverHandler, gcpProjectID string, serviceName string) (map[string]string, error) {
	descFilter := fmt.Sprintf(`metric.type = starts_with("custom.googleapis.com/%s")`, serviceName)
	descReq := &monitoringpb.ListMetricDescriptorsRequest{
		Name:   fmt.Sprintf("projects/%s", gcpProjectID),
		Filter: descFilter,
	}
	descIt := sd.Client.ListMetricDescriptors(ctx, descReq)

	customMetrics := make(map[string]string)
	for {
		resp, err := descIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		customMetrics[resp.DisplayName] = resp.Type
	}

	return customMetrics, nil
}

func (vm *VanityMetricHandler) GetVanityMetricJSON(ctx context.Context, sd *metrics.StackDriverHandler, gcpProjectID string, buyerID string, startTime time.Time, endTime time.Time) ([]byte, error) {
	// Create a interval and duration for all vanity metrics
	tsInterval := &monitoringpb.TimeInterval{EndTime: timestamppb.New(endTime), StartTime: timestamppb.New(startTime)}
	duration := endTime.Sub(startTime)

	// Create max aggregation (used for Counters)
	maxAgg := &monitoringpb.Aggregation{
		AlignmentPeriod:    durationpb.New(duration),
		PerSeriesAligner:   monitoringpb.Aggregation_Aligner(14), // Get summed values per alignment period
		CrossSeriesReducer: monitoringpb.Aggregation_Reducer(4),  // Sum across each alignment period
	}

	// Create the final returned map
	metricsMap := make(map[string]float64)

	// Get all vanity metric names for the buyer ID
	vanityMetricTypes, err := vm.GetCustomMetricTypes(ctx, sd, gcpProjectID, buyerID)
	if err != nil {
		level.Error(vm.logger).Log("err", err)
		return nil, err
	}

	// Get the values for each vanity metric
	for displayName, metricType := range vanityMetricTypes {
		// Ensure the metric is a vanity metric to show the customer
		if _, ok := vm.displayMap[displayName]; !ok {
			continue
		}

		tsFilter := vm.GetTimeSeriesFilter(metricType)
		tsName := vm.GetTimeSeriesName(gcpProjectID, metricType)
		pointsList, err := vm.GetPointDetails(ctx, sd, tsName, tsFilter, tsInterval, maxAgg)
		if err != nil {
			errStr := fmt.Sprintf("Could not get point details for %s (%s)", displayName, metricType)
			level.Error(vm.logger).Log("err", errStr)
			return nil, errors.New(errStr)
		}

		// Take the sum of the list of points
		sumPointVal := int64(0)
		for _, points := range pointsList {
			for _, point := range points {
				sumPointVal += point.Value.GetInt64Value()
			}
		}

		floatPointVal := float64(sumPointVal)
		// Check if the a slice metric needs hours calculated
		if vm.hourMetricsMap[displayName] {
			seconds := time.Second * time.Duration(10*sumPointVal)
			hours := seconds.Hours()
			// Round to 3 decimal places
			floatPointVal = math.Round(hours*1000) / 1000
		}

		// Add metric value to the final map
		metricsMap[vm.displayMap[displayName]] = floatPointVal
	}

	// Encode the map of vanity metric names to their values
	ret_val, err := json.Marshal(metricsMap)
	if err != nil {
		return nil, err
	}

	return ret_val, nil
}

// Gets lists of points for a time series request
func (vm *VanityMetricHandler) GetPointDetails(ctx context.Context, sd *metrics.StackDriverHandler, name string, filter string, interval *monitoringpb.TimeInterval, aggregation *monitoringpb.Aggregation) ([][]*monitoringpb.Point, error) {
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:        name,
		Filter:      filter,
		Interval:    interval,
		Aggregation: aggregation,
		View:        monitoringpb.ListTimeSeriesRequest_TimeSeriesView(0),
	}
	it := sd.Client.ListTimeSeries(ctx, req)

	var pointSeries [][]*monitoringpb.Point
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		pointSeries = append(pointSeries, resp.GetPoints())
	}

	return pointSeries, nil
}

func (vm *VanityMetricHandler) GetTimeSeriesFilter(metricType string) string {
	return fmt.Sprintf(`metric.type = "%s"`, metricType)
}

func (vm *VanityMetricHandler) GetTimeSeriesName(gcpProjectID string, metricType string) string {
	return fmt.Sprintf(`projects/%s/metricDescriptors/%s`, gcpProjectID, metricType)
}

func (vm *VanityMetricHandler) SessionIDExists(sessionID string) (exists bool, err error) {
	fmt.Printf("Inside SessionIDExists()\n")
	conn := vm.redisSessionsMap.Get()
	fmt.Printf("Got connection for SessionIDExists\n")
	defer conn.Close()
	member := fmt.Sprintf("sid-%s", sessionID)
	fmt.Printf("Before ZSCORE\n")
	score, err := conn.Do("ZSCORE", redis.Args{}.Add(vm.redisSetName).Add(member)...)
	if err != nil {
		fmt.Printf("Error from ZSCORE: %v\n", err)
		return false, err
	}
	fmt.Printf("After ZSCORE\n")

	if score != nil {
		// If the sessionID exists, refresh its expiration time
		err = vm.AddSessionID(sessionID)
		if err != nil {
			fmt.Printf("Error from AddSessionID() in SessionIDExists(): %v\n", err)
			return true, err
		}
		return true, nil
	}

	return false, nil
}

func (vm *VanityMetricHandler) AddSessionID(sessionID string) error {
	fmt.Printf("Inside AddSessionID()\n")
	conn := vm.redisSessionsMap.Get()
	fmt.Printf("Got connection for AddSessionID()\n")
	defer conn.Close()

	// Refresh the expiration time for this key
	refreshedTime := time.Now().Add(vm.maxUserIdleTime).UnixNano()
	member := fmt.Sprintf("sid-%s", sessionID)
	fmt.Printf("Before ZADD\n")
	_, err := conn.Do("ZADD", redis.Args{}.Add(vm.redisSetName).Add(refreshedTime).Add(member)...)
	if err != nil {
		fmt.Printf("Error from AddSessionID(): %v\n", err)
		return err
	}
	fmt.Printf("After ZADD\n")
	fmt.Printf("Before ExpireOldSessions()\n")
	// Expire old set members
	err = vm.ExpireOldSessions(conn)
	fmt.Printf("After ExpireOldSessions(), err is %v\n", err)
	return err
}

func (vm *VanityMetricHandler) ExpireOldSessions(conn redis.Conn) error {
	fmt.Printf("Inside ExpireOldSessions()\n")
	currentTime := time.Now().UnixNano()
	fmt.Printf("Before ZREMRANGEBYSCORE\n")
	numRemoved, err := redis.Int(conn.Do("ZREMRANGEBYSCORE", redis.Args{}.Add(vm.redisSetName).Add("-inf").Add(fmt.Sprintf("(%d", currentTime))...))
	fmt.Printf("After ZREMRANGEBYSCORE, err is %v\n", err)
	if numRemoved != 0 {
		fmt.Printf("Removed %d SessionIDs from Redis\n", numRemoved)
		level.Debug(vm.logger).Log("msg", "Removed Session IDs from redis", "numRemoved", numRemoved)
	}

	return err
}
