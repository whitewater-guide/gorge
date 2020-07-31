package quebec

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "quebec",
	Description: "Québec: Ministère de l'Environnement et de la Lutte contre les changements climatiques",
	Mode:        core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsQuebec{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsQuebec); ok {
			return &scriptQuebec{
				name:              name,
				codesURL:          "https://www.cehq.gouv.qc.ca/suivihydro/default.asp#region",
				referenceListURL:  "https://wateroffice.ec.gc.ca/station_metadata/reference_index_download_e.html",
				stationURLFormat:  "http://www.cehq.gouv.qc.ca/suivihydro/graphique.asp?NoStation=%s",
				readingsURLFormat: "http://www.cehq.gouv.qc.ca/suivihydro/fichier_donnees.asp?NoStation=%s",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsQuebec{})
	},
}
