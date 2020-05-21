package main

// Generate Dockerfile from a Docker image
import (
	"flag"
	"fmt"
	"os"
)

func main() {
	imageIdOpt := flag.String("i", "", "-i [imageid|layerid]")
	imageNameOpt := flag.String("n", "", "-n [foobar:latest|foobar:1.1.2]")
	imageRepo := flag.String("r", "docker.io/library", "-r [02345511234.dkr.ap-southeast-2.aws.com/foobar|asia.gcr.io/google-containers]")
	loglevel := flag.String("l", "info", "-l [info|debug|warn|fatal|error]")
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
			if err = dir.pullImage(); err != nil {
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
}
