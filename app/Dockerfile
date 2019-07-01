# Usage:
# 1時間ごとにupdatedbによってディレクトリデータベースを更新する
# docker run --name app -it --rm -v `pwd`:/work u1and0/updatedb
# app(このファイルで作成したイメージによるコンテナ)はtail -fでcrondの結果を標準出力に出力し続ける
#
#
# appは/var/lib/mlocateディレクトリをさらしているので、
# --volumes-fromオプションによりappをマウントすると、
# updatedbによって作られたmlocate.dbファイルが
# /var/lib/mlocateディレクトリに作成/更新される
#
# 以下のコマンドはappの/var/lib/mlocateディレクトリを
# dbという名前のコンテナにマウントしたものを作成する。
# ```shell-session
# $ docker create --name db --volumes-from app busybox
# ```
#
#
# ```/etc/crontabs/root
# # do daily/weekly/monthly maintenance
# min   hour    day     month   weekday command
# */15    *       *       *       *       run-parts /etc/periodic/15min
# 0       *       *       *       *       run-parts /etc/periodic/hourly
# 0       2       *       *       *       run-parts /etc/periodic/daily
# 0       3       *       *       6       run-parts /etc/periodic/weekly
# 0       5       1       *       *       run-parts /etc/periodic/monthly
#
# SHELL=/bin/sh
# PATH=/sbin:/bin:/usr/sbin:/usr/bin
# * * * * * updatedb -vU /work
# ```
# hourly, dailyではスケジュールできない複雑なスケジュールは/etc/crontabs/rootにボリュームをアタッチする
#
# ===========
#
# 1. data volume用のコンテナdbを作る
# docker create --name db -v /var/lib/mlocate -v /ShareUsers:/ShareUsers:ro busybox
#
# このコマンドではdbコンテナの/varlib/mlocateを外部に晒して、
# ホストのShareUsersをdbコンテナにマウントする
#
#
# 2. updatedb用のコンテナappを作る
# docker run --name app\
#     --volumes-from db\
#     -e UPDATEDB_PATH=/ShareUsers/<path to the db root>\
#     -e OUTPUT=mlocatepersonal.db\
#     u1and0/upadtedb
#
# このコマンドではdbコンテナのボリュームを参照し、
# updatedbをかけるパスをUPDATEDB_PATHで指定している
# dbでマウントしているのでこのコンテナで再度マウントする必要はない
#
#
# 3. locateコマンドでファイル検索するコンテナwebを作る
# docker run --name web --volumes-from db u1and0/locate-server [OPTIONS]
# docker run --name web --rm -it\
#    --volumes-from db -e TZ='Asia/Tokyo'\
#    -e LOCATE_PATH='/var/lib/mlocate/mlocatepersonal.db:/var/lib/mlocate/mlocatecommon.db'\
#    -p 8081:8080 u1and0/locate-server\
#    sh -c "locate-server -s -r '\\DFS'"


FROM alpine

# locate, updatedbコマンドを使用可能にする
# tzdataは環境変数ENVを変えるとUTCから変更される
# ex) docker run --rm -it -e TZ='Asia/Tokyo' u1and0/updatedb
RUN apk --update --no-cache add mlocate tzdata

# work 下のディレクトリを毎時データベース化
ARG TASK="/etc/crontabs/root"
RUN echo "SHELL=/bin/sh" >> $TASK &&\
    echo "PATH=/sbin:/bin:/usr/sbin:/usr/bin" >> $TASK &&\
    echo "7 * * * * updatedb -vU \${UPDATEDB_PATH:-/} --output=/var/lib/mlocate/\${OUTPUT:-mlocate.db}" >> $TASK
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
      version="u1and0/updatedb:0.0.0"
