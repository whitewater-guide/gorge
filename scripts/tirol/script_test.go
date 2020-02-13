package tirol

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestTirol_ListGauges(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/csw")
		fmt.Fprint(w, `Stationsname;Stationsnummer;Gew‰sser;Parameter;Zeitstempel in ISO8601;Wert;Einheit;Seehˆhe;Rechtswert;Hochwert;EPSG-Code
Steeg;201012;Lech;W.RADAR;2019-12-27T14:15:00+0100;226.0;cm;1109.3;147008.0;233674.0;EPSG:31257
Wattens;201657;Wattenbach;W;2019-12-28T13:45:00+0100;30.4;cm;550.86;245119.15;240454.16;EPSG:31257
Wattens;201657;Wattenbach;W;2019-12-28T14:00:00+0100;-777;cm;550.86;245119.15;240454.16;EPSG:31257
`)
	}))
	defer ts.Close()
	s := scriptTirol{name: "tirol", csvURL: ts.URL}
	gauges, err := s.ListGauges()
	if assert.NoError(t, err) {
		assert.ElementsMatch(
			t,
			[]core.Gauge{
				{
					GaugeID:   core.GaugeID{Script: "tirol", Code: "201012"},
					Name:      "Lech / Steeg",
					URL:       "https://apps.tirol.gv.at/hydro/#/Wasserstand/?station=201012",
					LevelUnit: "cm",
					Location:  &core.Location{Latitude: 47.24192, Longitude: 10.2935, Altitude: 1109},
				},
				{
					GaugeID:   core.GaugeID{Script: "tirol", Code: "201657"},
					Name:      "Wattenbach / Wattens",
					URL:       "https://apps.tirol.gv.at/hydro/#/Wasserstand/?station=201657",
					LevelUnit: "cm",
					Location:  &core.Location{Latitude: 47.29604, Longitude: 11.59062, Altitude: 550},
				},
			},
			gauges,
		)
	}
}

func TestTirol_Harvest(t *testing.T) {
	a := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/csw")
		fmt.Fprint(w, `Stationsname;Stationsnummer;Gew‰sser;Parameter;Zeitstempel in ISO8601;Wert;Einheit;Seehˆhe;Rechtswert;Hochwert;EPSG-Code
Steeg;201012;Lech;W.RADAR;2019-12-27T14:15:00+0100;226.0;cm;1109.3;147008.0;233674.0;EPSG:31257
Wattens;201657;Wattenbach;W;2019-12-28T13:45:00+0100;30.4;cm;550.86;245119.15;240454.16;EPSG:31257
Wattens;201657;Wattenbach;W;2019-12-28T14:00:00+0100;-777;cm;550.86;245119.15;240454.16;EPSG:31257
`)
	}))
	defer ts.Close()
	loc, _ := time.LoadLocation("Europe/Vienna")
	s := scriptTirol{name: "tirol", csvURL: ts.URL}
	res, err := core.HarvestSlice(&s, core.StringSet{"201657": {}, "201658": {}}, 0)
	if a.NoError(err) && a.Len(res, 1) {
		a.Equal(res[0].GaugeID.Script, "tirol")
		a.Equal(res[0].GaugeID.Code, "201657")
		a.Equal(res[0].Level, nulltype.NullFloat64Of(30.4))
		a.Equal(res[0].Flow, nulltype.NullFloat64{})
		a.True(time.Date(2019, time.December, 28, 13, 45, 0, 0, loc).Equal(res[0].Timestamp.Time))
	}
}
