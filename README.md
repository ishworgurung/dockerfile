Description
--
A tool to reconstruct a best-effort `Dockerfile` of an image. It needs access to Docker daemon's socket.

Use it standalone:
--
```
$ git clone github.com/ishworgurung/dockerfile
$ cd dockerfile
$ go build -ldflags='-s -w'
$ ./dockerfile -h
Usage of ./dockerfile:
  -i string
    	-i [imageid|layerid]
  -l string
    	-l [info|debug|warn|fatal|error] (default "info")
  -n string
    	-n [foobar:latest|foobar:1.1.2]
  -p string
    	-p [password]
  -r string
    	-r [191229840194.dkr.ecr.us-west-2.amazonaws.com|asia.gcr.io/google-containers] (default "docker.io/library")
  -t	-t [true|false] (for image layer visualisation)
  -u string
    	-u [username]
```

Build docker image
--
```
$ docker build -t dockerfile:latest .
$ alias dockerfile-container="docker run --rm -v '/var/run/docker.sock:/var/run/docker.sock' dockerfile:latest"
```

Run
--
Print Dockerfile for locally built image:
```
$ dockerfile-container -n dockerfile:1.0.6
FROM debian:latest
ADD file:fb54c709daa205bf9d04eb3d90ba068db4c34dfe3b6ec0d7691f677286120903 in /
CMD ["bash"]
COPY file:7c8482c8c3e71151e1c9e4351eccecfd03edf52470090db7560c6adec166f8c5 in /usr/local/bin/dockerfile
ENTRYPOINT ["/usr/local/bin/dockerfile"]
```

Print Dockerfile for a fresh new image with the image layer visualisation turned on:
```
$ dockerfile-container -n ubuntu:focal -t
2020-05-22T10:22:49Z WRN tried listing the image name but could not find any image
2020-05-22T10:22:49Z INF pulling docker image 'docker.io/library/ubuntu:focal'
[==================================================>]     163MB/163MB
FROM ubuntu:focal
ADD file:a58c8b447951f9e30c92e7262a2effbb8b403c2e795ebaf58456f096b5b2a720 in /
RUN /bin/sh -c [ -z "$(apt-get indextargets)" ]
RUN /bin/sh -c set -xe 		\
    && echo '#!/bin/sh' > /usr/sbin/policy-rc.d 	\
    && echo 'exit 101' >> /usr/sbin/policy-rc.d 	\
    && chmod +x /usr/sbin/policy-rc.d 		\
    && dpkg-divert --local --rename --add /sbin/initctl 	\
    && cp -a /usr/sbin/policy-rc.d /sbin/initctl 	\
    && sed -i 's/^exit.*/exit 0/' /sbin/initctl 		\
    && echo 'force-unsafe-io' > /etc/dpkg/dpkg.cfg.d/docker-apt-speedup 		\
    && echo 'DPkg::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' > /etc/apt/apt.conf.d/docker-clean 	\
    && echo 'APT::Update::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' >> /etc/apt/apt.conf.d/docker-clean 	\
    && echo 'Dir::Cache::pkgcache ""; Dir::Cache::srcpkgcache "";' >> /etc/apt/apt.conf.d/docker-clean 		\
    && echo 'Acquire::Languages "none";' > /etc/apt/apt.conf.d/docker-no-languages 		\
    && echo 'Acquire::GzipIndexes "true"; Acquire::CompressionTypes::Order:: "gz";' > /etc/apt/apt.conf.d/docker-gzip-indexes 		\
    && echo 'Apt::AutoRemove::SuggestsImportant "false";' > /etc/apt/apt.conf.d/docker-autoremove-suggests
RUN /bin/sh -c mkdir -p /run/systemd \
    && echo 'docker' > /run/systemd/container
CMD ["/bin/bash"]
2020-05-22T10:22:58Z INF printing all local images layer tree
2020-05-22T10:22:58Z INF pulling image 'nate/dockviz:latest'
├─<missing> Virtual Size: 6.6 MB
│ └─<missing> Virtual Size: 6.6 MB
│   └─93b5259c1e18 Virtual Size: 6.6 MB Tags: nate/dockviz:latest
├─<missing> Virtual Size: 72.8 MB
│ └─<missing> Virtual Size: 73.9 MB
│   └─<missing> Virtual Size: 73.9 MB
│     └─<missing> Virtual Size: 73.9 MB
│       └─1d622ef86b13 Virtual Size: 73.9 MB Tags: ubuntu:focal
```

Print Dockerfile for an image hosted on Amazon ECR registry:
```
$ dockerfile-container -r 121712040114.dkr.ecr.ap-southeast-2.amazonaws.com     \
    -n foobar:latest                                                            \
    -u AWS                                                                      \
    -p $(aws ecr get-login-password --region ap-southeast-2)

FROM 121712040114.dkr.ecr.ap-southeast-2.amazonaws.com/foobar:latest
ADD file:e69d441d729412d24675dcd33e04580885df99981cec43de8c9b24015313ff8e in /
CMD ["/bin/sh"]
[ ... ]
```

Some old images do not have enough context on the history using Docker API:
```
$ dockerfile-container -r gcr.io/google-containers -n hugo:latest -l debug
2020-05-22T00:56:01+10:00 WRN tried listing the image name but could not find any image
2020-05-22T00:56:01+10:00 WRN the image could not be found in local disk
2020-05-22T00:56:01+10:00 INF pulling docker image 'gcr.io/google-containers/hugo:latest'
[==================================================>]     587MB/587MB
FROM gcr.io/google-containers/hugo:latest
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
/bin/sh
```

The hows
--
The tool `dockerfile` uses the Docker client Go API to fetch an image's history.

If the image has been stored locally, I use it to read the history of the image
and extract all the relevant Docker steps (build instructions) used to create
the image as a whole.

If the image has not been stored locally / ever downloaded, then it uses the
remote registry to fetch the image and extract all the relevant Docker steps
(build instructions) used to create the image as a whole.

Limitations
--
By design, Docker fetches only a leaf container image when pulled from a remote repository. 
Thus, it cannot get the true base image of a "remotely" pulled image. 

On the other hand, when locally building an image, it will have information on the tagged 
layers and thus able to get the tagged parent image layer that it was built from.
More details at https://windsock.io/explaining-docker-image-ids/ and https://stackoverflow.com/a/59894125.

Thanks
--
The idea has been inspired by `lukapeschke/dockerfile-from-image` and `chenzj/dfimage`. Much thanks to them.
Thanks to `nate/dockviz` for the handy tool to check image layers visualisation.
