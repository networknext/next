package messages

import (
	"github.com/hamba/avro"
	"os"
	"testing"
)

func TestAnalyticsSessionUpdateMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/session_update.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsSessionUpdateMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsSessionUpdateMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestAnalyticsSessionSummaryMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/session_summary.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsSessionSummaryMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsSessionSummaryMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestAnalyticsServerUpdateMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/server_update.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsServerUpdateMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsServerUpdateMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestAnalyticsServerInitMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/server_init.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsServerInitMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsServerInitMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestAnalyticsRelayUpdateMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/relay_update.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsRelayUpdateMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsRelayUpdateMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestRouteMatrixUpdateMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/route_matrix_update.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsRouteMatrixUpdateMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsRouteMatrixUpdateMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestClientRelayPingMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/client_relay_ping.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsClientRelayPingMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsClientRelayPingMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestServerRelayPingMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/server_relay_ping.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsServerRelayPingMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsServerRelayPingMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}

func TestRelayToRelayPingMessage(t *testing.T) {

	t.Parallel()

	schemaData, err := os.ReadFile("../../schemas/pubsub/relay_to_relay_ping.json")
	if err != nil {
		panic(err)
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		panic(err)
	}

	in := AnalyticsRelayToRelayPingMessage{}

	data, err := avro.Marshal(schema, in)
	if err != nil {
		panic(err)
	}

	out := AnalyticsRelayToRelayPingMessage{}
	err = avro.Unmarshal(schema, data, &out)
	if err != nil {
		panic(err)
	}
}
