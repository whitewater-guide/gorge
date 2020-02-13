package switzerland

import (
	"encoding/xml"
	"time"

	"github.com/mattn/go-nulltype"
)

var swissTimezone, _ = time.LoadLocation("CET")

type swissTime struct {
	time.Time
}

func (c *swissTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	err := d.DecodeElement(&v, &start)
	if err != nil {
		return err
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05", v, swissTimezone)
	if err != nil {
		return err
	}
	c.Time = t.UTC()
	return nil
}

type swissDataRoot struct {
	XMLName  xml.Name       `xml:"locations"`
	Stations []swissStation `xml:"station"`
}

type swissStation struct {
	XMLName       xml.Name         `xml:"station"`
	Code          string           `xml:"number,attr"`
	Name          string           `xml:"name,attr"`
	WaterBodyName string           `xml:"water-body-name,attr"`
	WaterBodyType string           `xml:"water-body-type,attr"`
	Easting       int              `xml:"easting,attr"`
	Northing      int              `xml:"northing,attr"`
	Parameters    []swissParameter `xml:"parameter"`
}

type swissParameter struct {
	XMLName  xml.Name             `xml:"parameter"`
	Type     int                  `xml:"type,attr"`
	Variant  int                  `xml:"variant,attr"`
	Name     string               `xml:"name,attr"`
	Unit     string               `xml:"unit,attr"`
	Datetime swissTime            `xml:"datetime"`
	Value    nulltype.NullFloat64 `xml:"value"`
}

func (sp *swissParameter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw struct {
		XMLName  xml.Name  `xml:"parameter"`
		Type     int       `xml:"type,attr"`
		Variant  int       `xml:"variant,attr"`
		Name     string    `xml:"name,attr"`
		Unit     string    `xml:"unit,attr"`
		Datetime swissTime `xml:"datetime"`
		Value    string    `xml:"value"`
	}
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}
	var val nulltype.NullFloat64
	if raw.Value != "NaN" {
		val.UnmarshalJSON([]byte(raw.Value)) //nolint:errcheck
	}
	sp.XMLName = raw.XMLName
	sp.Type = raw.Type
	sp.Variant = raw.Variant
	sp.Name = raw.Name
	sp.Unit = raw.Unit
	sp.Datetime = raw.Datetime
	sp.Value = val
	return nil
}
