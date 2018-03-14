
default: build install

build:
	protoc -I . -I ${GOPATH}src -I ./vendor --gogo_out="Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:." orm/orm.proto

install:
	go install

example: default
	mkdir -p gen/gorm/example
	mkdir -p gen/pb/example
	protoc -I . -I ${GOPATH}src -I ./vendor --go_out="plugins=grpc:${GOPATH}src" example/*.proto
	protoc -I . -I ${GOPATH}src -I ./vendor --gorm_out="package_import_path=github.com/infobloxopen/protoc-gen-gorm/example/gorm:${GOPATH}src" example/*.proto
