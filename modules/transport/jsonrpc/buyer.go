package jsonrpc

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigtable"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gomodule/redigo/redis"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/middleware"

	ghostarmy "github.com/networknext/backend/modules/ghost_army"
)

const (
	TopSessionsSize          = 1000
	MapPointByteCacheVersion = uint8(1)
)

var (
	ErrInsufficientPrivileges = errors.New("insufficient privileges")
)

type BuyersService struct {
	mu                      sync.Mutex
	mapPointsByteCache      []byte
	mapPointsBuyerByteCache map[string][]byte

	mapPointsCache        json.RawMessage
	mapPointsCompactCache json.RawMessage

	mapPointsBuyerCache        map[string]json.RawMessage
	mapPointsCompactBuyerCache map[string]json.RawMessage

	Env string

	UseBigtable     bool
	BigTableCfName  string
	BigTable        *storage.BigTable
	BigTableMetrics *metrics.BigTableMetrics

	BqClient *bigquery.Client

	RedisPoolTopSessions   *redis.Pool
	RedisPoolSessionMeta   *redis.Pool
	RedisPoolSessionSlices *redis.Pool
	RedisPoolSessionMap    *redis.Pool
	RedisPoolUserSessions  *redis.Pool

	Metrics *metrics.BuyerEndpointMetrics
	Storage storage.Storer
	Logger  log.Logger
}

type FlushSessionsArgs struct{}

type FlushSessionsReply struct{}

func (s *BuyersService) FlushSessions(r *http.Request, args *FlushSessionsArgs, reply *FlushSessionsReply) error {
	if !middleware.VerifyAllRoles(r, middleware.OpsRole) {
		return fmt.Errorf("FlushSessions(): %v", ErrInsufficientPrivileges)
	}

	topSessions := s.RedisPoolTopSessions.Get()
	defer topSessions.Close()
	if _, err := topSessions.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	sessionMeta := s.RedisPoolSessionMeta.Get()
	defer sessionMeta.Close()
	if _, err := sessionMeta.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	sessionSlices := s.RedisPoolSessionSlices.Get()
	defer sessionSlices.Close()
	if _, err := sessionSlices.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	sessionMap := s.RedisPoolSessionMap.Get()
	defer sessionMap.Close()
	if _, err := sessionMap.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	return nil
}

type UserSessionsArgs struct {
	UserID string `json:"user_id"`
}

type UserSessionsReply struct {
	Sessions   []transport.SessionMeta `json:"sessions"`
	TimeStamps []time.Time             `json:"time_stamps"`
}

