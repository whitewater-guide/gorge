package quebec2

import (
	"encoding/json"
	"time"

	"github.com/mattn/go-nulltype"
)

type q2sites struct {
	Sites []q2site `json:"Site"`
}

type q2stations struct {
	Stations []q2site `json:"Station"`
}

type q2site struct {
	CodeRegionQC string          `json:"CodeRegionQC"`
	Composition  []q2composition `json:"Composition"`
	RegionQC     string          `json:"RegionQC"`
	DateDebut    string          `json:"date debut"`
	DateFin      interface{}     `json:"date fin"`
	Identifiant  string          `json:"identifiant"`
	Nom          string          `json:"nom"`
	Xcoord       string          `json:"xcoord"`
	Ycoord       string          `json:"ycoord"`
	Zcoord       interface{}     `json:"zcoord"`
}

var pasTemps = map[string]string{
	// Sites only
	"Journalier": "j",
	// Both sites and stations
	"Horaire": "h",
}

var typePointDonnee = map[string]string{
	// For sites
	"Débit turbiné": "turbine",
	"Apport filtré": "filtre",
	"Débit total":   "total",
	"Débit déversé": "deverse",
	// For stations
	"Niveau": "niveau",
	"Débit":  "debit",
}

type q2composition struct {
	Donnees q2donnees `json:"Donnees"`
	// Station element: Anémomètre:48 Doppler:72 GMON:98 Girouette:48 Hygromètre:97 Limnimètre:290 Précipitomètre:94 SR50 au GMON:98 Thermomètre:223
	Element        string `json:"element"`
	NomUniteMesure string `json:"nom_unite_mesure"`
	// pasTemps: Journalier, Horaire,
	PasTemps   string `json:"pas_temps"`
	TypeMesure string `json:"type_mesure"`
	// Site typePointDonnee: Débit turbiné, Apport filtré, Débit total, Débit déversé,
	// Station typePointDonnee: Direction du vent.10 mètres:48 Débit:34 Humidité relative.2 mètres:97 Niveau:328 Précipitation:94 Température Maximum:112 Température Minimum:111 Vitesse du vent.10 mètres:48 Épaisseur de neige:98 Équivalent en eau de la neige:98
	TypePointDonnee string `json:"type_point_donnee"`
}

type q2donnees struct {
	values map[time.Time]nulltype.NullFloat64
}

func (m *q2donnees) UnmarshalJSON(b []byte) error {
	var raw map[string]string
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	m.values = make(map[time.Time]nulltype.NullFloat64, len(raw))
	for ts, val := range raw {
		t, err := time.Parse("2006/01/02 15:04:05Z", ts)
		if err != nil {
			return err
		}
		var v nulltype.NullFloat64
		err = v.UnmarshalJSON([]byte(val))
		if err != nil {
			return err
		}
		m.values[t] = v
	}
	return nil
}
