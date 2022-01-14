# Usage:
# ```
# $ docker run -d --rm --name locs_test u1and0/locate-server [options]
# ```

FROM golang:1.17.6-bullseye AS go_builder
RUN apt update &&\
    apt install -y git &&\
    go install github.com/u1and0/gocate@v0.3.1
WORKDIR /go/src/github.com/u1and0/locate-server
# For go module using go-pipeline
ENV GO111MODULE=on
COPY ./main.go /go/src/github.com/u1and0/locate-server/main.go
COPY ./go.mod /go/src/github.com/u1and0/locate-server/go.mod
COPY ./go.sum /go/src/github.com/u1and0/locate-server/go.sum
COPY ./cmd /go/src/github.com/u1and0/locate-server/cmd
RUN go build -o /go/bin/locate-server

FROM debian:bullseye-slim
RUN apt update && apt install -y  plocate tzdata &&\
    apt clean -y &&\
    rm -rf /var/lib/apt/lists/*
COPY --from=go_builder /go/bin/locate-server /usr/bin/locate-server
COPY --from=go_builder /go/bin/gocate /usr/bin/gocate
WORKDIR /var/www
COPY ./static /var/www/static
COPY ./templates /var/www/templates
ENTRYPOINT ["/usr/bin/locate-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Run locate-server"\
      version="v3.0.0r"
