package vanity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
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
	metrics              *metrics.VanityMetricMetrics
	vanityMetricDataChan chan *VanityMetrics
	buyerMetricMap       map[string]*metrics.VanityMetric
	mapMutex             sync.RWMutex
	subscriber           pubsub.Subscriber
}

// VanityMetrics is the struct for all desired vanity metrics
// passed internally before being written by the metrics handler.
// The metric counterpart is located at modules/metrics/vanity_metric.go
// Aggregation of vanity metrics takes place in AggregateVanityMetrics() in modules/transport/post_session_handler.go
type VanityMetrics struct {
	BuyerID                 uint64
	SlicesAccelerated       uint64
	SlicesLatencyReduced    uint64
	SlicesPacketLossReduced uint64
	SlicesJitterReduced 	uint64
	SessionsAccelerated 	uint64
}

// func NewVanityMetricHandler(vanityHandler *metrics.TSMetricHandler, vanityMetricMetrics *metrics.VanityMetricMetrics, vanitySubscriber pubsub.Subscriber, vanityLogger log.logger) VanityMetricHandler {
func NewVanityMetricHandler(vanityHandler metrics.Handler, vanityMetricMetrics *metrics.VanityMetricMetrics, chanBufferSize int, vanitySubscriber pubsub.Subscriber) *VanityMetricHandler {
	return &VanityMetricHandler{
		handler:              vanityHandler,
		metrics:              vanityMetricMetrics,
		vanityMetricDataChan: make(chan *VanityMetrics, chanBufferSize),
		buyerMetricMap:       make(map[string]*metrics.VanityMetric),
		mapMutex:             sync.RWMutex{},
		subscriber:           vanitySubscriber,
	}
}

func (v VanityMetrics) Size() uint64 {
	sum := 8 + // BuyerID
		8 + // SlicesAccelerated
		8 + // SlicesLatencyReduced
		8 + // SlicesPacketLossReduced
		8 + // SlicesJitterReduced
		8   // SessionsAccelerated

	return uint64(sum)
}

func (v VanityMetrics) MarshalBinary() ([]byte, error) {
	data := make([]byte, v.Size())
	index := 0

	encoding.WriteUint64(data, &index, v.BuyerID)

	encoding.WriteUint64(data, &index, v.SlicesAccelerated)
	encoding.WriteUint64(data, &index, v.SlicesLatencyReduced)
	encoding.WriteUint64(data, &index, v.SlicesPacketLossReduced)
	encoding.WriteUint64(data, &index, v.SlicesJitterReduced)
	encoding.WriteUint64(data, &index, v.SessionsAccelerated)

	return data, nil
}

