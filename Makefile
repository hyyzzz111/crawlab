
GO_PROJECT_PATH := $(shell pwd)/backend

GOBIN ?= $(GPPATH)/bin
GOPROXY ?= https://goproxy.io
GOPROXY_ENV = GOPROXY=$(GOPROXY)

MOD_OUTDATED_COMMAND := ${GOBIN}/go-mod-outdated
ENUMER_COMMAND :=${GOBIN}/enumer
GRPC_COMMAND :=${GOBIN}/protoc-gen-go

check-mod-outdated-command:
	@if [ ! -f $(MOD_OUTDATED_COMMAND) ]; then go install github.com/psampaz/go-mod-outdated;fi
mod-outdated: check-mod-outdated-command
	cd $(GO_PROJECT_PATH) && go list -u -m -json all | $(MOD_OUTDATED_COMMAND) -direct
mod-update: check-mod-outdated-command
	cd $(GO_PROJECT_PATH) && go list -u -m -json all | $(MOD_OUTDATED_COMMAND) -update -direct
mod-compatible: check-mod-outdated-command
	cd $(GO_PROJECT_PATH) && go list -u -m -json all | $(MOD_OUTDATED_COMMAND) -style markdown

check-enumber-command:
	 @if [ ! -f $(ENUMER_COMMAND) ]; then cd $(GO_PROJECT_PATH); go install github.com/yaziming/enumer;fi

check-grpc-command:
		 @if [ ! -f $(GRPC_COMMAND) ]; then cd $(GO_PROJECT_PATH); go install github.com/golang/protobuf/protoc-gen-go;fi

boot_devlopment:check-enumber-command check-mod-outdated-command check-grpc-command

go-generate: check-enumber-command
	cd $(GO_PROJECT_PATH) && go generate
