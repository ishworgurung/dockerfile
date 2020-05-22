package main

import (
	"flag"
	"fmt"
	"os"
)

// Generate Dockerfile from a Docker image.
//
// Docker by design, only fetches leaf container image *if* the
// container was pulled from a remote repository and as such
// cannot get the true base image.
// If however, the image was built locally, it will have full
// tagged layers and so it is possible to fetch the parent image that
// it was built from.
// More detail at https://windsock.io/explaining-docker-image-ids/ and https://stackoverflow.com/a/59894125.

func main() {
	imageIdOpt := flag.String("i", "", "-i [imageid|layerid]")
	imageNameOpt := flag.String("n", "", "-n [foobar:latest|foobar:1.1.2]")
	imageRepo := flag.String("r", "docker.io/library", "-r registry [191229840194.dkr.ecr.us-west-2.amazonaws.com|asia.gcr.io/google-containers]")
	loglevel := flag.String("l", "info", "-l [info|debug|warn|fatal|error]")
	username := flag.String("u", "", "-u [registry username]")
	password := flag.String("p", "", "-p [registry password]")
	layersTree := flag.Bool("t", false, "-t [true|false] (for image layer visualisation)")

	flag.Parse()

	var (
		err error
	)
	// Either image id or image name tag must be provided.
	// The Image repo is optional and defaults to `docker.io/library/`
	if len(*imageNameOpt) == 0 && len(*imageIdOpt) == 0 {
		println("either image name or image id should be provided")
		flag.Usage()
		os.Exit(127)
	}

	dir := newDockerImageClient(*imageRepo, *loglevel)
	if len(*imageNameOpt) > 0 {
		dir.imageName = *imageNameOpt
		// Search the user provided image name to get the image id
		dir.imageId, err = dir.getImageIdByName()
		if err != nil {
			dir.zlog.Warn().Msg(err.Error())
		}
		// Pull image from registry since it does not exist in the local disk
		if len(dir.imageId) == 0 {
			dir.zlog.Debug().Msg("the image could not be found in local disk")
			if err = dir.pullImage(*username, *password, dir.repo, dir.imageName); err != nil {
				dir.zlog.Fatal().Msg(err.Error())
			}
		}
	} else {
		// Search the user provided image id to get the image name.
		// Image pull does not happen here.
		dir.imageName, err = dir.getBaseImageTagByImageId(*imageIdOpt)
		if err != nil {
			dir.zlog.Error().Msg(err.Error())
		}
		if len(dir.imageName) == 0 {
			dir.zlog.Fatal().Msg("the image could not be found in local disk")
		}
		dir.imageId = *imageIdOpt
	}

	dir.imageId, err = dir.getImageIdByName()
	if err != nil {
		dir.zlog.Fatal().Msg(err.Error())
	}
	// Dockerfile re-construction
	bi, err := dir.getBaseImageTagByImageId(dir.imageId)
	if err != nil {
		dir.zlog.Fatal().Msg(err.Error())
	}
	dir.dockerfile, err = dir.dockerFile(bi)
	if err != nil {
		dir.zlog.Fatal().Msg(err.Error())
	}
	fmt.Printf("%s", dir.dockerfile)

	if *layersTree {
		dir.zlog.Info().Msg("printing all local images layer tree")
		// nate/dockviz:latest is public on docker.io/library
		if err := dir.runContainer(
			"docker.io", "nate/dockviz:latest", []string{"images", "-t"}, "", "", false); err != nil {
			dir.zlog.Error().Msg(err.Error())
		}
	}
}
