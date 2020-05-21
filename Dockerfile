FROM golang:latest as build
WORKDIR /src
COPY go.* /src/
COPY *.go /src/
RUN go build -v -ldflags='-w -s' -o /tmp/dockerfile

FROM ubuntu:latest as run
COPY --from=build /tmp/dockerfile /usr/local/bin/dockerfile
ENTRYPOINT ["/usr/local/bin/dockerfile"]

