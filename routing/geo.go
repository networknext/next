package routing

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
)

const (
	regexLocalhostIPs = `127\.0\.0\.1|localhost`
)

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

	// if the ip is localhost, return nothing so we can test on our dev machines
	matches, _ := regexp.Match(regexLocalhostIPs, []byte(ip.String()))
	localhostMatches, _ := regexp.Match(regexLocalhostIPs, ip) // For the "localhost" case

	if matches || localhostMatches {
		return Location{}, nil
	}

	res, err := mmdb.Reader.City(ip)
	if err != nil {
		return Location{}, err
	}

	if len(res.City.Names) <= 0 {
		return Location{}, fmt.Errorf("no location found for '%s'", ip.String())
	}

	return Location{
		Continent: res.Continent.Names["en"],
		Country:   res.Country.Names["en"],
		Region:    res.Subdivisions[0].Names["en"],
		City:      res.City.Names["en"],
		Latitude:  res.Location.Latitude,
		Longitude: res.Location.Longitude,
	}, nil
}

// LocateIPFunc allows anyone to define a custom function to lookup and net.IP for a routing.Location
type LocateIPFunc func(net.IP) (Location, error)

// LocateIP just invokes the func itself and allows this to satisfy the IPLocator interface
func (f LocateIPFunc) LocateIP(ip net.IP) (Location, error) {
	return f(ip)
}

var NullIsland = LocateIPFunc(func(ip net.IP) (Location, error) {
	return Location{}, nil
})

type GeoClient struct {
	RedisClient redis.Cmdable
	Namespace   string
}

func (c *GeoClient) Add(r Relay) error {
	geoloc := redis.GeoLocation{
		Name:      strconv.FormatUint(r.ID, 10),
		Latitude:  r.Datacenter.Location.Latitude,
		Longitude: r.Datacenter.Location.Longitude,
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
