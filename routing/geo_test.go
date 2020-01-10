package routing_test

import (
	"log"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/routing"
)

func TestGeoClient(t *testing.T) {
	t.Skip()

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{})

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

		// GeoAdd is really just sorted sets so we can get them with ZMEMBERS
		relays, err := redisServer.ZMembers(geoclient.Namespace)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(relays))
		assert.Equal(t, "1", relays[0])
		assert.Equal(t, "2", relays[1])
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

		log.Fatal(relays)
	})
}
