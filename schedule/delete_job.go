package schedule

import (
	"github.com/whitewater-guide/gorge/core"
)

// DeleteJob implements JobScheduler interface
func (s *SimpleScheduler) DeleteJob(jobID string) error {
	entries := s.Cron.Entries()
	removed := false
	for _, entry := range entries {
		job := entry.Job.(*harvestJob)
		if job.jobID == jobID {
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
