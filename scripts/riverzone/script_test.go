package riverzone

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
	return testutils.SetupFileServer(nil, &testutils.HeaderAuthorizer{Key: "X-Key"})
}

func TestRiverzone_Auth(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptRiverzone{
		name:                "riverzone",
		stationsEndpointURL: ts.URL + "/data.json",
		options:             optionsRiverzone{Key: "__bad__"},
	}
	_, err := s.ListGauges()
	assert.Error(t, err)
}

func TestRiverzone_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptRiverzone{
		name:                "riverzone",
		stationsEndpointURL: ts.URL + "/data.json",
		options:             optionsRiverzone{Key: testutils.TestAuthKey},
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "d5a4cf14-e62c-4fcd-b963-3cd6a1c075de",
			},
			FlowUnit:  "m3s",
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  47.36814,
				Longitude: 8.06223,
			},
			Name: "CH - Aargau - Suhre - Suhr",
			URL:  "https://www.ag.ch/app/hydrometrie/station/?id=11560",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "19ac0462-2c0c-454b-ac7b-5a053db3efab",
			},
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  42.13721,
				Longitude: 13.75299,
			},
			Name: "IT - Abruzzo - Aterno - Molina (AQ)",
			URL:  "http://www.himet.it/cgi-bin/meteo/gmaps/new_idrope.cgi",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestRiverzone_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptRiverzone{
		name:                "riverzone",
		stationsEndpointURL: ts.URL + "/data.json",
		options:             optionsRiverzone{Key: testutils.TestAuthKey},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "d5a4cf14-e62c-4fcd-b963-3cd6a1c075de",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 19, 5, 20, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(26.3),
			Flow:  nulltype.NullFloat64Of(1.793),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "d5a4cf14-e62c-4fcd-b963-3cd6a1c075de",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 19, 5, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(26.4),
			Flow:  nulltype.NullFloat64Of(1.805),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "19ac0462-2c0c-454b-ac7b-5a053db3efab",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 19, 5, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(30.0),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "19ac0462-2c0c-454b-ac7b-5a053db3efab",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 19, 5, 45, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(30.0),
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
