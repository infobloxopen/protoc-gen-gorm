GOPATH ?= $(HOME)/go
SRCPATH := $(patsubst %/,%,$(GOPATH))/src

default: vendor options install

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

install:
	go install

example: default
	protoc -I. -I$(SRCPATH) -I./vendor \
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="engine=postgres:$(SRCPATH)" \
		example/feature_demo/demo_types.proto example/feature_demo/demo_service.proto

	protoc -I. -I$(SRCPATH) -I./vendor \
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="$(SRCPATH)" \
		example/user/user.proto

test: example
	go test -v ./...
	go build ./example/user
	go build ./example/feature_demo
