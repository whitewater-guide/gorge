package norway

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

func TestNorway_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNorway{
		name:          "norway",
		urlBase:       ts.URL,
		jsonURLFormat: ts.URL + "/json/%s/%d/data.json",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "2.32",
			},
			FlowUnit: "m3/s",
			Name:     "Atnasjø",
			Location: &core.Location{
				Latitude:  61.85194,
				Longitude: 10.22212,
				Altitude:  701,
			},
			URL:      ts.URL + "/0002.00032.000/index.html",
			Timezone: "Europe/Oslo",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "16.128",
			},
			FlowUnit: "m3/s",
			Name:     "Austbygdåi",
			Location: &core.Location{
				Latitude:  59.99525,
				Longitude: 8.82766,
				Altitude:  230,
			},
			URL:      ts.URL + "/0016.00128.000/index.html",
			Timezone: "Europe/Oslo",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNorway_GetJSONUrl(t *testing.T) {
	tests := []struct {
		code     string
		since    int64
		version  int
		expected string
	}{
		{
			code:     "6.38",
			version:  1,
			expected: "http://h-web01.nve.no/chartserver/ShowData.aspx?req=getchart&ver=1.0&vfmt=json&time=-1;0&lang=no&chd=ds=htsr,da=29,id=6.38.0.1001.1,rt=0&nocache=81",
		},
		{
			code:     "62.10",
			version:  2,
			expected: "http://h-web01.nve.no/chartserver/ShowData.aspx?req=getchart&ver=1.0&vfmt=json&time=-1;0&lang=no&chd=ds=htsr,da=29,id=62.10.0.1001.2,rt=0&nocache=81",
		},
		{
			code:     "62.10",
			since:    1579804200,
			version:  2,
			expected: "http://h-web01.nve.no/chartserver/ShowData.aspx?req=getchart&ver=1.0&vfmt=json&time=20200123T1830;0&lang=no&chd=ds=htsr,da=29,id=62.10.0.1001.2,rt=0&nocache=81",
		},
	}
	for _, tst := range tests {
		s := scriptNorway{
			name:       "norway",
			randomSeed: 1,
			options: optionsNorway{
				Version: tst.version,
			},
		}
		assert.Equal(t, tst.expected, s.getJSONUrl(tst.code, tst.since))
	}
}

func TestNorway_Harvest_JSON(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNorway{
		name:          "norway",
		urlBase:       ts.URL,
		jsonURLFormat: ts.URL + "/json/%s/%d/data.json",
		options: optionsNorway{
			Version: 1,
		},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"6.38": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "6.38",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 22, 19, 50, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(13.69747),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "6.38",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 22, 19, 55, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(13.5778),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestNorway_Harvest_CSV(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNorway{
		name:          "norway",
		urlBase:       ts.URL,
		jsonURLFormat: ts.URL + "/json/%s/%d/data.json",
		options: optionsNorway{
			Version: 1,
			CSV:     &csvOptions{},
		},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"213.4": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "213.4",
			},
			Timestamp: core.HTime{
				Time: time.Date(2019, time.November, 25, 11, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.088),
			Flow:  nulltype.NullFloat64Of(1.206),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "213.4",
			},
			Timestamp: core.HTime{
				Time: time.Date(2019, time.November, 25, 13, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.094),
			Flow:  nulltype.NullFloat64Of(1.260),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestNorway_Harvest_HTML(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNorway{
		name:          "norway",
		urlBase:       ts.URL,
		jsonURLFormat: ts.URL + "/json/%s/%d/data.json",
		options: optionsNorway{
			Version: 1,
			HTML:    true,
		},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"2.32": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "2.32",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 23, 18, 0, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(5.014),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
