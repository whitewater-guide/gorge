package galicia

type entries struct {
	ListEstadoActual []entry `json:"listEstadoActual"`
}

type entry struct {
	Concello string `json:"concello"`
	// DataLocal  string  `json:"dataLocal"`
	DataUTC    string  `json:"dataUTC"`
	Estacion   string  `json:"estacion"`
	IDEstacion int     `json:"idEstacion"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Prov       string  `json:"prov"`
	// Provincia  string  `json:"provincia"`
	// Utmx        string  `json:"utmx"`
	// Utmy        string  `json:"utmy"`
	ValorCaudal float64 `json:"valorCaudal"`
	ValorNivel  float64 `json:"valorNivel"`
}
