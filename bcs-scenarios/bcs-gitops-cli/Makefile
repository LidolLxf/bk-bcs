MAKE:=make
bcs_edition?=inner_edition

# init the build information
ifdef HASTAG
        GITTAG=${HASTAG}
else
        GITTAG=$(shell git describe --always)
endif

VERSION=${GITTAG}-$(shell date +%y.%m.%d)
BUILDTIME = $(shell date '+%Y-%m-%d %T')
GITCOMMIT = $(shell git rev-parse HEAD)
GOVERSION = $(shell go version)

LDFLAG=-ldflags "-s -w -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITCOMMIT}' \
 -X 'github.com/Tencent/bk-bcs/bcs-common/common/version.GoVersion=${GOVERSION}'"

linux:
	mkdir -p ./bin
	go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o bin/powerapp-linux-amd64 ./main.go

darwin:
	mkdir -p ./bin
	go mod tidy && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${LDFLAG} -o bin/powerapp-darwin-arm64 ./main.go

clean:
	rm -rf ./bin
