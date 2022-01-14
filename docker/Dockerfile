FROM ubuntu:18.04 as build

ENV GOGSTASH_VERSION 0.1.21

ADD https://github.com/tsaikd/gogstash/releases/download/${GOGSTASH_VERSION}/gogstash-Linux-x86_64 /usr/local/bin/

RUN mv /usr/local/bin/gogstash-Linux-x86_64 /usr/local/bin/gogstash && chmod +x /usr/local/bin/gogstash

ENTRYPOINT exec gogstash --config /etc/gogstash.yaml
