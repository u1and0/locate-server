# Usage:
# ```
# $ docker run -d --rm --name locs_test u1and0/locate-server [options]
# ```

FROM golang:1.17.0-alpine3.14 AS go_official
COPY ./main.go /go/src/github.com/u1and0/locate-server/main.go
COPY ./go.mod /go/src/github.com/u1and0/locate-server/go.mod
COPY ./go.sum /go/src/github.com/u1and0/locate-server/go.sum
COPY ./cmd /go/src/github.com/u1and0/locate-server/cmd
WORKDIR /go/src/github.com/u1and0/locate-server
# For go module using go-pipeline
ENV GO111MODULE=on
RUN apk --update --no-cache add git &&\
    go build -o /go/bin/locate-server
RUN go get github.com/u1and0/gocate

FROM frolvlad/alpine-glibc:alpine-3.14_glibc-2.33
COPY --from=go_official /go/bin/locate-server /usr/bin/locate-server
COPY --from=go_official /go/bin/gocate /usr/bin/gocate
COPY ./js /usr/bin/js
RUN apk --update --no-cache add mlocate tzdata
ENTRYPOINT ["/usr/bin/locate-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running locate-server"\
      version="v2.3.0"
