gogstash grok filter module
=============================

## Synopsis

```yaml
filter:
  - type: grok
    # (optional) grok patterns, default: ["%{COMMONAPACHELOG}"]
    match: ["%{COMMONAPACHELOG}"]
    # (optional) message field to parse, default: "message"
    source: "message"
    # (optional) grok patterns file path, default: empty
    patterns_path: "path/to/file"
```

**NOTICE:** If you using yaml config file, `\` should be written in `\\` in match patterns. For example: `"\\[%{HTTPDATE:nginx.access.time}\\]"`.

## Faster grok parser

If you need faster grok parse speed (by using C code binding regexp library: [Onigmo](https://github.com/k-takata/Onigmo)), you can compile gogstash from source code.

A `Dockerfile` example:

```dockerfile
FROM golang:alpine

ARG version

RUN apk --update add --no-cache ca-certificates git tzdata build-base

# build onigmo
WORKDIR /src/build/
RUN git clone https://github.com/k-takata/Onigmo.git --depth=1 \
  && cd Onigmo && ./configure && make && make install

WORKDIR /go/src/github.com/tsaikd/gogstash
COPY . /go/src/github.com/tsaikd/gogstash
RUN sed -i -e 's/github.com\/vjeantet\/grok/github.com\/tengattack\/grok/' /go/src/github.com/tsaikd/gogstash/filter/grok/filtergrok.go \
  && go get -d -v ./...
RUN go build -ldflags "-X main.Version=$version"
```
