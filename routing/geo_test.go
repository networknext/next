package routing_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/routing"
)

type mockRoundTripper struct {
	Response *http.Response
}

func (mrt mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return mrt.Response, nil
}

func TestLocation(t *testing.T) {
	zeroloc := routing.Location{}
	assert.True(t, zeroloc.IsZero())

	loc := routing.Location{Latitude: 13, Longitude: 15}
	assert.False(t, loc.IsZero())
}

func TestNewMaxmindDBReader(t *testing.T) {
	t.Run("local file not found", func(t *testing.T) {
		_, err := routing.NewMaxmindReader(http.DefaultClient, "./file/not/found")
		assert.Error(t, err)
	})

	t.Run("local file found", func(t *testing.T) {
		_, err := routing.NewMaxmindReader(http.DefaultClient, "../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)
	})

	t.Run("url non-200 error code", func(t *testing.T) {
		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			}),
		)
		_, err := routing.NewMaxmindReader(http.DefaultClient, svr.URL)
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
		_, err := routing.NewMaxmindReader(http.DefaultClient, svr.URL)
		assert.Error(t, err)
		svr.Close()
	})

	t.Run("response gzipped but not tar", func(t *testing.T) {
		db, err := ioutil.ReadFile("../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)

		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				gw := gzip.NewWriter(w)
				defer gw.Close()

				gw.Write(db)
			}),
		)
		_, err = routing.NewMaxmindReader(http.DefaultClient, svr.URL)
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
		_, err := routing.NewMaxmindReader(http.DefaultClient, svr.URL)
		assert.Error(t, err)
		svr.Close()
	})

	t.Run("response gzipped tar mmdb", func(t *testing.T) {
		db, err := ioutil.ReadFile("../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)

		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				gw := gzip.NewWriter(w)
				tw := tar.NewWriter(gw)

				tw.WriteHeader(&tar.Header{
					Name: "GeoIP2-City-Test.mmdb",
					Size: int64(len(db)),
				})
				tw.Write(db)

				tw.Close()
				gw.Close()
			}),
		)
		_, err = routing.NewMaxmindReader(http.DefaultClient, svr.URL)
		assert.NoError(t, err)
		svr.Close()
	})
}

