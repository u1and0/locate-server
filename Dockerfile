# Usage:
# ```
# $ docker run -d --rm --name locs_test u1and0/locate-server [options]
# ```

FROM golang:1.17.0-alpine3.14 AS go_official
RUN apk --update --no-cache add git &&\
    go install github.com/u1and0/gocate@v0.3.0
WORKDIR /go/src/github.com/u1and0/locate-server
# For go module using go-pipeline
ENV GO111MODULE=on
COPY ./main.go /go/src/github.com/u1and0/locate-server/main.go
COPY ./go.mod /go/src/github.com/u1and0/locate-server/go.mod
COPY ./go.sum /go/src/github.com/u1and0/locate-server/go.sum
COPY ./cmd /go/src/github.com/u1and0/locate-server/cmd
RUN go build -o /go/bin/locate-server

FROM frolvlad/alpine-glibc:alpine-3.14_glibc-2.33
RUN apk --update --no-cache add mlocate tzdata
WORKDIR /var/www
COPY --from=go_official /go/bin/locate-server /usr/bin/locate-server
COPY --from=go_official /go/bin/gocate /usr/bin/gocate
COPY ./static /var/www/static
COPY ./templates /var/www/templates
ENTRYPOINT ["/usr/bin/locate-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Run locate-server"\
      version="v3.0.0"
