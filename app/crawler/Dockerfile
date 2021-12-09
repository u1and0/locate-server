
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

FROM alpine
RUN apk --update --no-cache add jq nkf
ARG TASK="/etc/crontabs/root"
COPY ./crawler.sh /usr/bin/crawler.sh
RUN echo "SHELL=/bin/sh" > $TASK &&\
    echo "PATH=/sbin:/bin:/usr/sbin:/usr/bin" >> $TASK &&\
    echo "2 4 * * * /usr/bin/crawler.sh">> $TASK
CMD ["crond", "&&", "tail", "-f"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="make database regularly by `updatedb` command"\
      version="u1and0/crawler:v0.1.0"
