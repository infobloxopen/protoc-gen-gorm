include Makefile.buf

GENTOOL_IMAGE := infoblox/atlas-gentool:dev-gengorm

GOPATH ?= $(HOME)/go
SRCPATH := $(patsubst %/,%,$(GOPATH))/src

PROJECT_ROOT := github.com/infobloxopen/protoc-gen-gorm

lint: $(BUF)
	buf lint

build: $(BUF)
	buf build

test: lint build
	go test -v ./...

generate: options/gorm.pb.go example/user/*.pb.go example/postgres_arrays/*.pb.go example/feature_demo/*.pb.go

options/gorm.pb.go: proto/options/gorm.proto
	buf generate --template proto/options/buf.gen.yaml --path proto/options

# TODO: gorm files are not being built by buf generate yet, use docker for now

example/feature_demo/*.pb.go: example/feature_demo/*.proto
	buf generate --template example/feature_demo/buf.gen.yaml --path example/feature_demo

example/user/*.pb.go: example/user/*.proto
	buf generate --template example/user/buf.gen.yaml --path example/user

example/postgres_arrays/*.pb.go: example/postgres_arrays/*.proto
	buf generate --template example/postgres_arrays/buf.gen.yaml --path example/postgres_arrays

install:
	go install -v .

gentool:
	docker build -f docker/Dockerfile -t $(GENTOOL_IMAGE) .
	docker image prune -f --filter label=stage=server-intermediate

generate-gentool: SRCROOT_ON_HOST      := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
generate-gentool: SRCROOT_IN_CONTAINER := /go/src/$(PROJECT_ROOT)
generate-gentool: DOCKER_RUNNER        := docker run --rm
generate-gentool: DOCKER_RUNNER        += -v $(SRCROOT_ON_HOST):$(SRCROOT_IN_CONTAINER)
generate-gentool: DOCKER_GENERATOR     := infoblox/atlas-gentool:dev-gengorm
generate-gentool: GENERATOR            := $(DOCKER_RUNNER) $(DOCKER_GENERATOR)
generate-gentool: #gentool
	$(DOCKER_RUNNER) \
		$(GENTOOL_IMAGE) \
		--go_out="plugins=grpc:$(DOCKERPATH)" \
		--gorm_out="engine=postgres,enums=string,gateway:$(DOCKERPATH)" \
			feature_demo/demo_multi_file.proto \
			feature_demo/demo_types.proto \
			feature_demo/demo_service.proto \
			feature_demo/demo_multi_file_service.proto
