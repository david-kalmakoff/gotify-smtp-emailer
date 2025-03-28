BUILDDIR=./build
GOTIFY_VERSION=v2.6.0
FILE_SUFFIX=-for-gotify-$(GOTIFY_VERSION)
PLUGIN_NAME=gotify-smtp-emailer
PLUGIN_ENTRY=plugin.go
GO_VERSION=`cat $(BUILDDIR)/gotify-server-go-version`
DOCKER_BUILD_IMAGE=gotify/build
DOCKER_WORKDIR=/proj
DOCKER_RUN=docker run --rm -v "$$PWD/.:${DOCKER_WORKDIR}" -v "`go env GOPATH`/pkg/mod/.:/go/pkg/mod:ro" -w ${DOCKER_WORKDIR}
DOCKER_GO_BUILD=go build -mod=readonly -a -installsuffix cgo -ldflags "$$LD_FLAGS" -buildmode=plugin 

test-build: build-linux-amd64 test

test:
	GOTIFY_VERSION=$(GOTIFY_VERSION) FILE_SUFFIX=$(FILE_SUFFIX) go test ./...

local: build-linux-amd64
	GOTIFY_VERSION=$(GOTIFY_VERSION) FILE_SUFFIX=$(FILE_SUFFIX) go run cmd/local/main.go

download-tools:
	GO111MODULE=on go install github.com/gotify/plugin-api/cmd/gomod-cap

create-build-dir:
	mkdir -p ${BUILDDIR} || true

update-go-mod: create-build-dir
	wget -LO ${BUILDDIR}/gotify-server.mod https://raw.githubusercontent.com/gotify/server/${GOTIFY_VERSION}/go.mod
	gomod-cap -from ${BUILDDIR}/gotify-server.mod -to go.mod
	rm ${BUILDDIR}/gotify-server.mod || true
	go mod tidy

get-gotify-server-go-version: create-build-dir
	rm ${BUILDDIR}/gotify-server-go-version || true
	wget -LO ${BUILDDIR}/gotify-server-go-version https://raw.githubusercontent.com/gotify/server/${GOTIFY_VERSION}/GO_VERSION

build-linux-amd64 : get-gotify-server-go-version update-go-mod
	${DOCKER_RUN} ${DOCKER_BUILD_IMAGE}:$(GO_VERSION)-linux-amd64 ${DOCKER_GO_BUILD} -o ${BUILDDIR}/${PLUGIN_NAME}-linux-amd64${FILE_SUFFIX}.so ${DOCKER_WORKDIR}

build-linux-arm-7: get-gotify-server-go-version update-go-mod
	${DOCKER_RUN} ${DOCKER_BUILD_IMAGE}:$(GO_VERSION)-linux-arm-7 ${DOCKER_GO_BUILD} -o ${BUILDDIR}/${PLUGIN_NAME}-linux-arm-7${FILE_SUFFIX}.so ${DOCKER_WORKDIR}

build-linux-arm64: get-gotify-server-go-version update-go-mod
	${DOCKER_RUN} ${DOCKER_BUILD_IMAGE}:$(GO_VERSION)-linux-arm64 ${DOCKER_GO_BUILD} -o ${BUILDDIR}/${PLUGIN_NAME}-linux-arm64${FILE_SUFFIX}.so ${DOCKER_WORKDIR}

build: build-linux-arm-7 build-linux-amd64 build-linux-arm64

.PHONY: build

build-all:
	cat ci/SUPPORTED_VERSIONS.txt | while read TARGET; do
		make GOTIFY_VERSION="$TARGET" FILE_SUFFIX="-for-gotify-$TARGET" build || exit 1
	done
