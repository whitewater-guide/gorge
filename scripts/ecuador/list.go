package ecuador

import (
	"regexp"

	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (s *scriptEcuador) parseList() (map[string]core.Gauge, error) {
	raw, err := core.Client.GetAsString(s.listURL1, nil)
	if err != nil {
		return nil, err
	}
	r := regexp.MustCompile(`\\"(.*)\\".*(H\d{4})"`)
	matches := r.FindAllStringSubmatch(string(raw), -1)
	result := make(map[string]core.Gauge)
	for _, m := range matches {
		name := cases.Title(language.Spanish).String(m[1])
		code := m[2]
		if code != "" {
			result[code] = core.Gauge{
				GaugeID: core.GaugeID{
					Code:   code,
					Script: s.name,
				},
				LevelUnit: "m",
				Name:      name,
				Timezone:  "America/Guayaquil",
			}
		}
	}

	return result, nil
}
