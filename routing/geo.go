package routing

import (
	"strconv"

	"github.com/go-redis/redis/v7"
)

type GeoClient struct {
	RedisClient redis.Cmdable
	Namespace   string
}

func (c *GeoClient) Add(r Relay) error {
	geoloc := redis.GeoLocation{
		Name:      strconv.FormatUint(r.ID, 10),
		Latitude:  r.Latitude,
		Longitude: r.Longitude,
	}

	return c.RedisClient.GeoAdd(c.Namespace, &geoloc).Err()
}

func (c *GeoClient) RelaysWithin(lat float64, long float64, radius float64, uom string) ([]Relay, error) {
	geoquery := redis.GeoRadiusQuery{
		Radius:    radius,
		Unit:      uom,
		WithCoord: true,
		Sort:      "ASC",
	}

	res := c.RedisClient.GeoRadius(c.Namespace, long, lat, &geoquery)
	if res.Err() != nil {
		return nil, res.Err()
	}

	geolocs, err := res.Result()
	if err != nil {
		return nil, err
	}

	relays := make([]Relay, len(geolocs))
	for idx, geoloc := range geolocs {
		id, err := strconv.ParseUint(geoloc.Name, 10, 64)
		if err != nil {
			return nil, err
		}

		relays[idx] = Relay{
			ID:        id,
			Latitude:  geoloc.Latitude,
			Longitude: geoloc.Longitude,
		}
	}

	return relays, nil
}
