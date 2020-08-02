package kuban

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

func TestKuban_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptKuban{
		name: "kuban",
		url:  ts.URL + "/data.html",
	}
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "kuban",
				Code:   "83137",
			},
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  43.47,
				Longitude: 42.24,
			},
			Name: "р.Кубань - с. им.Коста Хетагурова",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "kuban",
				Code:   "83395",
			},
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  44.38,
				Longitude: 39.06,
			},
			Name: "р.Псекупс - г. Горячий ключ",
		},
	}
	actual, err := s.ListGauges()
	if assert.NoError(t, err) && assert.Len(t, actual, 28) {
		assert.Equal(t, expected[0], actual[0])
		assert.Equal(t, expected[1], actual[27])
	}
}

func TestKuban_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptKuban{
		name: "kuban",
		url:  ts.URL + "/data.html",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "kuban",
				Code:   "83137",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.August, 2, 5, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(451),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "kuban",
				Code:   "83395",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.August, 2, 5, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(-55),
		},
	}
	if assert.NoError(t, err) && assert.Len(t, actual, 28) {
		assert.Equal(t, expected[0], actual[0])
		assert.Equal(t, expected[1], actual[27])
	}
}
