package riverzone

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

type optionsRiverzone struct {
	Key string `desc:"Auth key"`
}

type scriptRiverzone struct {
	name                string
	options             optionsRiverzone
	stationsEndpointURL string
	core.LoggingScript
}

func containsStr(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (s *scriptRiverzone) fetch(path string, dest interface{}) error {
	key := s.options.Key
	if key == "" {
		key = os.Getenv("RIVERZONE_KEY")
	}
	if key == "" {
		return fmt.Errorf("riverzone api key not found")
	}
	err := core.Client.GetAsJSON(
		s.stationsEndpointURL+path,
		&dest,
		&core.RequestOptions{
			Headers: map[string]string{"X-Key": key},
		},
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *scriptRiverzone) fetchStations() (*stationsResp, error) {
	var response *stationsResp
	err := s.fetch("", &response)
	return response, err
}

func (s *scriptRiverzone) fetchReadings() (*readingsResp, error) {
	var response *readingsResp
	err := s.fetch("/readings?from=60&to=60", &response)
	return response, err
}

func (s *scriptRiverzone) ListGauges() (core.Gauges, error) {
	stations, err := s.fetchStations()
	if err != nil {
		return nil, err
	}
	var result []core.Gauge

	for _, station := range stations.Stations {
		if !station.IsActive {
			continue
		}
		var flowUnit, levelUnit string
		if containsStr(station.Sensors, "level") {
			levelUnit = "cm"
		}
		if containsStr(station.Sensors, "flow") {
			flowUnit = "m3s"
		}
		g := core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   station.ID,
			},
			Name: strings.ReplaceAll(fmt.Sprintf("%s - %s - %s - %s", station.CountryCode, station.State, station.River.value, station.Name), "-  -", "-"),
			Location: &core.Location{
				Latitude:  core.TruncCoord(station.Latlng.Lat),
				Longitude: core.TruncCoord(station.Latlng.Lng),
				Altitude:  0,
			},
			FlowUnit:  flowUnit,
			LevelUnit: levelUnit,
			URL:       station.SourceLinks.value,
		}
		result = append(result, g)
	}

	return result, nil
}

func (s *scriptRiverzone) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	readings, err := s.fetchReadings()
	if err != nil {
		errs <- err
		return
	}
	count := 0
	s.GetLogger().Debugf("Harvested %d stations", len(readings.Readings))
	for id, reading := range readings.Readings {

		flowValues := make(map[core.HTime]*core.Measurement)
		if reading.M3s != nil {
			for _, reading := range reading.M3s {
				t := core.HTime{Time: reading.Timestamp.Time}
				flowValues[t] = &core.Measurement{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   id,
					},
					Flow:      reading.Value,
					Timestamp: t,
				}
			}
		}

		if reading.Cm != nil {
			for _, reading := range reading.Cm {
				t := core.HTime{Time: reading.Timestamp.Time}
				// Trying to find corresponding flow value:
				if flowValue, ok := flowValues[t]; ok {
					flowValue.Level = reading.Value
				} else {
					recv <- &core.Measurement{
						GaugeID: core.GaugeID{
							Script: s.name,
							Code:   id,
						},
						Level:     reading.Value,
						Timestamp: t,
					}
				}
			}
		}

		for _, v := range flowValues {
			count += 1
			recv <- v
		}
	}
	s.GetLogger().Debugf("Sent %d readings", count)
}
