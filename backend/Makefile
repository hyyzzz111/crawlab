SHELL := /bin/bash
BASEDIR = $(shell pwd)

# build with verison infos
versionDir = "apiserver/pkg/version"
gitTag = $(shell if [ "`git describe --tags --abbrev=0 2>/dev/null`" != "" ];then git describe --tags --abbrev=0; else git log --pretty=format:'%h' -n 1; fi)
buildDate = $(shell TZ=Asia/Shanghai date +%FT%T%z)
gitCommit = $(shell git log --pretty=format:'%H' -n 1)
gitTreeState = $(shell if git status|grep -q 'clean';then echo clean; else  echo dirty; fi)

ldflags="-w -X ${versionDir}.gitTag=${gitTag} -X ${versionDir}.buildDate=${buildDate} -X ${versionDir}.gitCommit=${gitCommit} -X ${versionDir}.gitTreeState=${gitTreeState}"

all: gotool
	@go build -v -ldflags ${ldflags} .
clean:
	rm -f apiserver
	find . -name "[._]*.s[a-w][a-z]" | xargs -i rm -f {}
gotool:
	gofmt -w .
	go tool vet . |& grep -v vendor;true
ca:
	openssl req -new -nodes -x509 -out conf/server.crt -keyout conf/server.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=127.0.0.1/emailAddress=xxxxx@qq.com"

help:
	@echo "make - compile the source code"
	@echo "make clean - remove binary file and vim swp files"
	@echo "make gotool - run go tool 'fmt' and 'vet'"
	@echo "make ca - generate ca files"

.PHONY: clean gotool ca help


GOBIN ?= $(GOPATH)/bin
GOPROXY ?= https://goproxy.io
GOPROXY_ENV = GOPROXY=$(GOPROXY)

MOD_OUTDATED_COMMAND := ${GOBIN}/go-mod-outdated
ENUMER_COMMAND :=$(GOBIN)/enumer
GRPC_COMMAND :=${GOBIN}/protoc-gen-go
GOPATH :=$(GOPATH)
check-mod-outdated-command:
	@if [ ! -f $(MOD_OUTDATED_COMMAND) ]; then cd $(GOPATH);go install github.com/psampaz/go-mod-outdated;fi
mod-outdated: check-mod-outdated-command
	cd $(BASEDIR) && go list -u -m -json all | $(MOD_OUTDATED_COMMAND) -direct
mod-outdated-update: check-mod-outdated-command
	cd $(BASEDIR) && go list -u -m -json all | $(MOD_OUTDATED_COMMAND) -update -direct
mod-outdated-all: check-mod-outdated-command
	cd $(BASEDIR) && go list -u -m -json all | $(MOD_OUTDATED_COMMAND) -style markdown
check-enumber-command:
	 @if [ ! -f $(ENUMER_COMMAND) ]; then  cd $(GOPATH);go install github.com/yaziming/enumer;fi
check-grpc-command:
	@if [ ! -f $(GRPC_COMMAND) ]; then cd $(GOPATH);go install github.com/golang/protobuf/protoc-gen-go;fi

boot_development:check-enumber-command check-mod-outdated-command check-grpc-command

go-generate: check-enumber-command

	cd $(BASEDIR) && go generate *.go

go-env:
	go env -w GO111MODULE=on && go env -w "GOPROXY=https://goproxy.cn,https://goproxy.io,direct"
#	&& go env -w GOPRIVATE=*.corp.example.com

run-master:
	go run $(BASEDIR)/app/master/main.go start --config="$(BASEDIR)/../statics/config.yml"

