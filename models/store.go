package models

import (
	"encoding/json"
	"fmt"
	"sync"
	"task/v1/conn"

	"github.com/tidwall/buntdb"
)

var runningJobs sync.Map

const (
	JobsPrefix       = "jobs"
	ExecutionsPrefix = "executions"
)

type Store struct {
	db *buntdb.DB
}

func NewStore() *Store {
	db := conn.BuntDb
	db.CreateIndex("name", JobsPrefix+"*", buntdb.IndexJSON("Name"))
	db.CreateIndex("execution", ExecutionsPrefix+"*", buntdb.IndexJSON("JobName"))

	return &Store{
		db: db,
	}
}

func (s *Store) GetJob(jobName string) (*Job, error) {
	var job Job
	err := s.db.View(func(tx *buntdb.Tx) error {
		key := fmt.Sprintf("%s:%s", JobsPrefix, jobName)
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(val), &job); err != nil {
			return err
		}
		return nil
	})
	return &job, err
}

func (s *Store) GetJobs() ([]*Job, error) {
	var job Job
	jobs := make([]*Job, 0)
	s.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("name", func(key, value string) bool {
			if err := json.Unmarshal([]byte(value), &job); err != nil {
				return false
			}

			jobs = append(jobs, &job)
			return true
		})
		return err
	})
	return jobs, nil
}

func (s *Store) SetJob(job *Job) error {
	err := s.db.Update(func(tx *buntdb.Tx) error {
		item, err := json.Marshal(job)
		if err != nil {
			return err
		}
		if _, _, err := tx.Set(fmt.Sprintf("%s:%s", JobsPrefix, job.Name), string(item), nil); err != nil {
			return err
		}
		return nil
	})

	return err
}

func (S *Store) UpdateJob() error {
	return nil
}

func (s *Store) DeleteJob(jobName string) error {
	err := s.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(fmt.Sprintf("%s:%s", JobsPrefix, jobName))
		return err
	})
	return err
}

func (s *Store) AddJobExecution(je *JobExecution) error {
	err := s.db.Update(func(tx *buntdb.Tx) error {
		item, err := json.Marshal(je)
		if err != nil {
			return err
		}
		if _, _, err := tx.Set(fmt.Sprintf("%s:%s", ExecutionsPrefix, je.JobName), string(item), nil); err != nil {
			return err
		}
		return nil
	})

	return err
}

func (s *Store) GetJobExecutions(jobName string) ([]*JobExecution, error) {
	var jobExecution JobExecution
	jobExecutions := make([]*JobExecution, 0)
	err := s.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("execution", func(key, value string) bool {
			if err := json.Unmarshal([]byte(value), &jobExecution); err != nil {
				return false
			}

			jobExecutions = append(jobExecutions, &jobExecution)
			return true
		})
		return err
	})

	return jobExecutions, err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func GetRunningJobs() int {
	sum := 0
	runningJobs.Range(func(key, value interface{}) bool {
		sum = sum + 1
		return true
	})

	return sum
}
