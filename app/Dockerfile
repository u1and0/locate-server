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

FROM golang:1.17.0-alpine3.14 AS go_official

# locate, updatedbコマンドを使用可能にする
# tzdataは環境変数ENVを変えるとUTCから変更される
# ex) docker run -d -e TZ='Asia/Tokyo' u1and0/updatedb
RUN apk --update --no-cache add mlocate tzdata git &&\
    go install github.com/u1and0/gocate@v0.3.0

ARG TASK="/etc/crontabs/root"
RUN echo "SHELL=/bin/sh" > $TASK &&\
    echo "PATH=/sbin:/bin:/usr/sbin:/usr/bin" >> $TASK &&\
    echo "3 0 * * * gocate -init" >> $TASK
# ${UPDATEDB_PATH}以下のディレクトリを定期的にデータベース化
# 指定しない場合${UPDATEDB_PATH}のデフォルトは/

# /var/lib/mlocateディレクトリに${OUTPUT}で指定したファイル名のデータベースを作成する
# 指定しない場合${OUTPUT}のデフォルトはmlocate.db

# updatedbデータベースファイルを格納するマウントポイント
VOLUME ["/var/lib/mlocate/"]

# cron起動
CMD ["crond", "&&", "tail", "-f"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="make database regularly by `updatedb` command"\
      version="u1and0/updatedb:v1.1.0"
