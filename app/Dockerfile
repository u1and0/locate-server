# Usage:
# n時間ごとにupdatedbによってディレクトリデータベースを更新する
# docker run --name app -it --rm -v `pwd`:/work u1and0/updatedb
# app(このファイルで作成したイメージによるコンテナ)はtail -fでcrondの結果を標準出力に出力し続ける
FROM alpine

# locate, updatedbコマンドを使用可能にする
# tzdataは環境変数ENVを変えるとUTCから変更される
# ex) docker run --rm -it -e TZ='Asia/Tokyo' u1and0/updatedb
RUN apk --update --no-cache add mlocate tzdata

# work 下のディレクトリを毎時データベース化
ARG TASK="/etc/crontabs/root"
RUN echo "SHELL=/bin/sh" > $TASK &&\
    echo "PATH=/sbin:/bin:/usr/sbin:/usr/bin" >> $TASK &&\
    echo "32 8 * * * updatedb -U \${UPDATEDB_PATH:-/} --output=/var/lib/mlocate/\${OUTPUT:-mlocate.db}" >> $TASK
# docker runするときに-e UPDATEDB_PATH=/work　とすればwork以下のディレクトリを
# var/lib/mlocate/mlocate.dbにまとめてくれる
# デフォルトはルート/
# 同時に-e OUTPUT=hoge.dbとすれば/var/lib/mlocateディレクトリにデータベースを作成してくれる.
# デフォルトmlocate.db

# updatedbデータベースファイルを格納するマウントポイント
VOLUME ["/var/lib/mlocate/"]

# cron起動
CMD ["crond", "&&", "tail", "-f"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="make database regularly by `updatedb` command"\
      version="u1and0/updatedb:v1.0.2"
