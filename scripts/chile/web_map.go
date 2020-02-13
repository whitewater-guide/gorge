package chile

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

type spatialReference struct {
	WKID       int `json:"wkid"`
	LatestWKID int `json:"latestWkid"`
}

type geometry struct {
	X                float64          `json:"x"`
	Y                float64          `json:"y"`
	SpatialReference spatialReference `json:"spatialReference"`
}

type attributes struct {
	Name string `json:"NOMBRE"`
	//Name   string  `json:"NOM_SSUBC"`
	Code     string  `json:"COD_BNA"`
	Northing float64 `json:"NORTE_84"`
	Easting  float64 `json:"ESTE_84"`
}

type feature struct {
	Geometry   geometry   `json:"geometry"`
	Attributes attributes `json:"attributes"`
	// Added by us
	LayerName string
}

type featureSet struct {
	Features []feature `json:"features"`
}

type layerDefinition struct {
	Name string `json:"name"`
}

type layer struct {
	FeatureSet      featureSet      `json:"featureSet"`
	LayerDefinition layerDefinition `json:"layerDefinition"`
}

type featureCollection struct {
	Layers []layer `json:"layers"`
}

type operationalLayer struct {
	FeatureCollection featureCollection `json:"featureCollection"`
}

type webmapPage struct {
	OperationalLayers []operationalLayer `json:"operationalLayers"`
}

func (s *scriptChile) getWebmapID() (string, error) {
	type webmapIDPageValues struct {
		Webmap string `json:"webmap"`
	}
	type webmapIDPage struct {
		Values webmapIDPageValues `json:"values"`
	}
	response := &webmapIDPage{}
	err := core.Client.GetAsJSON(s.webmapIDPageURL, response, nil)

	if err != nil {
		return "", err
	}
	return response.Values.Webmap, nil
}

func (s *scriptChile) getWepmapURL() (string, error) {
	webmapID, err := s.getWebmapID()
	if err != nil {
		return "", err
	}
	result := fmt.Sprintf(s.webmapURLFormat, webmapID)
	return result, nil
}

func (s *scriptChile) parseWebmap() (map[string]feature, error) {
	url, err := s.getWepmapURL()
	if err != nil {
		return nil, err
	}
	response := &webmapPage{}
	err = core.Client.GetAsJSON(url, response, nil)
	if err != nil {
		return nil, err
	}

	result := make(map[string]feature)
	for _, opl := range response.OperationalLayers {
		for _, l := range opl.FeatureCollection.Layers {
			for _, f := range l.FeatureSet.Features {
				f.LayerName = l.LayerDefinition.Name
				result[f.Attributes.Code] = f
			}
		}
	}
	return result, nil
}
