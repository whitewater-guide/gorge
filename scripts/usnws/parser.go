package usnws

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/sync/errgroup"
)

type attributes struct {
	Objectid  int     `json:"objectid"`
	Gaugelid  string  `json:"gaugelid"`
	Location  string  `json:"location"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Waterbody string  `json:"waterbody"`
	State     string  `json:"state"`
	Obstime   string  `json:"obstime"`
	Units     string  `json:"units"`
	Secunit   string  `json:"secunit"`
	URL       string  `json:"url"`
	Observed  string  `json:"observed"`
	Secvalue  string  `json:"secvalue"`
}

type response struct {
	Features []struct {
		Attributes attributes `json:"attributes"`
	} `json:"features"`
	ExceededTransferLimit bool `json:"exceededTransferLimit"`
}

type countResponse struct {
	Count int `json:"count"`
}

func (s *scriptUsnws) parseJson(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	var cntResp countResponse
	if err := core.Client.GetAsJSON(s.url+"?where=1%3D1&text=&objectIds=&time=&timeRelation=esriTimeRelationOverlaps&geometry=&geometryType=esriGeometryEnvelope&inSR=&spatialRel=esriSpatialRelIntersects&distance=&units=esriSRUnit_Foot&relationParam=&outFields=&returnGeometry=true&returnTrueCurves=false&maxAllowableOffset=&geometryPrecision=&outSR=&havingClause=&returnIdsOnly=false&returnCountOnly=true&orderByFields=&groupByFieldsForStatistics=&outStatistics=&returnZ=false&returnM=false&gdbVersion=&historicMoment=&returnDistinctValues=false&resultOffset=&resultRecordCount=&returnExtentOnly=false&sqlFormat=none&datumTransformation=&parameterValues=&rangeValues=&quantizationParameters=&featureEncoding=esriDefault&f=pjson", &cntResp, nil); err != nil {
		errs <- err
		return
	}
	s.GetLogger().Debugf("found %d features", cntResp.Count)
	jobs := make(chan int)
	g := new(errgroup.Group)
	for i := 0; i < s.numWorkers; i++ {
		g.Go(func() error {
			return s.worker(jobs, gauges, measurements)
		})
	}
	for offset := 0; offset < cntResp.Count; offset += s.pageSize {
		jobs <- offset
	}
	close(jobs)
	if err := g.Wait(); err != nil {
		errs <- err
	}
}

func (s *scriptUsnws) worker(jobs <-chan int, gauges chan<- *core.Gauge, measurements chan<- *core.Measurement) error {
	for offset := range jobs {
		var resp response
		// if err := core.Client.GetAsJSON(fmt.Sprintf("%s?f=json&where=(1%%3D1)%%20AND%%20(1%%3D1)&returnGeometry=false&spatialRel=esriSpatialRelIntersects&outFields=objectid,gaugelid,location,latitude,longitude,waterbody,state,obstime,units,secunit,url,observed,secvalue&orderByFields=objectid%%20ASC&outSR=102100&resultOffset=%d&resultRecordCount=%d", s.url, offset, s.pageSize), &resp, nil); err != nil {
		if err := core.Client.GetAsJSON(fmt.Sprintf("%s?f=json&where=(1%%3D1)%%20AND%%20(1%%3D1)&returnGeometry=false&spatialRel=esriSpatialRelIntersects&outFields=objectid,gaugelid,location,latitude,longitude,waterbody,state,obstime,units,secunit,url,observed,secvalue&orderByFields=objectid%%20ASC&outSR=4326&resultOffset=%d&resultRecordCount=%d", s.url, offset, s.pageSize), &resp, nil); err != nil {
			return err
		}
		for _, feat := range resp.Features {
			if gauges != nil {
				if g := s.attributesToGauge(feat.Attributes); g != nil {
					gauges <- g
				}
			} else if measurements != nil {
				if m := s.attributesToMeasurement(feat.Attributes); m != nil {
					measurements <- m
				}
			}
		}
	}
	return nil
}

func (s *scriptUsnws) attributesToGauge(attrs attributes) *core.Gauge {
	flowU, levelU, _ := getUnits(attrs)
	if flowU == "" && levelU == "" {
		return nil
	}

	zone, err := core.USCoordinateToTimezone(attrs.State, attrs.Latitude, attrs.Longitude)
	if err != nil {
		zone = "UTC"
	}
	return &core.Gauge{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   attrs.Gaugelid,
		},
		Name:      fmt.Sprintf("%s / %s / %s", attrs.Waterbody, attrs.Location, attrs.State),
		URL:       attrs.URL,
		LevelUnit: levelU,
		FlowUnit:  flowU,
		Location: &core.Location{
			Latitude:  core.TruncCoord(attrs.Latitude),
			Longitude: core.TruncCoord(attrs.Longitude),
		},
		Timezone: zone,
	}
}

func (s *scriptUsnws) attributesToMeasurement(attrs attributes) *core.Measurement {
	flowU, levelU, flowPrimary := getUnits(attrs)
	if flowU == "" && levelU == "" {
		return nil
	}
	// 2023-10-01 18:30:00",
	obstime := strings.TrimSpace(attrs.Obstime)
	if obstime == "" || obstime == "N/A" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", obstime, time.UTC)
	if err != nil {
		s.GetLogger().Warnf("failed to parse obstime %s", obstime)
		return nil
	}
	flow, level := nulltype.NullFloat64{}, nulltype.NullFloat64{}
	vPrim, vSec := strings.TrimSpace(attrs.Observed), strings.TrimSpace(attrs.Secvalue)
	if vPrim != "" && vPrim != "-999.00" {
		if f, err := strconv.ParseFloat(vPrim, 64); err != nil {
			s.GetLogger().Warnf("failed to parse observer '%s'", vPrim)
		} else if flowPrimary {
			flow.Set(f)
		} else {
			level.Set(f)
		}
	}
	if vSec != "" && vSec != "-999.00" {
		if f, err := strconv.ParseFloat(vSec, 64); err != nil {
			s.GetLogger().Warnf("failed to parse secvalue '%s'", vSec)
		} else if flowPrimary {
			level.Set(f)
		} else {
			flow.Set(f)
		}
	}

	return &core.Measurement{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   attrs.Gaugelid,
		},
		Timestamp: core.HTime{
			Time: t,
		},
		Level: level,
		Flow:  flow,
	}
}

func getUnits(attrs attributes) (flowUnit string, levelUnit string, flowPrimary bool) {
	// known units values: "cfs", "cfs*", "ft", "ft*", "kcfs", "kcfs*"
	// known secunit values: " ", "ft", "kcfs"
	switch attrs.Units {
	case "cfs", "kcfs":
		flowUnit = attrs.Units
		levelUnit = attrs.Secunit
		flowPrimary = true
	case "ft":
		flowUnit = attrs.Secunit
		levelUnit = attrs.Units
	case "cfs*", "kcfs*":
		// https://water.weather.gov/ahps2/hydrograph.php?wfo=slc&gage=lctu1
		flowPrimary = true
		flowUnit = "kcfs"
		levelUnit = attrs.Secunit
	case "ft*":
		// https://water.weather.gov/ahps2/hydrograph.php?wfo=boi&gage=andi1
		flowPrimary = true
		flowUnit = "kcfs"
		levelUnit = attrs.Secunit
	}
	levelUnit = strings.TrimSpace(levelUnit)
	flowUnit = strings.TrimSpace(flowUnit)
	return
}
