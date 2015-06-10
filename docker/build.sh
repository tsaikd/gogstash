#!/bin/bash

set -e

PN="${BASH_SOURCE[0]##*/}"
PD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd "${PD}/.." >/dev/null

githash="$(git rev-parse HEAD | cut -c1-6)"
buildtime="$(date -Iseconds)"
cachedir="/tmp/gogstash-cache"

if [ -d "${cachedir}" ] ; then
	if [ "$(stat -c %Y "${cachedir}")" -lt "$(date +%s -d -1day)" ] ; then
		docker pull golang:latest
		rm -rf "${cachedir}" || true
	fi
fi

echo "[$(date -Iseconds)] build binary"
docker run --rm \
	-e BUILDTIME="${buildtime}" \
	-e GITHASH="${githash}" \
	-w "/go/src/github.com/tsaikd/gogstash" \
	-v "${PWD}:/go/src/github.com/tsaikd/gogstash" \
	-v "${cachedir}/go/src:/go/src" \
	-v "${cachedir}/go/bin:/go/bin" \
	golang:latest \
	"./docker/build-in-docker.sh"

echo "[$(date -Iseconds)] finish"

popd >/dev/null

