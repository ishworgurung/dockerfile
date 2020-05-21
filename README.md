dockerfile
--
A tool to reconstruct the `Dockerfile` of an image. It needs access to Docker daemon.

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
    -r string
      	-r [02345511234.dkr.ap-southeast-2.aws.com/foobar|asia.gcr.io/google-containers] (default "docker.io/library")
```

Build docker image
--
```
$ docker build -t dockerfile:latest .
```

Run
--
Alias it off, and run it
```
$ alias dockerfile-container="docker run --rm -v '/var/run/docker.sock:/var/run/docker.sock' dockerfile:latest"
```

For locally stored image:
```
$ dockerfile-container -n ubuntu:focal
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
```

Pull from gcr.io remote registry:
```
$ dockerfile-container -r asia.gcr.io/google-containers -n addon-builder:latest -l debug
2020-05-22T00:20:31+10:00 WRN tried listing the image name but could not find any image
2020-05-22T00:20:31+10:00 WRN the image could not be found locally
2020-05-22T00:20:31+10:00 INF pulling docker image 'asia.gcr.io/google-containers/addon-builder:latest'
[==================================================>]  451.6MB/451.6MB
FROM asia.gcr.io/google-containers/addon-builder:latest
bazel build ...
bazel build ...
/tmp/installer.sh
RUN /bin/sh -c apt-get -y update \
    &&     apt-get -y install         apt-transport-https         ca-certificates         curl         make         software-properties-common \
    &&     curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - \
    &&     apt-key fingerprint 0EBFCD88 \
    &&     add-apt-repository        "deb [arch=amd64] https://download.docker.com/linux/ubuntu        xenial        edge" \
    &&     apt-get -y update \
    &&     apt-get -y install docker-ce=5:18.09.6~3-0~ubuntu-xenial
ENTRYPOINT ["/usr/bin/docker"]
ARG GOPATH=/workspace/go
ARG GOROOT=/usr/local/go
ENV GOPATH=/workspace/go
ENV GOROOT=/usr/local/go
ARG INSTALL_GOPATH=/builder/go
ENV INSTALL_GOPATH=/builder/go
RUN /bin/sh -c mkdir -pv   /k8s-addon-builder   /builder   "${GOPATH}"   "${INSTALL_GOPATH}"
COPY dir:d22a299ad7a66301436a4f21c9db61700bb8800028ff29621a13211fdf431484 in /usr/local/go 
ENV PATH=/k8s-addon-builder:/builder/google-cloud-sdk/bin:/usr/local/go/bin:/workspace/go/bin:/builder/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
WORKDIR /workspace/go/src/github.com/GoogleCloudPlatform/k8s-addon-builder
COPY dir:6130feacd88863d26ae57249bcb35e06328f7246624ca6eabc2a765127990d9e in ./ 
RUN /bin/sh -c apt-get update   \
    && apt-get install -y --no-install-recommends     build-essential     git     make     jq     wget     python-pip     python-yaml     unzip   \
    && (GOPATH="${INSTALL_GOPATH}"; go get -v -u github.com/golang/dep/cmd/dep)   \
    && make build-static   \
    && cp ply builder-tools/* /k8s-addon-builder   \
    && pip install -r /k8s-addon-builder/requirements.txt   \
    && git config --system credential.helper gcloud.sh   \
    && wget -q -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip   \
    && unzip -p protoc.zip bin/protoc > /usr/local/bin/protoc   \
    && chmod +x /usr/local/bin/protoc   \
    && wget -qO- https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz | tar zxv -C /builder   \
    && CLOUDSDK_PYTHON="python2.7" /builder/google-cloud-sdk/install.sh     --usage-reporting=false     --bash-completion=false     --disable-installation-options   \
    && apt-get upgrade -y   \
    && apt-get dist-upgrade   \
    && rm -rf     /var/lib/apt/lists/*     ~/.config/gcloud
ENTRYPOINT []
LABEL GCB_BUILD_ID=47b1fa03-bde9-484c-a17e-328238b433dd
LABEL GCB_PROJECT_ID=gke-release

```

Image built from `mcr.microsoft.com/dotnet/core/runtime-deps:2.1-focal`:
```
$ dockerfile-container -n ishwor-dotnetcore:2.1 
FROM mcr.microsoft.com/dotnet/core/runtime-deps:2.1-focal
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
RUN /bin/sh -c apt-get update     \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends         ca-certificates                 libc6         libgcc1         libgssapi-krb5-2         libicu66         libssl1.1         libstdc++6         zlib1g     \
    && rm -rf /var/lib/apt/lists/*
ENV ASPNETCORE_URLS=http://+:80 DOTNET_RUNNING_IN_CONTAINER=true
RUN /bin/sh -c apt-get update     \
    && apt-get install -y --no-install-recommends         curl     \
    && rm -rf /var/lib/apt/lists/*
ENV DOTNET_VERSION=2.1.18
RUN /bin/sh -c curl -SL --output dotnet.tar.gz https://dotnetcli.azureedge.net/dotnet/Runtime/$DOTNET_VERSION/dotnet-runtime-$DOTNET_VERSION-linux-x64.tar.gz     \
    && dotnet_sha512='97e8060b97ead693278af4897e33b7d13065ac19fc657c0d1aa184ffc73aab0354a07615e2f15f0bd8280c6212183bba5ffc61e30302539639d1a5b5b6466dbe'     \
    && echo "$dotnet_sha512 dotnet.tar.gz" | sha512sum -c -     \
    && mkdir -p /usr/share/dotnet     \
    && tar -zxf dotnet.tar.gz -C /usr/share/dotnet     \
    && rm dotnet.tar.gz     \
    && ln -s /usr/share/dotnet/dotnet /usr/bin/dotnet
```


*Potential bug?* - Older images do not give enough context on the history using Docker API:
```
$ dockerfile-container -r gcr.io/google-containers -n hugo:latest -l debug
2020-05-22T00:56:01+10:00 WRN tried listing the image name but could not find any image
2020-05-22T00:56:01+10:00 WRN the image could not be found in local disk
2020-05-22T00:56:01+10:00 INF pulling docker image 'gcr.io/google-containers/hugo:latest'
[==================================================>]     587B/587B1MB
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

The tool `dockerfile` uses the Docker Go API to fetch an image's history.

If the image has been stored locally, I use it to read the history of the image
and extract all the relevant Docker steps (build instructions) used to create
the image as a whole.

If the image has not been stored locally / ever downloaded, then it uses the
remote registry to fetch the image and extract all the relevant Docker steps
(build instructions) used to create the image as a whole.

The idea has been inspired by `lukapeschke/dockerfile-from-image` and `chenzj/dfimage`.
Much thanks to them.
