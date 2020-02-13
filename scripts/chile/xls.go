package chile

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/net/html"
)

func (s *scriptChile) loadXLS(code string, since int64, retry bool) (string, error) {
	tz, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return "", nil
	}
	period := "1d"
	if since == 0 {
		period = "3m"
	}
	t := time.Now().In(tz)
	cookieErr := core.Client.EnsureCookie("http://dgasatel.mop.cl", !retry)
	if cookieErr != nil {
		s.GetLogger().Warn("cookie error", cookieErr)
	}
	values := url.Values{
		"accion":         {"refresca"},
		"chk_estacion1a": {code + "_1", code + "_12"},
		"chk_estacion1b": {""},
		"chk_estacion2a": {""},
		"chk_estacion2b": {""},
		"chk_estacion3a": {""},
		"chk_estacion3b": {""},
		"estacion1":      {code},
		"estacion2":      {"-1"},
		"estacion3":      {"-1"},
		"fecha_fin":      {t.Format("02/01/2006")},
		"fecha_finP":     {t.Format("02/01/2006")},
		"fecha_ini":      {t.Format("02/01/2006")},
		"period":         {period},
		"tiporep":        {"I"},
	}
	html, _, err := core.Client.PostFormAsString(s.xlsURL, values, nil)

	if !strings.Contains(html, "tabla para resultados numerados") {
		if retry {
			return s.loadXLS(code, since, false)
		}
		s.GetLogger().WithFields(logrus.Fields{
			"since":     since,
			"cookieErr": cookieErr != nil,
			"values":    values.Encode(),
		}).Warn("missing data table in XLS response")
	}

	return html, err
}

type columnIndices struct {
	timestamp int
	flow      int
	level     int
}

func newIndices() columnIndices {
	return columnIndices{-1, -1, -1}
}

func badIndices(ind columnIndices) bool {
	return ind.timestamp < 0 || ind.flow < 0 || ind.level < 0
}

/**
 * Takes raw xls file content and extracts strings that is html table with needed data
 */
func extractDataTable(raw string) (string, error) {
	startComment := "<!-- tabla para resultados numerados -->"
	endComment := "<!-- tabla con pixel gif transparente"
	startIndex := strings.Index(raw, startComment)
	if startIndex == -1 {
		return "", fmt.Errorf("cannot find tabla para resultados numerados")
	}
	rest := raw[startIndex+len(startComment):]
	endIndex := strings.Index(rest, endComment)
	if endIndex == -1 {
		return "", fmt.Errorf("cannot find end of tabla para resultados numerados")
	}
	return strings.TrimSpace(rest[:endIndex]), nil
}

func findColumnIndices(header *html.Node) (columnIndices, error) {
	ind := newIndices()
	for i, c := 0, header.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode && c.Data != "th" {
			continue
		}
		var thTxt strings.Builder
		err := html.Render(&thTxt, c)
		if err != nil {
			return ind, err
		}
		if strings.Contains(thTxt.String(), "Fecha-Hora") {
			ind.timestamp = i
		}
		if strings.Contains(thTxt.String(), "AltLM") {
			ind.level = i
		}
		if strings.Contains(thTxt.String(), "Caudal") {
			ind.flow = i
		}
		i++
	}
	return ind, nil
}

func findTableHeaderRow(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "tr" && len(node.Attr) == 1 && node.Attr[0].Key == "bgcolor" && node.Attr[0].Val == "D5E1F4" {
		return node
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		found := findTableHeaderRow(c)
		if found != nil {
			return found
		}
	}
	return nil
}

func parseDataRow(tr *html.Node, ind columnIndices, tz *time.Location) (result core.Measurement, err error) {
	for i, c := 0, tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "td" {
			continue
		}
		if i == ind.timestamp {
			if c.FirstChild == nil {
				err = fmt.Errorf("cannot find first child for timestamp")
				return
			}
			timeStr := c.FirstChild.Data
			var t time.Time
			t, err = time.ParseInLocation("02/01/2006 15:04", timeStr, tz)
			if err != nil {
				return
			}
			result.Timestamp = core.HTime{Time: t.UTC()}
		}
		if i == ind.flow {
			if c.FirstChild == nil {
				err = fmt.Errorf("cannot find first child for flow")
				return
			}
			err = result.Flow.UnmarshalJSON([]byte(c.FirstChild.Data))
			if err != nil {
				return
			}
		}
		if i == ind.level {
			if c.FirstChild == nil {
				err = fmt.Errorf("cannot find first child for level")
				return
			}
			err = result.Level.UnmarshalJSON([]byte(c.FirstChild.Data))
			if err != nil {
				return
			}
		}
		i++
	}
	return
}

func (s *scriptChile) parseXLS(recv chan<- *core.Measurement, errs chan<- error, code string, since int64) {
	rawDoc, err := s.loadXLS(code, since, true)
	if err != nil {
		errs <- err
		return
	}
	tableStr, err := extractDataTable(rawDoc)
	if err != nil {
		errs <- err
		return
	}
	doc, err := html.Parse(strings.NewReader(tableStr))
	if err != nil {
		errs <- err
		return
	}
	header := findTableHeaderRow(doc)
	if header == nil {
		errs <- fmt.Errorf("cannot find data table header")
		return
	}
	indices, err := findColumnIndices(header)
	if err != nil {
		errs <- err
		return
	}
	if badIndices(indices) {
		errs <- fmt.Errorf("could not find flow or level column")
		return
	}
	thead := header.Parent
	tbody := thead.NextSibling
	for tbody.Type != html.ElementNode || tbody.Data != "tbody" {
		tbody = tbody.NextSibling
	}
	// tbody contains rows with actual data
	santiago, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return
	}
	for c := tbody.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tr" {
			m, err := parseDataRow(c, indices, santiago)
			if err != nil {
				s.GetLogger().Error(err)
				continue
			}
			m.Script = s.name
			m.Code = code
			recv <- &m
		}
	}
}
