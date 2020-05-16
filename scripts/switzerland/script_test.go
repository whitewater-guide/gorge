package switzerland

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/testutils"

	"github.com/stretchr/testify/assert"
)

type BasicAuthorizer struct{}

// Authorize implements Authorizer interface
func (a *BasicAuthorizer) Authorize(req *http.Request) bool {
	if req.URL.Path != "/gauges.xml" {
		return true
	}
	user, pass, ok := req.BasicAuth()
	return ok && user == "user" && pass == "password"
}

func setupTestServer() *httptest.Server {
	return testutils.SetupFileServer(nil, &BasicAuthorizer{})
}

func TestSwitzerland_BasicAuth(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptSwitzerland{
		name:             "switzerland",
		xmlURL:           ts.URL + "/gauges.xml",
		gaugePageURLBase: ts.URL + "/",
		options:          optionsSwitzerland{Username: "foo", Password: "bar"},
	}
	_, err := s.ListGauges()
	assert.Error(t, err)
}

func TestSwitzerland_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptSwitzerland{
		name:             "switzerland",
		xmlURL:           ts.URL + "/gauges.xml",
		gaugePageURLBase: ts.URL + "/",
		options:          optionsSwitzerland{Username: "user", Password: "password"},
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID:   core.GaugeID{Script: "switzerland", Code: "2007"},
			LevelUnit: "m ü. M.",
			Name:      "Lac de Joux - Le Pont (lake)",
			URL:       "https://www.hydrodaten.admin.ch/en/2007.html",
			Location:  &core.Location{Latitude: 46.66532, Longitude: 6.32402, Altitude: 1004},
		},
		core.Gauge{
			GaugeID:   core.GaugeID{Script: "switzerland", Code: "2009"},
			LevelUnit: "m ü. M.",
			FlowUnit:  "m3/s",
			Name:      "Rhône - Porte du Scex",
			URL:       "https://www.hydrodaten.admin.ch/en/2009.html",
			Location:  &core.Location{Latitude: 46.34956, Longitude: 6.88861, Altitude: 377},
		},
		core.Gauge{
			GaugeID:  core.GaugeID{Script: "switzerland", Code: "2011"},
			FlowUnit: "m3/s",
			Name:     "Rhône - Sion",
			URL:      "https://www.hydrodaten.admin.ch/en/2011.html",
			Location: &core.Location{Latitude: 46.21908, Longitude: 7.3579, Altitude: 484},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
func TestSwitzerland_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptSwitzerland{
		name:             "switzerland",
		xmlURL:           ts.URL + "/gauges.xml",
		gaugePageURLBase: ts.URL + "/",
		options:          optionsSwitzerland{Username: "user", Password: "password"},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"2007": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID:   core.GaugeID{Script: "switzerland", Code: "2007"},
			Timestamp: core.HTime{Time: time.Date(2020, time.January, 16, 23, 0, 0, 0, time.UTC)},
			Level:     nulltype.NullFloat64Of(1003.64),
		},
		&core.Measurement{
			GaugeID:   core.GaugeID{Script: "switzerland", Code: "2009"},
			Timestamp: core.HTime{Time: time.Date(2020, time.January, 17, 7, 40, 0, 0, time.UTC)},
			Level:     nulltype.NullFloat64Of(374.84),
			Flow:      nulltype.NullFloat64Of(118),
		},
		&core.Measurement{
			GaugeID:   core.GaugeID{Script: "switzerland", Code: "2011"},
			Timestamp: core.HTime{Time: time.Date(2020, time.January, 17, 7, 50, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(46),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
