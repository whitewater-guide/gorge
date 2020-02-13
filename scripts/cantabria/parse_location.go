package cantabria

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

const preLen = len("var myCenter=new google.maps.LatLng(")
const postLen = len(");\n")

// http://saih.chminosil.es/index.php?url=/datos/ficha/estacion:N015
func (s *scriptCantabria) parseGaugeLocation(code string) (result core.Location) {
	resp, err := core.Client.Get(s.gaugeURLBase+code, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "var myCenter") {
			latLng := strings.Split(line[preLen+2:len(line)-postLen+1], ", ")
			latitude, _ := strconv.ParseFloat(latLng[0], 64)
			longitude, _ := strconv.ParseFloat(latLng[1], 64)
			result.Latitude = core.TruncCoord(latitude)
			result.Longitude = core.TruncCoord(longitude)
			return
		}
	}
	return
}
