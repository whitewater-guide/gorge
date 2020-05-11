package finland

import (
	"time"
)

type fTime struct {
	time.Time
}

var finTz, _ = time.LoadLocation("Europe/Helsinki")

func (ft *fTime) UnmarshalJSON(b []byte) (err error) {
	t, err := time.ParseInLocation(`"2006-01-02T15:04:05"`, string(b), finTz)
	ft.Time = t.UTC()
	return
}

type stationsList struct {
	Value []station `json:"value"`
	Next  string    `json:"odata.nextLink"`
}

type station struct {
	Lat       string `json:"KoordLat"`
	Lng       string `json:"KoordLong"`
	KuntaNimi string `json:"KuntaNimi"`
	Nro       string `json:"Nro"`
	PaikkaID  int    `json:"Paikka_Id"`
	Nimi      string `json:"Nimi"`
	SuureID   int    `json:"Suure_Id"`
}

type virtaamaList struct {
	Value []virtaama `json:"value"`
}

type virtaama struct {
	Aika fTime  `json:"Aika"`
	Arvo string `json:"Arvo"`
}
