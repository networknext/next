package routing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	"github.com/oschwald/geoip2-golang"
)

const (
	LocationVersion = 1

	MaxContinentLength   = 16
	MaxCountryLength     = 64
	MaxCountryCodeLength = 16
	MaxRegionLength      = 64
	MaxCityLength        = 128
	MaxISPNameLength     = 64

	MaxLocationSize = 128
)

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
	Latitude    float32 `json:"latitude"`
	Longitude   float32 `json:"longitude"`
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

	if version == 0 {
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
	}

	if version == 0 {
		var lat float64
		if !encoding.ReadFloat64(data, &index, &lat) {
			return errors.New("[Location] invalid read at latitude")
		}
		l.Latitude = float32(lat)

		var long float64
		if !encoding.ReadFloat64(data, &index, &long) {
			return errors.New("[Location] invalid read at longitude")
		}
		l.Longitude = float32(long)

		if !encoding.ReadString(data, &index, &l.ISP, math.MaxInt32) {
			return errors.New("[Location] invalid read at ISP")
		}
	} else {
		if !encoding.ReadFloat32(data, &index, &l.Latitude) {
			return errors.New("[Location] invalid read at latitude")
		}

		if !encoding.ReadFloat32(data, &index, &l.Longitude) {
			return errors.New("[Location] invalid read at longitude")
		}

		if !encoding.ReadString(data, &index, &l.ISP, MaxISPNameLength) {
			return errors.New("[Location] invalid read at ISP")
		}
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
	encoding.WriteFloat32(data, &index, l.Latitude)
	encoding.WriteFloat32(data, &index, l.Longitude)
	encoding.WriteString(data, &index, l.ISP, MaxISPNameLength)
	encoding.WriteUint32(data, &index, uint32(l.ASN))

	return data, nil
}

func (l *Location) Serialize(stream encoding.Stream) error {
	var version int32
	if stream.IsWriting() {
		version = int32(LocationVersion)
	}
	stream.SerializeInteger(&version, 0, LocationVersion)

	stream.SerializeFloat32(&l.Latitude)
	stream.SerializeFloat32(&l.Longitude)

	stream.SerializeString(&l.ISP, MaxISPNameLength)

	var asn uint64
	if stream.IsWriting() {
		asn = uint64(l.ASN)
	}
	stream.SerializeUint64(&asn)
	if stream.IsReading() {
		l.ASN = int(asn)
	}

	return stream.Error()
}

func (l Location) Size() uint64 {
	ispLength := len(l.ISP)
	if ispLength > MaxISPNameLength {
		ispLength = MaxISPNameLength
	}

	return uint64(4 + 4 + 4 + 4 + ispLength + 4)
}

func WriteLocation(entry *Location) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during Location packet entry write: %v\n", r)
		}
	}()

	buffer := make([]byte, MaxLocationSize)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadLocation(entry *Location, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
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

	var lat float64
	if lat, err = strconv.ParseFloat(values[index], 32); err != nil {
		return fmt.Errorf("[Location] failed to read latitude from redis data: %v", err)
	}
	l.Latitude = float32(lat)
	index++

	var long float64
	if long, err = strconv.ParseFloat(values[index], 32); err != nil {
		return fmt.Errorf("[Location] failed to read longitude from redis data: %v", err)
	}
	l.Longitude = float32(long)
	index++

	l.ISP = values[index]
	index++

	return nil
}

// MaxmindDB embeds the unofficial MaxmindDB reader so we can satisfy the IPLocator interface
type MaxmindDB struct {
	CityFile string
	IspFile  string

	cityReader *geoip2.Reader
	ispReader  *geoip2.Reader
}

func (mmdb *MaxmindDB) Sync(ctx context.Context, metrics *metrics.MaxmindSyncMetrics) error {
	metrics.Invocations.Add(1)
	durationStart := time.Now()

	if err := mmdb.OpenCity(ctx, mmdb.CityFile); err != nil {
		metrics.ErrorMetrics.FailedToSync.Add(1)
		return fmt.Errorf("could not open maxmind db uri: %v", err)
	}
	if err := mmdb.OpenISP(ctx, mmdb.IspFile); err != nil {
		metrics.ErrorMetrics.FailedToSyncISP.Add(1)
		return fmt.Errorf("could not open maxmind db isp uri: %v", err)
	}

	duration := time.Since(durationStart)
	metrics.DurationGauge.Set(float64(duration.Milliseconds()))

	return nil
}

func (mmdb *MaxmindDB) OpenCity(ctx context.Context, file string) error {
	reader, err := mmdb.openMaxmindDB(ctx, file)
	if err != nil {
		return err
	}

	mmdb.cityReader = reader
	return nil
}

func (mmdb *MaxmindDB) OpenISP(ctx context.Context, file string) error {
	reader, err := mmdb.openMaxmindDB(ctx, file)
	if err != nil {
		return err
	}

	mmdb.ispReader = reader
	return nil
}

func (mmdb *MaxmindDB) openMaxmindDB(ctx context.Context, file string) (*geoip2.Reader, error) {
	// If there is a local file at this uri then just open it
	if _, err := os.Stat(file); err != nil {
		return nil, err
	}

	return geoip2.Open(file)
}

// LocateIP queries the Maxmind geoip2.Reader for the net.IP and parses the response into a routing.Location
func (mmdb *MaxmindDB) LocateIP(ip net.IP) (Location, error) {
	if mmdb.cityReader == nil {
		return Location{}, errors.New("not configured with a Maxmind City DB")
	}
	if mmdb.ispReader == nil {
		return Location{}, errors.New("not configured with a Maxmind ISP DB")
	}

	cityres, err := mmdb.cityReader.City(ip)
	if err != nil {
		return Location{}, err
	}

	ispres, err := mmdb.ispReader.ISP(ip)
	if err != nil {
		return Location{}, err
	}

	fmt.Println("Using DB files to look up location")

	return Location{
		Latitude:  float32(cityres.Location.Latitude),
		Longitude: float32(cityres.Location.Longitude),
		ISP:       ispres.ISP,
		ASN:       int(ispres.AutonomousSystemNumber),
	}, nil
}

// LocateIPFunc allows anyone to define a custom function to lookup a net.IP for a routing.Location
type LocateIPFunc func(net.IP) (Location, error)

// LocateIP just invokes the func itself and allows this to satisfy the IPLocator interface
func (f LocateIPFunc) LocateIP(ip net.IP) (Location, error) {
	return f(ip)
}

var LocationNullIsland = Location{
	ISP: "Water",
}

var NullIsland = LocateIPFunc(func(ip net.IP) (Location, error) {
	return LocationNullIsland, nil
})
