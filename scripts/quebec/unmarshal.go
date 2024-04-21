package quebec

import "github.com/mattn/go-nulltype"

type qJson struct {
	Diffusion []diffusion `json:"diffusion"`
}

type diffusion struct {
	DateDonnee  string               `json:"dateDonnee"`
	HeureDonnee string               `json:"heureDonnee"`
	TypeDonnee  string               `json:"typeDonnee"`
	Donnee      nulltype.NullFloat64 `json:"donnee"`
	// TypeDebit   string               `json:"typeDebit"`
}
