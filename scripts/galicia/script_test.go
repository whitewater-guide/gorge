package galicia

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/testutils"
)

func setupTestServer() *httptest.Server {
	return testutils.SetupFileServer(nil, nil)
}

func TestGalicia_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGalicia{
		name: "galicia",
		url:  ts.URL + "/galicia.json",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30431",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  43.47774,
				Longitude: -7.33423,
			},
			Name:     "Masma",
			URL:      "http://www2.meteogalicia.gal/servizos/AugasdeGalicia/estacionsinfo.asp?Nest=30431",
			Timezone: "Europe/Madrid",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30433",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  43.55815,
				Longitude: -7.37617,
			},
			Name:     "Ouro",
			URL:      "http://www2.meteogalicia.gal/servizos/AugasdeGalicia/estacionsinfo.asp?Nest=30433",
			Timezone: "Europe/Madrid",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestGalicia_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGalicia{
		name: "galicia",
		url:  ts.URL + "/galicia.json",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30431",
			},
			Timestamp: core.HTime{
				Time: time.Date(2018, time.February, 19, 19, 40, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.07),
			Flow:  nulltype.NullFloat64Of(8.689578),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30433",
			},
			Timestamp: core.HTime{
				Time: time.Date(2018, time.February, 19, 19, 40, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.63),
			Flow:  nulltype.NullFloat64Of(7.1654706),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
