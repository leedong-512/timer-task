package scheduler

import (
	"log"
	"strings"
	"sync"
	"task/v1/extcron"
	"task/v1/models"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	Cron        *cron.Cron
	Started     bool
	EntryJobMap sync.Map
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		// Cron:        cron.New(cron.WithSeconds()),
		Cron:        cron.New(cron.WithParser(extcron.NewParser())),
		Started:     false,
		EntryJobMap: sync.Map{},
	}
}

func (s *Scheduler) Start(jobs []*models.Job) error {
	s.Cron = cron.New(cron.WithParser(extcron.NewParser()))
	for _, job := range jobs {
		s.AddJob(job)
	}

	s.Cron.Start()
	s.Started = true
	return nil
}

func (s *Scheduler) Stop() {
	if s.Started {
		s.Cron.Stop()
		s.Started = false
	}
}

func (s *Scheduler) AddJob(job *models.Job) error {
	if _, ok := s.EntryJobMap.Load(job.Name); ok {
		s.RemoveJob(job.Name)
	}
	if job.Disabled {
		return nil
	}

	schedule := job.Schedule
	if job.Timezone != "" &&
		!strings.HasPrefix(schedule, "@") &&
		!strings.HasPrefix(schedule, "TZ=") &&
		!strings.HasPrefix(schedule, "CRON_TZ=") {
		schedule = "CRON_TZ=" + job.Timezone + " " + schedule
	}
	entryId, err := s.Cron.AddJob(schedule, job)
	if err != nil {
		log.Fatal(err)
	}

	s.EntryJobMap.Store(job.Name, entryId)

	/*store := models.NewStore()
	if err := store.SetJob(job); err != nil {
		return err
	}*/
	return nil
}

func (s *Scheduler) UpdateJob(job *models.Job) error {
	if entryId, ok := s.EntryJobMap.Load(job.Name); ok {
		s.Cron.Remove(entryId.(cron.EntryID))
		s.EntryJobMap.Delete(job.Name)
	}

	entryId, err := s.Cron.AddJob(job.Schedule, job)
	if err != nil {
		log.Fatal(err)
	}

	s.EntryJobMap.Store(job.Name, entryId)
	return nil
}

func (s *Scheduler) RemoveJob(jobName string) error {
	if entryId, ok := s.EntryJobMap.Load(jobName); ok {
		s.Cron.Remove(entryId.(cron.EntryID))
		s.EntryJobMap.Delete(jobName)

		store := models.NewStore()
		return store.DeleteJob(jobName)
	}
	return nil
}

func (s *Scheduler) GetJob(jobName string) (*models.Job, error) {
	store := models.NewStore()
	return store.GetJob(jobName)
}

func (s *Scheduler) GetJobs() ([]*models.Job, error) {
	store := models.NewStore()
	return store.GetJobs()
}

func (s *Scheduler) SingleRunJob(jobName string) error {
	store := models.NewStore()
	job, err := store.GetJob(jobName)
	if err != nil {
		return err
	}
	go job.Run()

	return nil
}

func (s *Scheduler) GetJobExecutions(jobName string) ([]*models.JobExecution, error) {
	store := models.NewStore()
	return store.GetJobExecutions(jobName)
}
