package schedule

import (
	"fmt"
	"math"
	"sort"

	"github.com/fatih/structs"
	"github.com/robfig/cron/v3"
	"github.com/whitewater-guide/gorge/core"
)

// AddJob implements JobScheduler interface
func (s *SimpleScheduler) AddJob(description core.JobDescription) error {
	mode, err := s.Registry.GetMode(description.Script)
	if err != nil {
		return err
	}
	nGauges := len(description.Gauges)
	if nGauges == 0 {
		return (&core.Error{Msg: "job gauge codes must be specified"}).With("description", structs.Map(description))
	}
	if mode == core.AllAtOnce {
		_, err := cron.ParseStandard(description.Cron)
		if err != nil {
			return core.WrapErr(err, "bad job cron").With("description", structs.Map(description))
		}
		options, err := s.Registry.ParseJSONOptions(description.Script, description.Options)
		if err != nil {
			return core.WrapErr(err, "failed to parse options").With("description", structs.Map(description))
		}
		_, err = s.Cron.AddJob(description.Cron, &harvestJob{
			database: s.Database,
			cache:    s.Cache,
			logger:   s.Logger,
			registry: s.Registry,
			cron:     description.Cron,
			jobID:    description.ID,
			script:   description.Script,
			codes:    core.GaugesCodes(description.Gauges),
			options:  options,
		})
		if err != nil {
			return core.WrapErr(err, "failed to schedule harvest job").With("description", structs.Map(description))
		}
	} else if mode == core.OneByOne {
		j, step := 0, 59.0/(float64(nGauges))
		// sort codes first
		codes := make([]string, nGauges)
		for k := range description.Gauges {
			codes[j] = k
			j++
		}
		sort.Strings(codes)
		// Ensure transactional behaviour when adding cronjobs for gauges:
		// if we cannot add one gauge job, cancel the whol batch
		var tErr error
		var entryIDs []cron.EntryID
		for i, code := range codes {
			gaugeOpts := description.Gauges[code]
			minute := int(math.Ceil(float64(float64(i) * step)))
			spec := fmt.Sprintf("%d * * * *", minute)
			options, err := s.Registry.ParseJSONOptions(description.Script, description.Options, gaugeOpts)
			if err != nil {
				tErr = core.WrapErr(err, "failed to parse options").With("description", structs.Map(description))
				break
			}
			eid, err := s.Cron.AddJob(spec, &harvestJob{
				database: s.Database,
				cache:    s.Cache,
				logger:   s.Logger,
				registry: s.Registry,
				cron:     spec,
				jobID:    description.ID,
				script:   description.Script,
				codes:    core.StringSet{code: {}},
				options:  options,
			})
			if err != nil {
				tErr = core.WrapErr(err, "failed to schedule harvest job").With("description", structs.Map(description))
			}
			entryIDs = append(entryIDs, eid)
		}
		if tErr != nil {
			// rollback already added jobs
			for _, eid := range entryIDs {
				s.Cron.Remove(eid)
			}
			return tErr
		}
	}
	return nil
}
