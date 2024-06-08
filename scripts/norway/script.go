package norway

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNorway struct {
	ApiKey       string `desc:"API Key, default to env variable NVE_API_KEY" json:"version"`
	BatchSize    int    `desc:"Batch size for requesting multiple stations at once" json:"batchSize"`
	IgnoreLegacy bool   `desc:"Do not support station ids from previous version of this script" json:"ignoreLegacy"`
}

// GetBatchSize Implements core.BatchableOptions interface
func (o optionsNorway) GetBatchSize() int {
	return o.BatchSize
}

type scriptNorway struct {
	name         string
	urlBase      string
	apiKey       string
	ignoreLegacy bool
	core.LoggingScript
}

func (s *scriptNorway) ListGauges() (core.Gauges, error) {
	var resp statiosResponse
	err := core.Client.GetAsJSON(
		s.urlBase+"/Stations?Active=1",
		&resp,
		&core.RequestOptions{Headers: map[string]string{
			"X-API-Key": s.apiKey,
		}},
	)
	if err != nil {
		return nil, err
	}
	result := []core.Gauge{}
	for _, station := range resp.Data {
		gauge, ok := core.Gauge{}, false
		for _, series := range station.SeriesList {
			for _, resolution := range series.ResolutionList {
				if resolution.Method == "Instantaneous" {
					if series.Parameter == 1000 {
						ok = true
						gauge.LevelUnit = series.Unit
					} else if series.Parameter == 1001 {
						ok = true
						gauge.FlowUnit = series.Unit
					}
				}
			}
		}
		if ok {
			gauge.GaugeID = core.GaugeID{
				Script: s.name,
				Code:   getOurStationId(!s.ignoreLegacy, station.StationID),
			}
			gauge.Name = fmt.Sprintf("%s - %s", station.RiverName, station.StationName)
			gauge.URL = fmt.Sprintf("https://sildre.nve.no/station/%s", station.StationID)
			gauge.Location = &core.Location{
				Latitude:  station.Latitude,
				Longitude: station.Longitude,
				Altitude:  float64(station.Masl),
			}
			gauge.Timezone = "Europe/Oslo"
			result = append(result, gauge)
		}
	}
	return result, nil
}

func (s *scriptNorway) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	params := url.Values{}
	params.Add("StationId", strings.Join(getTheirStationIds(!s.ignoreLegacy, codes.Slice()), ","))
	params.Add("Parameter", "1000,1001")
	params.Add("ResolutionTime", "0")
	params.Add("ReferenceTime", "PT1H/")

	resp := observationsResp{}
	err := core.Client.GetAsJSON(
		fmt.Sprintf("%s/Observations?%s", s.urlBase, params.Encode()),
		&resp,
		&core.RequestOptions{Headers: map[string]string{
			"X-API-Key": s.apiKey,
		}},
	)
	if err != nil {
		errs <- err
		return
	}
	measurements := make(map[string]*core.Measurement)

	for _, obsList := range resp.Data {
		for _, o := range obsList.Observations {
			ts, err := time.ParseInLocation("2006-01-02T15:04:05Z", o.Time, time.UTC)
			if err != nil {
				continue
			}
			key := fmt.Sprintf("%s-%s", obsList.StationID, o.Time)
			m, ok := measurements[key]
			if !ok {
				m = &core.Measurement{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   getOurStationId(!s.ignoreLegacy, obsList.StationID),
					},
					Timestamp: core.HTime{Time: ts},
				}
			}
			if obsList.Parameter == 1000 {
				m.Level = o.Value
			} else if obsList.Parameter == 1001 {
				m.Flow = o.Value
			} else {
				continue
			}
			measurements[key] = m
		}
	}

	for _, m := range measurements {
		recv <- m
	}
}
