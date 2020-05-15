package nzniwa

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptNzniwa) fetchMeasurements(measurements chan<- *core.Measurement, errs chan<- error) {
	var data niwaList
	err := core.Client.GetAsJSON(s.flowURL, &data, nil)
	if err != nil {
		errs <- err
		return
	}
	for _, el := range data.Indicators[0].Indicators {
		measurements <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   el.LocationIdentifier,
			},
			Timestamp: core.HTime{
				Time: el.LastUpdated.Time,
			},
			Flow: el.Value,
		}
	}
}

func (s *scriptNzniwa) fetchGauges(gauges chan<- *core.Gauge) error {
	var data niwaList
	err := core.Client.GetAsJSON(s.flowURL, &data, nil)
	if err != nil {
		return err
	}
	go func() {
		for _, el := range data.Indicators[0].Indicators {
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprint(el.LocationID),
				},
				Name: el.Location,
				Location: &core.Location{
					Latitude:  core.TruncCoord(el.LocX),
					Longitude: core.TruncCoord(el.LocY),
				},
				FlowUnit: "m3/s",
				URL:      fmt.Sprintf("https://hydrowebportal.niwa.co.nz/Data/Location/Summary/Location/%s/Interval/Latest", el.LocationIdentifier),
			}
		}
		if gauges != nil {
			close(gauges)
		}
	}()
	return nil
}

func (s *scriptNzniwa) parseLocation(locationID string) (string, string, error) {
	url := s.locationURL + locationID
	req, err := http.NewRequest("POST", url, strings.NewReader("sort=Identifier-asc&page=1&pageSize=100&group=&filter=&timezone=0"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Content-Type", "text/plain")
	if err != nil {
		return "", "", err
	}
	resp, err := core.Client.Do(req, nil)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	var data niwaLocation
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", "", err
	}

	for _, el := range data.Data {
		if el.ParameterID == 327 {
			unit := el.Unit
			switch unit {
			case "Litres per second":
				unit = "l/s"
			case "Cubic Metres Per Second":
				unit = "m3/s"
			}
			return unit, el.LocIdentifier, nil
		}
	}
	return "", "", errors.New("not found")
}

func (s *scriptNzniwa) gaugePageWorker(gauges <-chan *core.Gauge, results chan<- *core.Gauge, wg *sync.WaitGroup) {
	for g := range gauges {
		unit, id, err := s.parseLocation(g.GaugeID.Code)
		if err != nil {
			fmt.Println(err)
			s.GetLogger().WithField("code", g.GaugeID.Code).Error(err)
			continue
		}
		g.GaugeID.Code = id
		g.FlowUnit = unit
		results <- g
	}
	wg.Done()
}
