package cantabria

import (
	"bufio"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
)

const href = "<a target=\"_blank\" href=\"https://www.chcantabrico.es/web/chcmovil/evolucion-de-niveles?cod_estacion="
const hrefLen = len(href)
const valor1 = "<span class=\"texto_verde\">"
const valor2 = "</span>"

type tableEntry struct {
	gauge       *core.Gauge
	measurement *core.Measurement
}

func (s *scriptCantabria) parseTable() (<-chan *tableEntry, <-chan error, error) {
	location, err := time.LoadLocation("CET")
	if err != nil {
		return nil, nil, err
	}
	resp, err := core.Client.Get(s.listURL, nil)
	if err != nil {
		return nil, nil, err
	}
	out := make(chan *tableEntry)
	errCh := make(chan error)
	go func() {
		defer close(out)
		defer close(errCh)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(splitTable)

		ind := -1
		var fecha, hora string
		var timestamp time.Time
		var tdStack []string
		for scanner.Scan() {
			ind++
			text := scanner.Text()
			if ind == 0 {
				fecha = strings.TrimSpace(text)
				continue
			} else if ind == 1 {
				hora = strings.TrimSpace(text)
				timestamp, err = time.ParseInLocation("02-01-2006 15:04", fecha+" "+hora, location)
				if err != nil {
					errCh <- core.WrapErr(err, "failed to parse timestamp").With("text", strings.TrimSpace(text))
					return
				}
				continue
			}
			tdStack = append(tdStack, text)
			hrefInd := strings.Index(text, href)
			if hrefInd != -1 {
				code := text[hrefInd+hrefLen : hrefInd+hrefLen+4]
				valorText := tdStack[len(tdStack)-3]
				valorInd := strings.Index(valorText, valor1) + len(valor1)
				valorEnd := strings.Index(valorText, valor2)
				valor, _ := strconv.ParseFloat(valorText[valorInd:valorEnd], 64)
				station := strings.TrimSpace(tdStack[len(tdStack)-4])
				river := strings.TrimSpace(tdStack[len(tdStack)-5])
				gauge := &core.Gauge{
					GaugeID: core.GaugeID{
						Code:   code,
						Script: s.name,
					},
					Name:      river + " - " + station,
					URL:       "https://www.chcantabrico.es/sistema-automatico-de-informacion-detalle-estacion?cod_estacion=" + code,
					LevelUnit: "m",
				}
				measurement := &core.Measurement{
					GaugeID: core.GaugeID{
						Code:   code,
						Script: s.name,
					},
					Level:     nulltype.NullFloat64Of(valor),
					Timestamp: core.HTime{Time: timestamp.UTC()},
				}
				out <- &tableEntry{gauge: gauge, measurement: measurement}
				tdStack = nil
			}
		}
		if err := scanner.Err(); err != nil {
			s.GetLogger().Errorf("scanner error: %w", err)
		}
	}()
	return out, errCh, nil
}
