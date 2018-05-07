GOPATH ?= $(HOME)/go
SRCPATH := $(patsubst %/,%,$(GOPATH))/src

default: build install

build:
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
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="$(SRCPATH)" \
		example/feature_demo/demo_types.proto example/feature_demo/point_service.proto

	protoc -I. -I$(SRCPATH) -I./vendor \
		-I$(SRCPATH)/github.com/google/protobuf/src/ \
		-I$(SRCPATH)/github.com/grpc-ecosystem/grpc-gateway \
		-I$(SRCPATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="$(SRCPATH)" \
		example/contacts/contacts.proto

test: example
	go test ./...
	go build ./example/contacts
	go build ./example/feature_demo
