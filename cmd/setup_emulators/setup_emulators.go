package main

import (
    "context"
    "encoding/json"
    "io/ioutil"
    "os"

    "cloud.google.com/go/bigquery"
    "cloud.google.com/go/pubsub"
    "github.com/networknext/backend/modules/core"
    "github.com/networknext/backend/modules/envvar"
    "google.golang.org/api/option"
)

const (
    REPEATED = "REPEATED"
    REQUIRED = "REQUIRED"
)

type BigquerySchema struct {
    Projects []BigqueryProject `json:"projects"`
}

type BigqueryProject struct {
    ID       string            `json:"id"`
    Datasets []BigqueryDataset `json:"datasets"`
}

type BigqueryDataset struct {
    ID     string          `json:"id"`
    Tables []BigqueryTable `json:"tables"`
}

type BigqueryTable struct {
    ID      string           `json:"id"`
    Columns []BigqueryColumn `json:"columns"`
}

type BigqueryColumn struct {
    Name string             `json:"name"`
    Type bigquery.FieldType `json:"type"`
    Mode string             `json:"mode"`
}

type PubsubMessageType struct {
    Topic        string
    Subscription string
}

func main() {
    ctx := context.Background()

    googleProjectID := envvar.GetString("GOOGLE_PROJECT_ID", "local")

    pubsubSetupClient, err := pubsub.NewClient(ctx, googleProjectID)
    if err != nil {
        core.Error("failed to create pubsub setup client: %v", err)
        os.Exit(1)
    }

    pubsubMessages := []PubsubMessageType{
        {
            Topic:        envvar.GetString("COST_MATRIX_STATS_PUBSUB_TOPIC", "cost_matrix_stats"),
            Subscription: envvar.GetString("COST_MATRIX_STATS_PUBSUB_SUBSCRIPTION", "cost_matrix_stats"),
        },
        {
            Topic:        envvar.GetString("ROUTE_MATRIX_STATS_PUBSUB_TOPIC", "route_matrix_stats"),
            Subscription: envvar.GetString("ROUTE_MATRIX_STATS_PUBSUB_SUBSCRIPTION", "route_matrix_stats"),
        },
        {
            Topic:        envvar.GetString("PING_STATS_PUBSUB_TOPIC", "ping_stats"),
            Subscription: envvar.GetString("PING_STATS_PUBSUB_SUBSCRIPTION", "ping_stats"),
        },
        {
            Topic:        envvar.GetString("RELAY_STATS_PUBSUB_TOPIC", "relay_stats"),
            Subscription: envvar.GetString("RELAY_STATS_PUBSUB_SUBSCRIPTION", "relay_stats"),
        },
        {
            Topic:        envvar.GetString("BILLING_PUBSUB_TOPIC", "billing"),
            Subscription: envvar.GetString("BILLING_PUBSUB_SUBSCRIPTION", "billing"),
        },
        {
            Topic:        envvar.GetString("SUMMARY_PUBSUB_TOPIC", "summary"),
            Subscription: envvar.GetString("SUMMARY_PUBSUB_SUBSCRIPTION", "summary"),
        },
    }

    for i := 0; i < len(pubsubMessages); i++ {
        messageType := pubsubMessages[i]

        pubsubSetupClient.CreateTopic(ctx, messageType.Topic)
        pubsubSetupClient.CreateSubscription(ctx, messageType.Subscription, pubsub.SubscriptionConfig{
            Topic: pubsubSetupClient.Topic(messageType.Topic),
        })
    }

    pubsubSetupClient.Close()

    // ----------------

    clientOptions := []option.ClientOption{
        option.WithEndpoint("http://127.0.0.1:9050"),
        option.WithoutAuthentication(),
    }

    schemaFile, err := os.Open(envvar.GetString("BIGQUERY_SCHEMA_FILE", "./testdata/bigquery_emulator/happy_path_tables.json"))
    if err != nil {
        core.Error("failed to open schema file: %v", err)
        return
    }

    defer schemaFile.Close()

    schemaBytes, err := ioutil.ReadAll(schemaFile)
    if err != nil {
        core.Error("failed to process schema file")
        return
    }

    bigquerySchema := BigquerySchema{}

    if err := json.Unmarshal([]byte(schemaBytes), &bigquerySchema); err != nil {
        core.Error("failed to unmarshal schema")
        return
    }

    for _, project := range bigquerySchema.Projects {

        bigqueryClient, err := bigquery.NewClient(ctx, project.ID, clientOptions...)
        if err != nil {
            core.Error("failed to create bigquery client for %s project", project.ID)
            continue
        }
        defer bigqueryClient.Close()

        for _, dataset := range project.Datasets {

            bigqueryClient.Dataset(dataset.ID).Create(ctx, &bigquery.DatasetMetadata{})

            for _, table := range dataset.Tables {

                tableMetaData := bigquery.TableMetadata{
                    Schema: make(bigquery.Schema, len(table.Columns)),
                }

                for i, column := range table.Columns {
                    tableMetaData.Schema[i] = &bigquery.FieldSchema{
                        Name:     column.Name,
                        Type:     column.Type,
                        Required: column.Mode == REQUIRED,
                        Repeated: column.Mode == REPEATED,
                    }
                }

                bigqueryClient.Dataset(dataset.ID).Table(table.ID).Create(ctx, &tableMetaData)
            }
        }
    }
}
