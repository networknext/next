package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/networknext/backend/modules-old/transport"
	"github.com/networknext/backend/modules-old/transport/middleware"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/rs/cors"
)

const (
	TOP_SESSIONS_COUNT            = 10
	STATS_DATASTORE               = "live_stats"
	TOP_SESSIONS_DATASTORE        = "top_sessions"
	LIVE_SESSION_COUNTS_DATASTORE = "live_session_counts"
)

var landingPageStatsMutex sync.RWMutex
var landingPageStats LookerStats

var liveSessionCountsMutex sync.RWMutex
var liveSessionCounts LiveSessionCounts

var topSessionsListMutex sync.RWMutex
var topSessionsList []TopSession

var lookerStatsRefreshInterval time.Duration
var liveStatsRefreshInterval time.Duration

var PLATFORM_TYPES = []string{
	"PS4",
	"PS5",
	"XBOX",
	"Switch",
	"Linux",
	"Mac",
	"PC",
}

var CONNECTION_TYPES = []string{
	"Wired",
	"WiFi",
	"Mobile",
}

func main() {
	service := common.CreateService("website_cruncher")

	lookerStatsRefreshInterval = envvar.GetDuration("LOOKER_STATS_REFRESH_INTERVAL", time.Hour*24)
	liveStatsRefreshInterval = envvar.GetDuration("LIVE_STATS_REFRESH_INTERVAL", time.Second*10)

	core.Log("looker stats refresh interval: %s", lookerStatsRefreshInterval)
	core.Log("live stats refresh interval: %s", liveStatsRefreshInterval)

	service.UseLooker()

	service.LeaderElection(false)

	StartRedisDataCollection(service)

	// StartLookerDataCollection(service)

	service.Router.HandleFunc("/stats", getAllStats())
	service.Router.HandleFunc("/sessions/counts", getLiveSessionCounts())
	service.Router.HandleFunc("/sessions/list", getTopSessionsList())

	service.StartWebServer()

	service.WaitForShutdown()
}

