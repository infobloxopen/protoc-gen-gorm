include Makefile.buf

GENTOOL_IMAGE := infoblox/atlas-gentool:dev-gengorm

GOPATH ?= $(HOME)/go
SRCPATH := $(patsubst %/,%,$(GOPATH))/src

PROJECT_ROOT := github.com/infobloxopen/protoc-gen-gorm

lint: $(BUF)
	buf lint

build: $(BUF)
	buf build

test: build
	go test -v ./...

regenerate: clean-gen generate

clean-gen:
	cd example/postgres_arrays && rm -f *.pb.gorm.go && rm -f *.pb.go
	cd example/user && rm -f *.pb.gorm.go && rm -f *.pb.go
	cd example/feature_demo && rm -f *.pb.gorm.go && rm -f *.pb.go
	cd options && rm -f *.pb.gorm.go && rm -f *.pb.go
	cd types && rm -f types.pb.go

generate: build options/gorm.pb.go types/types.pb.go install example/user/*.pb.go example/postgres_arrays/*.pb.go example/feature_demo/*.pb.go

options/gorm.pb.go: proto/options/gorm.proto
	buf generate --template proto/options/buf.gen.yaml --path proto/options

types/types.pb.go: proto/types/types.proto
	buf generate --template proto/types/buf.gen.yaml --path proto/types

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
generate-gentool: DOCKERPATH           := /go/src
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
			feature_demo/demo_multi_file_service.proto \
			feature_demo/demo_service.proto \
			feature_demo/demo_types.proto
	$(DOCKER_RUNNER) \
		$(GENTOOL_IMAGE) \
		--go_out="plugins=grpc:$(DOCKERPATH)" \
		--gorm_out="engine=postgres,enums=string,gateway:$(DOCKERPATH)" \
			user/user.proto
	$(DOCKER_RUNNER) \
		$(GENTOOL_IMAGE) \
		--go_out="plugins=grpc:$(DOCKERPATH)" \
		--gorm_out="engine=postgres,enums=string,gateway:$(DOCKERPATH)" \
			postgres_arrays/postgres_arrays.proto

build-local:
	rm -rf example/feature_demo/github.com/
	rm -rf example/feature_demo/google.golang.org
	go install
	protoc --proto_path . \
	-I./proto/ \
	-I./third_party/proto/ \
	-I=. example/feature_demo/demo_multi_file.proto \
	example/feature_demo/demo_service.proto --gorm_out="engine=postgres,enums=string,gateway:./example/feature_demo" --go_out=./example/feature_demo

build-user-local:
	rm -rf example/user/github.com/
	rm -rf example/user/google.golang.org
	go install
	protoc --proto_path . \
	-I./proto/ \
	-I./third_party/proto/ \
	example/user/user.proto --gorm_out="engine=postgres,enums=string,gateway:./example/user" --go_out=./example/user

build-postgres-local:
	rm -rf example/postgres_arrays/github.com/
	rm -rf example/postgres_arrays/google.golang.org
	go install
	protoc --proto_path . \
	-I./proto/ \
	-I./third_party/proto/ \
	example/postgres_arrays/postgres_arrays.proto --gorm_out="engine=postgres,enums=string,gateway:./example/postgres_arrays" --go_out=./example/postgres_arrays
