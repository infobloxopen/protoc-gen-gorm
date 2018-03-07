
default: build install

build:
	protoc -I . -I ${GOPATH}/src -I ./vendor --gogo_out="Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:." orm/orm.proto

install:
	go install

example: default
	mkdir -p pb/orm
	protoc -I . -I ${GOPATH}/src -I ./vendor --go_out="plugins=grpc:./pb" --gorm_out=./pb *.proto
