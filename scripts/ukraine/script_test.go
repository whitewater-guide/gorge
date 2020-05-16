package ukraine

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

func TestUkraine_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUkraine{
		name: "ukraine",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauge{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "42136",
		},
		LevelUnit: "cm",
		Location: &core.Location{
			Latitude:  48.36876,
			Longitude: 24.55166,
		},
		Name: "Прут Татарів",
		URL:  "https://meteo.gov.ua/ua/33345/hydrostorm",
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 192)
		assert.Contains(t, actual, expected)
	}
}

func TestUkraine_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUkraine{
		name: "ukraine",
		url:  ts.URL,
	}
	now := time.Now().UTC().Truncate(time.Hour)
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "42136",
		},
		Timestamp: core.HTime{
			Time: now,
		},
		Level: nulltype.NullFloat64Of(136),
	}
	if assert.NoError(t, err) {
		assert.Contains(t, actual, expected)
	}
}
