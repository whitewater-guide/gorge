package schedule

import (
	"github.com/whitewater-guide/gorge/core"
)

// ListNext implements JobScheduler interface
func (s *SimpleScheduler) ListNext(jobID string) map[string]core.HTime {
	result := make(map[string]core.HTime)
	for _, entry := range s.Cron.Entries() {
		job, ok := entry.Job.(*harvestJob)
		if !ok {
			continue
		}
		if jobID == "" {
			current, ok := result[job.jobID]
			if !ok || current.After(entry.Next) {
				result[job.jobID] = core.HTime{Time: entry.Next}
			}
		} else if len(job.codes) == 1 {
			code, _ := job.codes.Only()
			result[code] = core.HTime{Time: entry.Next}
		}
	}
	return result
}
