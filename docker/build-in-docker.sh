#!/bin/bash

set -e

PN="${BASH_SOURCE[0]##*/}"
PD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
projorg="tsaikd"
projname="gogstash"

renice 15 $$
cd "${PD}/.."

if ! type gobuilder &>/dev/null ; then
	go get -v "github.com/tsaikd/gobuilder"
fi

gobuilder --test --all

go test ./config/...

if [ "${GITHUB_TOKEN}" ] ; then
	echo "[$(date -Iseconds)] ${projname} release on github"
	if ! type github-release &>/dev/null ; then
		go get -v "github.com/aktau/github-release"
	fi
	version="$(./${projname} -v | grep -Eo "version [0-9\.]+" | cut -c9-)"
	github-release release \
		--user "${projorg}" \
		--repo "${projname}" \
		--tag "${version}"
	github-release upload \
		--user "${projorg}" \
		--repo "${projname}" \
		--tag "${version}" \
		--name "${projname}-$(uname -s)-$(uname -m)" \
		--file "${projname}"
fi
