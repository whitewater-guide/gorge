package riverzone

import (
	"context"
	"fmt"
	"os"

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

func (s *scriptRiverzone) fetchStations() (*stations, error) {
	key := s.options.Key
	if key == "" {
		key = os.Getenv("RIVERZONE_KEY")
	}
	if key == "" {
		return nil, fmt.Errorf("riverzone api key not found")
	}
	var response stations
	err := core.Client.GetAsJSON(
		s.stationsEndpointURL+"?status=enabled",
		&response,
		&core.RequestOptions{
			Headers: map[string]string{"X-Key": key},
		},
	)

	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *scriptRiverzone) ListGauges() (core.Gauges, error) {
	stations, err := s.fetchStations()
	if err != nil {
		return nil, err
	}
	var result []core.Gauge

	for _, station := range stations.Stations {
		if !station.Enabled {
			continue
		}
		var flowUnit, levelUnit string
		if station.Readings.Cm != nil {
			levelUnit = "cm"
		}
		if station.Readings.M3s != nil {
			flowUnit = "m3s"
		}
		g := core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   station.ID,
			},
			Name: fmt.Sprintf("%s - %s - %s - %s", station.CountryCode, station.State, station.RiverName, station.StationName),
			Location: &core.Location{
				Latitude:  core.TruncCoord(station.Latlng.Lat),
				Longitude: core.TruncCoord(station.Latlng.Lng),
				Altitude:  0,
			},
			FlowUnit:  flowUnit,
			LevelUnit: levelUnit,
			URL:       station.SourceLink,
		}
		result = append(result, g)
	}

	return result, nil
}

func (s *scriptRiverzone) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	stations, err := s.fetchStations()
	if err != nil {
		errs <- err
		return
	}
	for _, station := range stations.Stations {
		if !station.Enabled {
			continue
		}

		flowValues := make(map[core.HTime]*core.Measurement)
		if station.Readings.M3s != nil {
			for _, reading := range station.Readings.M3s {
				if reading.Value.Float64Value() == 0.0 {
					continue
				}
				t := core.HTime{Time: reading.Timestamp.Time}
				flowValues[t] = &core.Measurement{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   station.ID,
					},
					Flow:      reading.Value,
					Timestamp: t,
				}
			}
		}

		if station.Readings.Cm != nil {
			for _, reading := range station.Readings.Cm {
				if reading.Value.Float64Value() == 0.0 {
					continue
				}
				t := core.HTime{Time: reading.Timestamp.Time}
				// Trying to find corresponding flow value:
				if flowValue, ok := flowValues[t]; ok {
					flowValue.Level = reading.Value
				} else {
					recv <- &core.Measurement{
						GaugeID: core.GaugeID{
							Script: s.name,
							Code:   station.ID,
						},
						Level:     reading.Value,
						Timestamp: t,
					}
				}
			}
		}

		for _, v := range flowValues {
			recv <- v
		}
	}
}
