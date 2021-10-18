package galicia

import (
	"fmt"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
)

type item struct {
	gauge       *core.Gauge
	measurement *core.Measurement
}

func (s *scriptGalicia) fetchList() ([]item, error) {
	var data galiciaData
	err := core.Client.GetAsJSON(s.url, &data, nil)
	if err != nil {
		return nil, err
	}

	result := make([]item, len(data.ListaAforos))

	for i, entry := range data.ListaAforos {
		var flowValue, levelValue nulltype.NullFloat64
		flowUnit, levelUnit := "m3s", "m"
		for _, medida := range entry.ListaMedidas {
			switch medida.CodParametro {
			case 1:
				levelUnit = medida.Unidade
				levelValue = medida.Valor
			case 4:
				flowUnit = medida.Unidade
				flowValue = medida.Valor
			}
		}
		gauge := &core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   fmt.Sprintf("%d", entry.Ide),
			},
			Name: entry.NomeEstacion,
			Location: &core.Location{
				Latitude:  entry.Latitude,
				Longitude: entry.Lonxitude,
			},
			FlowUnit:  flowUnit,
			LevelUnit: levelUnit,
			URL:       "http://www2.meteogalicia.gal/servizos/AugasdeGalicia/estacionsinfo.asp?Nest=" + fmt.Sprintf("%d", entry.Ide),
			Timezone:  "Europe/Madrid",
		}
		measurement := &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   fmt.Sprintf("%d", entry.Ide),
			},
			Flow:      flowValue,
			Level:     levelValue,
			Timestamp: core.HTime{Time: entry.DataUTC.Time},
		}
		result[i] = item{gauge: gauge, measurement: measurement}
	}
	return result, nil
}
