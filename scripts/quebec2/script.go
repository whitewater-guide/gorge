package quebec2

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/sync/errgroup"
)

type optionsQuebec2 struct{}
type scriptQuebec2 struct {
	name    string
	urlBase string
	core.LoggingScript
}

func (s *scriptQuebec2) parseLocation(site q2site) (*core.Location, error) {
	lat, err := strconv.ParseFloat(site.Ycoord, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Ycoord '%s'", site.Ycoord)
	}
	lon, err := strconv.ParseFloat(site.Xcoord, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Xcoord '%s'", site.Xcoord)
	}
	return &core.Location{Latitude: lat, Longitude: lon}, nil
}

func (s *scriptQuebec2) fetchData() ([]q2site, []q2site, error) {
	g := new(errgroup.Group)
	var sitez q2sites
	var stationz q2stations
	g.Go(func() error {
		return core.Client.GetAsJSON(s.urlBase+"Donnees_VUE_CENTRALES_ET_OUVRAGES.json", &sitez, nil)
	})
	g.Go(func() error {
		return core.Client.GetAsJSON(s.urlBase+"Donnees_VUE_STATIONS_ET_TARAGES.json", &stationz, nil)
	})
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}
	return sitez.Sites, stationz.Stations, nil
}

// For sites, create multiple gauges per site, for different typePointDonnee (e.g. Débit turbiné and Débit total)
func (s *scriptQuebec2) parseSites(sites []q2site, gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	for _, site := range sites {
		for _, comp := range site.Composition {
			code := fmt.Sprintf("%s-%s-%s", site.Identifiant, typePointDonnee[comp.TypePointDonnee], pasTemps[comp.PasTemps])
			gaugeId := core.GaugeID{
				Script: s.name,
				Code:   code,
			}

			if gauges != nil {
				loc, err := s.parseLocation(site)
				if err != nil {
					s.GetLogger().Errorf("failed to parse locations for %s: %s", site.Nom, err)
					continue
				}

				gauges <- &core.Gauge{
					GaugeID:  gaugeId,
					Name:     fmt.Sprintf("%s - %s", site.Nom, comp.Element),
					URL:      "https://www.hydroquebec.com/generation/flows-water-level.html",
					FlowUnit: "m3/s",
					Location: loc,
					Timezone: "America/Toronto",
				}
			}
			if measurements != nil {
				for ts, val := range comp.Donnees.values {
					measurements <- &core.Measurement{
						GaugeID: gaugeId,
						Timestamp: core.HTime{
							Time: ts,
						},
						Flow: val,
					}
				}
			}
		}
	}
}

// For stations we compose one or two station's compositions to provide flow and level simultaneously
func (s *scriptQuebec2) parseStations(stations []q2site, gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	for _, station := range stations {
		// There're also some strange stations which have two identical sensors for (Doppler+Debit)
		// 1-7300, 1-7290, and 1-7298
		// They're not needed anyway
		if station.Identifiant == "1-7300" || station.Identifiant == "1-7290" || station.Identifiant == "1-7298" {
			continue
		}

		var flowComp *q2composition
		var levelComp *q2composition

		for _, comp := range station.Composition {
			cmp := comp
			// skip temperature, precipation, snow, etc.
			// For Niveau (water level) there're two type of element value Limnimètre and Doppler
			// In current data set, if there's a Doppler, there is always also a Limnimètre
			// But Doppler also measures Débit (discharge), so if there are both Limnimètre and Doppler, we prefer Doppler
			if comp.TypePointDonnee == "Niveau" && (levelComp == nil || levelComp.Element == "Limnimètre") {
				levelComp = &cmp
			} else if comp.TypePointDonnee == "Débit" && flowComp == nil {
				flowComp = &cmp
			}
		}

		if flowComp == nil && levelComp == nil {
			continue
		}
		var flowUnit, levelUnit string
		if flowComp != nil {
			flowUnit = strings.ReplaceAll(flowComp.NomUniteMesure, "³", "3")
		}
		if levelComp != nil {
			levelUnit = levelComp.NomUniteMesure
		}

		gaugeId := core.GaugeID{
			Script: s.name,
			Code:   station.Identifiant,
		}

		if gauges != nil {
			loc, err := s.parseLocation(station)
			if err != nil {
				s.GetLogger().Errorf("failed to parse locations for %s: %s", station.Nom, err)
				continue
			}

			gauges <- &core.Gauge{
				GaugeID:   gaugeId,
				Name:      station.Nom,
				URL:       "https://www.hydroquebec.com/generation/flows-water-level.html",
				FlowUnit:  flowUnit,
				LevelUnit: levelUnit,
				Location:  loc,
				Timezone:  "America/Toronto",
			}
		}

		if measurements != nil {
			byTime := map[time.Time]*core.Measurement{}
			if flowComp != nil {
				for ts, val := range flowComp.Donnees.values {
					appendMeasurement(byTime, gaugeId, ts, val, true)
				}
			}
			if levelComp != nil {
				for ts, val := range levelComp.Donnees.values {
					appendMeasurement(byTime, gaugeId, ts, val, false)
				}
			}

			for _, m := range byTime {
				measurements <- m
			}
		}
	}
}

func appendMeasurement(to map[time.Time]*core.Measurement, gaugeId core.GaugeID, ts time.Time, value nulltype.NullFloat64, isFlow bool) {
	m := to[ts]
	if m == nil {
		m = &core.Measurement{
			GaugeID: gaugeId,
			Timestamp: core.HTime{
				Time: ts,
			},
		}
	}
	if isFlow {
		m.Flow = value
	} else {
		m.Level = value
	}
	to[ts] = m
}

func (s *scriptQuebec2) ListGauges() (core.Gauges, error) {
	sites, stations, err := s.fetchData()
	if err != nil {
		return nil, err
	}
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.parseSites(sites, gaugesCh, nil, errCh)
		s.parseStations(stations, gaugesCh, nil, errCh)
	}()
	return core.GaugeSinkToSlice(gaugesCh, errCh)
}

func (s *scriptQuebec2) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	sites, stations, err := s.fetchData()
	if err != nil {
		errs <- err
		return
	}
	s.parseSites(sites, nil, recv, errs)
	s.parseStations(stations, nil, recv, errs)
}
