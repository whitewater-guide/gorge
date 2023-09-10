package usnws

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

var (
	reCode = regexp.MustCompile(`<b>NWSLID:\s*</b>\s*(.*)\s*<br />`)
	reName = regexp.MustCompile(`<b>Location:\s*</b>\s*(.*)\s*<br />`)
	reVal  = regexp.MustCompile(`<b>Latest Observation Value:\s*</b>\s*(.*)\s*<br />`)
	reTime = regexp.MustCompile(`<b>UTC Observation Time:\s*</b>\s*(.*)\s*<br />`)
	reLoc  = regexp.MustCompile(`<b>Lat/Long:\s*</b>\s*(.*)\s*<br />`)
	reHref = regexp.MustCompile(`<a href="(.*)">Link to Gauge Hydrograph</a>`)
)

type description struct {
	XMLName xml.Name `xml:"description"`
	Text    string   `xml:",cdata"`
}

func (s *scriptUsnws) parseKmz(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	client := core.NewClient(core.ClientOptions{
		UserAgent: "whitewater.guide robot",
		Timeout:   300,
	}, s.GetLogger())
	req, _ := http.NewRequest("GET", s.kmzUrl, nil)
	s.GetLogger().Debugf("fetching %s", s.kmzUrl)
	resp, err := client.Do(req, nil)
	if err != nil {
		errs <- err
		return
	}
	defer resp.Body.Close()
	s.GetLogger().Debug("fetched")
	zipFile, err := os.CreateTemp("", "usnws")
	if err != nil {
		errs <- fmt.Errorf("failed to create tmp file: %w", err)
		return
	}
	defer os.Remove(zipFile.Name())
	if _, err = io.Copy(zipFile, resp.Body); err != nil {
		errs <- fmt.Errorf("failed to write tmp file: %w", err)
		return
	}
	s.GetLogger().Debugf("saved temp zip %s", zipFile.Name())

	// Open a zip archive for reading.
	r, err := zip.OpenReader(zipFile.Name())
	if err != nil {
		errs <- fmt.Errorf("failed to open tmp file: %w", err)
		return
	}
	defer r.Close()
	if len(r.File) != 1 {
		errs <- fmt.Errorf("expected 1 file inside kmz, found many: %d", len(r.File))
		return
	}
	kmlReader, err := r.File[0].Open()
	if err != nil {
		errs <- fmt.Errorf("failed to read kml file: %w", err)
		return
	}
	defer kmlReader.Close()

	decoder := xml.NewDecoder(kmlReader)
	s.GetLogger().Debug("created xml decoder")
	var descr description
	for {
		t, err := decoder.Token()
		if err != nil || t == nil {
			if err != io.EOF {
				s.GetLogger().Errorf("xml token error: %s", err)
			}
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "description" {
				if err := decoder.DecodeElement(&descr, &se); err == nil {
					s.parseEntry(descr.Text, gauges, measurements)
				} else {
					s.GetLogger().Errorf("decoder error: %s", err)
				}
			}
		default:
		}
	}
}

func (s *scriptUsnws) parseEntry(text string, gauges chan<- *core.Gauge, measurements chan<- *core.Measurement) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanLines)
	var g core.Gauge
	g.Script = s.name
	var m core.Measurement
	mOk := false
	for scanner.Scan() {
		line := scanner.Text()
		if matches := reCode.FindStringSubmatch(line); len(matches) > 0 {
			g.Code = strings.TrimSpace(matches[1])
		} else if matches := reName.FindStringSubmatch(line); len(matches) > 0 {
			g.Name = strings.TrimSpace(matches[1])
		} else if matches := reVal.FindStringSubmatch(line); len(matches) > 0 {
			line = strings.TrimSpace(matches[1])
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				var v nulltype.NullFloat64
				vStr, unit := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

				if vF, err := strconv.ParseFloat(vStr, 64); err == nil {
					v = nulltype.NullFloat64Of(vF)
				} else {
					s.GetLogger().Warnf("cannot parse value '%s'", line)
				}

				switch unit {
				case "ft":
					m.Level = v
					g.LevelUnit = unit
				case "kcfs":
					m.Flow = v
					g.FlowUnit = unit
				default:
					s.GetLogger().Warnf("unknown unit '%s'", unit)
					m.Level = v
					g.LevelUnit = unit
				}
				mOk = true
			} else if line != "N/A" {
				// when the value is N/A, it's impossible to find out unit even from other lines, such as flood threshold
				s.GetLogger().Warnf("cannot parse value '%s'", line)
				continue
			}
		} else if matches := reTime.FindStringSubmatch(line); len(matches) > 0 {
			if t, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(matches[1])); err == nil {
				m.Timestamp = core.HTime{Time: t}
			} else {
				mOk = false
				if matches[1] != "N/A" {
					s.GetLogger().Warnf("cannot parse time '%s'", matches[1])
				}
			}

		} else if matches := reLoc.FindStringSubmatch(line); len(matches) > 0 {
			parts := strings.Split(strings.TrimSpace(matches[1]), ",")
			if len(parts) != 2 {
				s.GetLogger().Warnf("cannot parse location '%s'", matches[1])
				continue
			}
			g.Location = &core.Location{}
			if lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
				if lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					g.Location = &core.Location{Latitude: lat, Longitude: lon}
					zone, err := core.CoordinateToTimezone(lat, lon)
					if err != nil {
						s.GetLogger().Warnf("cannot find timezone for (%f, %f)", lat, lon)
						zone = "UTC"
					}
					g.Timezone = zone
				} else {
					s.GetLogger().Warnf("cannot parse longtitude '%s'", parts[1])
				}
			} else {
				s.GetLogger().Warnf("cannot parse latitude '%s'", parts[0])
			}
		} else if matches := reHref.FindStringSubmatch(line); len(matches) > 0 {
			g.URL = strings.TrimSpace(matches[1])
		}
	}
	if gauges != nil && g.Code != "" {
		gauges <- &g
	}
	if measurements != nil && g.Code != "" && mOk {
		m.GaugeID = g.GaugeID
		measurements <- &m
	}
}
