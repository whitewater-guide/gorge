package schedule

import (
	"github.com/whitewater-guide/gorge/core"
)

// DeleteJob implements core.JobScheduler interface
func (s *simpleScheduler) DeleteJob(jobID string) error {
	entries := s.Cron.Entries()
	removed := false
	for _, entry := range entries {
		job, ok := entry.Job.(*harvestJob)
		if ok && job.jobID == jobID {
			s.Cron.Remove(entry.ID)
			s.Logger.Debugf("deleted job  entry %s for job %s", job.cron, jobID)
			removed = true
		}
	}
	if removed {
		return nil
	}
	return (&core.Error{Msg: "specified job is not scheduled"}).With("job_id", jobID)
}
