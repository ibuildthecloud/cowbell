package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func f(t func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := t(rw, req); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		}
	})
}

func NewRouter(s *Server, token string) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// API framework routes
	router.Methods("GET").Path("/").Handler(f(s.Info))
	router.Methods("GET").Path("/" + token + "/jobs").Handler(f(s.ListJobs))
	router.Methods("GET").Path("/" + token + "/jobs/{id}/output").Handler(f(s.GetJobOutput))

	router.Methods("POST").Path("/" + token + "/scale/{stack}/{service}/{scale}").Handler(f(s.Scale))
	router.Methods("POST").Path("/" + token + "/upgrade/{stack}").Handler(f(s.Upgrade))
	router.Methods("POST").Path("/" + token + "/upgrade/{stack}/{service}").Handler(f(s.Upgrade))
	router.Methods("POST").Path("/" + token + "/reploy/{stack}").Handler(f(s.Redeploy))
	router.Methods("POST").Path("/" + token + "/reploy/{stack}/{service}").Handler(f(s.Redeploy))

	return router
}
