package canada

import "strings"

func getProvinces(provinces string) (result map[string]bool) {
	if provinces == "" {
		provinces = "AB,BC,MB,NB,NL,NS,NT,NU,ON,PE,QC,SK,YT"
	}
	result = make(map[string]bool)
	codes := strings.Split(provinces, ",")
	for _, v := range codes {
		result[v] = true
	}
	return
}