func getTopSessionsList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		middleware.CORSControlHandlerFunc(envvar.GetList("ALLOWED_ORIGINS", []string{}), w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(currentTopSessionsList()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func currentTopSessionsList() []TopSession {

	landingPageStatsMutex.RLock()
	topSessions := topSessionsList
	landingPageStatsMutex.RUnlock()

	return topSessions

}

func getLiveSessionCounts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		middleware.CORSControlHandlerFunc(envvar.GetList("ALLOWED_ORIGINS", []string{}), w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(currentLiveSessionCounts()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func currentLiveSessionCounts() LiveSessionCounts {

	landingPageStatsMutex.RLock()
	counts := liveSessionCounts
	landingPageStatsMutex.RUnlock()

	return counts

}

func getAllStats() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		middleware.CORSControlHandlerFunc(envvar.GetList("ALLOWED_ORIGINS", []string{}), w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(currentStats()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func currentStats() LookerStats {

	landingPageStatsMutex.RLock()
	stats := landingPageStats
	landingPageStatsMutex.RUnlock()

	return stats

}

// TODO - move to handlers or middleware or something - will be useful elsewhere
func CORSControlHandlerFunc(allowedOrigins []string, w http.ResponseWriter, r *http.Request) {
	cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods: []string{
			http.MethodPost,
			http.MethodGet,
			http.MethodOptions,
		},
	}).HandlerFunc(w, r)
}

// -----------------------------------------------------------------------------------------

// TODO: verify necessary / clean up
type LiveSessionCounts struct {
	TotalSessions int32 `json:"total_sessions"`
	TotalOnNext   int32 `json:"total_on_next"`
}

type Stats struct {
	RTT        float32 `json:"rtt"`
	Jitter     float32 `json:"jitter"`
	PacketLoss float32 `json:"packet_loss"`
}

type Envelope struct {
	Up   int32 `json:"up"`
	Down int32 `json:"down"`
}
type SessionSlice struct {
	Timestamp time.Time `json:"timestamp"`
	Next      Stats     `json:"next"`
	Direct    Stats     `json:"direct"`
	Envelope  Envelope  `json:"envelope"`
}
type SessionMeta struct {
	ISP            string `json:"isp"`
	Datacenter     string `json:"datacenter"`
	Platform       string `json:"platform"`
	ConnectionType string `json:"connection"`
	DirectRTT      int32  `json:"direct_rtt"`
	NextRTT        int32  `json:"next_rtt"`
	Improvement    int32  `json:"improvement"`
}
type SessionPoint struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type TopSession struct {
	Slices []SessionSlice `json:"slices"`
	Meta   SessionMeta    `json:"meta"`
}

func StartRedisDataCollection(service *common.Service) {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword := envvar.GetString("REDIS_PASSWORD", "")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHostname,
		Password: redisPassword,
	})

	ticker := time.NewTicker(liveStatsRefreshInterval)

	ctx := service.Context

	go func(ctx context.Context, client *redis.Client) {
		for {

			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:
				_, err := redisClient.Ping(ctx).Result()
				if err != nil {
					core.Error("failed to ping redis: %v", err)
					continue
				}

				// Top Sessions

				topSessions := make([]TopSession, TOP_SESSIONS_COUNT)         // TODO: use this structure for all data collection after transport is removed
				sessions := make([]transport.SessionMeta, TOP_SESSIONS_COUNT) // TODO: don't use transport structs

				minutes := time.Now().Unix() / 60

				topSessionsA, err := client.ZRevRange(ctx, fmt.Sprintf("s-%d", minutes-1), 0, TOP_SESSIONS_COUNT).Result()
				if err != nil {
					core.Error("failed to fetch top sessions group A: %v", err)
					continue
				}

				topSessionsB, err := client.ZRevRange(ctx, fmt.Sprintf("s-%d", minutes), 0, TOP_SESSIONS_COUNT).Result()
				if err != nil {
					core.Error("failed to fetch top sessions group B: %v", err)
					continue
				}

				core.Debug("%+v", topSessionsA)
				core.Debug("%+v", topSessionsB)

				metaPipeline := redisClient.Pipeline()
				defer metaPipeline.Close()

				cmdOutputs := make([]*redis.StringCmd, 0)

				sessionIDsRetreivedMap := make(map[string]bool)
				for _, sessionID := range topSessionsA {
					cmd := metaPipeline.Get(ctx, fmt.Sprintf("sm-%s", sessionID))
					cmdOutputs = append(cmdOutputs, cmd)
					sessionIDsRetreivedMap[sessionID] = true
				}
				for _, sessionID := range topSessionsB {
					if _, ok := sessionIDsRetreivedMap[sessionID]; !ok {
						cmd := metaPipeline.Get(ctx, fmt.Sprintf("sm-%s", sessionID))
						cmdOutputs = append(cmdOutputs, cmd)
						sessionIDsRetreivedMap[sessionID] = true
					}
				}

				core.Debug("%+v", sessionIDsRetreivedMap)

				cmds, err := metaPipeline.Exec(ctx)
				if err != nil {
					core.Error("failed to exec redis pipeline: %v", err)
					for _, cmd := range cmdOutputs {
						core.Debug("meta cmd err: %v", cmd.Err())
					}
					continue
				}

				var sessionMetasNext []transport.SessionMeta // TODO: avoid using transport structs
				var meta transport.SessionMeta               // TODO: avoid using transport structs
				for i := 0; i < len(sessionIDsRetreivedMap); i++ {
					metaString := cmds[i].String()

					core.Debug("meta string: %s", metaString)

					if metaString == "" {
						core.Error("meta data string is empty: %v", cmds[i].Err())
						continue
					}

					splitMetaStrings := strings.Split(metaString, "|")
					if err := meta.ParseRedisString(splitMetaStrings); err != nil {
						core.Error("failed to parse meta data string: %v", err)
						continue
					}

					sessionMetasNext = append(sessionMetasNext, meta)
				}

				sort.Slice(sessionMetasNext, func(i, j int) bool {
					return sessionMetasNext[i].DeltaRTT > sessionMetasNext[j].DeltaRTT
				})

				if len(sessionMetasNext) > TOP_SESSIONS_COUNT {
					sessions = sessionMetasNext[:TOP_SESSIONS_COUNT]
				} else {
					sessions = sessionMetasNext
				}

				var slice transport.SessionSlice // TODO: don't use transport
				for i := 0; i < len(sessions); i++ {

					currentSession := sessions[i]
					sessionID := currentSession.ID

					topSessions[i] = TopSession{
						Meta: SessionMeta{
							ISP:            currentSession.Location.ISP,
							Datacenter:     currentSession.DatacenterName,
							Platform:       PLATFORM_TYPES[currentSession.Platform],
							ConnectionType: CONNECTION_TYPES[currentSession.Connection],
							DirectRTT:      int32(currentSession.DirectRTT),
							NextRTT:        int32(currentSession.NextRTT),
							Improvement:    int32(currentSession.DirectRTT - currentSession.NextRTT),
						},
					}

					slices, err := redisClient.LRange(ctx, fmt.Sprintf("ss-%016x", sessionID), 0, -1).Result()
					if err != nil {
						core.Error("failed to look up slice data for session %016x: %v", sessionID, err)
						continue
					}

					topSessions[i].Slices = make([]SessionSlice, len(slices))

					for j := 0; j < len(slices); j++ {

						sliceStrings := strings.Split(slices[j], "|")
						if err := slice.ParseRedisString(sliceStrings); err != nil {
							core.Error("failed to parse slice string: %v", err)
							continue
						}

						topSessions[i].Slices[j] = SessionSlice{
							Timestamp: slice.Timestamp,
							Next: Stats{
								RTT:        float32(slice.Next.RTT),
								Jitter:     float32(slice.Next.Jitter),
								PacketLoss: float32(slice.Next.PacketLoss),
							},
							Direct: Stats{
								RTT:        float32(slice.Direct.RTT),
								Jitter:     float32(slice.Direct.Jitter),
								PacketLoss: float32(slice.Direct.PacketLoss),
							},
							Envelope: Envelope{
								Up:   int32(slice.Envelope.Up),
								Down: int32(slice.Envelope.Down),
							},
						}
					}
				}

				topSessionsListMutex.Lock()
				topSessionsList = topSessions
				topSessionsListMutex.Unlock()

				// Total Counts

				liveSessionCounts := LiveSessionCounts{}

				countsPipeline := redisClient.Pipeline()
				defer countsPipeline.Close()

				firstNextCounts, _, err := countsPipeline.Scan(ctx, 0, fmt.Sprintf("n-*-%d", minutes-1), -1).Result()
				if err != nil {
					core.Error("failed to get first set of counts: %v", err)
					continue
				}
				secondNextCounts, _, err := countsPipeline.Scan(ctx, 0, fmt.Sprintf("n-*-%d", minutes), -1).Result()
				if err != nil {
					core.Error("failed to get second set of counts: %v", err)
					continue
				}

				core.Debug("%+v", firstNextCounts)
				core.Debug("%+v", secondNextCounts)

				for i := 0; i < len(firstNextCounts); i++ {
					// TODO
				}

				for i := 0; i < len(secondNextCounts); i++ {
					// TODO
				}

				firstTotalCounts, err := countsPipeline.HGetAll(ctx, fmt.Sprintf("c-*-%d", minutes-1)).Result()
				if err != nil {
					core.Error("failed to get first set of counts: %v", err)
					continue
				}
				secondTotalCounts, err := countsPipeline.HGetAll(ctx, fmt.Sprintf("c-*-%d", minutes)).Result()
				if err != nil {
					core.Error("failed to get second set of counts: %v", err)
					continue
				}

				core.Debug("%+v", firstTotalCounts)
				core.Debug("%+v", secondTotalCounts)

				for i := 0; i < len(firstTotalCounts); i++ {
					// TODO
				}

				for i := 0; i < len(secondTotalCounts); i++ {
					// TODO
				}

				if err := updateDataStore(service, currentStats(), topSessions, liveSessionCounts); err != nil {
					core.Error("failed to update data store with new top sessions and counts: %v", err)
					continue
				}
			}
		}
	}(ctx, redisClient)
}

// ------------------------------------------------------------------------------------------

// TODO: update / verify necessary
type LookerStats struct {
	UniquePlayers             int32 `json:"unique_players"`
	AcceleratedPlayTime       int32 `json:"accelerated_play_time"`
	AcceleratedBandwidth      int32 `json:"accelerated_bandwidth"`
	UniquePlayersDelta        int32 `json:"unique_players_delta"`
	AcceleratedPlayTimeDelta  int32 `json:"accelerated_play_time_delta"`
	AcceleratedBandwidthDelta int32 `json:"accelerated_bandwidth_delta"`
}

func StartLookerDataCollection(service *common.Service) {

	ticker := time.NewTicker(lookerStatsRefreshInterval)

	go func() {

		for {

			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				stats := LookerStats{}

				// TODO: update stats using Looker

				if err := updateDataStore(service, stats, currentTopSessionsList(), currentLiveSessionCounts()); err != nil {
					core.Error("failed to update data store with new looker stats: %v", err)
					continue
				}
			}
		}
	}()
}

