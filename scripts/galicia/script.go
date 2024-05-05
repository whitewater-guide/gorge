package galicia

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type optionsGalicia struct {
	Code string `desc:"Auth code"`
}
type scriptGalicia struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptGalicia) ListGauges() (core.Gauges, error) {
	l, err := s.fetch()
	if err != nil {
		return nil, err
	}
	result := make([]core.Gauge, len(l))
	for i, item := range l {
		name := strings.ReplaceAll(fmt.Sprintf("%s @ %s", item.Estacion, item.Concello), "_", " ")
		name = cases.Title(language.Spanish).String(strings.ToLower(name))
		result[i] = core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   gaugeCode(item.IDEstacion),
			},
			Name:      fmt.Sprintf("[%s] %s", item.Prov, name),
			LevelUnit: "cm",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  item.Lat,
				Longitude: item.Lon,
			},
			URL:      fmt.Sprintf("https://servizos.meteogalicia.gal/mgafos/estacionshistorico/historico.action?idEst=%d", item.IDEstacion),
			Timezone: "Europe/Madrid",
		}
	}
	return result, nil
}

func (s *scriptGalicia) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	list, err := s.fetch()
	if err != nil {
		errs <- err
		return
	}
	for _, i := range list {
		t, err := time.ParseInLocation("2006-01-02T15:04:05", i.DataUTC, time.UTC)
		if err != nil {
			s.GetLogger().WithField("station", i.IDEstacion).Warnf("cannot parse time '%s': %s", i.DataUTC, err)
			continue
		}
		level, flow := nulltype.NullFloat64Of(i.ValorNivel), nulltype.NullFloat64{}
		if i.ValorCaudal != -9999.0 {
			flow = nulltype.NullFloat64Of(i.ValorCaudal)
		}

		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   gaugeCode(i.IDEstacion),
			},
			Timestamp: core.HTime{Time: t},
			Level:     level,
			Flow:      flow,
		}
	}
}

func (s *scriptGalicia) fetch() ([]entry, error) {
	var data entries
	err := core.Client.GetAsJSON(s.url, &data, nil)
	if err != nil {
		return nil, err
	}
	return data.ListEstadoActual, nil
}
