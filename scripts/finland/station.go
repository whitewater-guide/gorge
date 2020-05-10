package finland

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptFinland) fetchMeasurement(code string, measurements chan<- *core.Measurement, errs chan<- error) {
	weekAgo := time.Now().In(finTz).Add(-24 * 7 * time.Hour).Format("2006-01-02T15:04:05")
	path := fmt.Sprintf("%s/Virtaama?$top=1&$filter=Paikka_Id%%20eq%%20%s%%20and%%20Aika%%20ge%%20datetime%%27%s%%27&$orderby=Aika%%20desc&$select=Aika,Arvo", s.url, code, weekAgo)
	var data virtaamaList
	err := core.Client.GetAsJSON(path, &data, nil)
	if err != nil {
		errs <- err
		return
	}
	for _, val := range data.Value {
		flow, err := strconv.ParseFloat(val.Arvo, 64)
		if err == nil {
			measurements <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   code,
				},
				Timestamp: core.HTime{
					Time: val.Aika.Time,
				},
				Flow: nulltype.NullFloat64Of(flow),
			}
		}
	}

}
