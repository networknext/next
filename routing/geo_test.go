package routing_test

import (
	"net"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/routing"
)

func TestIPLocator(t *testing.T) {
	t.Run("Maxmind", func(t *testing.T) {
		mmreader, err := geoip2.Open("../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)

		mmdb := routing.MaxmindDB{
			Reader: mmreader,
		}

		{
			expected := routing.Location{
				Continent: "Europe",
				Country:   "United Kingdom",
				Region:    "England",
				City:      "London",
				Latitude:  51.5142,
				Longitude: -0.0931,
			}

			actual, err := mmdb.LocateIP(net.ParseIP("81.2.69.160"))
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		}

		{
			actual, err := mmdb.LocateIP(net.ParseIP("0.0.0.0"))
			assert.EqualError(t, err, "no location found for '0.0.0.0'")

			assert.Equal(t, routing.Location{}, actual)
		}

		{
			mmdb := routing.MaxmindDB{}

			actual, err := mmdb.LocateIP(net.ParseIP("0.0.0.0"))
			assert.EqualError(t, err, "not configured with a Maxmind DB")

			assert.Equal(t, routing.Location{}, actual)
		}
	})
}

func TestGeoClient(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	geoclient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "TESTING",
	}

	t.Run("Add", func(t *testing.T) {
		r1 := routing.Relay{
			ID:        1,
			Latitude:  38.115556,
			Longitude: 13.361389,
		}
		err := geoclient.Add(r1)
		assert.NoError(t, err)

		r2 := routing.Relay{
			ID:        2,
			Latitude:  37.502669,
			Longitude: 15.087269,
		}
		err = geoclient.Add(r2)
		assert.NoError(t, err)
	})

	t.Run("RelaysWithin", func(t *testing.T) {
		r1 := routing.Relay{
			ID:        1,
			Latitude:  38.115556,
			Longitude: 13.361389,
		}
		err := geoclient.Add(r1)
		assert.NoError(t, err)

		r2 := routing.Relay{
			ID:        2,
			Latitude:  37.502669,
			Longitude: 15.087269,
		}
		err = geoclient.Add(r2)
		assert.NoError(t, err)

		relays, err := geoclient.RelaysWithin(37, 15, 200, "km")
		assert.NoError(t, err)

		assert.Equal(t, 2, len(relays))
		assert.Equal(t, r2.ID, relays[0].ID)
		assert.Equal(t, r1.ID, relays[1].ID)
	})
}
