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
	"github.com/whitewater-guide/gorge/scripts/quebec"
	"github.com/whitewater-guide/gorge/scripts/riverzone"
	"github.com/whitewater-guide/gorge/scripts/russia1"
	"github.com/whitewater-guide/gorge/scripts/sepa"
	"github.com/whitewater-guide/gorge/scripts/switzerland"
	"github.com/whitewater-guide/gorge/scripts/testscripts"
	"github.com/whitewater-guide/gorge/scripts/tirol"
	"github.com/whitewater-guide/gorge/scripts/ukraine"
)

// Registry is used both by server and cli
// All the scripts must be registered here
var Registry = core.NewRegistry()

func init() {
	Registry.Register(testscripts.AllAtOnce)
	Registry.Register(testscripts.OneByOne)
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
	Registry.Register(ukraine.Descriptor)
	Registry.Register(norway.Descriptor)
	Registry.Register(quebec.Descriptor)
	Registry.Register(riverzone.Descriptor)
	Registry.Register(russia1.Descriptor)
	Registry.Register(sepa.Descriptor)
	Registry.Register(switzerland.Descriptor)
	Registry.Register(tirol.Descriptor)
}
