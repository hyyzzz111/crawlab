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
goproxy := $(shell go env GOPROXY)

ENUMER_COMMAND :=$(GOBIN)/enumer
COMMAND_GEN_GRPC :=${GOBIN}/protoc-gen-go
GOPATH :=$(GOPATH)
COMMAND_GOBIN := GOBIN/gobin
COMMAN_MOD_OUTDATED := ${GOBIN}/go-mod-outdated

go-env:
	@if [ ! $(goproxy) ]; then go env -w GO111MODULE=on && go env -w "GOPROXY=https://goproxy.cn,https://goproxy.io,direct";fi;
pre: go-env
	@if [ ! -f $(COMMAND_GOBIN) ];  then GO111MODULE=off go get -u github.com/myitcv/gobin;fi
	@if [ ! -f $(COMMAN_MOD_OUTDATED) ]; then COMMAND_GOBIN github.com/psampaz/go-mod-outdated;fi
	@if [ ! -f $(COMMAND_GEN_GRPC) ]; then COMMAND_GOBIN github.com/golang/protobuf/protoc-gen-go;fi

outdated: pre
	@go list -u -m -json all | $(COMMAN_MOD_OUTDATED) -direct -update -style markdon


go-generate: pre
	go generate *.go

run-master:
	go run $(BASEDIR)/app/master/main.go start --config="$(BASEDIR)/../statics/config.yml"

