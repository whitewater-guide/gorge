package quebec

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var tz, _ = time.LoadLocation("America/Toronto")

func parseValue(str string) (nulltype.NullFloat64, error) {
	s := strings.Replace(str, ",", ".", -1)
	s = strings.Replace(s, "*", "", -1)
	s = strings.Replace(s, ",", ".", -1)
	s = strings.Replace(s, "Â\u00a0", "", -1)
	result := nulltype.NullFloat64{}
	err := result.UnmarshalJSON([]byte(s))
	return result, err
}

func (s *scriptQuebec) parseReadings(recv chan<- *core.Measurement, errs chan<- error, r io.Reader, tz *time.Location, code string) {
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	levelInd, flowInd := -1, -1
	for {
		line, err := reader.Read()
		logger := s.GetLogger().WithField("row", strings.Join(line, ", "))
		if err == io.EOF {
			break
		} else if e, ok := err.(*csv.ParseError); ok && e.Err == csv.ErrFieldCount {
			continue
		} else if err != nil {
			logger.Errorf("csv line error: %v", err)
			continue
		}
		if len(line) == 1 && line[0] == "<!DOCTYPE html>" {
			logger.Error("received html instead of csv")
			errs <- fmt.Errorf("received html instead of csv")
			return
		}
		if len(line) < 3 || len(line) > 5 {
			logger.Errorf("unexpected csv format with %d rows insteas of 3 or 4", len(line))
			continue
		}
		if levelInd == -1 && flowInd == -1 {
			for i, v := range line {
				vt := strings.TrimSpace(v)
				if vt == "Niveau" {
					levelInd = i
				}
				if vt == "Débit" || vt == "DÈbit" || vt == "DÃ©bit" {
					flowInd = i
				}
			}
			continue
		}
		ts, err := time.ParseInLocation("2006-01-02T15:04", line[0]+"T"+line[1], tz)
		if err != nil {
			logger.Error("failed to parse time")
			continue
		}
		var level, flow nulltype.NullFloat64
		if levelInd != -1 {
			level, err = parseValue(line[levelInd])
			if err != nil {
				logger.Error("failed to parse level")
				continue
			}
		}
		if flowInd != -1 {
			flow, err = parseValue(line[flowInd])
			if err != nil {
				logger.Error("failed to parse flow")
				continue
			}
		}
		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   code,
			},
			Level:     level,
			Flow:      flow,
			Timestamp: core.HTime{Time: ts.UTC()},
		}
	}
}

func (s *scriptQuebec) getReadings(recv chan<- *core.Measurement, errs chan<- error, code string) {
	if err := s.getReadingsJson(recv, errs, code); err != nil {
		s.GetLogger().WithField("code", code).Debugf("failed to get json: %s", err)
		s.getReadingsCSV(recv, errs, code)
	}
}

func (s *scriptQuebec) getReadingsJson(recv chan<- *core.Measurement, errs chan<- error, code string) error {
	var dest qJson
	if err := core.Client.GetAsJSON(fmt.Sprintf(s.readingsJSONFormat, code), &dest, &core.RequestOptions{SkipCookies: true, RetryErrors: true}); err != nil {
		return err
	}
	all := map[int64]core.Measurement{}
	for _, d := range dest.Diffusion {
		t, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", d.DateDonnee, d.HeureDonnee), tz)
		if err != nil {
			errs <- err
			continue
		}
		m := all[t.Unix()]
		m.Timestamp = core.HTime{Time: t.UTC()}
		m.GaugeID = core.GaugeID{
			Script: s.name,
			Code:   code,
		}
		if d.TypeDonnee == "D" {
			m.Flow = d.Donnee
			all[t.Unix()] = m
		} else if d.TypeDonnee == "N" {
			m.Level = d.Donnee
			all[t.Unix()] = m
		}
	}
	for _, m := range all {
		v := m
		recv <- &v
	}
	return nil
}

func (s *scriptQuebec) getReadingsCSV(recv chan<- *core.Measurement, errs chan<- error, code string) {
	// will set-cookies every time, until max headers length overflows
	// the workaround is to ignore cookies entirely
	resp, err := core.Client.Get(fmt.Sprintf(s.readingsCSVFormat, code), &core.RequestOptions{SkipCookies: true, RetryErrors: true})
	if err != nil {
		errs <- err
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		s.GetLogger().
			WithField("statusCode", resp.StatusCode).
			WithField("requestHeaders", resp.Request.Header).
			WithField("responseHeaders", resp.Header).
			Debug("request failed")
		errs <- fmt.Errorf("readings request returned %d", resp.StatusCode)
		return
	}
	reader := transform.NewReader(resp.Body, charmap.Windows1252.NewDecoder())
	s.parseReadings(recv, errs, reader, tz, code)
}
