package ip2location

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"

	"github.com/networknext/next/modules/core"

	"github.com/oschwald/maxminddb-golang"
)

type City struct {
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
}

type ISP struct {
	ISP string `maxminddb:"isp"`
}

func Bash(command string) error {

	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		return fmt.Errorf("could not run bash")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: failed to run command: %v", err)
		os.Exit(1)
	}

	cmd.Wait()

	return nil
}

func DownloadDatabases_MaxMind(licenseKey string) error {

	dir, err := ioutil.TempDir("/tmp", "database-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	core.Log("downloading isp database")

	err = Bash(fmt.Sprintf("curl 'https://download.maxmind.com/app/geoip_download?edition_id=GeoIP2-ISP&license_key=%s&suffix=tar.gz' --output %s/GeoIP2-ISP.tar.gz", licenseKey, dir))
	if err != nil {
		return err
	}

	core.Log("downloading city database")

	err = Bash(fmt.Sprintf("rm -f GeoIP2-City.tar.gz && curl 'https://download.maxmind.com/app/geoip_download?edition_id=GeoIP2-City&license_key=%s&suffix=tar.gz' --output %s/GeoIP2-City.tar.gz", licenseKey, dir))
	if err != nil {
		return err
	}

	core.Log("decompressing databases")

	Bash(fmt.Sprintf("cd %s && tar -zxf GeoIP2-ISP.tar.gz", dir))
	Bash(fmt.Sprintf("cd %s && tar -zxf GeoIP2-City.tar.gz", dir))

	err = Bash(fmt.Sprintf("mv %s/GeoIP2-ISP_*/GeoIP2-ISP.mmdb %s", dir, dir))
	if err != nil {
		return err
	}

	err = Bash(fmt.Sprintf("mv %s/GeoIP2-City_*/GeoIP2-City.mmdb %s", dir, dir))
	if err != nil {
		return err
	}

	core.Log("validating isp database")

	isp_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-ISP.mmdb", dir))
	if err != nil {
		return fmt.Errorf("failed to load isp database: %v", err)
	}

	core.Log("validating city database")

	city_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-City.mmdb", dir))
	if err != nil {
		return fmt.Errorf("failed to load city database: %v", err)
	}

	core.Log("copying database files to app dir")

	err = Bash(fmt.Sprintf("cp %s/GeoIP2-*.mmdb .", dir))
	if err != nil {
		return fmt.Errorf("failed to copy databases: %v", err)
	}

	_ = isp_db
	_ = city_db

	return nil
}

func DownloadDatabases_CloudStorage(bucketName string) error {

	/*
		dir, err := ioutil.TempDir("/tmp", "database-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)

		core.Log("downloading isp database")

		// ..

		core.Log("downloading city database")

		// ...

		core.Log("validating isp database")

		isp_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-ISP.mmdb", dir))
		if err != nil {
			return fmt.Errorf("failed to load isp database: %v", err)
		}

		core.Log("validating city database")

		city_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-City.mmdb", dir))
		if err != nil {
			return fmt.Errorf("failed to load city database: %v", err)
		}

		core.Log("copying database files to app dir")

		err = Bash(fmt.Sprintf("cp %s/GeoIP2-*.mmdb .", dir))
		if err != nil {
			return fmt.Errorf("failed to copy databases: %v", err)
		}

		_ = isp_db
		_ = city_db
	*/

	return nil
}

func LoadDatabases() (*maxminddb.Reader, *maxminddb.Reader, error) {

	isp_db, err := maxminddb.Open("GeoIP2-ISP.mmdb")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load isp database: %v", err)
	}

	core.Log("loaded ip2location isp file")

	city_db, err := maxminddb.Open("GeoIP2-City.mmdb")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load city database: %v", err)
	}

	core.Log("loaded ip2location city file")

	return isp_db, city_db, nil
}

func GetLocation(city_db *maxminddb.Reader, ip net.IP) (float32, float32) {
	var city City
	err := city_db.Lookup(ip, &city)
	if err == nil {
		return float32(city.Location.Latitude), float32(city.Location.Longitude)
	} else {
		return 0, 0
	}
}

func GetISP(isp_db *maxminddb.Reader, ip net.IP) string {
	var isp ISP
	err := isp_db.Lookup(ip, &isp)
	if err == nil {
		return isp.ISP
	} else {
		return "Unknown"
	}
}
