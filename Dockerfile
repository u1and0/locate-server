# Usage:
# ```
# $ docker run -d --rm --name locs_test u1and0/locate-server sh -c "locate-server [options]"
# ```

FROM golang:1.12.6-alpine3.10 AS go_official
RUN apk add git &&\
    go get -u github.com/u1and0/locate-server

FROM frolvlad/alpine-glibc
COPY --from=go_official /go/bin/locate-server /usr/bin/locate-server
RUN apk add mlocate
CMD locate-server

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Running locate-server"\
      version="v0.0.0"
