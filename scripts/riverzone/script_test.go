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
	return testutils.SetupFileServer(map[string]string{
		"/readings": "readings.json",
		"/":         "stations.json",
	}, &testutils.HeaderAuthorizer{Key: "X-Key"})
}

func TestRiverzone_Auth(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptRiverzone{
		name:                "riverzone",
		stationsEndpointURL: ts.URL,
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
		stationsEndpointURL: ts.URL,
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
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "0d1135e2-d33d-544f-8227-0ee1a1f0e411",
			},
			LevelUnit: "cm",
			FlowUnit:  "m3s",
			Location: &core.Location{
				Latitude:  62.57209,
				Longitude: 9.15903,
			},
			Name: "NO - Driva - Grensehølen",
			URL:  "http://www2.nve.no/h/hd/plotreal/Q/0109.00020.000/index.html",
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
		stationsEndpointURL: ts.URL,
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
				Time: time.Date(2021, time.May, 16, 9, 40, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(34.2),
			Flow:  nulltype.NullFloat64Of(2.89),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "d5a4cf14-e62c-4fcd-b963-3cd6a1c075de",
			},
			Timestamp: core.HTime{
				Time: time.Date(2021, time.May, 16, 9, 50, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(34.2),
			Flow:  nulltype.NullFloat64Of(2.89),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "19ac0462-2c0c-454b-ac7b-5a053db3efab",
			},
			Timestamp: core.HTime{
				Time: time.Date(2021, time.May, 16, 9, 45, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(25.0),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "riverzone",
				Code:   "19ac0462-2c0c-454b-ac7b-5a053db3efab",
			},
			Timestamp: core.HTime{
				Time: time.Date(2021, time.May, 16, 10, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(26.0),
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
