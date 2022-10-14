package main

// todo: need to rebulid this off the new service architecture, only the leader does the uptime check
// should write the uptime check results out to analytics service via pubsub

import (
	"net/http"
	"os"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/messages"
	"github.com/russellcardullo/go-pingdom/pingdom"
)

var googleProjectId string
var maxUptimeStatsChannelSize int
var maxUptimeMessageBytes int
var pingdomAPIToken string
var serverBackendHostname string
var portalFrontendHostname string
var uptimeStatsPubsubTopic string
var pingInterval time.Duration

func main() {
	service := common.CreateService("pingdom")

	googleProjectId = envvar.GetString("GOOGLE_PROJECT_ID", "")
	pingdomAPIToken = envvar.GetString("PINGDOM_API_TOKEN", "")
	maxUptimeMessageBytes = envvar.GetInt("MAX_UPTIME_STATS_MESSAGE_BYTES", messages.MaxUptimeStatsMessageBytes)
	maxUptimeStatsChannelSize = envvar.GetInt("MAX_UPTIME_STATS_CHANNEL_SIZE", 10*1024)
	serverBackendHostname = envvar.GetString("SERVER_BACKEND_HOSTNAME", "")
	portalFrontendHostname = envvar.GetString("PORTAL_FRONTEND_HOSTNAME", "")
	uptimeStatsPubsubTopic = envvar.GetString("UPTIME_STATS_PUBSUB_TOPIC", "uptime_stats")
	// pingdom has a resoultion of 1 minute and this keeps us from going over API quota
	pingInterval = envvar.GetDuration("PINGDOM_API_PING_INTERVAL", time.Minute*1)

	core.Log("google project Id: %s", googleProjectId)
	core.Log("pingdom api token: %s", pingdomAPIToken)
	core.Log("max uptime stats channel size: %d", maxUptimeStatsChannelSize)
	core.Log("server backend hostname: %s", serverBackendHostname)
	core.Log("portal frontend hostname: %s", portalFrontendHostname)

	service.LeaderElection(true)

	StartUptimeCheck(service)

	service.StartWebServer()

	service.WaitForShutdown()
}

func StartUptimeCheck(service *common.Service) {

	pingdomClient, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
		APIToken: pingdomAPIToken,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	})
	if err != nil {
		core.Error("failed to set up pingdom client: %v", err)
		os.Exit(1)
	}

	producerConfig := common.GooglePubsubConfig{
		ProjectId:          googleProjectId,
		Topic:              uptimeStatsPubsubTopic,
		MessageChannelSize: maxUptimeStatsChannelSize,
	}

	producer, err := common.CreateGooglePubsubProducer(service.Context, producerConfig)
	if err != nil {
		core.Error("could not create google pubsub producer: %v", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case <-service.Context.Done():
			return
		case <-ticker.C:

			if !service.IsLeader() {
				continue
			}

			responses, err := pingdomClient.Checks.List()
			if err != nil {
				core.Error("failed to get check list from pingdom: %v", err)
				os.Exit(1)
			}

			for _, response := range responses {

				uptimeMessage := messages.UptimeStatsMessage{
					Timestamp:    uint64(time.Now().Unix()),
					ServiceName:  response.Name,
					Up:           response.Status == "up",
					ResponseTime: int(response.LastResponseTime),
				}

				messageBuffer := make([]byte, maxUptimeMessageBytes)

				message := uptimeMessage.Write(messageBuffer)

				producer.MessageChannel <- message
			}
		}
	}
}
