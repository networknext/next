package main

import (
	"fmt"
	// todo
	/*
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/envvar"
	"github.com/rs/cors"
	*/
)

func main() {
	fmt.Printf("hello website cruncher world. old modules are no longer allowed\n")
}

/*
const (
	TOP_SESSIONS_COUNT            = 10
	STATS_DATASTORE               = "live_stats"
	TOP_SESSIONS_DATASTORE        = "top_sessions"
	LIVE_SESSION_COUNTS_DATASTORE = "live_session_counts"
)

var landingPageStatsMutex sync.RWMutex
var landingPageStats common.LookerStats

var landingPageTopSessionsMutex sync.RWMutex
var landingPageTopSessions []TopSession

var landingPageSessionCountsMutex sync.RWMutex
var landingPageSessionCounts LiveSessionCounts

var lookerStatsRefreshInterval time.Duration
var redisStatsRefreshInterval time.Duration

func main() {
	service := common.CreateService("website_cruncher")

	lookerStatsRefreshInterval = envvar.GetDuration("LOOKER_STATS_REFRESH_INTERVAL", time.Second*30)
	redisStatsRefreshInterval = envvar.GetDuration("REDIS_STATS_REFRESH_INTERVAL", time.Second*10)

	core.Log("looker stats refresh interval: %s", lookerStatsRefreshInterval)
	core.Log("redis stats refresh interval: %s", redisStatsRefreshInterval)

	service.LeaderElection(false)

	service.UpdateLeaderStore([]common.DataStoreConfig{
		{
			Name: STATS_DATASTORE,
			Data: make([]byte, 0),
		},
		{
			Name: TOP_SESSIONS_DATASTORE,
			Data: make([]byte, 0),
		},
		{
			Name: LIVE_SESSION_COUNTS_DATASTORE,
			Data: make([]byte, 0),
		},
	})

	service.LoadLeaderStore()

	service.LoadDatabase()

	service.UseLooker()

	service.Router.HandleFunc("/stats", getAllStats())
	service.Router.HandleFunc("/sessions/counts", getLiveSessionCounts())
	service.Router.HandleFunc("/sessions/list", getTopSessionsList())

	service.StartWebServer()

	StartRedisDataCollection(service)

	StartLookerDataCollection(service)

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

	landingPageTopSessionsMutex.RLock()
	topSessions := landingPageTopSessions
	landingPageTopSessionsMutex.RUnlock()

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

	landingPageSessionCountsMutex.RLock()
	counts := landingPageSessionCounts
	landingPageSessionCountsMutex.RUnlock()

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

func currentStats() common.LookerStats {

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

	ticker := time.NewTicker(redisStatsRefreshInterval)

	ctx := service.Context

	buyerMap := service.Database().BuyerMap

	go func(ctx context.Context, client *redis.Client, buyerMap map[uint64]database.Buyer) {
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

				metaPipeline := redisClient.Pipeline()

				sessionIdsRetreivedMap := make(map[string]bool)
				for _, sessionId := range topSessionsA {
					metaPipeline.Get(ctx, fmt.Sprintf("sm-%s", sessionId))
					sessionIdsRetreivedMap[sessionId] = true
				}
				for _, sessionId := range topSessionsB {
					if _, ok := sessionIdsRetreivedMap[sessionId]; !ok {
						metaPipeline.Get(ctx, fmt.Sprintf("sm-%s", sessionId))
						sessionIdsRetreivedMap[sessionId] = true
					}
				}

				cmds, err := metaPipeline.Exec(ctx)
				if err != nil {
					core.Error("failed to exec redis pipeline: %v", err)
					continue
				}

				var sessionMetasNext []transport.SessionMeta // TODO: avoid using transport structs
				var meta transport.SessionMeta               // TODO: avoid using transport structs
				for _, cmd := range cmds {
					metaString := cmd.(*redis.StringCmd).Val()

					if metaString == "" {
						core.Error("meta data string is empty: %v", cmd.Err())
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

				// todo: we don't want to use old module transport stuff here
				var slice transport.SessionSlice
				for i := 0; i < len(sessions); i++ {

					currentSession := sessions[i]
					sessionId := currentSession.ID

					topSessions[i] = TopSession{
						Meta: SessionMeta{
							ISP:            currentSession.Location.ISP,
							Datacenter:     currentSession.DatacenterName,
							Platform:       transport.PlatformTypeText(currentSession.Platform),     // TODO: don't use transport
							ConnectionType: transport.ConnectionTypeText(currentSession.Connection), // TODO: don't use transport
							DirectRTT:      int32(currentSession.DirectRTT),
							NextRTT:        int32(currentSession.NextRTT),
							Improvement:    int32(currentSession.DirectRTT - currentSession.NextRTT),
						},
					}

					slices, err := redisClient.LRange(ctx, fmt.Sprintf("ss-%016x", sessionId), 0, -1).Result()
					if err != nil {
						core.Error("failed to look up slice data for session %016x: %v", sessionId, err)
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

				metaPipeline.Close()

				// Total Counts

				liveSessionCounts := LiveSessionCounts{}
				firstNextCount := 0
				secondNextCount := 0
				firstTotalCount := 0
				secondTotalCount := 0

				countsPipeline := redisClient.Pipeline()

				for _, buyer := range buyerMap {
					buyerId := fmt.Sprintf("%016x", buyer.Id)

					countsPipeline.HLen(ctx, fmt.Sprintf("n-%s-%d", buyerId, minutes-1))
					if err != nil {
						core.Error("failed to get first set of next counts: %v", err)
						continue
					}
					countsPipeline.HLen(ctx, fmt.Sprintf("n-%s-%d", buyerId, minutes))
					if err != nil {
						core.Error("failed to get second set of next counts: %v", err)
						continue
					}
				}

				// TODO: go back over this and see if we can avoid the string parsing
				cmds, err = countsPipeline.Exec(ctx)
				if err != nil {
					core.Error("failed to exec redis pipeline: %v", err)
					continue
				}

				for _, cmd := range cmds {
					countString := cmd.String()
					count := cmd.(*redis.IntCmd).Val()

					if countString == "" {
						core.Error("count output is empty: %v", cmd.Err())
						continue
					}

					if strings.Contains(countString, fmt.Sprintf("%d", minutes-1)) {
						firstNextCount += int(count)
					} else {
						secondNextCount += int(count)
					}
				}

				liveSessionCounts.TotalOnNext = int32(firstNextCount)
				if secondNextCount > firstNextCount {
					liveSessionCounts.TotalOnNext = int32(secondNextCount)
				}

				for _, buyer := range buyerMap {
					buyerId := fmt.Sprintf("%016x", buyer.Id)
					countsPipeline.HVals(ctx, fmt.Sprintf("c-%s-%d", buyerId, minutes-1))
					if err != nil {
						core.Error("failed to get first set of next counts: %v", err)
						continue
					}
					countsPipeline.HVals(ctx, fmt.Sprintf("c-%s-%d", buyerId, minutes))
					if err != nil {
						core.Error("failed to get second set of next counts: %v", err)
						continue
					}
				}

				cmds, err = countsPipeline.Exec(ctx)
				if err != nil {
					core.Error("failed to exec redis pipeline: %v", err)
					continue
				}

				for _, cmd := range cmds {
					countString := cmd.String()
					counts := cmd.(*redis.StringSliceCmd).Val()

					if countString == "" {
						core.Error("count output is empty: %v", cmd.Err())
						continue
					}

					for _, val := range counts {
						if val == "" {
							continue
						}

						parseInt, err := strconv.ParseInt(val, 10, 64)
						if err != nil {
							core.Error("failed to parse int: %v", err)
							continue
						}

						if strings.Contains(countString, fmt.Sprintf("%d", minutes-1)) {
							firstTotalCount += int(parseInt)
						} else {
							secondTotalCount += int(parseInt)
						}
					}
				}

				totalCounts := int32(firstTotalCount)

				if secondTotalCount > firstTotalCount {
					totalCounts = int32(secondTotalCount)
				}

				liveSessionCounts.TotalSessions = totalCounts

				countsPipeline.Close()

				if err := updateDataStore(service, currentStats(), topSessions, liveSessionCounts); err != nil {
					core.Error("failed to update data store with new top sessions and counts: %v", err)
					continue
				}
			}
		}
	}(ctx, redisClient, buyerMap)
}

// ------------------------------------------------------------------------------------------

func StartLookerDataCollection(service *common.Service) {

	ticker := time.NewTicker(lookerStatsRefreshInterval)

	go func(service *common.Service) {

		for {

			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				stats, err := service.FetchWebsiteStats()
				if err != nil {
					core.Error("failed to fetch website stats from Looker: %v", err)
					continue
				}

				// TODO: add in extrapolation

				if err := updateDataStore(service, stats, currentTopSessionsList(), currentLiveSessionCounts()); err != nil {
					core.Error("failed to update data store with new looker stats: %v", err)
					continue
				}
			}
		}
	}(service)
}

func updateDataStore(service *common.Service, stats common.LookerStats, topSessionsList []TopSession, liveSessionCounts LiveSessionCounts) error {

	var statsBuffer bytes.Buffer
	statsEncoder := gob.NewEncoder(&statsBuffer)
	if err := statsEncoder.Encode(stats); err != nil {
		return fmt.Errorf("failed to encode new looker stats")
	}

	var topSessionsBuffer bytes.Buffer
	sessionsListEncoder := gob.NewEncoder(&topSessionsBuffer)
	if err := sessionsListEncoder.Encode(topSessionsList); err != nil {
		return fmt.Errorf("failed to encode top sessions")
	}

	var sessionCountsBuffer bytes.Buffer
	sessionCountsEncoder := gob.NewEncoder(&sessionCountsBuffer)
	if err := sessionCountsEncoder.Encode(liveSessionCounts); err != nil {
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

	if len(dataStores) == 0 {
		return fmt.Errorf("no data stores returned from redis")
	}

	newLookerStats := common.LookerStats{}
	newTopSesssions := make([]TopSession, 10)
	newSessionCounts := LiveSessionCounts{}

	statsDecoder := gob.NewDecoder(bytes.NewBuffer(dataStores[0].Data))

	if err := statsDecoder.Decode(&newLookerStats); err != nil {
		return fmt.Errorf("could not decode live stats data: %v", err)
	}

	landingPageStatsMutex.Lock()
	landingPageStats = newLookerStats
	landingPageStatsMutex.Unlock()

	listDecoder := gob.NewDecoder(bytes.NewBuffer(dataStores[1].Data))

	if err := listDecoder.Decode(&newTopSesssions); err != nil {
		return fmt.Errorf("could not decode top sessions data")
	}

	landingPageTopSessionsMutex.Lock()
	landingPageTopSessions = newTopSesssions
	landingPageTopSessionsMutex.Unlock()

	countsDecoder := gob.NewDecoder(bytes.NewBuffer(dataStores[2].Data))

	if err := countsDecoder.Decode(&newSessionCounts); err != nil {
		return fmt.Errorf("could not decode session counts")
	}

	landingPageSessionCountsMutex.Lock()
	landingPageSessionCounts = newSessionCounts
	landingPageSessionCountsMutex.Unlock()

	return nil
}

// -----------------------------------------------------------------------------------------
*/