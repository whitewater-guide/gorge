package tirol

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

// https://epsg.io/31257.proj4
const epsg31257 = "+proj=tmerc +lat_0=0 +lon_0=10.33333333333333 +k=1 +x_0=150000 +y_0=-5000000 +ellps=bessel +towgs84=577.326,90.129,463.919,5.137,1.474,5.297,2.4232 +units=m +no_defs"

type csvRaw struct {
	name      string // Stationsname;
	code      string // Stationsnummer;
	river     string // Gewässer; - means body of water, not necessary river
	parameter string // Parameter;
	timestamp string // Zeitstempel in ISO8601;
	value     string // Wert;
	unit      string // Einheit;
	elevation string // Seehˆhe;
	easting   string // Rechtswert;
	northing  string // Hochwert;
	epsg      string // EPSG-Code
}

func fromRow(row []string) csvRaw {
	return csvRaw{
		name:      row[0],
		code:      row[1],
		river:     row[2],
		parameter: row[3],
		timestamp: row[4],
		value:     row[5],
		unit:      row[6],
		elevation: row[7],
		easting:   row[8],
		northing:  row[9],
		epsg:      row[10],
	}
}

func (s *scriptTirol) getGauge(raw csvRaw) (result core.Gauge, err error) {

	x, _ := strconv.ParseFloat(raw.easting, 64)
	y, _ := strconv.ParseFloat(raw.northing, 64)
	z, _ := strconv.ParseFloat(raw.elevation, 64)

	x, y, err = core.ToEPSG4326(x, y, epsg31257)

	if err != nil {
		return
	}

	result = core.Gauge{
		GaugeID: core.GaugeID{
			Code:   raw.code,
			Script: s.name,
		},
		Name:      fmt.Sprintf("%s / %s", raw.river, raw.name),
		URL:       "https://apps.tirol.gv.at/hydro/#/Wasserstand/?station=" + raw.code,
		LevelUnit: raw.unit, // all the data is supposed to be in cm
		FlowUnit:  "",

		Location: &core.Location{
			Longitude: x,
			Latitude:  y,
			Altitude:  math.Trunc(z),
		},
	}
	return
}

func (s *scriptTirol) getMeasurement(raw csvRaw) (core.Measurement, error) {
	t, err := time.Parse("2006-01-02T15:04:05-0700", raw.timestamp)
	if err != nil {
		return core.Measurement{}, err
	}
	level, err := strconv.ParseFloat(raw.value, 64)
	if err != nil {
		return core.Measurement{}, err
	}
	return core.Measurement{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   raw.code,
		},
		Level:     core.NPrecision(level, 2),
		Flow:      nulltype.NullFloat64{},
		Timestamp: core.HTime{Time: t},
	}, nil
}
