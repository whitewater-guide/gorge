package georgia

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open("./test_data/page.html")
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestGeorgia_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGeorgia{
		name: "georgia",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauge{
		GaugeID: core.GaugeID{
			Script: "georgia",
			Code:   "3c598e973566b29359a3821cf3dceceb",
		},
		LevelUnit: "cm",
		Name:      "Acharistskali - keda",
		URL:       ts.URL,
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 20)
		assert.Contains(t, actual, expected)
	}
}

func TestGeorgia_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGeorgia{
		name: "georgia",
		url:  ts.URL,
	}
	now := time.Now().UTC().Truncate(time.Hour)
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "georgia",
			Code:   "3c598e973566b29359a3821cf3dceceb",
		},
		Timestamp: core.HTime{
			Time: now,
		},
		Level: nulltype.NullFloat64Of(64),
	}
	expected2 := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "georgia",
			Code:   "89f8e0acaa48a39321c3d7f69ac5b5ad",
		},
		Timestamp: core.HTime{
			Time: now,
		},
		Level: nulltype.NullFloat64Of(156),
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 20)
		assert.Contains(t, actual, expected)
		assert.Contains(t, actual, expected2)
	}
}
