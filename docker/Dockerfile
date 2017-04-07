FROM busybox:ubuntu-14.04

MAINTAINER tsaikd <tsaikd@gmail.com>

ENV GOGSTASH_VERSION 0.1.14

ADD https://github.com/tsaikd/gogstash/releases/download/${GOGSTASH_VERSION}/gogstash-Linux-x86_64 /usr/local/bin/gogstash

RUN chmod +x /usr/local/bin/gogstash

CMD ["gogstash"]
