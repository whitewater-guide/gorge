package quebec2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/whitewater-guide/gorge/core"
)

type optionsQuebec2 struct{}
type scriptQuebec2 struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptQuebec2) ListGauges() (core.Gauges, error) {
	var list q2data
	err := core.Client.GetAsJSON(s.url, &list, nil)
	if err != nil {
		return nil, err
	}
	gauges := core.Gauges{}

	for _, site := range list.Site {
		lat, err := strconv.ParseFloat(site.Ycoord, 64)
		if err != nil {
			s.GetLogger().Errorf("failed to parse Ycoord for %s", site.Nom)
			continue
		}
		lon, err := strconv.ParseFloat(site.Xcoord, 64)
		if err != nil {
			s.GetLogger().Errorf("failed to parse Xcoord for %s", site.Nom)
			continue
		}
		for _, comp := range site.Composition {
			g := core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprintf("%s-%s-%s", site.Identifiant, typePointDonnee[comp.TypePointDonnee], pasTemps[comp.PasTemps]),
				},
				Name:     fmt.Sprintf("%s - %s", site.Nom, comp.Element),
				URL:      "https://www.hydroquebec.com/generation/flows-water-level.html",
				FlowUnit: "m3/s",
				Location: &core.Location{
					Latitude:  lat,
					Longitude: lon,
				},
				Timezone: "America/Toronto",
			}
			gauges = append(gauges, g)
		}
	}

	return gauges, nil
}

func (s *scriptQuebec2) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	var list q2data
	err := core.Client.GetAsJSON(s.url, &list, nil)
	if err != nil {
		errs <- err
		return
	}

	for _, site := range list.Site {
		for _, comp := range site.Composition {
			for ts, val := range comp.Donnees.values {
				recv <- &core.Measurement{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   fmt.Sprintf("%s-%s-%s", site.Identifiant, typePointDonnee[comp.TypePointDonnee], pasTemps[comp.PasTemps]),
					},
					Timestamp: core.HTime{
						Time: ts,
					},
					Flow: val,
				}
			}
		}
	}
}
