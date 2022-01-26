# n時間ごとにupdatedbによってディレクトリデータベースを更新する
# app(このファイルで作成したイメージによるコンテナ)はtail -fでcrondの結果を標準出力に出力し続ける
#
# Usage:
# docker run --name app -t --rm\
#   -v `pwd`:/work\
#   -e TZ='Asia/Tokyo'\
#   -e UPDATEDB_PATH=/path/to/work\
#   -e OUTPUT=mlocatework.db\
#   u1and0/updatedb

FROM u1and0/plocate:latest
RUN pacman -Syu --noconfirm go &&\
    go install github.com/u1and0/gocate@latest
RUN ln -s /root/go/bin/gocate /usr/bin

CMD : "Write crontab" &&\
    echo "${CRON_CYCLE} gocate -init -U ${UPDATEDB_PATH} -o /var/lib/plocate " | crontab - &&\
    : "Execute updatedb at first time" &&\
    gocate -init -U ${UPDATEDB_PATH} &&\
    : "Up cron daemon" &&\
    crond &&\
    : "Wait for some command (NOT exit container)" &&\
    tail -f /dev/null

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="make database regularly by `updatedb` command"\
      version="u1and0/updatedb:v1.1.0r"
