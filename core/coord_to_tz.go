package core

import (
	"os"
	"path"

	timezone "github.com/evanoberholster/timezoneLookup/v2"
	"github.com/ringsaturn/tzf"
)

var tz *timezone.Timezonecache

func load() (*timezone.Timezonecache, error) {
	if tz == nil {
		// This is for tests, because they run in tested package's directory
		tzDbDir := os.Getenv("TIMEZONE_DB_DIR")
		f, err := os.Open(path.Join(tzDbDir, "timezone.data"))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		tz = &timezone.Timezonecache{}
		if err = tz.Load(f); err != nil {
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
	result, err := db.Search(lat, lon)
	if err != nil {
		return "", err
	}
	name := result.Name
	if name == "" {
		// fallback, alternative lib, can discover America/St_John
		f, err := tzf.NewDefaultFinder()
		if err != nil {
			return "", err
		}
		name = f.GetTimezoneName(lon, lat)
	}
	return name, nil
}

func CloseTimezoneDb() {
	if tz != nil {
		(*tz).Close()
		tz = nil
	}
}
