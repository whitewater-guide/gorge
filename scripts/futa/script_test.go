package futa

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

func TestFuta_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptFuta{
		name:    "futa",
		dataURL: ts.URL + "/hoyweb.txt",
	}
	gauges, err := s.ListGauges()
	if assert.NoError(t, err) {
		assert.ElementsMatch(
			t,
			[]core.Gauge{
				{
					GaugeID:  core.GaugeID{Script: "futa", Code: "futa00"},
					Name:     "Futaleufu Hidroelectrica",
					URL:      "http://www.chfutaleufu.com.ar/default.asp",
					FlowUnit: "m3/s",
					Timezone: "America/Santiago",
				},
			},
			gauges,
		)
	}
}

func TestFuta_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptFuta{
		name:    "futa",
		dataURL: ts.URL + "/hoyweb.txt",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "futa",
				Code:   "futa00",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.January, 15, 3, 59, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(215),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "futa",
				Code:   "futa00",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.January, 15, 4, 59, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(228),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
