GOPATH ?= $(HOME)/go
SRCPATH := $(patsubst %/,%,$(GOPATH))/src

default: build install

build:
	protoc -I. -I$(SRCPATH) -I./vendor \
		--gogo_out="Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:." \
		options/gorm.proto

install:
	go install

example: default
	protoc -I. -I$(SRCPATH) -I./vendor \
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="$(SRCPATH)" \
		example/feature_demo/*.proto

	protoc -I. -I$(SRCPATH) -I./vendor \
		-I$(SRCPATH)/github.com/google/protobuf/src/ \
		-I$(SRCPATH)/github.com/grpc-ecosystem/grpc-gateway \
		-I$(SRCPATH)/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out="plugins=grpc:$(SRCPATH)" --gorm_out="$(SRCPATH)" \
		example/contacts/contacts.proto

	goimports -w ./example

test: example
	go test ./...
	go build ./example/contacts
	go build ./example/feature_demo
