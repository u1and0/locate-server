# Usage:
# ```
# $ docker run -d --rm --name locs_test u1and0/locate-server [options]
# ```

FROM golang:1.12.7-alpine3.10 AS go_official
COPY ./main.go /go/src/github.com/u1and0/locate-server/main.go
WORKDIR /go/src/github.com/u1and0/locate-server
RUN go build -o /go/bin/locate-server

FROM frolvlad/alpine-glibc
COPY --from=go_official /go/bin/locate-server /usr/bin/locate-server
RUN apk add mlocate tzdata
ENTRYPOINT ["/usr/bin/locate-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running locate-server"\
      version="v0.3.0"
