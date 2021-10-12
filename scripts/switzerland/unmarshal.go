package switzerland

import (
	"encoding/xml"
	"strings"
	"time"
)

var swissTimezone, _ = time.LoadLocation("Europe/Zurich")

type swissTime struct {
	time.Time
}

func (c *swissTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	err := d.DecodeElement(&v, &start)
	if err != nil {
		return err
	}
	parts := strings.Split(v, "+")
	tstr := parts[0]
	var z string
	if len(parts) > 1 {
		z = parts[1]
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05", tstr, swissTimezone)
	if z == "01:00" {
		t = t.Add(time.Hour)
	}
	if err != nil {
		return err
	}
	c.Time = t.UTC()
	return nil
}

type locations struct {
	XMLName        xml.Name  `xml:"locations"`
	Text           string    `xml:",chardata"`
	Xsi            string    `xml:"xsi,attr"`
	ExportTime     string    `xml:"export-time,attr"`
	SchemaLocation string    `xml:"schemaLocation,attr"`
	Station        []station `xml:"station"`
}

type station struct {
	Text          string      `xml:",chardata"`
	Name          string      `xml:"name,attr"`
	Easting       int         `xml:"easting,attr"`
	Northing      int         `xml:"northing,attr"`
	Number        string      `xml:"number,attr"`
	WaterBodyName string      `xml:"water-body-name,attr"`
	WaterBodyType string      `xml:"water-body-type,attr"`
	Parameter     []parameter `xml:"parameter"`
}

type parameter struct {
	Text      string    `xml:",chardata"`
	Name      string    `xml:"name,attr"`
	Unit      string    `xml:"unit,attr"`
	FieldName string    `xml:"field-name,attr"`
	Datetime  swissTime `xml:"datetime"`
	Value     value     `xml:"value"`
	Max24h    value     `xml:"max-24h"`
	Mean24h   string    `xml:"mean-24h"`
	Min24h    string    `xml:"min-24h"`
}

type value struct {
	Text             string `xml:",chardata"`
	HqClass          string `xml:"hq-class,attr"`
	QuantileClass    string `xml:"quantile-class,attr"`
	WarnLevelClass   string `xml:"warn-level-class,attr"`
	TemperatureClass string `xml:"temperature-class,attr"`
}