func (s *BuyersService) UserSessions(r *http.Request, args *UserSessionsArgs, reply *UserSessionsReply) error {
	if args.UserID == "" {
		err := fmt.Errorf("UserSessions() user id is required")
		level.Error(s.Logger).Log("err", err)
		return err
	}
	reply.Sessions = make([]transport.SessionMeta, 0)
	sessionIDs := make([]string, 0)

	var sessionSlice transport.SessionSlice

	// Raw user input
	userID := args.UserID

	// Hex of the userID in case it's a signed decimal hash
	var hexUserID string
	{
		userIDInt, err := strconv.Atoi(userID)
		if err == nil {
			// userID was an int that we need to convert to hex and lookup
			hexUserID = fmt.Sprintf("%016x", userIDInt)
		}
	}

	// Hash the ID
	hash := fnv.New64a()
	_, err := hash.Write([]byte(userID))
	if err != nil {
		err = fmt.Errorf("UserSessions() error writing 64a hash: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}
	userHash := fmt.Sprintf("%016x", hash.Sum64())

	// Fetch live sessions if there are any
	liveSessions, err := s.FetchCurrentTopSessions(r, "")
	if err != nil {
		err = fmt.Errorf("UserSessions() failed to fetch live sessions")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	for _, session := range liveSessions {
		// Check both the ID, hex of ID, and the hash just in case the ID is actually a hash from the top sessions table or decimal hash
		if userHash == fmt.Sprintf("%016x", session.UserHash) || userID == fmt.Sprintf("%016x", session.UserHash) || hexUserID == fmt.Sprintf("%016x", session.UserHash) {
			sessionSlicesClient := s.RedisPoolSessionSlices.Get()
			defer sessionSlicesClient.Close()

			slices, err := redis.Strings(sessionSlicesClient.Do("LRANGE", fmt.Sprintf("ss-%016x", session.ID), "0", "0"))
			if err != nil && err != redis.ErrNil {
				err = fmt.Errorf("UserSessions() failed getting session slices: %v", err)
				level.Error(s.Logger).Log("err", err)
				err = fmt.Errorf("UserSessions() failed getting session slices")
				return err
			}

			// If a slice exists, add the session and the timestamp
			if len(slices) > 0 {
				sliceString := strings.Split(slices[0], "|")
				if err := sessionSlice.ParseRedisString(sliceString); err != nil {
					err = fmt.Errorf("UserSessions() SessionSlice parsing error: %v", err)
					level.Error(s.Logger).Log("err", err)
					return err
				}

				sessionIDs = append(sessionIDs, fmt.Sprintf("%016x", session.ID))

				buyer, err := s.Storage.Buyer(session.BuyerID)
				if err != nil {
					err = fmt.Errorf("UserSessions() failed to fetch buyer: %v", err)
					level.Error(s.Logger).Log("err", err)
					return err
				}

				if !middleware.VerifyAllRoles(r, s.SameBuyerRole(buyer.CompanyCode)) {
					session.Anonymise()
				}

				reply.Sessions = append(reply.Sessions, session)
				reply.TimeStamps = append(reply.TimeStamps, sessionSlice.Timestamp.UTC())
			} else {
				// Increment counter
				s.Metrics.NoSlicesFailure.Add(1)
			}
		}
	}

	if s.UseBigtable {
		// Create the filters to use for reading rows
		chainFilter := bigtable.ChainFilters(bigtable.ColumnFilter("meta"),
			bigtable.LatestNFilter(1),
		)

		// Fetch historic sessions by user hash if there are any
		rowsByHash, err := s.BigTable.GetRowsWithPrefix(context.Background(), fmt.Sprintf("%s#", userHash), bigtable.RowFilter(chainFilter), bigtable.LimitRows(100))
		if err != nil {
			s.BigTableMetrics.ReadMetaFailureCount.Add(1)
			err = fmt.Errorf("UserSessions() failed to fetch historic user sessions: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		s.BigTableMetrics.ReadMetaSuccessCount.Add(1)

		// Fetch historic sessions by user ID if there are any
		rowsByID, err := s.BigTable.GetRowsWithPrefix(context.Background(), fmt.Sprintf("%s#", userID), bigtable.RowFilter(chainFilter), bigtable.LimitRows(100))
		if err != nil {
			s.BigTableMetrics.ReadMetaFailureCount.Add(1)
			err = fmt.Errorf("UserSessions() failed to fetch historic user sessions: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		s.BigTableMetrics.ReadMetaSuccessCount.Add(1)

		// Fetch historic sessions by hex user ID if there are any
		rowsByHexID, err := s.BigTable.GetRowsWithPrefix(context.Background(), fmt.Sprintf("%s#", hexUserID), bigtable.RowFilter(chainFilter), bigtable.LimitRows(100))
		if err != nil {
			s.BigTableMetrics.ReadMetaFailureCount.Add(1)
			err = fmt.Errorf("UserSessions() failed to fetch historic user sessions: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

		s.BigTableMetrics.ReadMetaSuccessCount.Add(1)

		liveIDString := strings.Join(sessionIDs, ",")

		if len(rowsByHash) > 0 {
			if err = s.GetHistoricalSlices(r, reply, rowsByHash, liveIDString, sessionSlice); err != nil {
				level.Error(s.Logger).Log("err", err)
				return err
			}
		} else if len(rowsByID) > 0 {
			if err = s.GetHistoricalSlices(r, reply, rowsByID, liveIDString, sessionSlice); err != nil {
				level.Error(s.Logger).Log("err", err)
				return err
			}
		} else if len(rowsByHexID) > 0 {
			if err = s.GetHistoricalSlices(r, reply, rowsByHexID, liveIDString, sessionSlice); err != nil {
				level.Error(s.Logger).Log("err", err)
				return err
			}
		}
	}

	return nil
}

func (s *BuyersService) GetHistoricalSlices(r *http.Request, reply *UserSessionsReply, rows []bigtable.Row, liveIDString string, sessionSlice transport.SessionSlice) error {
	var sessionMeta transport.SessionMeta

	for _, row := range rows {
		if err := sessionMeta.UnmarshalBinary(row[s.BigTableCfName][0].Value); err != nil {
			return err
		}
		if !strings.Contains(liveIDString, fmt.Sprintf("%016x", sessionMeta.ID)) {
			sliceRows, err := s.BigTable.GetRowsWithPrefix(context.Background(), fmt.Sprintf("%016x#", sessionMeta.ID), bigtable.RowFilter(bigtable.ColumnFilter("slices")))
			if err != nil {
				s.BigTableMetrics.ReadSliceFailureCount.Add(1)
				err = fmt.Errorf("GetHistoricalSlices() failed to fetch historic slice information: %v", err)
				return err
			}
			s.BigTableMetrics.ReadSliceSuccessCount.Add(1)

			// If a slice exists, add the session and the timestamp
			if len(sliceRows) > 0 {
				sessionSlice.UnmarshalBinary(sliceRows[0][s.BigTableCfName][0].Value)

				buyer, err := s.Storage.Buyer(sessionMeta.BuyerID)
				if err != nil {
					err = fmt.Errorf("GetHistoricalSlices() failed to fetch buyer: %v", err)
					level.Error(s.Logger).Log("err", err)
					return err
				}

				if !middleware.VerifyAllRoles(r, s.SameBuyerRole(buyer.CompanyCode)) {
					sessionMeta.Anonymise()
				}

				reply.Sessions = append(reply.Sessions, sessionMeta)
				reply.TimeStamps = append(reply.TimeStamps, sessionSlice.Timestamp.UTC())
			} else {
				// Increment counter
				s.Metrics.NoSlicesFailure.Add(1)
			}
		}
	}

	return nil
}

type TotalSessionsArgs struct {
	CompanyCode string `json:"company_code"`
}

type TotalSessionsReply struct {
	Direct int `json:"direct"`
	Next   int `json:"next"`
}

func (s *BuyersService) TotalSessions(r *http.Request, args *TotalSessionsArgs, reply *TotalSessionsReply) error {
	if r.Body != nil {
		defer r.Body.Close()
	}

	redisClient := s.RedisPoolSessionMap.Get()
	defer redisClient.Close()
	minutes := time.Now().Unix() / 60

	ghostArmyBuyerID := ghostarmy.GhostArmyBuyerID(s.Env)
	var ghostArmyScalar uint64 = 50
	if v, ok := os.LookupEnv("GHOST_ARMY_SCALER"); ok {
		if v, err := strconv.ParseUint(v, 10, 64); err == nil {
			ghostArmyScalar = v
		}
	}

	var ghostArmyNextScaler uint64 = 5
	if v, ok := os.LookupEnv("GHOST_ARMY_NEXT_SCALER"); ok {
		if v, err := strconv.ParseUint(v, 10, 64); err == nil {
			ghostArmyNextScaler = v
		}
	}

	switch args.CompanyCode {
	case "":
		buyers := s.Storage.Buyers()

		var firstNextCount int
		var secondNextCount int
		var ghostArmyNextCount int

		for _, buyer := range buyers {
			stringID := fmt.Sprintf("%016x", buyer.ID)
			redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", stringID, minutes-1))
			redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", stringID, minutes))
		}
		redisClient.Flush()

		for _, buyer := range buyers {
			firstCount, err := redis.Int(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed getting total session count next: %v", err)
				level.Error(s.Logger).Log("err", err)
				err = fmt.Errorf("TotalSessions() failed getting total session count next")
				return err
			}
			firstNextCount += firstCount

			secondCount, err := redis.Int(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed getting total session count next: %v", err)
				level.Error(s.Logger).Log("err", err)
				err = fmt.Errorf("TotalSessions() failed getting total session count next")
				return err
			}
			secondNextCount += secondCount

			if buyer.ID == ghostArmyBuyerID {
				if firstCount > secondCount {
					ghostArmyNextCount = firstCount * int(ghostArmyNextScaler)
				} else {
					ghostArmyNextCount = secondCount * int(ghostArmyNextScaler)
				}
				firstNextCount += ghostArmyNextCount
				secondNextCount += ghostArmyNextCount
			}
		}
		reply.Next = firstNextCount
		if secondNextCount > firstNextCount {
			reply.Next = secondNextCount
		}

		var firstTotalCount int
		var secondTotalCount int

		for _, buyer := range buyers {
			stringID := fmt.Sprintf("%016x", buyer.ID)
			redisClient.Send("HVALS", fmt.Sprintf("c-%s-%d", stringID, minutes-1))
			redisClient.Send("HVALS", fmt.Sprintf("c-%s-%d", stringID, minutes))
		}
		redisClient.Flush()

		for _, buyer := range buyers {
			firstCounts, err := redis.Strings(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to receive first session count: %v", err)
				level.Error(s.Logger).Log("err", err)
				err = fmt.Errorf("TotalSessions() failed to receive first session count")
				return err
			}

			for i := 0; i < len(firstCounts); i++ {
				firstCount, err := strconv.ParseUint(firstCounts[i], 10, 32)
				if err != nil {
					err = fmt.Errorf("TotalSessions() failed to parse first session count: %v", err)
					level.Error(s.Logger).Log("err", err)
					return err
				}

				firstTotalCount += int(firstCount)
			}

			secondCounts, err := redis.Strings(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to receive second session count: %v", err)
				level.Error(s.Logger).Log("err", err)
				err = fmt.Errorf("TotalSessions() failed to receive second session count")
				return err
			}

			for i := 0; i < len(secondCounts); i++ {
				secondCount, err := strconv.ParseUint(secondCounts[i], 10, 32)
				if err != nil {
					err = fmt.Errorf("TotalSessions() failed to parse second session count: %v", err)
					level.Error(s.Logger).Log("err", err)
					return err
				}

				secondTotalCount += int(secondCount)
			}

			if buyer.ID == ghostArmyBuyerID {
				// scale by next values because ghost army data contains 0 direct
				// if ghost army is turned off then this number will be 0 and have no effect
				firstTotalCount += ghostArmyNextCount*int(ghostArmyScalar) + ghostArmyNextCount
				secondTotalCount += ghostArmyNextCount*int(ghostArmyScalar) + ghostArmyNextCount
			}
		}

		totalCount := firstTotalCount
		if secondTotalCount > firstTotalCount {
			totalCount = secondTotalCount
		}

		reply.Direct = totalCount - reply.Next

	default:
		var ghostArmyNextCount int

		buyer, err := s.Storage.BuyerWithCompanyCode(args.CompanyCode)
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting company with code: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		buyerID := fmt.Sprintf("%016x", buyer.ID)
		if !middleware.VerifyAllRoles(r, s.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("TotalSessions(): %v", ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}

		redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", buyerID, minutes-1))
		redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", buyerID, minutes))
		redisClient.Send("HVALS", fmt.Sprintf("c-%s-%d", buyerID, minutes-1))
		redisClient.Send("HVALS", fmt.Sprintf("c-%s-%d", buyerID, minutes))
		redisClient.Flush()

		firstNextCount, err := redis.Int(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer session next counts: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("TotalSessions() failed getting buyer session next counts")
			return err
		}

		secondNextCount, err := redis.Int(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer session next counts: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("TotalSessions() failed getting buyer session next counts")
			return err
		}

		if buyer.ID == ghostArmyBuyerID {
			if firstNextCount > secondNextCount {
				ghostArmyNextCount = firstNextCount * int(ghostArmyNextScaler)
			} else {
				ghostArmyNextCount = secondNextCount * int(ghostArmyNextScaler)
			}
			firstNextCount += ghostArmyNextCount
			secondNextCount += ghostArmyNextCount
		}

		reply.Next = firstNextCount
		if secondNextCount > firstNextCount {
			reply.Next = secondNextCount
		}

		var firstTotalCount int
		var secondTotalCount int

		firstCounts, err := redis.Strings(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer first session total counts: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("TotalSessions() failed getting buyer first session total counts")
			return err
		}

		for i := 0; i < len(firstCounts); i++ {
			firstCount, err := strconv.ParseUint(firstCounts[i], 10, 32)
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to parse buyer first session count: %v", err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

			firstTotalCount += int(firstCount)
		}

		secondCounts, err := redis.Strings(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer second session total counts: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("TotalSessions() failed getting buyer second session total counts")
			return err
		}

		for i := 0; i < len(secondCounts); i++ {
			secondCount, err := strconv.ParseUint(secondCounts[i], 10, 32)
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to parse buyer second session count: %v", err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

			secondTotalCount += int(secondCount)
		}

		if buyer.ID == ghostArmyBuyerID {
			// scale by next values because ghost army data contains 0 direct
			// if ghost army is turned off then this number will be 0 and have no effect
			firstTotalCount += firstNextCount*int(ghostArmyScalar) + firstNextCount
			secondTotalCount += secondNextCount*int(ghostArmyScalar) + secondNextCount
		}

		totalCount := firstTotalCount
		if secondTotalCount > firstTotalCount {
			totalCount = secondTotalCount
		}

		reply.Direct = totalCount - reply.Next
	}

	return nil
}

type TopSessionsArgs struct {
	CompanyCode string `json:"company_code"`
}

type TopSessionsReply struct {
	Sessions []transport.SessionMeta `json:"sessions"`
}

// TopSessions generates the top sessions sorted by improved RTT
func (s *BuyersService) TopSessions(r *http.Request, args *TopSessionsArgs, reply *TopSessionsReply) error {
	sessions, err := s.FetchCurrentTopSessions(r, args.CompanyCode)
	if err != nil {
		err = fmt.Errorf("TopSessions() failed to fetch top sessions: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}
	reply.Sessions = sessions
	return nil
}

type SessionDetailsArgs struct {
	SessionID string `json:"session_id"`
}

type SessionDetailsReply struct {
	Meta   transport.SessionMeta    `json:"meta"`
	Slices []transport.SessionSlice `json:"slices"`
}

func (s *BuyersService) SessionDetails(r *http.Request, args *SessionDetailsArgs, reply *SessionDetailsReply) error {
	var err error
	var historic bool = false

	if args.SessionID == "" {
		err = fmt.Errorf("SessionDetails() session ID is required")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	sessionMetaClient := s.RedisPoolSessionMeta.Get()
	defer sessionMetaClient.Close()

	metaString, err := redis.String(sessionMetaClient.Do("GET", fmt.Sprintf("sm-%s", args.SessionID)))
	// Use bigtable if error from redis or requesting historic information
	if s.UseBigtable && (err != nil || metaString == "") {
		metaRows, err := s.BigTable.GetRowWithRowKey(context.Background(), fmt.Sprintf("%s", args.SessionID), bigtable.RowFilter(bigtable.ColumnFilter("meta")))
		if err != nil {
			s.BigTableMetrics.ReadMetaFailureCount.Add(1)
			err = fmt.Errorf("SessionDetails() failed to fetch historic meta information from bigtable: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		if len(metaRows) == 0 {
			s.BigTableMetrics.ReadMetaFailureCount.Add(1)
			err = fmt.Errorf("SessionDetails() failed to fetch historic meta information: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		s.BigTableMetrics.ReadMetaSuccessCount.Add(1)

		// Set historic to true when bigtable should be used
		historic = true

		for _, row := range metaRows {
			reply.Meta.UnmarshalBinary(row[0].Value)
		}
	}

	if !historic {
		metaStringsSplit := strings.Split(metaString, "|")
		if err := reply.Meta.ParseRedisString(metaStringsSplit); err != nil {
			err = fmt.Errorf("SessionDetails() SessionMeta unmarshaling error: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
	}

	buyer, err := s.Storage.Buyer(reply.Meta.BuyerID)
	if err != nil {
		err = fmt.Errorf("SessionDetails() failed to fetch buyer: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	if !middleware.VerifyAllRoles(r, s.SameBuyerRole(buyer.CompanyCode)) {
		reply.Meta.Anonymise()
	}

	var slice transport.SessionSlice
	reply.Slices = make([]transport.SessionSlice, 0)

	if !historic {
		sessionSlicesClient := s.RedisPoolSessionSlices.Get()
		defer sessionSlicesClient.Close()

		slices, err := redis.Strings(sessionSlicesClient.Do("LRANGE", fmt.Sprintf("ss-%s", args.SessionID), "0", "-1"))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("SessionDetails() failed getting session slices: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("SessionDetails() failed getting session slices")
			return err
		}

		for i := 0; i < len(slices); i++ {
			sliceStrings := strings.Split(slices[i], "|")
			if err := slice.ParseRedisString(sliceStrings); err != nil {
				err = fmt.Errorf("SessionDetails() SessionSlice parsing error: %v", err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

			reply.Slices = append(reply.Slices, slice)
		}
	} else {
		sliceRows, err := s.BigTable.GetRowsWithPrefix(context.Background(), fmt.Sprintf("%s#", args.SessionID), bigtable.RowFilter(bigtable.ColumnFilter("slices")))
		if err != nil {
			s.BigTableMetrics.ReadSliceFailureCount.Add(1)
			err = fmt.Errorf("SessionDetails() failed to fetch historic slice information from bigtable: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		if len(sliceRows) == 0 {
			s.BigTableMetrics.ReadSliceFailureCount.Add(1)
			err = fmt.Errorf("SessionDetails() failed to fetch historic slice information: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		s.BigTableMetrics.ReadSliceSuccessCount.Add(1)

		for _, row := range sliceRows {
			slice.UnmarshalBinary(row[s.BigTableCfName][0].Value)
			reply.Slices = append(reply.Slices, slice)
		}
	}

	sort.Slice(reply.Meta.NearbyRelays, func(i, j int) bool {
		return reply.Meta.NearbyRelays[i].ClientStats.RTT < reply.Meta.NearbyRelays[j].ClientStats.RTT
	})

	sort.Slice(reply.Slices, func(i, j int) bool {
		return reply.Slices[i].Timestamp.Before(reply.Slices[j].Timestamp)
	})

	return nil
}

type MapPointsArgs struct {
	CompanyCode string `json:"company_code"`
}

type MapPointsReply struct {
	Points json.RawMessage `json:"map_points"`
}

type MapPointsByteReply struct {
	Points mapPointsByte `json:"map_points"`
}

type mapPointsByte struct {
	Version     uint8   `json:"version"`
	GreenPoints []point `json:"green_points"`
	BluePoints  []point `json:"blue_points"`
}

type point struct {
	Latitude      float32 `json:"latitude"`
	Longitude     float32 `json:"longitude"`
	OnNetworkNext bool    `json:"on_network_next"`
}

func (s *BuyersService) GenerateMapPointsPerBuyer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	buyers := s.Storage.Buyers()

	// slices to hold all the final map points
	mapPointsBuyers := make(map[string][]transport.SessionMapPoint, 0)
	mapPointsBuyersCompact := make(map[string][][]interface{}, 0)
	mapPointsGlobal := make([]transport.SessionMapPoint, 0)
	mapPointsGlobalCompact := make([][]interface{}, 0)

	s.mapPointsBuyerCache = make(map[string]json.RawMessage, 0)
	s.mapPointsCompactBuyerCache = make(map[string]json.RawMessage, 0)

	for _, buyer := range buyers {
		isLargeCustomer := buyer.InternalConfig.LargeCustomer
		directPointStrings, nextPointStrings, err := s.getDirectAndNextMapPointStrings(&buyer)
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer %s: %v", buyer.CompanyCode, err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer")
			return err
		}

		var point transport.SessionMapPoint
		for _, directPointString := range directPointStrings {
			directSplitStrings := strings.Split(directPointString, "|")
			if err := point.ParseRedisString(directSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to parse direct map point for buyer %s: %v", buyer.CompanyCode, err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

			sessionID := fmt.Sprintf("%016x", point.SessionID)

			if point.Latitude != 0 && point.Longitude != 0 {
				mapPointsBuyers[buyer.CompanyCode] = append(mapPointsBuyers[buyer.CompanyCode], point)
				mapPointsGlobal = append(mapPointsGlobal, point)

				mapPointsBuyersCompact[buyer.CompanyCode] = append(mapPointsBuyersCompact[buyer.CompanyCode], []interface{}{point.Longitude, point.Latitude, isLargeCustomer, sessionID})
				mapPointsGlobalCompact = append(mapPointsGlobalCompact, []interface{}{point.Longitude, point.Latitude, isLargeCustomer, sessionID})
			}
		}

		for _, nextPointString := range nextPointStrings {
			nextSplitStrings := strings.Split(nextPointString, "|")
			if err := point.ParseRedisString(nextSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to next parse map point for buyer %s: %v", buyer.CompanyCode, err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

			sessionID := fmt.Sprintf("%016x", point.SessionID)

			if point.Latitude != 0 && point.Longitude != 0 {
				mapPointsBuyers[buyer.CompanyCode] = append(mapPointsBuyers[buyer.CompanyCode], point)
				mapPointsGlobal = append(mapPointsGlobal, point)

				mapPointsBuyersCompact[buyer.CompanyCode] = append(mapPointsBuyersCompact[buyer.CompanyCode], []interface{}{point.Longitude, point.Latitude, true, sessionID})
				mapPointsGlobalCompact = append(mapPointsGlobalCompact, []interface{}{point.Longitude, point.Latitude, true, sessionID})
			}
		}

		s.mapPointsBuyerCache[buyer.CompanyCode], err = json.Marshal(mapPointsBuyers[buyer.CompanyCode])
		if err != nil {
			return err
		}

		s.mapPointsCompactBuyerCache[buyer.CompanyCode], err = json.Marshal(mapPointsBuyersCompact[buyer.CompanyCode])
		if err != nil {
			return err
		}
	}

	// marshal the map points slice to local cache
	s.mapPointsCache, err = json.Marshal(mapPointsGlobal)
	if err != nil {
		return err
	}

	s.mapPointsCompactCache, err = json.Marshal(mapPointsGlobalCompact)
	if err != nil {
		return err
	}
	return nil
}

func (s *BuyersService) GenerateMapPointsPerBuyerBytes() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var mapPointsGlobal mapPointsByte
	var mapPointsBuyer mapPointsByte

	s.mapPointsBuyerByteCache = make(map[string][]byte, 0)
	s.mapPointsByteCache = make([]byte, 0)

	buyers := s.Storage.Buyers()

	for _, buyer := range buyers {
		stringID := fmt.Sprintf("%016x", buyer.ID)

		mapPointsBuyer.GreenPoints = make([]point, 0)
		mapPointsBuyer.BluePoints = make([]point, 0)

		directPointStrings, nextPointStrings, err := s.getDirectAndNextMapPointStrings(&buyer)
		if err != nil {
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer %s: %v", stringID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

		var currentPoint transport.SessionMapPoint
		for _, directPointString := range directPointStrings {
			directSplitStrings := strings.Split(directPointString, "|")
			if err := currentPoint.ParseRedisString(directSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to parse direct map point for buyer %s: %v", stringID, err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

			if currentPoint.Latitude != 0 && currentPoint.Longitude != 0 {
				bytePoint := point{
					Latitude:      float32(currentPoint.Latitude),
					Longitude:     float32(currentPoint.Longitude),
					OnNetworkNext: false,
				}

				mapPointsGlobal.BluePoints = append(mapPointsGlobal.BluePoints, bytePoint)
				mapPointsBuyer.BluePoints = append(mapPointsGlobal.BluePoints, bytePoint)
			}
		}

		for _, nextPointString := range nextPointStrings {
			nextSplitStrings := strings.Split(nextPointString, "|")
			if err := currentPoint.ParseRedisString(nextSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to next parse map point for buyer %s: %v", stringID, err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

			if currentPoint.Latitude != 0 && currentPoint.Longitude != 0 {
				bytePoint := point{
					Latitude:      float32(currentPoint.Latitude),
					Longitude:     float32(currentPoint.Longitude),
					OnNetworkNext: false,
				}

				bytePoint.OnNetworkNext = true
				mapPointsGlobal.GreenPoints = append(mapPointsGlobal.GreenPoints, bytePoint)
				mapPointsBuyer.GreenPoints = append(mapPointsGlobal.GreenPoints, bytePoint)
			}
		}

		// Write entries to byte cache
		s.mapPointsBuyerByteCache[stringID] = WriteMapPointCache(&mapPointsBuyer)
	}
	s.mapPointsByteCache = WriteMapPointCache(&mapPointsGlobal)

	return nil
}

func (s *BuyersService) getDirectAndNextMapPointStrings(buyer *routing.Buyer) ([]string, []string, error) {
	minutes := time.Now().Unix() / 60

	redisClient := s.RedisPoolSessionMap.Get()
	defer redisClient.Close()

	stringID := fmt.Sprintf("%016x", buyer.ID)
	redisClient.Send("HGETALL", fmt.Sprintf("d-%s-%d", stringID, minutes-1))
	redisClient.Send("HGETALL", fmt.Sprintf("d-%s-%d", stringID, minutes))
	redisClient.Send("HGETALL", fmt.Sprintf("n-%s-%d", stringID, minutes-1))
	redisClient.Send("HGETALL", fmt.Sprintf("n-%s-%d", stringID, minutes))
	redisClient.Flush()

	directMap, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		// Redis is down - return empty values for map
		return []string{}, []string{}, nil
	}

	directMapB, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		// Redis is down - return empty values for map
		return []string{}, []string{}, nil
	}

	for k, v := range directMapB {
		directMap[k] = v
	}

	nextMap, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		// Redis is down - return empty values for map
		return []string{}, []string{}, nil
	}

	nextMapB, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		// Redis is down - return empty values for map
		return []string{}, []string{}, nil
	}

	for k, v := range nextMapB {
		nextMap[k] = v
	}

	direct := make([]string, len(directMap))
	var index int
	for _, v := range directMap {
		direct[index] = v
		index++
	}

	next := make([]string, len(nextMap))
	index = 0
	for _, v := range nextMap {
		next[index] = v
		index++
	}

	return direct, next, nil
}

func WriteMapPointCache(points *mapPointsByte) []byte {
	var length uint32
	numGreenPoints := uint32(len(points.GreenPoints))
	numBluePoints := uint32(len(points.BluePoints))

	length = 1 + 4 + 4 + (1+4+4)*numGreenPoints + (1+4+4)*numBluePoints

	data := make([]byte, length)
	index := 0
	encoding.WriteUint8(data, &index, MapPointByteCacheVersion)

	encoding.WriteUint32(data, &index, numGreenPoints)

	for _, point := range points.GreenPoints {
		encoding.WriteFloat32(data, &index, point.Latitude)
		encoding.WriteFloat32(data, &index, point.Longitude)
		encoding.WriteBool(data, &index, point.OnNetworkNext)
	}

	encoding.WriteUint32(data, &index, numBluePoints)

	for _, point := range points.BluePoints {
		encoding.WriteFloat32(data, &index, point.Latitude)
		encoding.WriteFloat32(data, &index, point.Longitude)
		encoding.WriteBool(data, &index, point.OnNetworkNext)
	}
	return data
}

func ReadMapPointsCache(points *mapPointsByte, data []byte) bool {
	var numGreenPoints uint32
	var numBluePoints uint32

	index := 0
	if !encoding.ReadUint8(data, &index, &points.Version) {
		return false
	}
	if points.Version != 1 {
		return false
	}
	if !encoding.ReadUint32(data, &index, &numGreenPoints) {
		return false
	}

	var latitude float32
	var longitude float32
	var onNetworkNext bool
	points.GreenPoints = make([]point, numGreenPoints)

	for i := 0; i < int(numGreenPoints); i++ {
		if !encoding.ReadFloat32(data, &index, &latitude) {
			return false
		}
		if !encoding.ReadFloat32(data, &index, &longitude) {
			return false
		}
		if !encoding.ReadBool(data, &index, &onNetworkNext) {
			return false
		}
		points.GreenPoints[i] = point{
			Latitude:      latitude,
			Longitude:     longitude,
			OnNetworkNext: onNetworkNext,
		}
	}

	if !encoding.ReadUint32(data, &index, &numBluePoints) {
		return false
	}

	points.BluePoints = make([]point, numBluePoints)

	for i := 0; i < int(numBluePoints); i++ {
		if !encoding.ReadFloat32(data, &index, &latitude) {
			return false
		}
		if !encoding.ReadFloat32(data, &index, &longitude) {
			return false
		}
		if !encoding.ReadBool(data, &index, &onNetworkNext) {
			return false
		}
		points.BluePoints[i] = point{
			Latitude:      latitude,
			Longitude:     longitude,
			OnNetworkNext: onNetworkNext,
		}
	}

	return true
}

func (s *BuyersService) SessionMap(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		reply.Points = s.mapPointsCompactCache
	default:
		if !middleware.VerifyAllRoles(r, s.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		reply.Points = s.mapPointsCompactBuyerCache[args.CompanyCode]
	}

	return nil
}

func (s *BuyersService) SessionMapPoints(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		reply.Points = s.mapPointsCache
	default:
		if !middleware.VerifyAllRoles(r, s.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		reply.Points = s.mapPointsBuyerCache[args.CompanyCode]
	}

	return nil
}

// SessionMapPoints returns the locally cached JSON from GenerateSessionMapPoints
func (s *BuyersService) SessionMapPointsByte(r *http.Request, args *MapPointsArgs, reply *MapPointsByteReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		ReadMapPointsCache(&reply.Points, s.mapPointsCache)
	default:
		if !middleware.VerifyAllRoles(r, s.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		ReadMapPointsCache(&reply.Points, s.mapPointsBuyerCache[args.CompanyCode])
	}

	return nil
}

func (s *BuyersService) SessionMapByte(r *http.Request, args *MapPointsArgs, reply *MapPointsByteReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		ReadMapPointsCache(&reply.Points, s.mapPointsByteCache)
	default:
		if !middleware.VerifyAllRoles(r, s.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		ReadMapPointsCache(&reply.Points, s.mapPointsBuyerByteCache[args.CompanyCode])
	}

	return nil
}

type GameConfigurationArgs struct {
	NewPublicKey string `json:"new_public_key"`
}

type GameConfigurationReply struct {
	GameConfiguration gameConfiguration `json:"game_config"`
}

type gameConfiguration struct {
	PublicKey string `json:"public_key"`
}

func (s *BuyersService) GameConfiguration(r *http.Request, args *GameConfigurationArgs, reply *GameConfigurationReply) error {
	var err error
	var buyer routing.Buyer

	companyCode, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
	if !ok {
		err := fmt.Errorf("GameConfiguration(): user is not assigned to a company")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	if companyCode == "" {
		err = fmt.Errorf("GameConfiguration(): failed to parse company code")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err = fmt.Errorf("GameConfiguration(): %v", ErrInsufficientPrivileges)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	reply.GameConfiguration.PublicKey = ""

	buyer, err = s.Storage.BuyerWithCompanyCode(companyCode)
	// Buyer not found
	if err != nil {
		return nil
	}

	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()

	return nil
}

type BuyerInformationArgs struct {
	NewPublicKey string `json:"new_public_key"`
}

type BuyerInformationReply struct {
	PublicKey string `json:"public_key"`
}

func (s *BuyersService) UpdateBuyerInformation(r *http.Request, args *BuyerInformationArgs, reply *BuyerInformationReply) error {
	var err error
	var buyerID uint64
	var buyer routing.Buyer

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err = fmt.Errorf("UpdateBuyerInformation(): %v", ErrInsufficientPrivileges)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	ctx := context.Background()

	companyCode, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
	if !ok {
		err := fmt.Errorf("UpdateBuyerInformation(): user is not assigned to a company")
		level.Error(s.Logger).Log("err", err)
		return err
	}
	if companyCode == "" {
		err = fmt.Errorf("UpdateBuyerInformation(): failed to parse company code")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	if args.NewPublicKey == "" {
		err = fmt.Errorf("UpdateBuyerInformation() new public key is required")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	buyer, err = s.Storage.BuyerWithCompanyCode(companyCode)

	// Buyer found
	if buyer.ID != 0 {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "PublicKey", args.NewPublicKey); err != nil {
			err = fmt.Errorf("UpdateBuyerInformation() buyer update failed: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
	} else {
		// New Buyer
		byteKey, err := base64.StdEncoding.DecodeString(args.NewPublicKey)
		if err != nil {
			err = fmt.Errorf("UpdateBuyerInformation() could not decode public key string")
			level.Error(s.Logger).Log("err", err)
			return err
		}

		buyerID = binary.LittleEndian.Uint64(byteKey[0:8])

		// Create new buyer
		err = s.Storage.AddBuyer(ctx, routing.Buyer{
			CompanyCode: companyCode,
			ID:          buyerID,
			Live:        false,
			PublicKey:   byteKey[8:],
		})
		if err != nil {
			err = fmt.Errorf("UpdateBuyerInformation() failed to add buyer")
			level.Error(s.Logger).Log("err", err)
			return err
		}

		// Check if buyer is associated with the ID and everything worked
		buyer, err = s.Storage.Buyer(buyerID)
		if err != nil {
			err = fmt.Errorf("UpdateBuyerInformation() buyer creation failed: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
	}

	// Set reply
	reply.PublicKey = buyer.EncodedPublicKey()

	return nil
}

func (s *BuyersService) UpdateGameConfiguration(r *http.Request, args *GameConfigurationArgs, reply *GameConfigurationReply) error {
	var err error
	var buyerID uint64
	var buyer routing.Buyer

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err = fmt.Errorf("UpdateGameConfiguration(): %v", ErrInsufficientPrivileges)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	ctx := context.Background()

	companyCode, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
	if !ok {
		err := fmt.Errorf("UpdateGameConfiguration(): user is not assigned to a company")
		level.Error(s.Logger).Log("err", err)
		return err
	}
	if companyCode == "" {
		err = fmt.Errorf("UpdateGameConfiguration(): failed to parse company code")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	if args.NewPublicKey == "" {
		err = fmt.Errorf("UpdateGameConfiguration() new public key is required")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	buyer, err = s.Storage.BuyerWithCompanyCode(companyCode)

	byteKey, err := base64.StdEncoding.DecodeString(args.NewPublicKey)
	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() could not decode public key string")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	buyerID = binary.LittleEndian.Uint64(byteKey[0:8])

	// Buyer not found
	if buyer.ID == 0 {

		// Create new buyer
		err = s.Storage.AddBuyer(ctx, routing.Buyer{
			CompanyCode: companyCode,
			ID:          buyerID,
			Live:        false,
			PublicKey:   byteKey[8:],
		})

		if err != nil {
			err = fmt.Errorf("UpdateGameConfiguration() failed to add buyer")
			level.Error(s.Logger).Log("err", err)
			return err
		}

		// Check if buyer is associated with the ID and everything worked
		if buyer, err = s.Storage.Buyer(buyerID); err != nil {
			err = fmt.Errorf("UpdateGameConfiguration() buyer creation failed: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

		// Setup reply
		reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()

		return nil
	}

	live := buyer.Live

	if err = s.Storage.RemoveBuyer(ctx, buyer.ID); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() failed to remove buyer")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	err = s.Storage.AddBuyer(ctx, routing.Buyer{
		CompanyCode: companyCode,
		ID:          buyerID,
		Live:        live,
		PublicKey:   byteKey[8:],
	})

	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() buyer update failed: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	// Check if buyer is associated with the ID and everything worked
	if buyer, err = s.Storage.Buyer(buyerID); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() buyer update check failed: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	// Set reply
	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()

	return nil
}

type BuyerListArgs struct{}

type BuyerListReply struct {
	Buyers []buyerAccount `json:"buyers"`
}

type buyerAccount struct {
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
	ID          string `json:"id"`
	IsLive      bool   `json:"is_live"`
}

func (s *BuyersService) Buyers(r *http.Request, args *BuyerListArgs, reply *BuyerListReply) error {
	reply.Buyers = make([]buyerAccount, 0)
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	for _, b := range s.Storage.Buyers() {
		id := fmt.Sprintf("%016x", b.ID)
		customer, err := s.Storage.Customer(b.CompanyCode)
		if err != nil {
			err = fmt.Errorf("Buyers() buyer is not assigned to customer: %v", b.ID)
			level.Error(s.Logger).Log("err", err)
			continue
		}
		account := buyerAccount{
			CompanyName: customer.Name,
			CompanyCode: b.CompanyCode,
			ID:          id,
			IsLive:      b.Live,
		}
		if middleware.VerifyAllRoles(r, s.SameBuyerRole(b.CompanyCode)) {
			reply.Buyers = append(reply.Buyers, account)
		}
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].CompanyName < reply.Buyers[j].CompanyName
	})

	return nil
}

type DatacenterMapsArgs struct {
	ID    uint64 `json:"buyer_id"`
	HexID string `json:"hexBuyerID"`
}

type DatacenterMapsFull struct {
	Alias          string
	DatacenterName string
	DatacenterID   string
	BuyerName      string
	BuyerID        string
	SupplierName   string
}

type DatacenterMapsReply struct {
	DatacenterMaps []DatacenterMapsFull
}

func (s *BuyersService) DatacenterMapsForBuyer(r *http.Request, args *DatacenterMapsArgs, reply *DatacenterMapsReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	var buyerID uint64
	var err error

	if args.HexID != "" {
		buyerID, err = strconv.ParseUint(args.HexID, 16, 64)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() could not parse hex buyer ID: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
	} else {
		buyerID = args.ID
	}

	var dcm map[uint64]routing.DatacenterMap

	dcm = s.Storage.GetDatacenterMapsForBuyer(buyerID)

	var replySlice []DatacenterMapsFull
	for _, dcMap := range dcm {
		buyer, err := s.Storage.Buyer(dcMap.BuyerID)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() could not parse buyer")
			level.Error(s.Logger).Log("err", err)
			return err
		}
		datacenter, err := s.Storage.Datacenter(dcMap.DatacenterID)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() could not parse datacenter")
			level.Error(s.Logger).Log("err", err)
			return err
		}

		customer, err := s.Storage.Customer(buyer.CompanyCode)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() buyer is not associated with a company")
			level.Error(s.Logger).Log("err", err)
			continue
		}

		dcmFull := DatacenterMapsFull{
			Alias:          dcMap.Alias,
			DatacenterName: datacenter.Name,
			DatacenterID:   fmt.Sprintf("%016x", dcMap.DatacenterID),
			BuyerName:      customer.Name,
			BuyerID:        fmt.Sprintf("%016x", dcMap.BuyerID),
		}

		replySlice = append(replySlice, dcmFull)
	}

	reply.DatacenterMaps = replySlice
	return nil

}

type JSRemoveDatacenterMapArgs struct {
	DatacenterHexID string `json:"hexDatacenterID"`
	BuyerHexID      string `json:"hexBuyerID"`
	Alias           string `json:"alias"`
}

type JSRemoveDatacenterMapReply struct {
	DatacenterMap routing.DatacenterMap
}

func (s *BuyersService) JSRemoveDatacenterMap(r *http.Request, args *JSRemoveDatacenterMapArgs, reply *JSRemoveDatacenterMapReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	buyerID, err := strconv.ParseUint(args.BuyerHexID, 16, 64)
	if err != nil {
		return fmt.Errorf("Unable to parse BuyerID: %s", args.BuyerHexID)
	}

	datacenterID, err := strconv.ParseUint(args.DatacenterHexID, 16, 64)
	if err != nil {
		return fmt.Errorf("Unable to parse DatacenterID: %s", args.BuyerHexID)
	}

	dcMap := routing.DatacenterMap{
		Alias:        args.Alias,
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
	}

	return s.Storage.RemoveDatacenterMap(ctx, dcMap)

}

type RemoveDatacenterMapArgs struct {
	DatacenterMap routing.DatacenterMap
}

type RemoveDatacenterMapReply struct{}

func (s *BuyersService) RemoveDatacenterMap(r *http.Request, args *RemoveDatacenterMapArgs, reply *RemoveDatacenterMapReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	return s.Storage.RemoveDatacenterMap(ctx, args.DatacenterMap)

}

// UpdateDatacenterMapArgs: HexBuyerID and HexDatacenterID are the combined primary
// key needed to look up the existing datacenter map
type UpdateDatacenterMapArgs struct {
	HexBuyerID      string `json:"hexBuyerID"`
	HexDatacenterID string `json:"hexDatacenterID"`
	Field           string `json:"field"`
	Value           string `json:"value"`
}

type UpdateDatacenterMapReply struct{}

func (s *BuyersService) UpdateDatacenterMap(r *http.Request, args *UpdateDatacenterMapArgs, reply *UpdateDatacenterMapReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	datacenterID, err := strconv.ParseUint(args.HexDatacenterID, 16, 64)
	if err != nil {
		return fmt.Errorf("Value: %v is not a valid hex ID", args.Value)
	}

	buyerID, err := strconv.ParseUint(args.HexBuyerID, 16, 64)
	if err != nil {
		return fmt.Errorf("Value: %v is not a valid hex ID", args.Value)
	}

	switch args.Field {
	case "HexDatacenterID", "Alias":
		err = s.Storage.UpdateDatacenterMap(context.Background(), buyerID, datacenterID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateDatacenterMap() error updating datacenter map: %v", err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist or is not editable on the DatacenterMap type", args.Field)
	}

	return nil
}

type JSAddDatacenterMapArgs struct {
	HexBuyerID      string `json:"hexBuyerID"`
	HexDatacenterID string `json:"hexDatacenterID"`
	Alias           string `json:"alias"`
}

type JSAddDatacenterMapReply struct{}

func (s *BuyersService) JSAddDatacenterMap(r *http.Request, args *JSAddDatacenterMapArgs, reply *JSAddDatacenterMapReply) error {

	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	buyerID, err := strconv.ParseUint(args.HexBuyerID, 16, 64)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	datacenterID, err := strconv.ParseUint(args.HexDatacenterID, 16, 64)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	dcMap := routing.DatacenterMap{
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
		Alias:        args.Alias,
	}

	return s.Storage.AddDatacenterMap(ctx, dcMap)

}

type AddDatacenterMapArgs struct {
	DatacenterMap routing.DatacenterMap
}

type AddDatacenterMapReply struct{}

func (s *BuyersService) AddDatacenterMap(r *http.Request, args *AddDatacenterMapArgs, reply *AddDatacenterMapReply) error {

	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	return s.Storage.AddDatacenterMap(ctx, args.DatacenterMap)

}

// SameBuyerRole checks the JWT for the correct passed in buyerID
func (s *BuyersService) SameBuyerRole(companyCode string) middleware.RoleFunc {
	return func(req *http.Request) (bool, error) {
		if middleware.VerifyAnyRole(req, middleware.AdminRole, middleware.OpsRole) {
			return true, nil
		}
		if middleware.VerifyAllRoles(req, middleware.AnonymousRole) {
			return false, nil
		}
		if companyCode == "" {
			return false, fmt.Errorf("SameBuyerRole(): buyerID is required")
		}
		requestCompanyCode, ok := req.Context().Value(middleware.Keys.CompanyKey).(string)
		if !ok {
			err := fmt.Errorf("SameBuyerRole(): user is not assigned to a company")
			level.Error(s.Logger).Log("err", err)
			return false, err
		}
		if requestCompanyCode == "" {
			err := fmt.Errorf("SameBuyerRole(): failed to parse company code")
			level.Error(s.Logger).Log("err", err)
			return false, err
		}

		return companyCode == requestCompanyCode, nil
	}
}

func (s *BuyersService) FetchCurrentTopSessions(r *http.Request, companyCodeFilter string) ([]transport.SessionMeta, error) {
	var err error
	var topSessionsA []string
	var topSessionsB []string

	sessions := make([]transport.SessionMeta, 0)

	minutes := time.Now().Unix() / 60

	topSessionsClient := s.RedisPoolTopSessions.Get()
	defer topSessionsClient.Close()

	// get the top session IDs globally or for a buyer from the sorted set
	switch companyCodeFilter {
	case "":
		// Get top sessions from the past 2 minutes sorted by greatest to least improved RTT
		topSessionsA, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("s-%d", minutes-1), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions A: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions A")
			return sessions, err
		}
		topSessionsB, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("s-%d", minutes), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions B: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions B")
			return sessions, err
		}
	default:
		if !middleware.VerifyAllRoles(r, s.SameBuyerRole(companyCodeFilter)) {
			err := fmt.Errorf("FetchCurrentTopSessions(): %v", ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return sessions, err
		}

		buyer, err := s.Storage.BuyerWithCompanyCode(companyCodeFilter)
		if err != nil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting buyer with code: %v", err)
			level.Error(s.Logger).Log("err", err)
			return sessions, err
		}
		buyerID := fmt.Sprintf("%016x", buyer.ID)

		topSessionsA, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("sc-%s-%d", buyerID, minutes-1), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions A for buyer ID %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions A for buyer ID %016x", buyerID)
			return sessions, err
		}
		topSessionsB, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("sc-%s-%d", buyerID, minutes), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions B for buyer ID %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions B for buyer ID %016x", buyerID)
			return sessions, err
		}
	}

	sessionMetaClient := s.RedisPoolSessionMeta.Get()
	defer sessionMetaClient.Close()

	sessionIDsRetreivedMap := make(map[string]bool)
	for _, sessionID := range topSessionsA {
		sessionMetaClient.Send("GET", fmt.Sprintf("sm-%s", sessionID))
		sessionIDsRetreivedMap[sessionID] = true
	}
	for _, sessionID := range topSessionsB {
		if _, ok := sessionIDsRetreivedMap[sessionID]; !ok {
			sessionMetaClient.Send("GET", fmt.Sprintf("sm-%s", sessionID))
			sessionIDsRetreivedMap[sessionID] = true
		}
	}
	sessionMetaClient.Flush()

	var sessionMetasNext []transport.SessionMeta
	var sessionMetasDirect []transport.SessionMeta
	var meta transport.SessionMeta
	for i := 0; i < len(sessionIDsRetreivedMap); i++ {
		metaString, err := redis.String(sessionMetaClient.Receive())
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions meta: %v", err)
			level.Error(s.Logger).Log("err", err)
			err = fmt.Errorf("FetchCurrentTopSessions() failed getting top sessions meta")
			return sessions, err
		}

		splitMetaStrings := strings.Split(metaString, "|")
		if err := meta.ParseRedisString(splitMetaStrings); err != nil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed to parse redis string into meta: %v", err)
			level.Error(s.Logger).Log("err", err, "redisString", metaString)
			continue
		}

		buyer, err := s.Storage.Buyer(meta.BuyerID)
		if err != nil {
			err = fmt.Errorf("FetchCurrentTopSessions() failed to fetch buyer: %v", err)
			level.Error(s.Logger).Log("err", err)
			return sessions, err
		}

		if !middleware.VerifyAllRoles(r, s.SameBuyerRole(buyer.CompanyCode)) {
			meta.Anonymise()
		}

		// Split the sessions metas into two slices so we can sort them separately.
		// This is necessary because if we were to force sessions next, then sorting
		// by improvement won't always put next sessions on top.
		if meta.OnNetworkNext {
			sessionMetasNext = append(sessionMetasNext, meta)
		} else {
			sessionMetasDirect = append(sessionMetasDirect, meta)
		}
	}

	// These sorts are necessary because we are combining two ZREVRANGEs from two separate minute buckets.
	sort.Slice(sessionMetasNext, func(i, j int) bool {
		return sessionMetasNext[i].DeltaRTT > sessionMetasNext[j].DeltaRTT
	})

	sort.Slice(sessionMetasDirect, func(i, j int) bool {
		return sessionMetasDirect[i].DirectRTT > sessionMetasDirect[j].DirectRTT
	})

	sessionMetas := append(sessionMetasNext, sessionMetasDirect...)

	if len(sessionMetas) > TopSessionsSize {
		sessions = sessionMetas[:TopSessionsSize]
		return sessions, err
	}

	sessions = sessionMetas
	return sessions, err
}

type JSInternalConfig struct {
	RouteSelectThreshold       int64 `json:"routeSelectThreshold"`
	RouteSwitchThreshold       int64 `json:"routeSwitchThreshold"`
	MaxLatencyTradeOff         int64 `json:"maxLatencyTradeOff"`
	RTTVeto_Default            int64 `json:"rttVeto_Default"`
	RTTVeto_Multipath          int64 `json:"rttVeto_Multipath"`
	RTTVeto_PacketLoss         int64 `json:"rttVeto_PacketLoss"`
	MultipathOverloadThreshold int64 `json:"multipathOverloadThreshold"`
	TryBeforeYouBuy            bool  `json:"tryBeforeYouBuy"`
	ForceNext                  bool  `json:"forceNext"`
	LargeCustomer              bool  `json:"largeCustomer"`
	Uncommitted                bool  `json:"uncommitted"`
	MaxRTT                     int64 `json:"maxRTT"`
	HighFrequencyPings         bool  `json:"highFrequencyPings"`
	RouteDiversity             int64 `json:"routeDiversity"`
	MultipathThreshold         int64 `json:"multipathThreshold"`
	EnableVanityMetrics        bool  `json:"enableVanityMetrics"`
}

type InternalConfigArg struct {
	BuyerID string `json:"buyerID"`
}

type InternalConfigReply struct {
	InternalConfig JSInternalConfig
}

func (s *BuyersService) InternalConfig(r *http.Request, arg *InternalConfigArg, reply *InternalConfigReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	buyerID, err := strconv.ParseUint(arg.BuyerID, 16, 64)
	if err != nil {
		level.Error(s.Logger).Log("err", err)
		return err
	}

	ic, err := s.Storage.InternalConfig(buyerID)
	if err != nil {
		err = fmt.Errorf("InternalConfig() no InternalConfig stored for buyer %s", arg.BuyerID)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	jsonIC := JSInternalConfig{
		RouteSelectThreshold:       int64(ic.RouteSelectThreshold),
		RouteSwitchThreshold:       int64(ic.RouteSwitchThreshold),
		MaxLatencyTradeOff:         int64(ic.MaxLatencyTradeOff),
		RTTVeto_Default:            int64(ic.RTTVeto_Default),
		RTTVeto_Multipath:          int64(ic.RTTVeto_Multipath),
		RTTVeto_PacketLoss:         int64(ic.RTTVeto_PacketLoss),
		MultipathOverloadThreshold: int64(ic.MultipathOverloadThreshold),
		TryBeforeYouBuy:            ic.TryBeforeYouBuy,
		ForceNext:                  ic.ForceNext,
		LargeCustomer:              ic.LargeCustomer,
		Uncommitted:                ic.Uncommitted,
		MaxRTT:                     int64(ic.MaxRTT),
		HighFrequencyPings:         ic.HighFrequencyPings,
		RouteDiversity:             int64(ic.RouteDiversity),
		MultipathThreshold:         int64(ic.MultipathThreshold),
		EnableVanityMetrics:        ic.EnableVanityMetrics,
	}

	reply.InternalConfig = jsonIC
	return nil
}

type JSAddInternalConfigArgs struct {
	BuyerID        string           `json:"buyerID"`
	InternalConfig JSInternalConfig `json:"internalConfig"`
}

type JSAddInternalConfigReply struct{}

func (s *BuyersService) JSAddInternalConfig(r *http.Request, arg *JSAddInternalConfigArgs, reply *JSAddInternalConfigReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	buyerID, err := strconv.ParseUint(arg.BuyerID, 16, 64)
	if err != nil {
		level.Error(s.Logger).Log("err", err)
		return err
	}

	ic := core.InternalConfig{
		RouteSelectThreshold:       int32(arg.InternalConfig.RouteSelectThreshold),
		RouteSwitchThreshold:       int32(arg.InternalConfig.RouteSwitchThreshold),
		MaxLatencyTradeOff:         int32(arg.InternalConfig.MaxLatencyTradeOff),
		RTTVeto_Default:            int32(arg.InternalConfig.RTTVeto_Default),
		RTTVeto_Multipath:          int32(arg.InternalConfig.RTTVeto_Multipath),
		RTTVeto_PacketLoss:         int32(arg.InternalConfig.RTTVeto_PacketLoss),
		MultipathOverloadThreshold: int32(arg.InternalConfig.MultipathOverloadThreshold),
		TryBeforeYouBuy:            arg.InternalConfig.TryBeforeYouBuy,
		ForceNext:                  arg.InternalConfig.ForceNext,
		LargeCustomer:              arg.InternalConfig.LargeCustomer,
		Uncommitted:                arg.InternalConfig.Uncommitted,
		MaxRTT:                     int32(arg.InternalConfig.MaxRTT),
		HighFrequencyPings:         arg.InternalConfig.HighFrequencyPings,
		RouteDiversity:             int32(arg.InternalConfig.RouteDiversity),
		MultipathThreshold:         int32(arg.InternalConfig.MultipathThreshold),
		EnableVanityMetrics:        arg.InternalConfig.EnableVanityMetrics,
	}

	err = s.Storage.AddInternalConfig(context.Background(), ic, buyerID)
	if err != nil {
		err = fmt.Errorf("JSAddInternalConfig() error adding internal config for buyer %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	return nil
}

type UpdateInternalConfigArgs struct {
	BuyerID    uint64 `json:"buyerID"`
	HexBuyerID string `json:"hexBuyerID"`
	Field      string `json:"field"`
	Value      string `json:"value"`
}

type UpdateInternalConfigReply struct{}

func (s *BuyersService) UpdateInternalConfig(r *http.Request, args *UpdateInternalConfigArgs, reply *UpdateInternalConfigReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	var err error
	var buyerID uint64

	if args.HexBuyerID == "" {
		buyerID = args.BuyerID
	} else {
		buyerID, err = strconv.ParseUint(args.HexBuyerID, 16, 64)
		if err != nil {
			return fmt.Errorf("Can not parse HexBuyerID: %s", args.HexBuyerID)
		}
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	case "RouteSelectThreshold", "RouteSwitchThreshold", "MaxLatencyTradeOff",
		"RTTVeto_Default", "RTTVeto_PacketLoss", "RTTVeto_Multipath",
		"MultipathOverloadThreshold", "MaxRTT", "RouteDiversity", "MultipathThreshold":
		newInt, err := strconv.ParseInt(args.Value, 10, 32)
		if err != nil {
			return fmt.Errorf("Value: %v is not a valid integer type", args.Value)
		}
		newInt32 := int32(newInt)
		err = s.Storage.UpdateInternalConfig(context.Background(), buyerID, args.Field, newInt32)
		if err != nil {
			err = fmt.Errorf("UpdateInternalConfig() error updating internal config for buyer %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	case "TryBeforeYouBuy", "ForceNext", "LargeCustomer", "Uncommitted",
		"HighFrequencyPings", "EnableVanityMetrics":
		newValue, err := strconv.ParseBool(args.Value)
		if err != nil {
			return fmt.Errorf("Value: %v is not a valid boolean type", args.Value)
		}

		err = s.Storage.UpdateInternalConfig(context.Background(), buyerID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateInternalConfig() error updating internal config for buyer %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist on the InternalConfig type", args.Field)
	}

	return nil
}

type RemoveInternalConfigArg struct {
	BuyerID string `json:"buyerID"`
}

type RemoveInternalConfigReply struct{}

func (s *BuyersService) RemoveInternalConfig(r *http.Request, arg *RemoveInternalConfigArg, reply *RemoveInternalConfigReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	buyerID, err := strconv.ParseUint(arg.BuyerID, 16, 64)
	if err != nil {
		level.Error(s.Logger).Log("err", err)
		return err
	}

	err = s.Storage.RemoveInternalConfig(context.Background(), buyerID)
	if err != nil {
		err = fmt.Errorf("RemoveInternalConfig() error removing internal config for buyer %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	return nil
}

type JSRouteShader struct {
	DisableNetworkNext        bool            `json:"disableNetworkNext"`
	SelectionPercent          int64           `json:"selectionPercent"`
	ABTest                    bool            `json:"abTest"`
	ProMode                   bool            `json:"proMode"`
	ReduceLatency             bool            `json:"reduceLatency"`
	ReduceJitter              bool            `json:"reduceJitter"`
	ReducePacketLoss          bool            `json:"reducePacketLoss"`
	Multipath                 bool            `json:"multipath"`
	AcceptableLatency         int64           `json:"acceptableLatency"`
	LatencyThreshold          int64           `json:"latencyThreshold"`
	AcceptablePacketLoss      float64         `json:"acceptablePacketLoss"`
	BandwidthEnvelopeUpKbps   int64           `json:"bandwidthEnvelopeUpKbps"`
	BandwidthEnvelopeDownKbps int64           `json:"bandwidthEnvelopeDownKbps"`
	BannedUsers               map[string]bool `json:"bannedUsers"`
}
type RouteShaderArg struct {
	BuyerID string `json:"buyerID"`
}

type RouteShaderReply struct {
	RouteShader JSRouteShader
}

func (s *BuyersService) RouteShader(r *http.Request, arg *RouteShaderArg, reply *RouteShaderReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	buyerID, err := strconv.ParseUint(arg.BuyerID, 16, 64)
	if err != nil {
		level.Error(s.Logger).Log("err", err)
		return err
	}

	rs, err := s.Storage.RouteShader(buyerID)
	if err != nil {
		err = fmt.Errorf("RouteShader() error retrieving route shader for buyer %s: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	jsonRS := JSRouteShader{
		DisableNetworkNext:        rs.DisableNetworkNext,
		SelectionPercent:          int64(rs.SelectionPercent),
		ABTest:                    rs.ABTest,
		ProMode:                   rs.ProMode,
		ReduceLatency:             rs.ReduceLatency,
		ReduceJitter:              rs.ReduceJitter,
		ReducePacketLoss:          rs.ReducePacketLoss,
		Multipath:                 rs.Multipath,
		AcceptableLatency:         int64(rs.AcceptableLatency),
		LatencyThreshold:          int64(rs.LatencyThreshold),
		AcceptablePacketLoss:      float64(rs.AcceptablePacketLoss),
		BandwidthEnvelopeUpKbps:   int64(rs.BandwidthEnvelopeUpKbps),
		BandwidthEnvelopeDownKbps: int64(rs.BandwidthEnvelopeDownKbps),
	}

	reply.RouteShader = jsonRS
	return nil
}

type JSAddRouteShaderArgs struct {
	BuyerID     string        `json:"buyerID"`
	RouteShader JSRouteShader `json:"routeShader"`
}

type JSAddRouteShaderReply struct{}

func (s *BuyersService) JSAddRouteShader(r *http.Request, arg *JSAddRouteShaderArgs, reply *JSAddRouteShaderReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	buyerID, err := strconv.ParseUint(arg.BuyerID, 16, 64)
	if err != nil {
		level.Error(s.Logger).Log("err", err)
		return err
	}

	rs := core.RouteShader{
		DisableNetworkNext:        arg.RouteShader.DisableNetworkNext,
		SelectionPercent:          int(arg.RouteShader.SelectionPercent),
		ABTest:                    arg.RouteShader.ABTest,
		ProMode:                   arg.RouteShader.ProMode,
		ReduceLatency:             arg.RouteShader.ReduceLatency,
		ReduceJitter:              arg.RouteShader.ReduceJitter,
		ReducePacketLoss:          arg.RouteShader.ReducePacketLoss,
		Multipath:                 arg.RouteShader.Multipath,
		AcceptableLatency:         int32(arg.RouteShader.AcceptableLatency),
		LatencyThreshold:          int32(arg.RouteShader.LatencyThreshold),
		AcceptablePacketLoss:      float32(arg.RouteShader.AcceptablePacketLoss),
		BandwidthEnvelopeUpKbps:   int32(arg.RouteShader.BandwidthEnvelopeUpKbps),
		BandwidthEnvelopeDownKbps: int32(arg.RouteShader.BandwidthEnvelopeDownKbps),
	}

	err = s.Storage.AddRouteShader(context.Background(), rs, buyerID)
	if err != nil {
		err = fmt.Errorf("AddRouteShader() error adding route shader for buyer %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	return nil
}

type RemoveRouteShaderArg struct {
	BuyerID string `json:"buyerID"`
}

type RemoveRouteShaderReply struct{}

func (s *BuyersService) RemoveRouteShader(r *http.Request, arg *RemoveRouteShaderArg, reply *RemoveRouteShaderReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	buyerID, err := strconv.ParseUint(arg.BuyerID, 16, 64)
	if err != nil {
		level.Error(s.Logger).Log("err", err)
		return err
	}

	err = s.Storage.RemoveRouteShader(context.Background(), buyerID)
	if err != nil {
		err = fmt.Errorf("RemoveRouteShader() error removing route shader for buyer %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	return nil
}

type UpdateRouteShaderArgs struct {
	BuyerID    uint64 `json:"buyerID"`
	HexBuyerID string `json:"hexBuyerID"`
	Field      string `json:"field"`
	Value      string `json:"value"`
}

type UpdateRouteShaderReply struct{}

func (s *BuyersService) UpdateRouteShader(r *http.Request, args *UpdateRouteShaderArgs, reply *UpdateRouteShaderReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	var err error
	var buyerID uint64

	if args.HexBuyerID == "" {
		buyerID = args.BuyerID
	} else {
		buyerID, err = strconv.ParseUint(args.HexBuyerID, 16, 64)
		if err != nil {
			return fmt.Errorf("Can not parse HexBuyerID: %s", args.HexBuyerID)
		}
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	case "AcceptableLatency", "LatencyThreshold", "BandwidthEnvelopeUpKbps",
		"BandwidthEnvelopeDownKbps":
		newInt, err := strconv.ParseInt(args.Value, 10, 32)
		if err != nil {
			return fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid integer type", args.Value)
		}
		newInt32 := int32(newInt)
		err = s.Storage.UpdateRouteShader(context.Background(), buyerID, args.Field, newInt32)
		if err != nil {
			err = fmt.Errorf("UpdateRouteShader() error updating route shader for buyer %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	case "SelectionPercent":
		newInt, err := strconv.ParseInt(args.Value, 10, 64) // 64 bits is a guess, should be
		if err != nil {
			return fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid integer type", args.Value)
		}
		newInteger := int(newInt)
		err = s.Storage.UpdateRouteShader(context.Background(), buyerID, args.Field, newInteger)
		if err != nil {
			err = fmt.Errorf("UpdateRouteShader() error updating route shader for buyer %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	case "AcceptablePacketLoss":
		newFloat, err := strconv.ParseFloat(args.Value, 64)
		if err != nil {
			return fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid float type", args.Value)
		}
		newFloat32 := float32(newFloat)
		err = s.Storage.UpdateRouteShader(context.Background(), buyerID, args.Field, newFloat32)
		if err != nil {
			err = fmt.Errorf("UpdateRouteShader() error updating route shader for buyer %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	case "DisableNetworkNext", "ABTest", "ProMode", "ReduceLatency",
		"ReduceJitter", "ReducePacketLoss", "Multipath":
		newValue, err := strconv.ParseBool(args.Value)
		if err != nil {
			return fmt.Errorf("BuyersService.UpdateRouteShader Value: %v is not a valid boolean type", args.Value)
		}

		fmt.Printf("newValue: %T\n", newValue)
		err = s.Storage.UpdateRouteShader(context.Background(), buyerID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateRouteShader() error updating route shader for buyer %016x: %v", buyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist on the RouteShader type", args.Field)
	}

	return nil
}

// BannedUser CRUD endpoints
type GetBannedUserArg struct {
	BuyerID uint64
}

type GetBannedUserReply struct {
	BannedUsers []string // hex strings
}

func (s *BuyersService) GetBannedUsers(r *http.Request, arg *GetBannedUserArg, reply *GetBannedUserReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	var userList []string

	bannedUsers, err := s.Storage.BannedUsers(arg.BuyerID)
	if err != nil {
		err = fmt.Errorf("GetBannedUsers() error retrieving banned users for buyer %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	for userID := range bannedUsers {
		userList = append(userList, fmt.Sprintf("%016x", userID))
	}

	reply.BannedUsers = userList
	return nil
}

type BannedUserArgs struct {
	BuyerID uint64
	UserID  uint64
}

type BannedUserReply struct{}

func (s *BuyersService) AddBannedUser(r *http.Request, arg *BannedUserArgs, reply *BannedUserReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	err := s.Storage.AddBannedUser(context.Background(), arg.BuyerID, arg.UserID)
	if err != nil {
		err = fmt.Errorf("AddBannedUser() error adding banned user for buyer %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	return nil
}

func (s *BuyersService) RemoveBannedUser(r *http.Request, arg *BannedUserArgs, reply *BannedUserReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	err := s.Storage.RemoveBannedUser(context.Background(), arg.BuyerID, arg.UserID)
	if err != nil {
		err = fmt.Errorf("RemoveBannedUser() error removing banned user for buyer %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	return nil
}

type BuyerArg struct {
	BuyerID uint64
}

type BuyerReply struct {
	Buyer routing.Buyer
}

func (s *BuyersService) Buyer(r *http.Request, arg *BuyerArg, reply *BuyerReply) error {

	var b routing.Buyer
	var err error

	b, err = s.Storage.Buyer(arg.BuyerID)
	if err != nil {
		err = fmt.Errorf("Buyer() error retrieving buyer for ID %016x: %v", arg.BuyerID, err)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	reply.Buyer = b

	return nil
}

type UpdateBuyerArgs struct {
	BuyerID    uint64 `json:"buyerID"`
	HexBuyerID string `json:"hexBuyerID"` // needed for external (non-go) clients
	Field      string `json:"field"`
	Value      string `json:"value"`
}

type UpdateBuyerReply struct{}

func (s *BuyersService) UpdateBuyer(r *http.Request, args *UpdateBuyerArgs, reply *UpdateBuyerReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	var buyerID uint64
	var err error
	if args.BuyerID != 0 {
		buyerID = args.BuyerID
	} else {
		buyerID, err = strconv.ParseUint(args.HexBuyerID, 16, 64)
		if err != nil {
			return fmt.Errorf("BuyersService.UpdateBuyer could not parse hexBuyerID: %s", args.HexBuyerID)
		}
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	case "Live", "Debug":
		newValue, err := strconv.ParseBool(args.Value)
		if err != nil {
			return fmt.Errorf("BuyersService.UpdateBuyer Value: %v is not a valid boolean type", args.Value)
		}

		err = s.Storage.UpdateBuyer(context.Background(), buyerID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateBuyer() error updating record for buyer %016x: %v", args.BuyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
	case "ShortName", "PublicKey":
		err := s.Storage.UpdateBuyer(context.Background(), buyerID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateBuyer() error updating record for buyer %016x: %v", args.BuyerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the Buyer type", args.Field)
	}

	return nil
}