func TestIPLocator(t *testing.T) {
	t.Run("IPStack", func(t *testing.T) {

		{
			ipstack := routing.IPStack{
				Client: &http.Client{
					Transport: &mockRoundTripper{
						Response: &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
								"continent_name": "Europe",
								"country_name": "United Kingdom",
								"region_name": "England",
								"city": "Stroud",
								"latitude": 51.750999450683594,
								"longitude": -2.296999931335449,
								"connection": {
									"isp": "Andrews & Arnold Ltd"
								}
							}`)),
						},
					},
				},
				AccessKey: "2a3640e34301da9ab257c59243b0d7c6",
			}

			expected := routing.Location{
				Continent: "Europe",
				Country:   "United Kingdom",
				Region:    "England",
				City:      "Stroud",
				Latitude:  51.750999450683594,
				Longitude: -2.296999931335449,
				ISP:       "Andrews & Arnold Ltd",
			}

			actual, err := ipstack.LocateIP(net.ParseIP("81.2.69.160"))
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		}

		{
			ipstack := routing.IPStack{
				Client: &http.Client{
					Transport: &mockRoundTripper{
						Response: &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
								"latitude": 0,
								"longitude": 0
							}`)),
						},
					},
				},
				AccessKey: "2a3640e34301da9ab257c59243b0d7c6",
			}

			actual, err := ipstack.LocateIP(net.ParseIP("127.0.0.1"))
			assert.NoError(t, err)

			assert.Equal(t, routing.LocationNullIsland, actual)
		}

		{
			ipstack := routing.IPStack{
				Client: &http.Client{
					Transport: &mockRoundTripper{
						Response: &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
								"latitude": 0,
								"longitude": 0
							}`)),
						},
					},
				},
				AccessKey: "2a3640e34301da9ab257c59243b0d7c6",
			}

			actual, err := ipstack.LocateIP(([]byte)("localhost"))
			assert.NoError(t, err)

			assert.Equal(t, routing.LocationNullIsland, actual)
		}

		{
			ipstack := routing.IPStack{
				Client: &http.Client{
					Transport: &mockRoundTripper{
						Response: &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
								"latitude": 0,
								"longitude": 0
							}`)),
						},
					},
				},
				AccessKey: "2a3640e34301da9ab257c59243b0d7c6",
			}

			actual, err := ipstack.LocateIP(net.ParseIP("0.0.0.0"))
			assert.NoError(t, err)

			assert.Equal(t, routing.LocationNullIsland, actual)
		}
	})

	t.Run("Maxmind", func(t *testing.T) {
		cityreader, err := geoip2.Open("../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)

		ispreader, err := geoip2.Open("../testdata/GeoIP2-ISP-Test.mmdb")
		assert.NoError(t, err)

		mmdb := routing.MaxmindDB{
			CityReader: cityreader,
			ISPReader:  ispreader,
		}

		{
			expected := routing.Location{
				Continent:   "Europe",
				Country:     "United Kingdom",
				CountryCode: "GB",
				Region:      "England",
				City:        "London",
				Latitude:    51.5142,
				Longitude:   -0.0931,
				ISP:         "Andrews & Arnold Ltd",
			}

			actual, err := mmdb.LocateIP(net.ParseIP("81.2.69.160"))
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		}

		{
			actual, err := mmdb.LocateIP(net.ParseIP("127.0.0.1"))
			assert.NoError(t, err)

			assert.Equal(t, routing.LocationNullIsland, actual)
		}

		{
			actual, err := mmdb.LocateIP(([]byte)("localhost"))
			assert.NoError(t, err)

			assert.Equal(t, routing.LocationNullIsland, actual)
		}

		{
			actual, err := mmdb.LocateIP(net.ParseIP("0.0.0.0"))
			assert.NoError(t, err)

			assert.Equal(t, routing.LocationNullIsland, actual)
		}

		{
			mmdb := routing.MaxmindDB{}
			actual, err := mmdb.LocateIP(net.ParseIP("0.0.0.0"))
			assert.EqualError(t, err, "not configured with a Maxmind City DB")

			assert.Equal(t, routing.Location{}, actual)
		}

		{
			mmdb := routing.MaxmindDB{
				CityReader: cityreader,
			}
			actual, err := mmdb.LocateIP(net.ParseIP("0.0.0.0"))
			assert.EqualError(t, err, "not configured with a Maxmind ISP DB")

			assert.Equal(t, routing.Location{}, actual)
		}

		// Fail to locate IP because the database cannot be read from
		{
			mmdb := routing.MaxmindDB{
				CityReader: cityreader,
				ISPReader:  ispreader,
			}

			cityreader.Close()
			ispreader.Close()

			actual, err := mmdb.LocateIP(net.ParseIP("0.0.0.0"))
			assert.NoError(t, err)

			assert.Equal(t, routing.LocationNullIsland, actual)
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
		err := geoclient.Add(1, 38.115556, 13.361389)
		assert.NoError(t, err)

		err = geoclient.Add(2, 37.502669, 15.087269)
		assert.NoError(t, err)
	})

	t.Run("RelaysWithin", func(t *testing.T) {
		err := geoclient.Add(1, 38.115556, 13.361389)
		assert.NoError(t, err)

		err = geoclient.Add(2, 37.502669, 15.087269)
		assert.NoError(t, err)

		relays, err := geoclient.RelaysWithin(37, 15, 200, "km")
		assert.NoError(t, err)

		assert.Equal(t, 2, len(relays))
		assert.Equal(t, uint64(2), relays[0].ID)
		assert.Equal(t, uint64(1), relays[1].ID)

		// Bad georadius call to redis
		_, err = geoclient.RelaysWithin(37, 15, 200, "invalid")
		assert.Error(t, err)

		// Unable to parse name from redis
		geoloc := redis.GeoLocation{
			Name:      "bad data",
			Latitude:  37.502669,
			Longitude: 15.087269,
		}

		err = geoclient.RedisClient.GeoAdd(geoclient.Namespace, &geoloc).Err()
		assert.NoError(t, err)

		_, err = geoclient.RelaysWithin(37, 15, 200, "km")
		assert.Error(t, err)
	})
}
