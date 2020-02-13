package switzerland

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestSwitzerland_GetLocation(t *testing.T) {
	assert := assert.New(t)
	loc, err := getLocation(swissStation{Easting: 575500, Northing: 197790})
	if assert.NoError(err) {
		assert.Equal(7.11691, loc.Longitude)
		assert.Equal(46.93074, loc.Latitude)
	}
}

func TestSwitzerland_StationConversion(t *testing.T) {
	station := swissStation{
		XMLName:       xml.Name{Local: "station"},
		Code:          "2011",
		Name:          "Sion",
		WaterBodyName: "Rhône",
		WaterBodyType: "river",
		Easting:       593770,
		Northing:      118630,
		Parameters: []swissParameter{
			swissParameter{
				XMLName:  xml.Name{Local: "parameter"},
				Name:     "Temperatur",
				Unit:     "°C",
				Type:     3,
				Variant:  20,
				Datetime: swissTime{time.Date(2019, time.February, 9, 12, 0, 0, 0, time.UTC)},
				Value:    nulltype.NullFloat64Of(4.2),
			},
			swissParameter{
				XMLName:  xml.Name{Local: "parameter"},
				Name:     "Abfluss m3/s",
				Unit:     "m3/s",
				Type:     10,
				Variant:  11,
				Datetime: swissTime{time.Date(2019, time.February, 9, 21, 40, 0, 0, time.UTC)},
				Value:    nulltype.NullFloat64Of(27),
			},
			swissParameter{
				XMLName:  xml.Name{Local: "parameter"},
				Name:     "Pegel m ü. M.",
				Unit:     "m ü. M.",
				Type:     2,
				Variant:  1,
				Datetime: swissTime{time.Date(2019, time.February, 9, 21, 40, 0, 0, time.UTC)},
				Value:    nulltype.NullFloat64Of(482.60),
			},
		},
	}
	s := scriptSwitzerland{name: "switzerland"}
	t.Run("to gauge", func(t *testing.T) {
		expected := core.Gauge{
			GaugeID: core.GaugeID{
				Code:   "2011",
				Script: "switzerland",
			},
			Name:      "Rhône - Sion",
			URL:       "https://www.hydrodaten.admin.ch/en/2011.html",
			LevelUnit: "m ü. M.",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Longitude: 7.3579,
				Latitude:  46.21908,
			},
		}
		actual, err := s.stationToGauge(&station)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, *actual)
		}
	})
	t.Run("to measurement", func(t *testing.T) {
		actual := s.stationToMeasurement(&station)
		expected := core.Measurement{
			GaugeID: core.GaugeID{
				Code:   "2011",
				Script: "switzerland",
			},
			Level:     nulltype.NullFloat64Of(482.60),
			Flow:      nulltype.NullFloat64Of(27),
			Timestamp: core.HTime{Time: time.Date(2019, time.February, 9, 21, 40, 0, 0, time.UTC)},
		}
		assert.Equal(t, expected, *actual)
	})
}