func updateDataStore(service *common.Service, stats LookerStats, topSessionsList []TopSession, liveSessionCounts LiveSessionCounts) error {

	var statsBuffer bytes.Buffer
	encoder := gob.NewEncoder(&statsBuffer)
	if err := encoder.Encode(stats); err != nil {
		return fmt.Errorf("failed to encode new looker stats")
	}

	var topSessionsBuffer bytes.Buffer
	encoder = gob.NewEncoder(&topSessionsBuffer)
	if err := encoder.Encode(topSessionsList); err != nil {
		return fmt.Errorf("failed to encode top sessions")
	}

	var sessionCountsBuffer bytes.Buffer
	encoder = gob.NewEncoder(&sessionCountsBuffer)
	if err := encoder.Encode(liveSessionCounts); err != nil {
		return fmt.Errorf("failed to encode session counts")
	}

	dataStores := []common.DataStoreConfig{
		{
			Name: STATS_DATASTORE,
			Data: statsBuffer.Bytes(),
		},
		{
			Name: TOP_SESSIONS_DATASTORE,
			Data: topSessionsBuffer.Bytes(),
		},
		{
			Name: LIVE_SESSION_COUNTS_DATASTORE,
			Data: sessionCountsBuffer.Bytes(),
		},
	}

	service.UpdateLeaderStore(dataStores)

	dataStores = service.LoadLeaderStore()

	newLookerStats := LookerStats{}
	newTopSesssions := make([]TopSession, 10)
	newSessionCounts := LiveSessionCounts{}

	decoder := gob.NewDecoder(bytes.NewBuffer(dataStores[0].Data))

	if err := decoder.Decode(&newLookerStats); err != nil {
		return fmt.Errorf("could not decode live stats data: %v", err)
	}

	landingPageStatsMutex.Lock()
	landingPageStats = newLookerStats
	landingPageStatsMutex.Unlock()

	decoder = gob.NewDecoder(bytes.NewBuffer(dataStores[1].Data))

	if err := decoder.Decode(&newTopSesssions); err != nil {
		return fmt.Errorf("could not decode top sessions data")
	}

	topSessionsListMutex.Lock()
	topSessionsList = newTopSesssions
	topSessionsListMutex.Unlock()

	decoder = gob.NewDecoder(bytes.NewBuffer(dataStores[2].Data))

	if err := decoder.Decode(&newSessionCounts); err != nil {
		return fmt.Errorf("could not decode session counts")
	}

	liveSessionCountsMutex.Lock()
	liveSessionCounts = newSessionCounts
	liveSessionCountsMutex.Unlock()

	return nil
}

// -----------------------------------------------------------------------------------------
