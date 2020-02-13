package catalunya

import (
	"time"
)

type cTime struct {
	time.Time
}

var timezone, _ = time.LoadLocation("CET")

func (ct *cTime) UnmarshalJSON(b []byte) (err error) {
	t, err := time.ParseInLocation(`"02/01/2006T15:04:05"`, string(b), timezone)
	ct.Time = t.UTC()
	return
}

type additionalInfo struct {
	TempsMostreigMin string `json:"Temps mostreig (min)"`
	RangMNim         string `json:"Rang mínim"`
	RangMXim         string `json:"Rang màxim"`
}

type componentAdditionalInfo struct {
	Comarca                string `json:"Comarca"`
	Provincia              string `json:"Província"`
	Riu                    string `json:"Riu"`
	DistricteFluvial       string `json:"Districte fluvial"`
	Subconca               string `json:"Subconca"`
	TermeMunicipal         string `json:"Terme municipal"`
	SuperficieConcaDrenada string `json:"Superfície conca drenada"`
	Conca                  string `json:"Conca"`
}

type componentTechnicalDetails struct {
	Producer     string `json:"producer"`
	Model        string `json:"model"`
	SerialNumber string `json:"serialNumber"`
	MacAddress   string `json:"macAddress"`
	Energy       string `json:"energy"`
	Connectivity string `json:"connectivity"`
}

type sensor struct {
	Sensor                    string                    `json:"sensor"`
	Description               string                    `json:"description"`
	DataType                  string                    `json:"dataType"`
	Location                  string                    `json:"location"`
	Type                      string                    `json:"type"`
	Unit                      string                    `json:"unit"`
	TimeZone                  string                    `json:"timeZone"`
	PublicAccess              bool                      `json:"publicAccess"`
	Component                 string                    `json:"component"`
	ComponentType             string                    `json:"componentType"`
	ComponentDesc             string                    `json:"componentDesc"`
	ComponentPublicAccess     bool                      `json:"componentPublicAccess"`
	AdditionalInfo            additionalInfo            `json:"additionalInfo"`
	ComponentAdditionalInfo   componentAdditionalInfo   `json:"componentAdditionalInfo"`
	ComponentTechnicalDetails componentTechnicalDetails `json:"componentTechnicalDetails"`
}

type provider struct {
	Provider   string   `json:"provider"`
	Permission string   `json:"permission"`
	Sensors    []sensor `json:"sensors"`
}

type observation struct {
	Value     float64 `json:"value,string"`
	Timestamp cTime   `json:"timestamp"`
	Location  string  `json:"location"`
}

type dataSensor struct {
	Sensor       string        `json:"sensor"`
	Observations []observation `json:"observations"`
}

type catalunyaList struct {
	Providers []provider `json:"providers"`
}

type catalunyaData struct {
	Sensors []dataSensor `json:"sensors"`
}
