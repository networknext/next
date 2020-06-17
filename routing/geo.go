package routing

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-redis/redis/v7"
	jsoniter "github.com/json-iterator/go"
	"github.com/oschwald/geoip2-golang"
)

const (
	regexLocalhostIPs = `127\.0\.0\.1|localhost`
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
	Continent string  `json:"continent"`
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ISP       string  `json:"isp"`
}

func (l *Location) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, l)
}

func (l Location) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(l)
}

// IsZero reports whether l represents the zero location lat/long 0,0 similar to how Time.IsZero works.
func (l *Location) IsZero() bool {
	return l.Latitude == 0 && l.Longitude == 0
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
	*geoip2.Reader
}

// LocateIP queries the Maxmind geoip2.Reader for the net.IP and parses the response into a routing.Location
func (mmdb *MaxmindDB) LocateIP(ip net.IP) (Location, error) {
	if mmdb.Reader == nil {
		return Location{}, errors.New("not configured with a Maxmind DB")
	}

	if isLocalHost(ip) {
		return LocationNullIsland, nil
	}

	res, err := mmdb.Reader.City(ip)
	if err != nil {
		return Location{}, err
	}

	if len(res.City.Names) <= 0 {
		return Location{}, fmt.Errorf("no location found for '%s'", ip.String())
	}

	continent := "unknown"
	if val, ok := res.Continent.Names["en"]; ok {
		continent = val
	}
	country := "unknown"
	if val, ok := res.Country.Names["en"]; ok {
		country = val
	}
	region := "unknown"
	if len(res.Subdivisions) > 0 {
		if val, ok := res.Subdivisions[0].Names["en"]; ok {
			region = val
		}
	}
	city := "unknown"
	if val, ok := res.City.Names["en"]; ok {
		city = val
	}

	return Location{
		Continent: continent,
		Country:   country,
		Region:    region,
		City:      city,
		Latitude:  res.Location.Latitude,
		Longitude: res.Location.Longitude,
		ISP:       "unknown",
	}, nil
}

// LocateIPFunc allows anyone to define a custom function to lookup and net.IP for a routing.Location
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

type GeoClient struct {
	RedisClient redis.Cmdable
	Namespace   string
}

func (c *GeoClient) Add(relayID uint64, latitude float64, longitude float64) error {
	geoloc := redis.GeoLocation{
		Name:      strconv.FormatUint(relayID, 10),
		Latitude:  latitude,
		Longitude: longitude,
	}

	return c.RedisClient.GeoAdd(c.Namespace, &geoloc).Err()
}

func (c *GeoClient) Remove(relayID uint64) error {
	return c.RedisClient.ZRem(c.Namespace, strconv.FormatUint(relayID, 10)).Err()
}

// uom can be one of the following: "m", "km", "mi", "ft"
func (c *GeoClient) RelaysWithin(lat float64, long float64, radius float64, uom string) ([]Relay, error) {
	geoquery := redis.GeoRadiusQuery{
		Radius:    radius,
		Unit:      uom,
		WithCoord: true,
		Sort:      "ASC",
	}

	res := c.RedisClient.GeoRadius(c.Namespace, long, lat, &geoquery)

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
			ID: id,
			Datacenter: Datacenter{
				Location: Location{
					Latitude:  geoloc.Latitude,
					Longitude: geoloc.Longitude,
				},
			},
		}
	}

	return relays, nil
}
