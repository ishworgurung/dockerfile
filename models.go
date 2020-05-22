package main

import (
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

// Pull image event response from Docker daemon
type DockerImagePullEvent struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

// Docker image client to hold all the relevant bits of information
type DockerImageClient struct {
	zlog       zerolog.Logger
	cli        *client.Client
	imageName  string
	imageId    string
	repo       string
	dockerfile string
	username   string
	password   string
}
