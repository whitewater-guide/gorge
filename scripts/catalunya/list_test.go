package catalunya

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestCatalunya_Convert(t *testing.T) {
	s := scriptCatalunya{name: "catalunya"}
	assert := assert.New(t)
	input := &sensor{
		Sensor:                "431652-001-ANA001",
		Description:           "Nivell riu",
		DataType:              "NUMBER",
		Location:              "41.233966843 1.33729486",
		Type:                  "0019",
		Unit:                  "cm",
		TimeZone:              "CET",
		PublicAccess:          true,
		Component:             "431652-001",
		ComponentType:         "aforament",
		ComponentDesc:         "Vilabella",
		ComponentPublicAccess: true,
		AdditionalInfo: additionalInfo{
			TempsMostreigMin: "5",
			RangMNim:         "0",
			RangMXim:         "10",
		},
		ComponentAdditionalInfo: componentAdditionalInfo{
			Comarca:                "ALT CAMP",
			Provincia:              "TARRAGONA",
			Riu:                    "RIU GAIÀ",
			DistricteFluvial:       "ACA",
			SuperficieConcaDrenada: "329,73 km²",
			Subconca:               "EL GAIÀ",
			TermeMunicipal:         "VILABELLA",
			Conca:                  "EL GAIÀ",
		},
		ComponentTechnicalDetails: componentTechnicalDetails{
			Producer:     "",
			Model:        "",
			SerialNumber: "",
			MacAddress:   "",
			Energy:       "",
			Connectivity: "",
		},
	}
	actual, err := s.convert(input)
	if assert.NoError(err) {
		assert.Equal(core.Gauge{
			GaugeID: core.GaugeID{
				Script: "catalunya",
				Code:   "431652-001-ANA001",
			},
			Name:      "Riu Gaià - Vilabella (cm)",
			LevelUnit: "cm",
			URL:       "http://aca-web.gencat.cat/sentilo-catalog-web/component/AFORAMENT-EST.431652-001/detail",
			Location: &core.Location{
				Latitude:  41.23396,
				Longitude: 1.33729,
			},
		}, *actual)
	}
}
