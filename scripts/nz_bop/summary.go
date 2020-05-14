package nz_bop

type summary struct {
	name       string
	gridRef    string
	siteNumber int
}

// Some stations's pages have missing fields, mostly locations
// This data was manually extracted from http://monitoring.boprc.govt.nz/MonitoredSites/summary.pdf
// to fill the gaps
var summaries = map[string]summary{
	// {
	// 	name:       "Tuapiro at Woodlands Road",
	// 	gridRef:    "T13: 661 057",
	// 	siteNumber: 13310,
	// },
	"202": {
		name:       "Kopurereroa at S.H.29 Bridge",
		gridRef:    "U14: 843 805",
		siteNumber: 14302,
	},
	"184": {
		name:       "Ngongotaha at S.H.5 Bridge ",
		gridRef:    "U15: 910 414",
		siteNumber: 1014641,
	},
}
