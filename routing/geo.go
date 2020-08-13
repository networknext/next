package routing

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/metrics"
	"github.com/oschwald/geoip2-golang"
)

const (
	LocationVersion = 0

	regexLocalhostIPs = `0\.0\.0\.0|127\.0\.0\.1|localhost`
)

func isLocalHost(ip net.IP) bool {
	// if the ip is localhost, return nothing so we can test on our dev machines
	matches, _ := regexp.Match(regexLocalhostIPs, []byte(ip.String()))
	localhostMatches, _ := regexp.Match(regexLocalhostIPs, ip) // For the "localhost" case

	return matches || localhostMatches
}

// IPLocator defines anything that returns a routing.Location given an net.IP
type IPLocator interface {
	LocateIP(net.IP) (Location, error)
}

// Location represents a lat/long on Earth with additional metadata
type Location struct {
	Continent   string  `json:"continent"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
	ASN         int     `json:"asn"`
}

func (l *Location) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[Location] invalid read at version number")
	}

	if version > LocationVersion {
		return fmt.Errorf("unknown location version: %d", version)
	}

	if !encoding.ReadString(data, &index, &l.Continent, math.MaxInt32) {
		return errors.New("[Location] invalid read at continent")
	}

	if !encoding.ReadString(data, &index, &l.Country, math.MaxInt32) {
		return errors.New("[Location] invalid read at country")
	}

	if !encoding.ReadString(data, &index, &l.CountryCode, math.MaxInt32) {
		return errors.New("[Location] invalid read at country code")
	}

	if !encoding.ReadString(data, &index, &l.Region, math.MaxInt32) {
		return errors.New("[Location] invalid read at region")
	}

	if !encoding.ReadString(data, &index, &l.City, math.MaxInt32) {
		return errors.New("[Location] invalid read at city")
	}

	if !encoding.ReadFloat64(data, &index, &l.Latitude) {
		return errors.New("[Location] invalid read at latitude")
	}

	if !encoding.ReadFloat64(data, &index, &l.Longitude) {
		return errors.New("[Location] invalid read at longitude")
	}

	if !encoding.ReadString(data, &index, &l.ISP, math.MaxInt32) {
		return errors.New("[Location] invalid read at ISP")
	}

	var asn uint32
	if !encoding.ReadUint32(data, &index, &asn) {
		return errors.New("[Location] invalid read at ASN")
	}
	l.ASN = int(asn)

	return nil
}

func (l Location) MarshalBinary() ([]byte, error) {
	data := make([]byte, l.Size())
	index := 0

	encoding.WriteUint32(data, &index, LocationVersion)
	encoding.WriteString(data, &index, l.Continent, uint32(len(l.Continent)))
	encoding.WriteString(data, &index, l.Country, uint32(len(l.Country)))
	encoding.WriteString(data, &index, l.CountryCode, uint32(len(l.CountryCode)))
	encoding.WriteString(data, &index, l.Region, uint32(len(l.Region)))
	encoding.WriteString(data, &index, l.City, uint32(len(l.City)))
	encoding.WriteFloat64(data, &index, l.Latitude)
	encoding.WriteFloat64(data, &index, l.Longitude)
	encoding.WriteString(data, &index, l.ISP, uint32(len(l.ISP)))
	encoding.WriteUint32(data, &index, uint32(l.ASN))

	return data, nil
}

func (l Location) Size() uint64 {
	return uint64(4 + 4 + len(l.Continent) + 4 + len(l.Country) + 4 + len(l.CountryCode) + 4 + len(l.Region) + 4 + len(l.City) + 8 + 8 + 4 + len(l.ISP) + 4)
}

// IsZero reports whether l represents the zero location lat/long 0,0 similar to how Time.IsZero works.
func (l *Location) IsZero() bool {
	return l.Latitude == 0 && l.Longitude == 0
}

func (l Location) RedisString() string {
	return fmt.Sprintf("%.2f|%.2f|%s", l.Latitude, l.Longitude, l.ISP)
}

func (l *Location) ParseRedisString(values []string) error {
	var index int
	var err error

	if l.Latitude, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Location] failed to read latitude from redis data: %v", err)
	}
	index++

	if l.Longitude, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[Location] failed to read longitude from redis data: %v", err)
	}
	index++

	l.ISP = values[index]
	index++

	return nil
}

type IPStack struct {
	*http.Client

	AccessKey string
}

type ipStackResponse struct {
	IP            string  `json:"ip"`
	Type          string  `json:"type"`
	ContinentCode string  `json:"continent_code"`
	ContinentName string  `json:"continent_name"`
	CountryCode   string  `json:"country_code"`
	CountryName   string  `json:"country_name"`
	RegionCode    string  `json:"region_code"`
	RegionName    string  `json:"region_name"`
	City          string  `json:"city"`
	Zip           string  `json:"zip"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Location      struct {
		GeonameID int    `json:"geoname_id"`
		Capital   string `json:"capital"`
		Languages []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Native string `json:"native"`
		} `json:"languages"`
		CountryFlag             string `json:"country_flag"`
		CountryFlagEmoji        string `json:"country_flag_emoji"`
		CountryFlagEmojiUnicode string `json:"country_flag_emoji_unicode"`
		CallingCode             string `json:"calling_code"`
		IsEU                    bool   `json:"is_eu"`
	} `json:"location"`
	TimeZone struct {
		ID               string `json:"id"`
		CurrentTime      string `json:"current_time"`
		GMTOffset        int    `json:"gmt_offset"`
		Code             string `json:"code"`
		IsDaylightSaving bool   `json:"is_daylight_saving"`
	} `json:"time_zone"`
	Currency struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		Plural       string `json:"plural"`
		Symbol       string `json:"symbol"`
		SymbolNative string `json:"symbol_native"`
	} `json:"currency"`
	Connection struct {
		ASN int    `json:"asn"`
		ISP string `json:"isp"`
	} `json:"connection"`
}

