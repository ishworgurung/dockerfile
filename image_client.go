package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

var (
	errLocalImageNotFound = errors.New("tried listing the image name but could not find any image")
)

func newDockerImageClient(repo string, loglevel string) DockerImageClient {
	l := strings.ToLower(loglevel)
	var ll zerolog.Level
	switch {
	case l == "debug":
		ll = zerolog.DebugLevel
	case l == "warn":
		ll = zerolog.WarnLevel
	case l == "info":
		ll = zerolog.InfoLevel
	case l == "error":
		ll = zerolog.ErrorLevel
	case l == "fatal":
		ll = zerolog.FatalLevel
	default:
		ll = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(ll)

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	zlog := zerolog.New(output).With().Timestamp().Logger()
	cli, err := client.NewEnvClient()
	if err != nil {
		zlog.Fatal().Msg(err.Error())
	}
	dic := DockerImageClient{
		zlog: zlog,
		cli:  cli,
		repo: repo + "/",
	}
	return dic
}

func (d *DockerImageClient) getImageIdByName() (string, error) {
	// TODO: a better way is to use a filter.
	imageList, err := d.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return "", err
	}
	for _, image := range imageList {
		for _, i := range image.RepoTags {
			// e.g. `asia.gcr.io/google-containers/ubuntu-slim:0.14` vs. `ubuntu:focal`
			// The first one is fully canonicalize whereas the second one is integrated
			// with Docker to use `docker.io/library/ubuntu:focal` internally. We look for both matches.
			if len(i) > 0 && (i == (d.repo+d.imageName) || (i == d.imageName)) {
				// Found
				return image.ID, nil
			}
		}
	}
	return "", errLocalImageNotFound
}

func (d *DockerImageClient) getBaseImageTagByImageId(imageId string) (string, error) {
	imageHistory, err := d.cli.ImageHistory(context.Background(), imageId)
	if err != nil {
		return "", err
	}
	var t string
	for _, ih := range imageHistory {
		if len(ih.Tags) == 0 {
			continue
		}
		t = ih.Tags[0]
	}
	return t, nil
}

func (d *DockerImageClient) dockerFile(base string) (string, error) {
	var (
		reservedInstructions = []string{
			"ENV",
			"EXPOSE",
			"ARG",
			"LABEL",
			"USER",
			"EXPOSE",
			"CMD",
			"MAINTAINER",
			"ENTRYPOINT",
			"STOPSIGNAL",
			"COPY",
			"VOLUME",
			"WORKDIR",
			"ONBUILD",
			"HEALTHCHECK",
			"SHELL",
		}
		cleaner = strings.NewReplacer(
			"/bin/sh -c #(nop) ", "",
			"/bin/sh -c", "RUN /bin/sh -c",
			"&&", "\\\n    &&",
		)
	)

	imageHistory, err := d.cli.ImageHistory(context.Background(), d.imageId)
	if err != nil {
		return "", err
	}
	df := fmt.Sprintf("FROM %s\n", base)
	// traverse image history slice backwards
	for i := len(imageHistory) - 1; i >= 0; i-- {
		history := imageHistory[i].CreatedBy
		if len(history) == 0 {
			continue
		}
		steps := cleaner.Replace(history)
		for _, e := range reservedInstructions {
			steps = strings.ReplaceAll(steps, " "+e, e)
		}
		df += fmt.Sprintf("%s\n", steps)
	}
	return df, nil
}

func (d *DockerImageClient) pullImage() error {
	if len(d.repo) == 0 {
		return errors.New("invalid repo")
	}
	if len(d.imageName) == 0 {
		return errors.New("invalid image name")
	}
	canonicalRepo := d.repo + d.imageName
	d.zlog.Info().Msgf("pulling docker image '%s'", canonicalRepo)
	i, err := d.cli.ImagePull(context.Background(), canonicalRepo, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	if i == nil {
		return fmt.Errorf("error occurred when trying to pull image tag '%s'", canonicalRepo)
	}
	defer i.Close()
	j := json.NewDecoder(i)
	pullEvent := &DockerImagePullEvent{}
	for {
		if err := j.Decode(&pullEvent); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if len(pullEvent.Progress) == 0 {
			continue
		}
		// progress bar
		fmt.Printf("\r%s", strings.Repeat(" ", 55))
		fmt.Printf("\r%s", pullEvent.Progress)
	}
	fmt.Println()
	return nil
}
