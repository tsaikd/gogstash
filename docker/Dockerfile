FROM busybox:1-glibc

MAINTAINER tsaikd <tsaikd@gmail.com>

ENV GOGSTASH_VERSION 0.1.18

ADD https://github.com/tsaikd/gogstash/releases/download/${GOGSTASH_VERSION}/gogstash-Linux-x86_64 /usr/local/bin/gogstash

RUN chmod +x /usr/local/bin/gogstash

CMD ["gogstash"]