func (ips *IPStack) LocateIP(ip net.IP) (Location, error) {
	if isLocalHost(ip) {
		return LocationNullIsland, nil
	}

	res, err := ips.Get(fmt.Sprintf("https://api.ipstack.com/%s?access_key=%s", ip.String(), ips.AccessKey))
	if err != nil {
		return Location{}, err
	}
	defer res.Body.Close()

	var ipstackres ipStackResponse
	if err := jsoniter.NewDecoder(res.Body).Decode(&ipstackres); err != nil {
		return Location{}, err
	}

	if ipstackres.Latitude == 0 && ipstackres.Longitude == 0 {
		return Location{}, fmt.Errorf("no location found for '%s'", ip.String())
	}

	return Location{
		Continent: ipstackres.ContinentName,
		Country:   ipstackres.CountryName,
		Region:    ipstackres.RegionName,
		City:      ipstackres.City,
		Latitude:  ipstackres.Latitude,
		Longitude: ipstackres.Longitude,
		ISP:       ipstackres.Connection.ISP,
	}, nil
}

// MaxmindDB embeds the unofficial MaxmindDB reader so we can satisfy the IPLocator interface
type MaxmindDB struct {
	HTTPClient *http.Client
	CityURI    string
	IspURI     string

	cityReader *geoip2.Reader
	ispReader  *geoip2.Reader
}

func (mmdb *MaxmindDB) Sync(ctx context.Context, metrics *metrics.MaxmindSyncMetrics) error {
	metrics.Invocations.Add(1)
	durationStart := time.Now()

	if err := mmdb.OpenCity(ctx, mmdb.HTTPClient, mmdb.CityURI); err != nil {
		metrics.ErrorMetrics.FailedToSync.Add(1)
		return fmt.Errorf("could not open maxmind db uri: %v", err)
	}
	if err := mmdb.OpenISP(ctx, mmdb.HTTPClient, mmdb.IspURI); err != nil {
		metrics.ErrorMetrics.FailedToSyncISP.Add(1)
		return fmt.Errorf("could not open maxmind db isp uri: %v", err)
	}

	duration := time.Since(durationStart)
	metrics.DurationGauge.Set(float64(duration.Milliseconds()))

	return nil
}

