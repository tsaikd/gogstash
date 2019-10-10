#!/bin/bash

set -eu

PN="${BASH_SOURCE[0]##*/}"
PD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

renice 15 $$
cd "${PD}/.."

docker build \
  --build-arg "GITHUB_TOKEN=${GITHUB_TOKEN:=}" \
  -t tsaikd/gogstash \
  -f docker/Dockerfile \
  .
