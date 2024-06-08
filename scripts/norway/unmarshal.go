package norway

import (
	"github.com/mattn/go-nulltype"
)

type statiosResponse struct {
	// CurrentLink string    `json:"currentLink"`
	// APIVersion string    `json:"apiVersion"`
	// License    string    `json:"license"`
	// CreatedAt  time.Time `json:"createdAt"`
	// QueryTime  string    `json:"queryTime"`
	ItemCount int           `json:"itemCount"`
	Data      []stationData `json:"data"`
}

type stationData struct {
	StationID   string  `json:"stationId"`
	StationName string  `json:"stationName"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	// UtmEastZ33               int          `json:"utmEast_Z33"`
	// UtmNorthZ33              int          `json:"utmNorth_Z33"`
	Masl      int    `json:"masl"`
	RiverName string `json:"riverName"`
	// CouncilNumber            string       `json:"councilNumber"`
	// CouncilName              string       `json:"councilName"`
	// CountyName               string       `json:"countyName"`
	// DrainageBasinKey         int          `json:"drainageBasinKey"`
	// Hierarchy                string       `json:"hierarchy"`
	// LakeArea                 float64      `json:"lakeArea"`
	// LakeName                 string       `json:"lakeName"`
	// LakeNo                   int          `json:"lakeNo"`
	// RegineNo                 string       `json:"regineNo"`
	// ReservoirNo              string       `json:"reservoirNo"`
	// ReservoirName            string       `json:"reservoirName"`
	// StationTypeName          string       `json:"stationTypeName"`
	// StationStatusName        string       `json:"stationStatusName"`
	// DrainageBasinArea        float64      `json:"drainageBasinArea"`
	// DrainageBasinAreaNorway  float64      `json:"drainageBasinAreaNorway"`
	// Gradient1085             float64      `json:"gradient1085"`
	// GradientBasin            interface{}  `json:"gradientBasin"`
	// GradientRiver            float64      `json:"gradientRiver"`
	// HeightMinimum            int          `json:"heightMinimum"`
	// HeightHypso10            int          `json:"heightHypso10"`
	// HeightHypso20            int          `json:"heightHypso20"`
	// HeightHypso30            int          `json:"heightHypso30"`
	// HeightHypso40            int          `json:"heightHypso40"`
	// HeightHypso50            int          `json:"heightHypso50"`
	// HeightHypso60            int          `json:"heightHypso60"`
	// HeightHypso70            int          `json:"heightHypso70"`
	// HeightHypso80            int          `json:"heightHypso80"`
	// HeightHypso90            int          `json:"heightHypso90"`
	// HeightMaximum            int          `json:"heightMaximum"`
	// LengthKmBasin            float64      `json:"lengthKmBasin"`
	// LengthKmRiver            float64      `json:"lengthKmRiver"`
	// PercentAgricul           float64      `json:"percentAgricul"`
	// PercentBog               float64      `json:"percentBog"`
	// PercentEffBog            interface{}  `json:"percentEffBog"`
	// PercentEffLake           float64      `json:"percentEffLake"`
	// PercentForest            float64      `json:"percentForest"`
	// PercentGlacier           int          `json:"percentGlacier"`
	// PercentLake              float64      `json:"percentLake"`
	// PercentMountain          int          `json:"percentMountain"`
	// PercentUrban             float64      `json:"percentUrban"`
	// UtmZoneGravi             interface{}  `json:"utmZoneGravi"`
	// UtmEastGravi             interface{}  `json:"utmEastGravi"`
	// UtmNorthGravi            interface{}  `json:"utmNorthGravi"`
	// UtmZoneInlet             interface{}  `json:"utmZoneInlet"`
	// UtmEastInlet             interface{}  `json:"utmEastInlet"`
	// UtmNorthInlet            interface{}  `json:"utmNorthInlet"`
	// UtmZoneOutlet            interface{}  `json:"utmZoneOutlet"`
	// UtmEastOutlet            interface{}  `json:"utmEastOutlet"`
	// UtmNorthOutlet           interface{}  `json:"utmNorthOutlet"`
	// AnnualRunoff             float64      `json:"annualRunoff"`
	// SpecificDischarge        float64      `json:"specificDischarge"`
	// RegulationArea           float64      `json:"regulationArea"`
	// AreaReservoirs           float64      `json:"areaReservoirs"`
	// VolumeReservoirs         float64      `json:"volumeReservoirs"`
	// RegulationPartReservoirs float64      `json:"regulationPartReservoirs"`
	// TransferAreaIn           float64      `json:"transferAreaIn"`
	// TransferAreaOut          int          `json:"transferAreaOut"`
	// ReservoirAreaIn          float64      `json:"reservoirAreaIn"`
	// ReservoirAreaOut         float64      `json:"reservoirAreaOut"`
	// ReservoirVolumeIn        float64      `json:"reservoirVolumeIn"`
	// ReservoirVolumeOut       float64      `json:"reservoirVolumeOut"`
	// RemainingArea            int          `json:"remainingArea"`
	// NumberReservoirs         int          `json:"numberReservoirs"`
	// FirstYearRegulation      int          `json:"firstYearRegulation"`
	// CatchmentRegTypeName     string       `json:"catchmentRegTypeName"`
	// Owner                    string       `json:"owner"`
	// QNumberOfYears           int          `json:"qNumberOfYears"`
	// QStartYear               int          `json:"qStartYear"`
	// QEndYear                 int          `json:"qEndYear"`
	// Qm                       interface{}  `json:"qm"`
	// Q5                       interface{}  `json:"q5"`
	// Q10                      interface{}  `json:"q10"`
	// Q20                      interface{}  `json:"q20"`
	// Q50                      interface{}  `json:"q50"`
	// Hm                       float64      `json:"hm"`
	// H5                       float64      `json:"h5"`
	// H10                      float64      `json:"h10"`
	// H20                      float64      `json:"h20"`
	// H50                      float64      `json:"h50"`
	// CulQm                    interface{}  `json:"culQm"`
	// CulQ5                    interface{}  `json:"culQ5"`
	// CulQ10                   interface{}  `json:"culQ10"`
	// CulQ20                   interface{}  `json:"culQ20"`
	// CulQ50                   interface{}  `json:"culQ50"`
	// CulHm                    float64      `json:"culHm"`
	// CulH5                    float64      `json:"culH5"`
	// CulH10                   float64      `json:"culH10"`
	// CulH20                   float64      `json:"culH20"`
	// CulH50                   float64      `json:"culH50"`
	SeriesList []seriesList `json:"seriesList"`
}

type resolutionList struct {
	// ResTime      int       `json:"resTime"`
	Method string `json:"method"`
	// TimeOffset   int       `json:"timeOffset"`
	// DataFromTime time.Time `json:"dataFromTime"`
	// DataToTime   time.Time `json:"dataToTime"`
}

type seriesList struct {
	ParameterName string `json:"parameterName"`
	Parameter     int    `json:"parameter"`
	// VersionNo      int              `json:"versionNo"`
	Unit string `json:"unit"`
	// SerieFrom      string           `json:"serieFrom"`
	// SerieTo        interface{}      `json:"serieTo"`
	ResolutionList []resolutionList `json:"resolutionList"`
}

type observationsResp struct {
	// CurrentLink string        `json:"currentLink"`
	// APIVersion  string        `json:"apiVersion"`
	// License     string        `json:"license"`
	// CreatedAt   time.Time     `json:"createdAt"`
	// QueryTime   string        `json:"queryTime"`
	// ItemCount   int           `json:"itemCount"`
	Data []observationsList `json:"data"`
}

type observationsList struct {
	StationID string `json:"stationId"`
	// StationName      string        `json:"stationName"`
	Parameter int `json:"parameter"`
	// ParameterName    string        `json:"parameterName"`
	// ParameterNameEng string        `json:"parameterNameEng"`
	// SerieVersionNo   int           `json:"serieVersionNo"`
	// Method           string        `json:"method"`
	// Unit             string        `json:"unit"`
	// ObservationCount int           `json:"observationCount"`
	Observations []observation `json:"observations"`
}

type observation struct {
	Time  string               `json:"time"`
	Value nulltype.NullFloat64 `json:"value"`
	// Correction int `json:"correction"`
	// Quality int `json:"quality"`
}
