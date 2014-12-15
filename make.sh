#!/bin/sh

set -e

gopkg="github.com/tsaikd/gogstash"
githash="$(git rev-parse HEAD | cut -c1-8)"
buildtime="$(date +%Y-%m-%d)"

LDFLAGS="${LDFLAGS} -X ${gopkg}/version.BUILDTIME ${buildtime}"
LDFLAGS="${LDFLAGS} -X ${gopkg}/version.GITCOMMIT ${githash}"

cd gogstash
go get -v
go build -ldflags "${LDFLAGS}"
./gogstash -V
