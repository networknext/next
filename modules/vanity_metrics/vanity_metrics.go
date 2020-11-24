package vanity_metrics

import (
	"context"
	"fmt"
	// "io/ioutil"
	// "os"
	// "os/signal"
	// "runtime"
	// "strconv"
	// "time"
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	// "github.com/gorilla/mux"
	// "google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/networknext/backend/modules/metrics"
	// "github.com/networknext/backend/modules/transport"
)

type VanityMetricHandler struct {
	Handler 	  	*metrics.StackDriverHandler
	Logger 			log.Logger
}

type VanityMetrics struct {
	NumSlicesGlobal			int
	NumSlicesPerCustomer	int
	NumSessionsGlobal		int
	NumSessionsPerCustomer	int
	NumPlayHours			int
}

func (vm *VanityMetricHandler) GetEmptyMetrics() ([]byte, error) {
	ret_val, err := json.Marshal(&VanityMetrics{})
	if err != nil {
		level.Error(vm.Logger).Log("err", err)
		return nil, err
	}

	return ret_val, nil
}


func (vm *VanityMetricHandler) ListCustomMetrics(ctx context.Context) ([]byte, error) {
	// filter := `metric.type = "custom.googleapis.com/server_backend/session_update.latency_worse"`
	// startTime := timestamppb.New(time.Now().Add(-10 * time.Minute))
	// // aggr := &monitoringpb.Aggregation
	// req := &monitoringpb.ListTimeSeriesRequest{
	// 	Name: "projects/network-next-v3-dev/metricDescriptors/custom.googleapis.com/server_backend/session_update.latency_worse",
	// 	Filter: filter,
	// 	Interval: &monitoringpb.TimeInterval{EndTime: timestamppb.Now(), StartTime: startTime},
	// 	View: monitoringpb.ListTimeSeriesRequest_TimeSeriesView(0),

	// }
	// it := vm.Handler.Client.ListTimeSeries(ctx, req)
	// for {
	// 	resp, err := it.Next()
	// 	if err == iterator.Done {
	// 		break
	// 	}
	// 	if err != nil {
	// 		level.Error(vm.Logger).Log("err", err)
	// 		return nil, err
	// 	}		
	// }

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