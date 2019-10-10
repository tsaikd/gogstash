FROM golang:1 AS version
WORKDIR /build
COPY .git .git
RUN git describe --tags | tee VERSION
RUN git describe --tags --abbrev=0 | tee TAG
RUN git rev-parse HEAD | tee GITCOMMIT

FROM golang:1 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=version /build/VERSION /build/TAG /build/GITCOMMIT ./
RUN CGO_ENABLED=0 go build -ldflags "\
	-X \"github.com/tsaikd/KDGoLib/version.NAME=gogstash\" \
	-X \"github.com/tsaikd/KDGoLib/version.BUILDTIME=$(date)\" \
	-X \"github.com/tsaikd/KDGoLib/version.VERSION=$(cat VERSION)\" \
	-X \"github.com/tsaikd/KDGoLib/version.GITCOMMIT=$(cat GITCOMMIT)\" \
	"
RUN ./gogstash version

FROM golang:1 AS release
WORKDIR /build
RUN go get -v "github.com/aktau/github-release"
COPY --from=version /build/VERSION /build/TAG /build/GITCOMMIT ./
COPY --from=build /build/gogstash .
ARG GITHUB_TOKEN
ARG GITHUB_USER=tsaikd
ARG GITHUB_REPO=gogstash
RUN if [ "$(cat VERSION)" != "$(cat TAG)" ]; then echo "--pre-release" > PRERELEASE ; else > PRERELEASE ; fi
RUN if [ "${GITHUB_TOKEN}" ]; then \
	github-release release \
	--tag "$(cat VERSION)" \
	--target "$(cat GITCOMMIT)" \
	--description "curl 'https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/download/$(cat VERSION)/${GITHUB_REPO}-$(uname -s)-$(uname -m)' -SLo ${GITHUB_REPO} && chmod +x ${GITHUB_REPO}" \
	$(cat PRERELEASE) ; \
	github-release upload \
	--tag "$(cat VERSION)" \
	--name "${GITHUB_REPO}-$(uname -s)-$(uname -m)" \
	--file "${GITHUB_REPO}" ; \
	fi

FROM alpine:3
COPY --from=build /build/gogstash /usr/local/bin/gogstash
RUN gogstash version
CMD gogstash
