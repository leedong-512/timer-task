package models

import (
	"fmt"
	"regexp"
	"task/v1/exector"
	"task/v1/extcron"
	"time"
)

type Job struct {
	Name          string
	Schedule      string
	Displayname   string
	Owner         string
	Exector       string
	Exectorconfig map[string]interface{}
	Disabled      bool
	Timezone      string
}

func NewJob() *Job { return &Job{} }

func (j *Job) All() []*Job {
	//var jobs []*Job
	//jobs = append(jobs, &Job{
	//	Name:          "job1",
	//	Schedule:      "@every 5h",
	//	Exector:       "shell",
	//	Exectorconfig: "php -v",
	//})
	//jobs = append(jobs, &Job{
	//	Name:          "job2",
	//	Schedule:      "@every 5h",
	//	Exector:       "shell",
	//	Exectorconfig: "go env",
	//})
	//jobs = append(jobs, &Job{
	//	Name:          "job3",
	//	Schedule:      "@every 5h",
	//	Exector:       "websocket",
	//	Exectorconfig: "php -v",
	//})

	jobs, err := NewStore().GetJobs()
	if err != nil {
		return nil
	}

	return jobs
}

func (j *Job) Run() {
	runningJobs.Store(j.Name, true)
	switch j.Exector {
	case "shell":
		//fmt.Println(j.Exectorconfig["command"].(string))
		sched := j.Schedule
		fmt.Println(sched)
		e := exector.NewExectorShell(j.Exectorconfig["command"].(string))
		e.Execute()
	case "http":
		//fmt.Println(reflect.TypeOf(j.Exectorconfig["data"]))
		fmt.Println(j.Exectorconfig["data"].(map[string]interface{})["Name"])
		e := exector.NewExectorHttp(j.Exectorconfig["url"].(string), j.Exectorconfig["method"].(string), j.Exectorconfig["data"].(map[string]interface{}))
		e.Execute()
	default:
		fmt.Println("定时任务：", j.Name, "执行失败", "暂时只支持shell，http两种执行器")
		return
	}
	runningJobs.Delete(j.Name)
}

func (j *Job) Validate() error {
	if j.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if valid, chr := isSlug(j.Name); !valid {
		return fmt.Errorf("name contains illegal character '%s'", chr)
	}
	// Validate schedule, allow empty schedule if parent job set.
	if j.Schedule != "" {
		if _, err := extcron.Parse(j.Schedule); err != nil {
			return fmt.Errorf("%s: %s", "err", err)
		}
	}

	/*if j.Concurrency != ConcurrencyAllow && j.Concurrency != ConcurrencyForbid && j.Concurrency != "" {
		return ErrWrongConcurrency
	}*/

	// An empty string is a valid timezone for LoadLocation
	if _, err := time.LoadLocation(j.Timezone); err != nil {
		return err
	}

	return nil
}

// isSlug determines whether the given string is a proper value to be used as
// key in the backend store (a "slug"). If false, the 2nd return value
// will contain the first illegal character found.
func isSlug(candidate string) (bool, string) {
	// Allow only lower case letters (unicode), digits, underscore and dash.
	illegalCharPattern, _ := regexp.Compile(`[^\p{Ll}0-9_-]`)
	whyNot := illegalCharPattern.FindString(candidate)
	return whyNot == "", whyNot
}

/*func (j *Job) runJob()  {
	store := NewStore()
	jobs, _ := store.GetJobs()
	for {

	}
}*/
