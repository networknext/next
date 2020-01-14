package routing

import (
	"net"
	"strconv"

	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
)

// IPLocator defines anything that returns a routing.Location given an net.IP
type IPLocator interface {
	LocateIP(net.IP) (Location, error)
}

// Location represents a lat/long on Earth with additional metadata
type Location struct {
	Continent string
	Country   string
	Region    string
	City      string
	Latitude  float64
	Longitude float64
}

// MaxmindDB embeds the unofficial MaxmindDB reader so we can satisfy the IPLocator interface
type MaxmindDB struct {
	*geoip2.Reader
}

// LocateIP queries the Maxmind geoip2.Reader for the net.IP and parses the response into a routing.Location
func (mmdb *MaxmindDB) LocateIP(ip net.IP) (Location, error) {
	res, err := mmdb.Reader.City(ip)
	if err != nil {
		return Location{}, err
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
