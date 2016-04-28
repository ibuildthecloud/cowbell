package compose

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rancher/go-rancher/client"
)

type Client struct {
	c *client.RancherClient
}

func (c *Client) downloadURL(composeURL, rancherComposeURL string) (string, string, error) {
	resp, err := http.Get(composeURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	compose, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	rancherCompose := []byte("{}")
	if rancherComposeURL != "" {
		resp, err := http.Get(rancherComposeURL)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()

		rancherCompose, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", "", err
		}
	}

	return string(compose), string(rancherCompose), nil
}

func (c *Client) getCurrentScale(stackName, serviceName string) (int64, error) {
	stack, err := c.getStack(stackName)
	if err != nil {
		return 0, err
	}

	service, err := c.getService(stack, serviceName)
	if err != nil {
		return 0, err
	}

	return service.Scale, nil
}

func (c *Client) getService(stack *client.Environment, service string) (*client.Service, error) {
	opts := client.NewListOpts()
	opts.Filters["name"] = service
	opts.Filters["removed_null"] = "true"
	opts.Filters["environmentId"] = stack.Id

	services, err := c.c.Service.List(opts)
	if err != nil {
		return nil, err
	}

	if len(services.Data) == 0 {
		return nil, fmt.Errorf("Failed to find stack %s", service)
	}

	if len(services.Data) > 1 {
		return nil, fmt.Errorf("Found %d services named %s", len(services.Data), service)
	}

	return &services.Data[0], nil
}

func (c *Client) getStack(stackName string) (*client.Environment, error) {
	opts := client.NewListOpts()
	opts.Filters["name"] = stackName
	opts.Filters["removed_null"] = "true"

	stacks, err := c.c.Environment.List(opts)
	if err != nil {
		return nil, err
	}

	if len(stacks.Data) == 0 {
		return nil, fmt.Errorf("Failed to find stack %s", stackName)
	}

	if len(stacks.Data) > 1 {
		return nil, fmt.Errorf("Found %d stacks named %s", len(stacks.Data), stackName)
	}

	return &stacks.Data[0], nil
}

func (c *Client) getTemplates(stackName string) (string, string, error) {
	stack, err := c.getStack(stackName)
	if err != nil {
		return "", "", err
	}

	config, err := c.c.Environment.ActionExportconfig(stack, &client.ComposeConfigInput{})
	if err != nil {
		return "", "", err
	}

	return config.DockerComposeConfig, config.RancherComposeConfig, nil
}
