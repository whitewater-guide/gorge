package norway

import (
	"fmt"
	"strings"
)

var oldIds = map[string]struct{}{
	"100.1":   {},
	"103.1":   {},
	"103.19":  {},
	"103.20":  {},
	"103.3":   {},
	"103.40":  {},
	"109.20":  {},
	"109.21":  {},
	"109.42":  {},
	"109.9":   {},
	"111.5":   {},
	"12.171":  {},
	"12.178":  {},
	"12.193":  {},
	"12.207":  {},
	"12.209":  {},
	"12.215":  {},
	"12.70":   {},
	"12.76":   {},
	"12.91":   {},
	"122.11":  {},
	"122.14":  {},
	"122.17":  {},
	"122.9":   {},
	"123.31":  {},
	"124.2":   {},
	"127.13":  {},
	"127.6":   {},
	"139.13":  {},
	"15.21":   {},
	"15.61":   {},
	"15.79":   {},
	"151.15":  {},
	"151.21":  {},
	"151.28":  {},
	"156.19":  {},
	"16.10":   {},
	"16.122":  {},
	"16.128":  {},
	"16.155":  {},
	"16.193":  {},
	"16.51":   {},
	"163.5":   {},
	"163.7":   {},
	"19.73":   {},
	"2.11":    {},
	"2.1151":  {},
	"2.129":   {},
	"2.25":    {},
	"2.265":   {},
	"2.267":   {},
	"2.268":   {},
	"2.275":   {},
	"2.28":    {},
	"2.280":   {},
	"2.284":   {},
	"2.290":   {},
	"2.291":   {},
	"2.303":   {},
	"2.32":    {},
	"2.439":   {},
	"2.461":   {},
	"2.479":   {},
	"2.481":   {},
	"2.578":   {},
	"2.595":   {},
	"2.605":   {},
	"2.614":   {},
	"213.4":   {},
	"22.4":    {},
	"25.32":   {},
	"27.24":   {},
	"27.25":   {},
	"311.4":   {},
	"311.460": {},
	"62.10":   {},
	"62.15":   {},
	"62.17":   {},
	"62.18":   {},
	"71.1":    {},
	"72.77":   {},
	"76.10":   {},
	"76.5":    {},
	"77.3":    {},
	"78.12":   {},
	"8.2":     {},
	"83.2":    {},
	"84.15":   {},
	"86.10":   {},
	"87.10":   {},
	"89.1":    {},
	"98.4":    {},
}

// getOurStationId checks converts gauge id to old format (with one dot) if this id was harvested by old version of script
func getOurStationId(supportLegacy bool, id string) string {
	if !supportLegacy {
		return id
	}
	parts := strings.Split(id, ".")
	if len(parts) == 3 && parts[2] == "0" {
		ourId := fmt.Sprintf("%s.%s", parts[0], parts[1])
		if _, ok := oldIds[ourId]; ok {
			return ourId
		}
	}
	return id
}

// getTheirStationId returns nve.no station id in x.x.x format
func getTheirStationId(supportLegacy bool, id string) string {
	if !supportLegacy {
		return id
	}
	parts := strings.Split(id, ".")
	if len(parts) == 2 {
		return id + ".0"
	}
	return id
}

func getTheirStationIds(supportLegacy bool, ids []string) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = getTheirStationId(supportLegacy, id)
	}
	return result
}
