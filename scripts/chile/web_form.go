package chile

import (
	"bufio"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/whitewater-guide/gorge/core"
)

const (
	optionsStart   = "<select name=\"estacion1\""
	optionsEnd     = "</select>"
	optionStart    = "<option value="
	optionEnd      = "</option>"
	optionStartLen = len(optionStart)
	optionEndLen   = len(optionEnd)
)

// Not all gauges contain flow and/or level data
// To check if given gauges are useful we simulate user expeience with http://dgasatel.mop.cl/filtro_paramxestac_new.asp website
// User can submit up to 3 gauges at a time
// When gauges are selected it's possible to select which parameters do we query next
// If this list of parameters contains level/flow then the gauge is useful
func (s *scriptChile) areGaugesUseful(ids []string, data map[string]bool) error {
	if len(ids) > 3 {
		return fmt.Errorf("no more than 3 ids at a time allowed, but received %d", len(ids))
	}
	estacion1 := ids[0]
	if estacion1 == "" {
		estacion1 = "-1"
	}
	estacion2 := ids[1]
	if estacion2 == "" {
		estacion2 = "-1"
	}
	estacion3 := ids[2]
	if estacion3 == "" {
		estacion3 = "-1"
	}
	tz, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return err
	}
	t := time.Now().In(tz)
	html, _, err := core.Client.PostFormAsString(s.selectFormURL, url.Values{
		"accion":     {"refresca"},
		"EsDL1":      {"0"},
		"EsDL2":      {"0"},
		"EsDL3":      {"0"},
		"estacion1":  {estacion1},
		"estacion2":  {estacion2},
		"estacion3":  {estacion3},
		"fecha_fin":  {t.Format("02/01/2006")},
		"fecha_finP": {t.Format("02/01/2006")},
		"fecha_ini":  {t.Format("02/01/2006")},
		"hora_fin":   {"0"},
		"tipo":       {"ANO"},
		"UserID":     {"nobody"},
	}, nil)
	if err != nil {
		return err
	}

	// sometimes a retry is needed
	if !strings.Contains(html, "DATOS EN TABLAS") {
		time.Sleep(10 * time.Second)
		return s.areGaugesUseful(ids, data)
	}

	// append "_1" to gauge id => value of "Nivel de agua" checkbox (level)
	// append "_12" to gauge id => value of "Caudal" checkbox (flow)
	for _, id := range ids {
		if id == "" {
			continue
		}
		data[id] = strings.Contains(html, id+"_1>") || strings.Contains(html, id+"_12>")
	}

	return nil
}

func splitOptions(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	dataStr := string(data)
	if start := strings.Index(dataStr, optionStart); start >= 0 {
		end := strings.Index(dataStr[start:], optionEnd)
		if end <= 0 {
			return
		}
		return end + start + optionEndLen, data[start+optionStartLen : start+end], nil
	}

	return
}

func parseOption(opt string) (id string, name string) {
	valAndRest := strings.Split(opt, ">")
	id = valAndRest[0]
	rest := strings.Join(valAndRest[1:], " ")
	valAndRest = strings.Split(rest, " ")
	name = strings.TrimSpace(strings.Join(valAndRest[1:], " "))
	return
}

/**
 * Parse dropdown select options and get gauge ids
 * Returns map where gauge id is the key and gauge name is the value
 */
func (s *scriptChile) getListedGauges() (map[string]string, error) {
	html, err := core.Client.GetAsString(s.selectFormURL, nil)
	if err != nil {
		return nil, err
	}
	optionsStartIndex := strings.Index(html, optionsStart)
	optionsEndIndex := strings.Index(html, optionsEnd) + len(optionsEnd)
	html = html[optionsStartIndex:optionsEndIndex]

	scanner := bufio.NewScanner(strings.NewReader(html))
	scanner.Split(splitOptions)
	result := make(map[string]string)
	for scanner.Scan() {
		value := scanner.Text()
		id, name := parseOption(value)
		if id != "\"-1\"" {
			result[id] = name
		}
	}

	return result, nil
}
