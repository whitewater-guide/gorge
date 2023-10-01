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

// only those state that have strictly one timezone
var stateTimeZones = map[string]string{
	"AL": "America/Chicago",
	"AZ": "America/Phoenix",
	"AR": "America/Chicago",
	"CA": "America/Los_Angeles",
	"CO": "America/Denver",
	"CT": "America/New_York",
	"DE": "America/New_York",
	"DC": "America/New_York",
	"GA": "America/New_York",
	"HI": "Pacific/Honolulu",
	"IL": "America/Chicago",
	"IA": "America/Chicago",
	"ME": "America/New_York",
	"MD": "America/New_York",
	"MA": "America/New_York",
	"MN": "America/Chicago",
	"MS": "America/Chicago",
	"MO": "America/Chicago",
	"MT": "America/Denver",
	"NV": "America/Los_Angeles",
	"NH": "America/New_York",
	"NJ": "America/New_York",
	"NM": "America/Denver",
	"NY": "America/New_York",
	"NC": "America/New_York",
	"OH": "America/New_York",
	"OK": "America/Chicago",
	"PA": "America/New_York",
	"RI": "America/New_York",
	"SC": "America/New_York",
	"UT": "America/Denver",
	"VT": "America/New_York",
	"VA": "America/New_York",
	"WA": "America/Los_Angeles",
	"WV": "America/New_York",
	"WI": "America/Chicago",
	"WY": "America/Denver",
}

func USCoordinateToTimezone(state string, lat float64, lon float64) (string, error) {
	if name, ok := stateTimeZones[state]; ok {
		return name, nil
	}
	return CoordinateToTimezone(lat, lon)
}

func CloseTimezoneDb() {
	if tz != nil {
		(*tz).Close()
		tz = nil
	}
}
