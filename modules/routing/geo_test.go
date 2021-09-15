package routing_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/modules/routing"
)

type mockRoundTripper struct {
	Response *http.Response
}

func (mrt mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return mrt.Response, nil
}

// Helper function to create a random string of a specified length
// Useful for testing constant string lengths
// Adapted from: https://stackoverflow.com/a/22892986
func generateRandomStringSequence(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func TestLocation(t *testing.T) {
	zeroloc := routing.Location{}
	assert.True(t, zeroloc.IsZero())

	loc := routing.Location{Latitude: 13, Longitude: 15}
	assert.False(t, loc.IsZero())
}

func testLocation() routing.Location {
	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	data := routing.Location{
		Continent:   generateRandomStringSequence(rand.Intn(routing.MaxContinentLength - 1)),
		Country:     generateRandomStringSequence(rand.Intn(routing.MaxCountryLength - 1)),
		CountryCode: generateRandomStringSequence(rand.Intn(routing.MaxCountryCodeLength - 1)),
		Region:      generateRandomStringSequence(rand.Intn(routing.MaxRegionLength - 1)),
		City:        generateRandomStringSequence(rand.Intn(routing.MaxCityLength - 1)),
		Latitude:    rand.Float32(),
		Longitude:   rand.Float32(),
		ISP:         generateRandomStringSequence(rand.Intn(routing.MaxISPNameLength - 1)),
		ASN:         rand.Int(),
	}

	// Zero out unused fields during serialization
	data.Continent = ""
	data.Country = ""
	data.CountryCode = ""
	data.Region = ""
	data.City = ""

	return data
}

func TestLocationSerialize(t *testing.T) {
	t.Parallel()

	t.Run("test serialize write", func(t *testing.T) {
		locData := testLocation()

		data, err := routing.WriteLocation(&locData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("test serialize read", func(t *testing.T) {
		locData := testLocation()

		data, err := routing.WriteLocation(&locData)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var readLocData routing.Location

		err = routing.ReadLocation(&readLocData, data)

		assert.NoError(t, err)
		assert.Equal(t, locData, readLocData)
	})
}

func TestNewMaxmindDBReader(t *testing.T) {
	t.Parallel()

	t.Run("local file not found", func(t *testing.T) {
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), "./file/not/found")
		assert.Error(t, err)
	})

	t.Run("local file found", func(t *testing.T) {
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), "../../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)
	})

	t.Run("url non-200 error code", func(t *testing.T) {
		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			}),
		)
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), "./file/not/found")
		assert.Error(t, err)
		svr.Close()
	})

	t.Run("response not gzipped", func(t *testing.T) {
		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not gzip data"))
			}),
		)
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), "./file/not/found")
		assert.Error(t, err)
		svr.Close()
	})

	t.Run("response gzipped but not tar", func(t *testing.T) {
		db, err := ioutil.ReadFile("../../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)

		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				gw := gzip.NewWriter(w)
				defer gw.Close()

				gw.Write(db)
			}),
		)
		r := routing.MaxmindDB{}
		err = r.OpenCity(context.Background(), "./file/not/found")
		assert.Error(t, err)
		svr.Close()
	})

	t.Run("response gzipped tar but no mmdb", func(t *testing.T) {
		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				gw := gzip.NewWriter(w)
				defer gw.Close()

				tw := tar.NewWriter(gw)
				defer tw.Close()

				tw.WriteHeader(&tar.Header{
					Name: "not-a-db",
				})
				tw.Write([]byte("just some text"))
			}),
		)
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), "./file/not/found")
		assert.Error(t, err)
		svr.Close()
	})
}

func TestIPLocator(t *testing.T) {
	t.Parallel()

	t.Run("Maxmind", func(t *testing.T) {
		mmdb := routing.MaxmindDB{}

		err := mmdb.OpenCity(context.Background(), "../../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)
		err = mmdb.OpenISP(context.Background(), "../../testdata/GeoIP2-ISP-Test.mmdb")
		assert.NoError(t, err)

		cityreader, err := geoip2.Open("../../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)

		ispreader, err := geoip2.Open("../../testdata/GeoIP2-ISP-Test.mmdb")
		assert.NoError(t, err)

		{
			expected := routing.Location{
				Latitude:  51.5142,
				Longitude: -0.0931,
				ISP:       "Andrews & Arnold Ltd",
			}

			actual, err := mmdb.LocateIP(net.ParseIP("81.2.69.160"), rand.Uint64())
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		}

		// Fail to locate IP because the database cannot be read from
		{
			mmdb := routing.MaxmindDB{}

			cityreader.Close()
			ispreader.Close()

			_, err := mmdb.LocateIP(net.ParseIP("81.2.69.160"), rand.Uint64())
			assert.EqualError(t, err, "not configured with a Maxmind City DB")

		}
	})
}
