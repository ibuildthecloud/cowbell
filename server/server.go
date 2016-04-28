package server

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rancher/cowbell/compose"
	"github.com/rancher/cowbell/job"
)

type jobSorter []*job.Job

func (js jobSorter) Len() int {
	return len(js)
}

func (js jobSorter) Swap(i, j int) {
	js[i], js[j] = js[j], js[i]
}

func (js jobSorter) Less(i, j int) bool {
	return js[i].ID < js[j].ID
}

type Server struct {
	jm *job.Manager
	c  *compose.Compose
}

func NewServer(jm *job.Manager, c *compose.Compose) *Server {
	return &Server{
		c:  c,
		jm: jm,
	}
}

func (s *Server) Info(rw http.ResponseWriter, req *http.Request) error {
	return infoTemplate.Execute(rw, nil)
}

func (s *Server) ListJobs(rw http.ResponseWriter, req *http.Request) error {
	jobs := s.jm.ListJobs()
	sort.Sort(jobSorter(jobs))
	return listTemplate.Execute(rw, map[string]interface{}{
		"jobs": jobs,
	})
}

func (s *Server) GetJobOutput(rw http.ResponseWriter, req *http.Request) error {
	id := mux.Vars(req)["id"]
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	job := s.jm.GetJob(int32(idNum))
	if job != nil {
		rw.Write(job.GetOutput())
	}

	return nil
}

func (s *Server) Scale(rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	return s.c.Scale(vars["stack"], vars["service"], vars["scale"])
}

func (s *Server) Upgrade(rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	return s.c.Upgrade(vars["stack"], vars["service"],
		req.URL.Query().Get("dockerComposeUrl"),
		req.URL.Query().Get("rancherComposeUrl"))
}

func (s *Server) Redeploy(rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	return s.c.Redeploy(vars["stack"], vars["service"])
}
