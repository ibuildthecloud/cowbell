package compose

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/cowbell/job"
	"github.com/rancher/go-rancher/client"
)

type Compose struct {
	c  *Client
	jm *job.Manager
}

func NewCompose(url, accessKey, secretKey string, jm *job.Manager) (*Compose, error) {
	rancherClient, err := client.NewRancherClient(&client.ClientOpts{
		Url:       url,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if err != nil {
		return nil, err
	}

	return &Compose{
		c: &Client{
			c: rancherClient,
		},
		jm: jm,
	}, nil
}

func (c *Compose) Scale(stack, service, scale string) error {
	job, err := c.newJob(stack, "", "")
	if err != nil {
		return err
	}

	newScale, err := c.c.getCurrentScale(stack, service)
	if err != nil {
		return job.SetErr(err)
	}

	if strings.EqualFold(scale, "up") {
		newScale++
	} else if strings.EqualFold(scale, "down") {
		newScale--
		if newScale < 1 {
			newScale = 1
		}
	} else {
		n, err := strconv.Atoi(scale)
		if err != nil {
			return job.SetErr(err)
		}
		newScale = int64(n)
	}

	return job.Run("-p", stack, "scale", fmt.Sprintf("%s=%d", service, newScale))
}

func (c *Compose) Upgrade(stack, service, dockerComposeURL, rancherComposeURL string) error {
	job, err := c.newJob(stack, dockerComposeURL, rancherComposeURL)
	if err != nil {
		return err
	}

	args := []string{"-p", stack, "up", "-d", "-u", "-c", "-p"}
	if service != "" {
		args = append(args, service)
	}
	return job.Run(args...)
}

func (c *Compose) Redeploy(stack, service string) error {
	job, err := c.newJob(stack, "", "")
	if err != nil {
		return err
	}

	args := []string{"-p", stack, "up", "-d", "--force-upgrade", "-c", "-p"}
	if service != "" {
		args = append(args, service)
	}
	return job.Run(args...)
}

func (c *Compose) newJob(stack, dockerComposeURL, rancherComposeURL string) (*job.Job, error) {
	job, err := c.jm.NewJob()

	if dockerComposeURL == "" {
		job.DockerCompose, job.RancherCompose, err = c.c.getTemplates(stack)
	} else {
		job.DockerCompose, job.RancherCompose, err = c.c.downloadURL(dockerComposeURL, rancherComposeURL)
	}
	if err != nil {
		return nil, job.SetErr(err)
	}
	return job, nil
}
