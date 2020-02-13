package galicia

import (
	"time"

	"github.com/mattn/go-nulltype"
)

type gTime struct {
	time.Time
}

func (gt *gTime) UnmarshalJSON(b []byte) (err error) {
	t, err := time.Parse(`"2006-01-02T15:04:05"`, string(b))
	gt.Time = t.UTC()
	return
}

type medida struct {
	CodParametro int                  `json:"codParametro"`
	Unidade      string               `json:"unidade"`
	Valor        nulltype.NullFloat64 `json:"valor"`
}

type aforo struct {
	DataUTC      gTime    `json:"dataUTC"`
	Ide          int      `json:"ide"`
	Latitude     float64  `json:"latitude,string"`
	Lonxitude    float64  `json:"lonxitude,string"`
	ListaMedidas []medida `json:"listaMedidas"`
	NomeEstacion string   `json:"nomeEstacion"`
}

type galiciaData struct {
	ListaAforos []aforo `json:"listaAforos"`
}
