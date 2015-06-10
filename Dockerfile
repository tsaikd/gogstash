FROM busybox:ubuntu-14.04

MAINTAINER tsaikd <tsaikd@gmail.com>

RUN mkdir -p /gogstash

ADD gogstash-Linux-x86_64 /gogstash/gogstash

WORKDIR /gogstash

CMD ["./gogstash"]

