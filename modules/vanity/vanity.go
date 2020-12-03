package vanity

import (
	"context"
	"fmt"
	"time"
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/encoding"
)

type VanityMetricHandler struct {
	Handler 	  	*metrics.StackDriverHandler
	Logger 			log.Logger
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

func (v VanityMetrics) UnmarshalBinary(data []byte) eror {
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