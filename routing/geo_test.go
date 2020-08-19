package routing_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

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
	t.Parallel()

	t.Run("local file not found", func(t *testing.T) {
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), http.DefaultClient, "./file/not/found")
		assert.Error(t, err)
	})

	t.Run("local file found", func(t *testing.T) {
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), http.DefaultClient, "../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)
	})

	t.Run("url non-200 error code", func(t *testing.T) {
		svr := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			}),
		)
		r := routing.MaxmindDB{}
		err := r.OpenCity(context.Background(), http.DefaultClient, svr.URL)
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
		err := r.OpenCity(context.Background(), http.DefaultClient, svr.URL)
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
		r := routing.MaxmindDB{}
		err = r.OpenCity(context.Background(), http.DefaultClient, svr.URL)
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
		err := r.OpenCity(context.Background(), http.DefaultClient, svr.URL)
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
		r := routing.MaxmindDB{}
		err = r.OpenCity(context.Background(), http.DefaultClient, svr.URL)
		assert.NoError(t, err)
		svr.Close()
	})
}

func TestIPLocator(t *testing.T) {
	t.Parallel()

	t.Run("Maxmind", func(t *testing.T) {
		mmdb := routing.MaxmindDB{}

		err := mmdb.OpenCity(context.Background(), nil, "../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)
		err = mmdb.OpenISP(context.Background(), nil, "../testdata/GeoIP2-ISP-Test.mmdb")
		assert.NoError(t, err)

		cityreader, err := geoip2.Open("../testdata/GeoIP2-City-Test.mmdb")
		assert.NoError(t, err)

		ispreader, err := geoip2.Open("../testdata/GeoIP2-ISP-Test.mmdb")
		assert.NoError(t, err)

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

		// Fail to locate IP because the database cannot be read from
		{
			mmdb := routing.MaxmindDB{}

			cityreader.Close()
			ispreader.Close()

			_, err := mmdb.LocateIP(net.ParseIP("81.2.69.160"))
			assert.EqualError(t, err, "not configured with a Maxmind City DB")

		}
	})
}
