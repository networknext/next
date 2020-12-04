package vanity

import (
	"context"
	"fmt"
	"time"
	"encoding/json"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/encoding"
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
	handler 	  	*metrics.StackDriverHandler
	metrics 		*metrics.VanityMetricMetrics
	subscriber		pubsub.Subscriber
	logger 			log.Logger
}

type VanityMetrics struct {
	BuyerID					uint64
	NumSlicesGlobal			uint32
	NumSlicesPerCustomer	uint32
	NumSessionsGlobal		uint32
	NumSessionsPerCustomer	uint32
	NumPlayHours			uint32
	RTTReduction			float32
	PacketLossReduction		float32
}

func NewVanityMetricHandler(vanityHandler *metrics.StackDriverHandler, vanityMetricMetrics *metrics.VanityMetricMetrics, vanitySubscriber pubsub.Subscriber, vanityLogger log.Logger) VanityMetricHandler {
	return &VanityMetricHandler{handler: vanityHandler, metrics: vanityMetricMetrics, subscriber: vanitySubscriber, logger: vanityLogger}
}

func (v VanityMetrics) Size() uint64 {
	sum = 8 + // BuyerID
		  4 + // NumSlicesGlobal
		  4 + // NumSlicesPerCustomer 
		  4 + // NumSessionsGlobal
		  4 + // NumSessionsPerCustomer
		  4 + // NumPlayHours
		  4 + // RTTReduction
		  4   // PacketLossReduction

	return sum
}

func (v VanityMetrics) MarshalBinary() ([]byte, error) {
	data := make([]byte, v.Size())
	index := 0

	encoding.WriteUint64(data &index, v.BuyerID)

	encoding.WriteUint32(data, &index, v.NumSlicesGlobal)
	encoding.WriteUint32(data, &index, v.NumSlicesPerCustomer)
	encoding.WriteUint32(data, &index, v.NumSessionsGlobal)
	encoding.WriteUint32(data, &index, v.NumSessionsPerCustomer)
	encoding.WriteUint32(data, &index, v.NumPlayHours)

	encoding.WriteFloat32(data, &index, v.RTTReduction)
	encoding.WriteFloat32(data, &index, v.PacketLossReduction)

	return data, nil
}

func (v VanityMetrics) UnmarshalBinary(data []byte) error {
	index := 0

	if !encoding.ReadUint64(data, &index, &v.BuyerID) {
		return erorrs.New("[VanityMetrics] invalid read at buyer ID")
	

	if !encoding.ReadUint32(data, &index, &v.NumSlicesGlobal) {
		return erorrs.New("[VanityMetrics] invalid read at num slices global")
	}

	if !encoding.ReadUint32(data, &index, &v.NumSlicesPerCustomer) {
		return erorrs.New("[VanityMetrics] invalid read at num slices per customer")
	}

	if !encoding.ReadUint32(data, &index, &v.NumSessionsGlobal) {
		return erorrs.New("[VanityMetrics] invalid read at num sessions global")
	}

	if !encoding.ReadUint32(data, &index, &v.NumSessionsPerCustomer) {
		return erorrs.New("[VanityMetrics] invalid read at num sesions per customer")
	}

	if !encoding.ReadUint32(data, &index, &v.NumPlayHours) {
		return erorrs.New("[VanityMetrics] invalid read at num play hours")
	}

	if !encoding.ReadUint32(data, &index, &v.NumSlicesGlobal) {
		return erorrs.New("[VanityMetrics] invalid read at num slices global")
	}

	if !encoding.ReadFloat32(data, &index, &v.RTTReduction) {
		return erorrs.New("[VanityMetrics] invalid read at RTT reduction")
	}

	if !encoding.ReadFloat32(data, &index, &v.PacketLossReduction) {
		return erorrs.New("[VanityMetrics] invalid read at packet loss reduction")
	}

	return nil
}

func (vm *VanityMetricHandler) Start(ctx context.Context, numVanityWriteGoroutines int) error {
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

	// Start the goroutines to write to StackDriver
	for i := 0; i < numVanityWriteGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each goroutine has its own buffer to avoid syncing
			vanityMetricDataBuffer := make([]VanityMetrics, 0)

			for {
				select {
				// Buffer up some vanity metric entries to insert into StackDriver
				case vanityData := <-vm.vanityMetricDataChan:
					vanityMetricDataBuffer = append(vanityMetricDataBuffer, vanityData)

					if err := vm.WriteToStackDriver(ctx, vanityMetricDataBuffer); err != nil {
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
		case pubsub.TopicVanityData:
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

// Writes messages to StackDriver
func (vm *VanityMetricHandler) WriteToStackDriver(ctx context.Context, vanityMetricDataBuffer []VanityMetrics) error {


}

// Returns a marshaled JSON of an empty VanityMetrics struct
func (vm *VanityMetricHandler) GetEmptyMetrics() ([]byte, error) {
	ret_val, err := json.Marshal(&VanityMetrics{})
	if err != nil {
		level.Error(vm.Logger).Log("err", err)
		return nil, err
	}

	return ret_val, nil
}

// Returns a marshaled JSON of all custom metrics tracked through Stackdriver
func (vm *VanityMetricHandler) ListCustomMetrics(ctx context.Context) ([]byte, error) {
	descFilter := `metric.type = starts_with("custom.googleapis.com/")`
	descReq := &monitoringpb.ListMetricDescriptorsRequest{
		Name: "projects/" + vm.Handler.ProjectID,
		Filter: descFilter,
	}
	descIt := vm.Handler.Client.ListMetricDescriptors(ctx, descReq)

	customMetrics := make(map[string]string)
	for {
		resp, err := descIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			level.Error(vm.Logger).Log("err", err)
			return nil, err
		}
		customMetrics[resp.DisplayName] = resp.Description
	}

	// Encode the map of custom metric names to descriptions
	ret_val, err := json.Marshal(customMetrics)
	if err != nil {
		level.Error(vm.Logger).Log("err", err)
		return nil, err
	}

	return ret_val, nil
}

// Gets points for a metric for the last N duration
// given a metricServiceName (i.e. server_backend)
// and a metricID (i.e. session_update.latency_worse)
func (vm *VanityMetricHandler) GetPointDetails(ctx context.Context, gcpProjectID string, metricServiceName string, metricID string, duration time.Duration) ([]byte, error) {
	filter := fmt.Sprintf(`metric.type = "custom.googleapis.com/%s/%s"`, metricServiceName, metricID)
	name := fmt.Sprintf(`projects/%s/metricDescriptors/custom.googleapis.com/%s/%s`, gcpProjectID, metricServiceName, metricID)
	startTime := timestamppb.New(time.Now().Add(duration))
	req := &monitoringpb.ListTimeSeriesRequest{
		Name: name,
		Filter: filter,
		Interval: &monitoringpb.TimeInterval{EndTime: timestamppb.Now(), StartTime: startTime},
		View: monitoringpb.ListTimeSeriesRequest_TimeSeriesView(0),

	}
	it := vm.Handler.Client.ListTimeSeries(ctx, req)

	metricDetails := make(map[string][]*monitoringpb.Point)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			level.Error(vm.Logger).Log("err", err)
			return nil, err
		}
		metricDetails[resp.GetMetric().GetType()] = resp.GetPoints()		
	}

	// Encode the map of custom metric names to descriptions
	ret_val, err := json.Marshal(metricDetails)
	if err != nil {
		level.Error(vm.Logger).Log("err", err)
		return nil, err
	}

	return ret_val, nil
}