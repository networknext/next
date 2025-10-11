package ip2location

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/networknext/next/modules/core"

	"github.com/oschwald/maxminddb-golang"
)

type City struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
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
	}

	cmd.Wait()

	return nil
}

func DownloadDatabases_MaxMind(licenseKey string) error {

	dir, err := os.MkdirTemp("/tmp", "database-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	core.Debug("downloading isp database")

	err = Bash(fmt.Sprintf("curl -J -L -u %s 'https://download.maxmind.com/geoip/databases/GeoIP2-ISP/download?suffix=tar.gz' --output %s/GeoIP2-ISP.tar.gz", licenseKey, dir))
	if err != nil {
		return err
	}

	core.Debug("downloading city database")

	err = Bash(fmt.Sprintf("curl -J -L -u %s 'https://download.maxmind.com/geoip/databases/GeoIP2-City/download?suffix=tar.gz' --output %s/GeoIP2-City.tar.gz", licenseKey, dir))
	if err != nil {
		return err
	}

	core.Debug("decompressing databases")

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

	core.Debug("validating isp database")

	isp_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-ISP.mmdb", dir))
	if err != nil {
		return fmt.Errorf("failed to load isp database: %v", err)
	}

	core.Debug("validating city database")

	city_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-City.mmdb", dir))
	if err != nil {
		return fmt.Errorf("failed to load city database: %v", err)
	}

	core.Debug("copying database files to app dir")

	err = Bash(fmt.Sprintf("cp %s/GeoIP2-*.mmdb .", dir))
	if err != nil {
		return fmt.Errorf("failed to copy databases: %v", err)
	}

	_ = isp_db
	_ = city_db

	return nil
}

func DownloadDatabases_CloudStorage(bucketName string) (error, *maxminddb.Reader, *maxminddb.Reader) {

	dir, err := os.MkdirTemp("/tmp", "database-")
	if err != nil {
		return err, nil, nil
	}

	core.Debug("downloading isp database")

	err = Bash(fmt.Sprintf("gsutil cp gs://%s/GeoIP2-ISP.mmdb %s", bucketName, dir))
	if err != nil {
		return fmt.Errorf("failed to download isp database: %v", err), nil, nil
	}

	err = Bash(fmt.Sprintf("gsutil cp gs://%s/GeoIP2-City.mmdb %s", bucketName, dir))
	if err != nil {
		return fmt.Errorf("failed to download isp database: %v", err), nil, nil
	}

	core.Debug("validating isp database")

	isp_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-ISP.mmdb", dir))
	if err != nil {
		return fmt.Errorf("failed to load isp database: %v", err), nil, nil
	}

	core.Debug("validating city database")

	city_db, err := maxminddb.Open(fmt.Sprintf("%s/GeoIP2-City.mmdb", dir))
	if err != nil {
		return fmt.Errorf("failed to load city database: %v", err), nil, nil
	}

	return nil, isp_db, city_db
}

func RemoveOldDatabaseFiles() {
	// IMPORTANT: WE need to cleanup ip2location database files lazily, because they are accessed via memory mapped files
	// If we delete them while in use, we get undefined behavior. Since we update ip2location dbs no more than once every 1hr,
	// it is safe to delete ip2location database files older than 2 hours
	core.Debug("looking for old ip2location files...")
	currentTime := time.Now()
	matches, _ := filepath.Glob("/tmp/database-*")
	var dirs []string
	for _, match := range matches {
		core.Debug("found '%s'", match)
		f, _ := os.Stat(match)
		if f.IsDir() && currentTime.Sub(f.ModTime()) > 2*time.Hour {
			dirs = append(dirs, match)
		}
	}
	for i := range dirs {
		if dirs[i][0] == '/' && dirs[i][1] == 't' && dirs[i][2] == 'm' && dirs[i][3] == 'p' && dirs[i][4] == '/' {
			core.Debug("removed old ip2location file '%s'", dirs[i])
			os.RemoveAll(dirs[i])
		}
	}
}

func LoadDatabases() (*maxminddb.Reader, *maxminddb.Reader, error) {

	isp_db, err := maxminddb.Open("GeoIP2-ISP.mmdb")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load isp database: %v", err)
	}

	core.Debug("loaded ip2location isp file")

	city_db, err := maxminddb.Open("GeoIP2-City.mmdb")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load city database: %v", err)
	}

	core.Debug("loaded ip2location city file")

	return isp_db, city_db, nil
}

func GetLocation(city_db *maxminddb.Reader, ip net.IP) (float32, float32) {
	var city City
	if city_db != nil && city_db.Lookup(ip, &city) == nil {
		return float32(city.Location.Latitude), float32(city.Location.Longitude)
	} else {
		return 0, 0
	}
}

func GetISPAndCountry(isp_db *maxminddb.Reader, city_db *maxminddb.Reader, ip net.IP) (string, string) {
	var isp ISP
	var city City
	if isp_db != nil && city_db != nil && isp_db.Lookup(ip, &isp) == nil && city_db.Lookup(ip, &city) == nil {
		return isp.ISP, city.Country.ISOCode
	} else {
		return "Unknown", "Unknown"
	}
}
