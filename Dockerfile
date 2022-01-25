# Usage:
# ```
# $ docker run -d --rm --name locs_test u1and0/locate-server [options]
# ```

FROM golang:1.17.6-bullseye AS go_builder
RUN apt update &&\
    apt install -y git &&\
    go install github.com/u1and0/gocate@latest
WORKDIR /work
# For go module using go-pipeline
ENV GO111MODULE=on
COPY ./main.go /work/main.go
COPY ./go.mod /work/go.mod
COPY ./go.sum /work/go.sum
COPY ./cmd /work/cmd
RUN go build -o /go/bin/locate-server

FROM u1and0/plocate
COPY --from=go_builder /go/bin/locate-server /usr/bin/locate-server
COPY --from=go_builder /go/bin/gocate /usr/bin/gocate
WORKDIR /var/www
COPY ./static /var/www/static
COPY ./templates /var/www/templates
ENTRYPOINT ["/usr/bin/locate-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Run locate-server"\
      version="v3.0.0r"
