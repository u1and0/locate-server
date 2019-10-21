# 10:30にtar.gzとしてバックアップを取る
#
# Usage:
# docker run --name backup -t
#     -v `pwd`/backup:/work
#     --volumes-from=db
#     u1and0/updatedb:backup_v1.0.0

FROM alpine

# /var/lib/mlocate/下のファイルのログを取る
ARG TASK="/etc/crontabs/root"
RUN echo "SHELL=/bin/sh" > $TASK &&\
    echo "PATH=/sbin:/bin:/usr/sbin:/usr/bin" >> $TASK &&\
    echo "30 1 * * * tar -cvzf /work/mlocate.tar.gz /var/lib/mlocate/*.db /var/lib/mlocate/*.log" >> $TASK

# cron起動
CMD ["crond", "&&", "tail", "-f"]

LABEL maintainer="u1and0 <e01.ando60@gmail.com>"\
      description="backup regularly"\
      version="u1and0/updatedb:backup_v1.0.0"
