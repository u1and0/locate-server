# Usage:
# ```
# $ docker run -d --rm --name locs_test u1and0/locate-server [options]
# ```

FROM archlinux:base-devel AS go_builder
RUN pacman-key --init &&\
    pacman-key --populate archlinux &&\
    pacman -Syu --noconfirm go git &&\
    : "Clear cache" &&\
    pacman -Qtdq | xargs -r pacman --noconfirm -Rcns
ENV GOPATH=/go
RUN go install github.com/u1and0/gocate@v0.3.2
WORKDIR /work
ENV GO111MODULE=on
COPY ./main.go /work/main.go
COPY ./go.mod /work/go.mod
COPY ./go.sum /work/go.sum
COPY ./cmd /work/cmd
RUN go build -o /go/bin/locate-server

FROM archlinux:base-devel
RUN pacman-key --init &&\
    pacman-key --populate archlinux &&\
    pacman -Syu --noconfirm plocate &&\
    : "Clear cache" &&\
    pacman -Qtdq | xargs -r pacman --noconfirm -Rcns
COPY --from=go_builder /go/bin/locate-server /usr/bin/locate-server
COPY --from=go_builder /go/bin/gocate /usr/bin/gocate
WORKDIR /var/www
COPY ./static /var/www/static
COPY ./templates /var/www/templates
ENTRYPOINT ["/usr/bin/locate-server"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="Run locate-server"\
      version="v3.0.0r"
