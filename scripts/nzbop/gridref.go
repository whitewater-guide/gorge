package nzbop

import (
	"errors"
	"math"
	"regexp"
	"strconv"

	"github.com/whitewater-guide/gorge/core"
)

var gridRegExp = regexp.MustCompile(`^\s*([A-Z]{1,2})(\d\d):?\s*(?:(\d\d)\s*(\d\d)|(\d\d\d)\s*(\d\d\d)|(\d\d\d\d)\s*(\d\d\d\d))\s*$`)
var errInvalidRef = errors.New("is not valid NZMS260 grid refrence")

func convertNZMS260(val string) (*core.Location, error) {
	matches := gridRegExp.FindAllStringSubmatch(val, -1)
	if len(matches) != 1 || len(matches[0]) != 9 {
		return nil, errInvalidRef
	}
	parts := matches[0]

	numberIndex, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, errInvalidRef
	}
	numberIndex = numberIndex - 1

	ind, mult := 0, 0.0
	if parts[3] != "" {
		ind, mult = 3, 1000.0
	} else if parts[5] != "" {
		ind, mult = 5, 100.0
	} else if parts[7] != "" {
		ind, mult = 7, 10.0
	} else {
		return nil, errInvalidRef
	}
	mape, err := strconv.ParseFloat(parts[ind], 64)
	if err != nil {
		return nil, errInvalidRef
	}
	mapn, err := strconv.ParseFloat(parts[ind+1], 64)
	if err != nil {
		return nil, errInvalidRef
	}
	mape = mape * mult
	mapn = mapn * mult

	letterIndex := int64([]rune(parts[1])[0] - []rune("A")[0])
	if letterIndex < 0 {
		return nil, errInvalidRef
	}
	letterIndex, numberIndex = numberIndex, letterIndex

	mapCentreE := 1970000.0 + (float64(numberIndex)+0.5)*40000.0
	mapCentreN := 6790000.0 - (float64(letterIndex)+0.5)*30000.0

	mape = mape + math.Round((mapCentreE-mape)/100000.0)*100000.0
	mapn = mapn + math.Round((mapCentreN-mapn)/100000.0)*100000.0

	lon, lat, err := core.ToEPSG4326(mape, mapn, "EPSG:27200")
	if err != nil {
		return nil, err
	}

	return &core.Location{
		Latitude:  lat,
		Longitude: lon,
	}, nil
}
