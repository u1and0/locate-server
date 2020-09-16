# Usage:
# docker run --name backup -t
#     -v `pwd`/backup:/work
#     --volumes-from=db
#     u1and0/updatedb:backup_v1.0.0

FROM alpine

# /var/lib/mlocate/下のファイルのバックアップを取る
# 毎分locate.logを/workにcopyする
# 毎日12:15にlocate-S.logを/workにcopyする
# 毎日12:16にmlocate*.dbを/workにtar.gzで圧縮する
ARG TASK="/etc/crontabs/root"
RUN echo "SHELL=/bin/sh" > $TASK &&\
    echo "PATH=/sbin:/bin:/usr/sbin:/usr/bin" >> $TASK &&\
    echo "* * * * * cp -f /var/lib/mlocate/locate.log /work/" >> $TASK &&\
    echo "* */12 * * * cp -f /var/lib/mlocate/locate-S.log /work/" >> $TASK &&\
    echo "* */12 * * * tar -cvzf /work/mlocate.tar.gz /var/lib/mlocate/*.db" >> $TASK

# cron起動
CMD ["crond", "&&", "tail", "-f"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="backup regularly"\
      version="u1and0/updatedb:backup_v1.1.0"
