package scripts

import (
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/scripts/canada"
	"github.com/whitewater-guide/gorge/scripts/cantabria"
	"github.com/whitewater-guide/gorge/scripts/catalunya"
	"github.com/whitewater-guide/gorge/scripts/chile"
	"github.com/whitewater-guide/gorge/scripts/ecuador"
	"github.com/whitewater-guide/gorge/scripts/finland"
	"github.com/whitewater-guide/gorge/scripts/galicia"
	"github.com/whitewater-guide/gorge/scripts/galicia2"
	"github.com/whitewater-guide/gorge/scripts/georgia"
	"github.com/whitewater-guide/gorge/scripts/norway"
	"github.com/whitewater-guide/gorge/scripts/nzbop"
	"github.com/whitewater-guide/gorge/scripts/nzcan"
	"github.com/whitewater-guide/gorge/scripts/nzmbh"
	"github.com/whitewater-guide/gorge/scripts/nzniwa"
	"github.com/whitewater-guide/gorge/scripts/nzstl"
	"github.com/whitewater-guide/gorge/scripts/nztrc"
	"github.com/whitewater-guide/gorge/scripts/nzwgn"
	"github.com/whitewater-guide/gorge/scripts/nzwko"
	"github.com/whitewater-guide/gorge/scripts/quebec"
	"github.com/whitewater-guide/gorge/scripts/riverzone"
	"github.com/whitewater-guide/gorge/scripts/russia1"
	"github.com/whitewater-guide/gorge/scripts/sepa"
	"github.com/whitewater-guide/gorge/scripts/switzerland"
	"github.com/whitewater-guide/gorge/scripts/testscripts"
	"github.com/whitewater-guide/gorge/scripts/tirol"
	"github.com/whitewater-guide/gorge/scripts/ukea"
	"github.com/whitewater-guide/gorge/scripts/ukraine"
	"github.com/whitewater-guide/gorge/scripts/usgs"
	"github.com/whitewater-guide/gorge/scripts/wales"
)

// Registry is used both by server and cli
// All the scripts must be registered here
var Registry = core.NewRegistry()

func init() {
	Registry.Register(testscripts.AllAtOnce)
	Registry.Register(testscripts.OneByOne)
	Registry.Register(testscripts.Batched)
	Registry.Register(testscripts.Broken)
	// Please sort scripts alphabetically
	Registry.Register(canada.Descriptor)
	Registry.Register(cantabria.Descriptor)
	Registry.Register(catalunya.Descriptor)
	Registry.Register(chile.Descriptor)
	Registry.Register(ecuador.Descriptor)
	Registry.Register(finland.Descriptor)
	Registry.Register(galicia.Descriptor)
	Registry.Register(galicia2.Descriptor)
	Registry.Register(georgia.Descriptor)
	Registry.Register(norway.Descriptor)
	Registry.Register(nzcan.Descriptor)
	Registry.Register(nzbop.Descriptor)
	Registry.Register(nzmbh.Descriptor)
	Registry.Register(nzniwa.Descriptor)
	Registry.Register(nzstl.Descriptor)
	Registry.Register(nztrc.Descriptor)
	Registry.Register(nzwgn.Descriptor)
	Registry.Register(nzwko.Descriptor)
	Registry.Register(quebec.Descriptor)
	Registry.Register(riverzone.Descriptor)
	Registry.Register(russia1.Descriptor)
	Registry.Register(sepa.Descriptor)
	Registry.Register(switzerland.Descriptor)
	Registry.Register(tirol.Descriptor)
	Registry.Register(ukea.Descriptor)
	Registry.Register(ukraine.Descriptor)
	Registry.Register(usgs.Descriptor)
	Registry.Register(wales.Descriptor)
}
