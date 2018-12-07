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
DOCKER_RUNNER        += -v $(SRCROOT_ON_HOST):$(SRCROOT_IN_CONTAINER)
DOCKER_GENERATOR     := infoblox/atlas-gentool:dev-gengorm
GENERATOR            := $(DOCKER_RUNNER) $(DOCKER_GENERATOR)

GENGORM_IMAGE      := $(IMAGE_REGISTRY)/atlas-gentool
GENGORM_DOCKERFILE := $(DOCKERFILE_PATH)/Dockerfile

.PHONY: default
default: vendor install

.PHONY: vendor
vendor:
	@dep ensure -vendor-only

.PHONY: vendor-update
vendor-update:
	@dep ensure

.PHONY: options
options:
	protoc -I. -I$(SRCPATH) -I./vendor \
		--gogo_out="Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:$(SRCPATH)" \
		options/gorm.proto

.PHONY: types
types:
	protoc --go_out=$(SRCPATH) types/types.proto

.PHONY: install
install:
	go install

.PHONY: example
example: default
	protoc -I. -I$(SRCPATH) -I./vendor -I./vendor/github.com/grpc-ecosystem/grpc-gateway \
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="engine=postgres,enums=string,gateway:$(SRCPATH)" \
		example/feature_demo/demo_multi_file.proto \
		example/feature_demo/demo_types.proto \
		example/feature_demo/demo_service.proto \
		example/feature_demo/demo_multi_file_service.proto

	protoc -I. -I$(SRCPATH) -I./vendor -I./vendor -I./vendor/github.com/grpc-ecosystem/grpc-gateway \
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="$(SRCPATH)" \
		example/user/user.proto

.PHONY: run-tests
run-tests:
	go test -v ./...
	go build ./example/user
	go build ./example/feature_demo

.PHONY: test
test: example run-tests

.PHONY: gentool
gentool: vendor
	@docker build -f $(GENGORM_DOCKERFILE) -t $(GENGORM_IMAGE):$(IMAGE_VERSION) .
	@docker tag $(GENGORM_IMAGE):$(IMAGE_VERSION) $(GENGORM_IMAGE):latest
	@docker image prune -f --filter label=stage=server-intermediate

.PHONY: gentool-example
gentool-example: gentool
	@$(GENERATOR) \
		--go_out="plugins=grpc:$(DOCKERPATH)" \
		--gorm_out="engine=postgres,enums=string,gateway:$(DOCKERPATH)" \
			example/feature_demo/demo_multi_file.proto \
			example/feature_demo/demo_types.proto \
			example/feature_demo/demo_service.proto \
			example/feature_demo/demo_multi_file_service.proto

	@$(GENERATOR) \
		--go_out="plugins=grpc:$(DOCKERPATH)" \
		--gorm_out="$(DOCKERPATH)" \
			example/user/user.proto

.PHONY: gentool-test
gentool-test: gentool-example run-tests

.PHONY: gentool-types
gentool-types:
	@$(GENERATOR) --go_out=$(DOCKERPATH) types/types.proto

.PHONY: gentool-options
gentool-options:
	@$(GENERATOR) \
                --gogo_out="Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:$(DOCKERPATH)" \
                options/gorm.proto
