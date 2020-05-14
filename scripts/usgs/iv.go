package usgs

import (
	"fmt"
	"strconv"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptUSGS) listInstantaneousValues(codes string, recv chan<- *core.Measurement, errs chan<- error) {
	var root ivRoot
	err := core.Client.GetAsJSON(fmt.Sprintf("%s/iv/?format=rdb&sites=%s&modifiedSince=PT1H&parameterCd=%s,%s&siteType=ST&siteStatus=active", s.url, codes, paramFlow, paramLevel), &root, nil)
	if err != nil {
		errs <- err
		return
	}
	// Flow and level for same station and timestamp will be present as 2 separate items
	byCodeAndTime := map[string]map[int64]core.Measurement{}
	for _, ts := range root.Value.TimeSeries {
		// code := ts.SourceInfo.SiteCode[0].AgencyCode + ":" + ts.SourceInfo.SiteCode[0].Value
		code := ts.SourceInfo.SiteCode[0].Value
		isFlow := ts.Variable.VariableCode[0].Value == paramFlow
		noDataValue := ts.Variable.NoDataValue
		for _, v := range ts.Values {
			if v.Value[0].Value == "" {
				continue
			}
			vf, err := strconv.ParseFloat(v.Value[0].Value, 64)
			if err != nil {
				continue
			}
			if vf == noDataValue.Float64Value() {
				continue
			}
			when := v.Value[0].DateTime
			byTime, ok := byCodeAndTime[code]
			if !ok {
				byTime = map[int64]core.Measurement{}
			}
			m, ok := byTime[when.Unix()]
			if !ok {
				m = core.Measurement{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   code,
					},
					Timestamp: core.HTime{
						Time: when.UTC(),
					},
				}
			}
			if isFlow {
				m.Flow = nulltype.NullFloat64Of(vf)
			} else {
				m.Level = nulltype.NullFloat64Of(vf)
			}
			byTime[when.Unix()] = m
			byCodeAndTime[code] = byTime
		}
	}
	for _, byTime := range byCodeAndTime {
		for _, m := range byTime {
			mm := m
			recv <- &mm
		}
	}
}
