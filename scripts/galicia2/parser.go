package galicia2

import (
	"bufio"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/im7mortal/UTM"
	"github.com/whitewater-guide/gorge/core"
)

const (
	colOpen     = "<td"
	colClose    = "</td>"
	colCloseLen = len(colClose)
	delim1      = "<!----------------------------------------------------------------------------- LINEA 1 ------------------------------------------------------------------>"
	delim2      = "<!----------------------------------------------------------------------------- LINEA 2 ------------------------------------------------------------------>"
)

type item struct {
	gauge       core.Gauge
	measurement core.Measurement
}

func splitColumns(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	dataStr := string(data)
	if start := strings.Index(dataStr, colOpen); start >= 0 {
		openTagEnd := strings.Index(dataStr[start:], ">")
		if openTagEnd <= 0 {
			return
		}
		openTagEnd += start
		end := strings.Index(dataStr[start:], colClose)
		//fmt.Println(start, end)
		if end <= 0 {
			return
		}
		end += start
		return end + colCloseLen, data[openTagEnd+1 : end], nil
	}

	return

}

func prettyName(name string) string {
	words := strings.Fields(html.UnescapeString(name))
	for i, word := range words {
		if word == "DE" || word == "EN" {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = word[:1] + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func (s *scriptGalicia2) parseTable() ([]item, error) {
	var result []item
	if !s.skipCookies {
		err := core.Client.EnsureCookie("http://saih.chminosil.es", false)
		if err != nil {
			return result, err
		}
	}
	resp, err := core.Client.Get(s.listURL, nil)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	core.Client.SaveCookies()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(splitColumns)

	location, err := time.LoadLocation("CET")
	if err != nil {
		return result, err
	}

	ind := 0
	var gauge core.Gauge
	var msm core.Measurement
	stationExp := regexp.MustCompile(`([A-Z]\d{3})\s-\s(.*)`)
	header := true
	for scanner.Scan() {
		if header {
			if ind == 6 {
				header = false
				ind = 0
			} else {
				ind++
			}
			continue
		}
		// There are 7 columns in a row
		switch ind {
		case 0:
			gauge = core.Gauge{LevelUnit: "m", Timezone: "Europe/Madrid"}
		case 1: // River name and code
			station := scanner.Text()
			parts := stationExp.FindStringSubmatch(station)
			gauge.Script = s.name
			gauge.Code = parts[1]
			gauge.Name = prettyName(parts[2])
			gauge.URL = fmt.Sprintf(s.gaugeURLFormat, parts[1])
		case 5: // Level
			levelStr := scanner.Text()
			levelStr = strings.Replace(levelStr, ",", ".", 1)
			msm = core.Measurement{GaugeID: gauge.GaugeID}
			msm.Level.UnmarshalJSON([]byte(levelStr)) //nolint:errcheck
		case 6: // timestamp
			t, _ := time.ParseInLocation("02/01/2006 15:04", scanner.Text(), location)
			msm.Timestamp = core.HTime{Time: t.UTC()}
			result = append(result, item{gauge: gauge, measurement: msm})
		}
		ind = (ind + 1) % 7
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, nil
}

// http://saih.chminosil.es/index.php?url=/datos/ficha/estacion:N015
func (s *scriptGalicia2) parseGaugePage(code string) (lat float64, lon float64, altitude float64) {
	html, err := core.Client.GetAsString(fmt.Sprintf(s.gaugeURLFormat, code), nil)
	if err != nil {
		return
	}
	bodyInd := strings.Index(html, delim1) + len(delim1)
	bodyEnd := strings.Index(html, delim2)
	html = html[bodyInd:bodyEnd]
	trEnd := strings.Index(html, "</tr>")
	html = html[trEnd+5:]
	trEnd = strings.Index(html, "</tr>")
	html = html[:trEnd+5]

	scanner := bufio.NewScanner(strings.NewReader(html))
	scanner.Split(splitColumns)
	i, coord := 0, [4]float64{}
	for scanner.Scan() {
		coord[i], _ = strconv.ParseFloat(scanner.Text(), 64)
		i++
	}
	altitude = coord[3]
	lat, lon, err = UTM.ToLatLon(coord[1], coord[2], int(coord[0]), "", true)
	if err != nil {
		fmt.Println(err)
	}
	lat = core.TruncCoord(lat)
	lon = core.TruncCoord(lon)
	return
}