func (mmdb *MaxmindDB) OpenCity(ctx context.Context, httpClient *http.Client, uri string) error {
	reader, err := mmdb.openMaxmindDB(ctx, httpClient, uri)
	if err != nil {
		return err
	}

	mmdb.cityReader = reader
	return nil
}

func (mmdb *MaxmindDB) OpenISP(ctx context.Context, httpClient *http.Client, uri string) error {
	reader, err := mmdb.openMaxmindDB(ctx, httpClient, uri)
	if err != nil {
		return err
	}

	mmdb.ispReader = reader
	return nil
}

func (mmdb *MaxmindDB) openMaxmindDB(ctx context.Context, httpClient *http.Client, uri string) (*geoip2.Reader, error) {
	var err error

	// If there is a local file at this uri then just open it
	if _, err := os.Stat(uri); err == nil {
		return geoip2.Open(uri)
	}

	// otherwise attempt to download it
	mmres, err := httpClient.Get(uri)
	if err != nil {
		return nil, err
	}
	defer mmres.Body.Close()

	if mmres.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %d: %s from Maxmind.com", mmres.StatusCode, http.StatusText(mmres.StatusCode))
	}

	gz, err := gzip.NewReader(mmres.Body)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	buf := bytes.NewBuffer(nil)
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(hdr.Name, "mmdb") {
			_, err := io.Copy(buf, tr)
			if err != nil {
				return nil, err
			}
		}
	}

	return geoip2.FromBytes(buf.Bytes())
}

// LocateIP queries the Maxmind geoip2.Reader for the net.IP and parses the response into a routing.Location
func (mmdb *MaxmindDB) LocateIP(ip net.IP) (Location, error) {
	if mmdb.cityReader == nil {
		return Location{}, errors.New("not configured with a Maxmind City DB")
	}
	if mmdb.ispReader == nil {
		return Location{}, errors.New("not configured with a Maxmind ISP DB")
	}

	if isLocalHost(ip) {
		return LocationNullIsland, nil
	}

	cityres, err := mmdb.cityReader.City(ip)
	if err != nil {
		return Location{}, err
	}

	continent := "unknown"
	if val, ok := cityres.Continent.Names["en"]; ok {
		continent = val
	}
	country := "unknown"
	if val, ok := cityres.Country.Names["en"]; ok {
		country = val
	}
	countryCode := "unknown"
	if cityres.Country.IsoCode != "" {
		countryCode = cityres.Country.IsoCode
	}
	region := "unknown"
	if len(cityres.Subdivisions) > 0 {
		if val, ok := cityres.Subdivisions[0].Names["en"]; ok {
			region = val
		}
	}
	city := "unknown"
	if val, ok := cityres.City.Names["en"]; ok {
		city = val
	}

	ispres, err := mmdb.ispReader.ISP(ip)
	if err != nil {
		return Location{}, err
	}

	return Location{
		Continent:   continent,
		CountryCode: countryCode,
		Country:     country,
		Region:      region,
		City:        city,
		Latitude:    cityres.Location.Latitude,
		Longitude:   cityres.Location.Longitude,
		ISP:         ispres.ISP,
		ASN:         int(ispres.AutonomousSystemNumber),
	}, nil
}

// LocateIPFunc allows anyone to define a custom function to lookup a net.IP for a routing.Location
type LocateIPFunc func(net.IP) (Location, error)

// LocateIP just invokes the func itself and allows this to satisfy the IPLocator interface
func (f LocateIPFunc) LocateIP(ip net.IP) (Location, error) {
	return f(ip)
}

var LocationNullIsland = Location{
	Continent: "Null Island",
	Country:   "Null Island",
	Region:    "Null Island",
	City:      "Null Island",
	ISP:       "Water",
}

var NullIsland = LocateIPFunc(func(ip net.IP) (Location, error) {
	return LocationNullIsland, nil
})
