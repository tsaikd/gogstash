#!/bin/bash

set -e

PN="${BASH_SOURCE[0]##*/}"
PD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

renice 15 $$
pushd "${PD}/.." >/dev/null

orgname="tsaikd"
projname="gogstash"
repo="github.com/${orgname}/${projname}"
githash="$(git rev-parse HEAD | cut -c1-6)"
cachedir="/tmp/${orgname}-${projname}-cache"
buildtoolimg="golang:1"

if [ -d "${cachedir}" ] ; then
	if [ "$(stat -c %Y "${cachedir}")" -lt "$(date +%s -d -7day)" ] ; then
		rm -rf "${cachedir}" || true
		docker pull "${buildtoolimg}"
	fi
else
	docker pull "${buildtoolimg}"
fi

echo "[$(date -Iseconds)] build ${projname} binary (${githash})"
docker run --rm \
	-e GITHUB_TOKEN="${GITHUB_TOKEN}" \
	-w "/go/src/${repo}" \
	-v "${PWD}:/go/src/${repo}" \
	-v "${cachedir}/go/src:/go/src" \
	-v "${cachedir}/go/bin:/go/bin" \
	"${buildtoolimg}" \
	"./docker/build-in-docker.sh"

echo "[$(date -Iseconds)] ${projname} finished (${githash})"

popd >/dev/null
