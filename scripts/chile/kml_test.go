package chile

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestChile_ParseKml(t *testing.T) {
	s := `<Placemark>
	<name>FLUVIOMETRICAS</name>
	<Snippet maxLines="0"></Snippet>
	<description><![CDATA[CODIGO BNA = 01001001-2<br />
NOMBRE = RIO CAQUENA EN NACIMIENTO<br />
ALTITUD = 4385<br />
VIGENCIA = SUSPENDIDA<br />
INICIO = Jul 1 1976<br />
SUSPENSION = Feb 11 2010<br />
PARAMETRO = CAUDAL (m3/s)<br />]]></description>
	<styleUrl>#FEATURES00</styleUrl>
	<Point>
		<coordinates>-69.19807010919931,-18.0804342517309,0</coordinates>
	</Point>
</Placemark>`
	var pm kmlPlacemark
	err := xml.Unmarshal([]byte(s), &pm)
	if assert.NoError(t, err) {
		expected := kmlPlacemark{
			Name: "FLUVIOMETRICAS",
			Snippet: kmlPmSnippet{
				MaxLines: "0",
			},
			Description: kmlPmDesc{
				codigo:     "01001001-2",
				nombre:     "Rio Caquena En Nacimiento",
				altitud:    4385,
				vigencia:   false,
				inicio:     "Jul 1 1976",
				suspension: "Feb 11 2010",
				parametro:  "CAUDAL (m3/s)",
			},
			StyleURL: "#FEATURES00",
			Point: kmlPmPoint{
				Coordinates: core.Location{
					Longitude: -69.19807,
					Latitude:  -18.08043,
				},
			},
		}
		assert.Equal(t, expected, pm)
	}
}
