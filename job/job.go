package job

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
)

type Manager struct {
	sync.Mutex

	counter int32
	jobs    map[int32]*Job
}

const (
	Preparing = "preparing"
	Running   = "running"
	Errored   = "error"
	Done      = "done"
)

type Job struct {
	ID             int32
	args           []string
	Created        time.Time
	Dir            string
	Err            error
	State          string
	DockerCompose  string
	RancherCompose string
	output         *bytes.Buffer
}

func NewJobManager() *Manager {
	return &Manager{
		jobs: map[int32]*Job{},
	}
}

func (jm *Manager) NewJob() (*Job, error) {
	tempdir, err := ioutil.TempDir("", "compose-")
	if err != nil {
		return nil, err
	}

	id := atomic.AddInt32(&jm.counter, 1)
	job := &Job{
		ID:      id,
		Dir:     tempdir,
		State:   Preparing,
		Created: time.Now(),
	}

	jm.Lock()
	defer jm.Unlock()

	jm.jobs[id] = job
	return job, nil
}

func (jm *Manager) GetJob(id int32) *Job {
	return jm.jobs[id]
}

func (jm *Manager) ListJobs() []*Job {
	jm.Lock()
	defer jm.Unlock()

	result := []*Job{}
	for _, j := range jm.jobs {
		result = append(result, j)
	}

	return result
}

func (jm *Manager) cleanup() {
	for {
		time.Sleep(24 * time.Hour)
		jm.purgeOldJobs()
	}
}

func (jm *Manager) purgeOldJobs() {
	jm.Lock()
	defer jm.Unlock()

	dayAgo := time.Now().Add(-24 * time.Hour)
	for _, job := range jm.jobs {
		if job.Created.Before(dayAgo) {
			err := job.Delete()
			if err != nil {
				logrus.Errorf("Failed to delete job %d: %v", job.ID, err)
			}
			delete(jm.jobs, job.ID)
		}
	}
}

func (j *Job) SetErr(err error) error {
	j.State = Errored
	j.Err = err
	return err
}

func (j *Job) Run(args ...string) error {
	j.args = args
	cmd, err := j.doRun(args...)
	if err != nil {
		j.State = Errored
		j.Err = err
		return err
	}

	j.State = Running

	go func() {
		if j.Err = cmd.Wait(); j.Err == nil {
			j.State = Done
		} else {
			j.State = Errored
		}
	}()

	return nil
}

func (j *Job) doRun(args ...string) (*exec.Cmd, error) {
	err := ioutil.WriteFile(path.Join(j.Dir, "docker-compose.yml"), []byte(j.DockerCompose), 0644)
	if err != nil {
		return nil, err
	}

	if j.RancherCompose != "" {
		err := ioutil.WriteFile(path.Join(j.Dir, "rancher-compose.yml"), []byte(j.RancherCompose), 0644)
		if err != nil {
			return nil, err
		}
	}

	j.output = &bytes.Buffer{}
	tee := io.MultiWriter(j.output, os.Stdout)
	cmd := exec.Command("rancher-compose", args...)
	cmd.Dir = j.Dir
	cmd.Stdout = tee
	cmd.Stderr = tee
	logrus.Infof("Running in %s command %s %v", cmd.Dir, cmd.Path, cmd.Args)
	return cmd, cmd.Start()
}

func (j *Job) GetOutput() []byte {
	return j.output.Bytes()
}

func (j *Job) Delete() error {
	if _, err := os.Stat(j.Dir); err == nil {
		return os.RemoveAll(j.Dir)
	}
	return nil
}
