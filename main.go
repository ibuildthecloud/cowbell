package main

import (
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/cowbell/compose"
	"github.com/rancher/cowbell/job"
	"github.com/rancher/cowbell/server"
)

func main() {
	token := os.Getenv("TOKEN")
	listen := os.Getenv("LISTEN")

	url := os.Getenv("CATTLE_URL")
	accessKey := os.Getenv("CATTLE_ACCESS_KEY")
	secretKey := os.Getenv("CATTLE_SECRET_KEY")

	if listen == "" {
		listen = ":8080"
	}

	if token == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		token = strconv.FormatInt(r.Int63(), 10)
	}

	logrus.Infof("Listening on %s with token %s", listen, token)

	jm := job.NewJobManager()
	compose, err := compose.NewCompose(url, accessKey, secretKey, jm)
	if err != nil {
		logrus.Fatalf("Failed to construct client to %s with access key %s: %v", url, accessKey, err)
	}

	router := server.NewRouter(server.NewServer(jm, compose), token)
	if err := http.ListenAndServe(listen, router); err != nil {
		logrus.Fatal(err)
	}
}
