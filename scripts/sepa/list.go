package sepa

import (
	"fmt"
	"math"
	"strconv"

	"github.com/whitewater-guide/gorge/core"
)

type csvRaw struct {
	sepaHydrologyOffice   string
	stationName           string
	locationCode          string
	nationalGridReference string
	catchmentName         string
	riverName             string
	gaugeDatum            string
	catchmentArea         string
	startDate             string
	endDate               string
	systemID              string
	lowestValue           string
	low                   string
	maxValue              string
	high                  string
	maxDisplay            string
	mean                  string
	units                 string
	webMessage            string
	nrfaLink              string
}

func gaugeFromRow(row []string) csvRaw {
	return csvRaw{
		sepaHydrologyOffice:   row[0],
		stationName:           row[1],
		locationCode:          row[2],
		nationalGridReference: row[3],
		catchmentName:         row[4],
		riverName:             row[5],
		gaugeDatum:            row[6],
		catchmentArea:         row[7],
		startDate:             row[8],
		endDate:               row[9],
		systemID:              row[10],
		lowestValue:           row[11],
		low:                   row[12],
		maxValue:              row[13],
		high:                  row[14],
		maxDisplay:            row[15],
		mean:                  row[16],
		units:                 row[17],
		webMessage:            row[18],
		nrfaLink:              row[19],
	}
}

func (s *scriptSepa) getGauge(raw csvRaw) (result core.Gauge, err error) {
	ref, err := parseGridRef(raw.nationalGridReference)
	if err != nil {
		return result, core.WrapErr(err, "failed to parse location").With("code", raw.locationCode)
	}

	x, y, err := core.ToEPSG4326(float64(ref.easting), float64(ref.northing), "EPSG:27700")
	if err != nil {
		return
	}
	alt, err := strconv.ParseFloat(raw.gaugeDatum, 64)
	if err != nil {
		return
	}
	result = core.Gauge{
		GaugeID: core.GaugeID{
			Code:   raw.locationCode,
			Script: s.name,
		},
		Name:      fmt.Sprintf("%s - %s", raw.riverName, raw.stationName),
		URL:       "http://apps.sepa.org.uk/waterlevels/default.aspx?sd=t&lc=" + raw.locationCode,
		LevelUnit: raw.units, // all the data is supposed to be in m
		FlowUnit:  "",

		Location: &core.Location{
			Longitude: x,
			Latitude:  y,
			Altitude:  math.Trunc(alt),
		},
		Timezone: "Europe/London",
	}
	return
}
