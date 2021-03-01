GOPATH ?= $(HOME)/go
SRCPATH := $(patsubst %/,%,$(GOPATH))/src

PROJECT_ROOT := github.com/infobloxopen/protoc-gen-gorm

DOCKERFILE_PATH := $(CURDIR)/docker
IMAGE_REGISTRY ?= infoblox
IMAGE_VERSION  ?= dev-gengorm

# configuration for the protobuf gentool
SRCROOT_ON_HOST      := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
SRCROOT_IN_CONTAINER := /go/src/$(PROJECT_ROOT)
DOCKERPATH           := /go/src
DOCKER_RUNNER        := docker run --rm
DOCKER_RUNNER        += -v $(SRCROOT_ON_HOST):$(SRCROOT_IN_CONTAINER) -w $(SRCROOT_IN_CONTAINER)
DOCKER_GENERATOR     := infoblox/docker-protobuf:latest
PROTOC_FLAGS         := -Ivendor -Iexample -Iproto \
		-Ivendor/github.com/grpc-ecosystem/grpc-gateway/v2 \
		--gorm_out="engine=postgres,enums=string,gateway,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor,Mprotoc-gen-openapiv2/options/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options:/go/src" \
		--go_out="Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor,Mprotoc-gen-openapiv2/options/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options:/go/src"
GENERATOR            := $(DOCKER_RUNNER) $(DOCKER_GENERATOR) $(PROTOC_FLAGS)

.PHONY: default
default: vendor install

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: options
options-proto:
	protoc $(PROTOC_FLAGS) \
		options/gorm.proto

options: options-proto
	go build ./options

.PHONY: types
types:
	protoc --go_out=$(SRCPATH) types/types.proto

.PHONY: install
install:
	go install


.PHONY: example
example:
	protoc -I. $(PROTOC_FLAGS) \
		example/user/user.proto

	protoc -I. $(PROTOC_FLAGS) \
		example/feature_demo/demo_types.proto \
		example/feature_demo/demo_multi_file_service.proto \
		example/feature_demo/demo_multi_file.proto \
		example/feature_demo/demo_service.proto

.PHONY: run-tests
run-tests:
	go mod tidy
	go mod vendor
	go test -v ./...
	go build ./example/user
	go build ./example/feature_demo

.PHONY: test
test: example run-tests


.PHONY: gentool-example
gentool-example:
	$(GENERATOR) -I. \
			example/feature_demo/demo_multi_file_service.proto \
			example/feature_demo/demo_multi_file.proto \
			example/feature_demo/demo_service.proto

gentool-demotypes:
	$(GENERATOR) -I. \
			example/feature_demo/demo_types.proto

gentool-user:
	$(GENERATOR) -I. \
			example/user/user.proto

.PHONY: gentool-test
gentool-test: gentool-user gentool-demotypes gentool-example
	$(MAKE) run-tests

.PHONY: gentool-types
gentool-types:
	@$(GENERATOR) --go_out=$(DOCKERPATH) types/types.proto

.PHONY: gentool-options
gentool-options:
	$(GENERATOR) \
                options/gorm.proto
