package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/routing"
	"google.golang.org/api/iterator"
)

type bqSession struct {
	SessionId      int
	UserId         string
	PlatformId     int
	DatacenterId   string
	Latitude       float64
	Longitude      float64
	Isp            string
	ConnectionType int
	ClientAddr     string
	ServerAddr     string

	SDK string

	NextRTT   float64
	DirectRTT float64
}

type BigQuery struct {
	Client *bigquery.Client
	Logger log.Logger
}

func (bq *BigQuery) Session(ctx context.Context, id uint64) (routing.SessionMeta, error) {
	return routing.SessionMeta{}, nil
}

func (bq *BigQuery) SessionsWithUserHash(ctx context.Context, hash uint64) ([]routing.SessionMeta, error) {
	sessions := make([]routing.SessionMeta, 0)

	q := bq.Client.Query(fmt.Sprintf(`
		SELECT DISTINCT 
			sessionId AS SessionID,
			userId as UserHash,
			platformId AS PlatformID,
			datacenterId AS DatacenterID,
			latitude AS Latitude,
			longitude AS Longitude,
			isp AS ISP,
			connectionType AS ConnectionType,
			clientIpAddress AS ClientAddr,
			serverIpAddress AS ServerAddr,

			CONCAT(versionMajor, '.', versionMinor, '.', versionPatch) AS SDK,

			nextRtt AS NextRTT,
			directRtt AS DirectRTT
		FROM v3_dev.billing
		WHERE timestampStart = '%s' AND userId = '%s'
		LIMIT 10
	`, time.Now().Add(-3*720*time.Hour).Format("2006-01-02"), strconv.FormatUint(hash, 10)))

	itr, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	for {
		var s bqSession
		err := itr.Next(&s)
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		sessions = append(sessions, routing.SessionMeta{
			ID:         fmt.Sprintf("%x", uint64(s.SessionId)),
			UserHash:   fmt.Sprintf("%x", hash),
			Datacenter: s.DatacenterId,
			Location: routing.Location{
				Latitude:  s.Latitude,
				Longitude: s.Longitude,
				ISP:       s.Isp,
			},
			ClientAddr: s.ClientAddr,
			ServerAddr: s.ServerAddr,
			SDK:        s.SDK,
			NextRTT:    s.NextRTT,
			DirectRTT:  s.DirectRTT,
			Connection: routing.ConnectionTypeText(int32(s.ConnectionType)),
		})
	}

	return sessions, nil
}

func (bq *BigQuery) Slices(ctx context.Context, id uint64) ([]routing.SessionSlice, error) {
	return nil, nil
}
