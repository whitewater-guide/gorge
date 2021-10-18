package tz

import (
	"os"

	timezone "github.com/evanoberholster/timezoneLookup"
)

var tz *timezone.TimezoneInterface

func load() (*timezone.TimezoneInterface, error) {
	if tz == nil {
		// This is for tests, because they run in tested package's directory
		tzDbDir := os.Getenv("TIMEZONE_DB_DIR")
		zones, err := timezone.LoadTimezones(timezone.Config{
			DatabaseType: "boltdb",             // memory or boltdb
			DatabaseName: tzDbDir + "timezone", // Name without suffix
			Snappy:       true,
			Encoding:     "msgpack", // json or msgpack
		})
		tz = &zones
		if err != nil {
			return nil, err
		}
	}
	return tz, nil
}

// CoordinateToTimezone returns IANA timezone for coordinate
func CoordinateToTimezone(lat float64, lon float64) (string, error) {
	db, err := load()
	if err != nil {
		return "", err
	}
	return (*db).Query(timezone.Coord{Lat: float32(lat), Lon: float32(lon)})
}

func CloseTimezoneDb() {
	if tz != nil {
		(*tz).Close()
		tz = nil
	}
}