func (v VanityMetrics) UnmarshalBinary(data []byte) error {
	index := 0

	if !encoding.ReadUint64(data, &index, &v.BuyerID) {
		return errors.New("[VanityMetrics] invalid read at buyer ID")
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

					if err := vm.UpdateMetrics(ctx, vanityMetricDataBuffer); err != nil {
						vm.metrics.UpdateVanityFailureCount.Add(1)
						errChan <- err
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
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

// Receive messages from ZeroMQ and insert them into the VanityMetricHandler's data channel
func (vm *VanityMetricHandler) ReceiveMessage(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()

	case messageInfo := <-vm.subscriber.ReceiveMessage():
		vm.metrics.ReceivedVanityCount.Add(1)

		if messageInfo.Err != nil {
			return &ErrReceiveMessage{err: messageInfo.Err}
		}

		switch messageInfo.Topic {
		case pubsub.TopicVanityMetricData:
			var vanityData VanityMetrics
			if err := vanityData.UnmarshalBinary(messageInfo.Message); err != nil {
				return &ErrUnmarshalMessage{err: err}
			}

			select {
			case vm.vanityMetricDataChan <- &vanityData:
			default:
				return &ErrChannelFull{}
			}
		default:
			return &ErrUnknownMessage{}
		}

		return nil
	}
}

// Updates the metrics per buyer
func (vm *VanityMetricHandler) UpdateMetrics(ctx context.Context, vanityMetricDataBuffer []*VanityMetrics) error {
	for j := range vanityMetricDataBuffer {
		buyerID := fmt.Sprintf("%016x", vanityMetricDataBuffer[j].BuyerID)

		// Get the counters / gauges / histograms per vanity metric for this buyer ID
		// Check the map first for quick look up
		var vanityMetricPerBuyer *metrics.VanityMetric
		var err error

		vm.mapMutex.RLock()
		vanityMetricPerBuyer, exists := vm.buyerMetricMap[buyerID]
		vm.mapMutex.RUnlock()

		if !exists {
			// Creates counters / gauges / histograms per vanity metric for this buyer ID,
			// or provides existing ones from a previous run
			vanityMetricPerBuyer, err = metrics.NewVanityMetric(ctx, vm.handler, buyerID)
			if err != nil {
				return err
			}

			// Store this vanity metric in the map for future look up
			vm.mapMutex.Lock()
			vm.buyerMetricMap[buyerID] = vanityMetricPerBuyer
			vm.mapMutex.Unlock()
		}

		// Update each metric's value
		// Writing to stack driver is taken care of by the the WriteLoop() started in cmd/vanity/vanity.go
		vanityMetricPerBuyer.SlicesAccelerated.Add(float64(vanityMetricDataBuffer[j].SlicesAccelerated))
		vanityMetricPerBuyer.SlicesLatencyReduced.Add(float64(vanityMetricDataBuffer[j].SlicesLatencyReduced))
		vanityMetricPerBuyer.SlicesPacketLossReduced.Add(float64(vanityMetricDataBuffer[j].SlicesPacketLossReduced))
		vanityMetricPerBuyer.SlicesJitterReduced.Add(float64(vanityMetricDataBuffer[j].SlicesJitterReduced))
		vanityMetricPerBuyer.SessionsAccelerated.Add(float64(vanityMetricDataBuffer[j].SessionsAccelerated))

		vm.metrics.UpdateVanitySuccessCount.Add(1)
	}

	vanityMetricDataBuffer = vanityMetricDataBuffer[0:]
	return nil
}

// Returns a marshaled JSON of an empty VanityMetrics struct
func (vm *VanityMetricHandler) GetEmptyMetrics() ([]byte, error) {
	ret_val, err := json.Marshal(&VanityMetrics{})
	if err != nil {
		return nil, err
	}

	return ret_val, nil
}

// Returns a marshaled JSON of all custom metrics tracked through Stackdriver
func (vm *VanityMetricHandler) ListCustomMetrics(ctx context.Context, sd *metrics.StackDriverHandler, gcpProjectID string) ([]byte, error) {
	descFilter := `metric.type = starts_with("custom.googleapis.com/")`
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
		customMetrics[resp.DisplayName] = resp.Description
	}

	// Encode the map of custom metric names to descriptions
	ret_val, err := json.Marshal(customMetrics)
	if err != nil {
		return nil, err
	}

	return ret_val, nil
}

// Gets points for a metric for the last N duration
// given a metricServiceName (i.e. server_backend)
// and a metricID (i.e. session_update.latency_worse)
func (vm *VanityMetricHandler) GetPointDetails(ctx context.Context, sd *metrics.StackDriverHandler, gcpProjectID string, metricServiceName string, metricID string, duration time.Duration) ([]byte, error) {
	filter := fmt.Sprintf(`metric.type = "custom.googleapis.com/%s/%s"`, metricServiceName, metricID)
	name := fmt.Sprintf(`projects/%s/metricDescriptors/custom.googleapis.com/%s/%s`, gcpProjectID, metricServiceName, metricID)
	startTime := timestamppb.New(time.Now().Add(duration))
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:     name,
		Filter:   filter,
		Interval: &monitoringpb.TimeInterval{EndTime: timestamppb.Now(), StartTime: startTime},
		View:     monitoringpb.ListTimeSeriesRequest_TimeSeriesView(0),
	}
	it := sd.Client.ListTimeSeries(ctx, req)

	metricDetails := make(map[string][]*monitoringpb.Point)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		metricDetails[resp.GetMetric().GetType()] = resp.GetPoints()
	}

	// Encode the map of custom metric names to descriptions
	ret_val, err := json.Marshal(metricDetails)
	if err != nil {
		return nil, err
	}

	return ret_val, nil
}
