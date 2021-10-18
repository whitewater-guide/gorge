package chile

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

type kmlPmDesc struct {
	codigo     string
	nombre     string
	altitud    float64
	vigencia   bool
	inicio     string
	suspension string
	parametro  string
}

type kmlPmSnippet struct {
	MaxLines string `xml:"maxLines,attr"`
}

type kmlPmPoint struct {
	Coordinates core.Location
}

type kmlPlacemark struct {
	Name        string       `xml:"name"`
	Snippet     kmlPmSnippet `xml:"Snippet"`
	Description kmlPmDesc    `xml:"description"`
	StyleURL    string       `xml:"styleUrl"`
	Point       kmlPmPoint   `xml:"Point"`
}

func (desc *kmlPmPoint) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw struct {
		Coordinates string `xml:"coordinates"`
	}
	err := d.DecodeElement(&raw, &start)
	if err != nil {
		return err
	}
	lla := strings.Split(raw.Coordinates, ",")
	if len(lla) != 3 {
		return errors.New("failed to split coordinate")
	}
	lon, err := strconv.ParseFloat(lla[0], 64)
	if err != nil {
		return fmt.Errorf("failed to parse longitude: %w", err)
	}
	lat, err := strconv.ParseFloat(lla[1], 64)
	if err != nil {
		return fmt.Errorf("failed to parse latitude: %w", err)
	}
	desc.Coordinates.Longitude = core.TruncCoord(lon)
	desc.Coordinates.Latitude = core.TruncCoord(lat)
	return nil
}

func (desc *kmlPmDesc) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	err := d.DecodeElement(&s, &start)
	if err != nil {
		return err
	}
	lines := strings.Split(s, "<br />")
	for _, l := range lines {
		kv := strings.Split(l, " = ")
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		switch k {
		case "CODIGO BNA":
			desc.codigo = val
		case "NOMBRE":
			desc.nombre = strings.Title(strings.ToLower(val))
		case "ALTITUD":
			alt, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("failed to parse altitude: %w", err)
			}
			if alt == -9999 {
				alt = 0
			}
			desc.altitud = alt
		case "VIGENCIA":
			desc.vigencia = val == "VIGENTE"
		case "INICIO":
			desc.inicio = val
		case "SUSPENSION":
			desc.suspension = val
		case "PARAMETRO":
			desc.parametro = val // everything is CAUDAL (m3/s)
		}
	}
	return nil
}

// generated with https://www.onlinetool.io/xmltogo/
type chileKml struct {
	XMLName  xml.Name `xml:"kml"`
	Text     string   `xml:",chardata"`
	Xmlns    string   `xml:"xmlns,attr"`
	Gx       string   `xml:"gx,attr"`
	Kml      string   `xml:"kml,attr"`
	Atom     string   `xml:"atom,attr"`
	Document struct {
		Text    string `xml:",chardata"`
		Name    string `xml:"name"`
		Open    string `xml:"open"`
		Snippet struct {
			Text     string `xml:",chardata"`
			MaxLines string `xml:"maxLines,attr"`
		} `xml:"Snippet"`
		Description string `xml:"description"`
		Style       []struct {
			Text      string `xml:",chardata"`
			ID        string `xml:"id,attr"`
			IconStyle struct {
				Text  string `xml:",chardata"`
				Color string `xml:"color"`
				Scale string `xml:"scale"`
				Icon  struct {
					Text string `xml:",chardata"`
					Href string `xml:"href"`
					X    string `xml:"x"`
					Y    string `xml:"y"`
					W    string `xml:"w"`
					H    string `xml:"h"`
				} `xml:"Icon"`
			} `xml:"IconStyle"`
			LabelStyle struct {
				Text  string `xml:",chardata"`
				Color string `xml:"color"`
			} `xml:"LabelStyle"`
		} `xml:"Style"`
		StyleMap []struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Pair []struct {
				Text     string `xml:",chardata"`
				Key      string `xml:"key"`
				StyleURL string `xml:"styleUrl"`
			} `xml:"Pair"`
		} `xml:"StyleMap"`
		Folder struct {
			Text      string         `xml:",chardata"`
			Name      string         `xml:"name"`
			Placemark []kmlPlacemark `xml:"Placemark"`
		} `xml:"Folder"`
	} `xml:"Document"`
}

func (s *scriptChile) getKMLGauges() (map[string]core.Gauge, error) {
	var data chileKml
	err := core.Client.GetAsXML("http://documentos.dga.cl/KML/01_Red_Hidrometrica/Fluviometricas_001.kml", &data, nil)
	if err != nil {
		return nil, err
	}
	result := make(map[string]core.Gauge)
	for _, pm := range data.Document.Folder.Placemark {
		if pm.Name != "FLUVIOMETRICAS" || !pm.Description.vigencia {
			continue
		}
		result[pm.Description.codigo] = core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   pm.Description.codigo,
			},
			Name:      pm.Description.nombre,
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  pm.Point.Coordinates.Latitude,
				Longitude: pm.Point.Coordinates.Longitude,
				Altitude:  pm.Description.altitud,
			},
			Timezone: "America/Santiago",
		}
	}
	return result, nil
}
