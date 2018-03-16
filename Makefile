
default: build install

build:
	protoc -I . -I ${GOPATH}src -I ./vendor --gogo_out="Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:." options/orm.proto

install:
	go install

example: default
	protoc -I . -I ${GOPATH}src -I ./vendor \
	 --go_out="plugins=grpc:${GOPATH}src" example/feature_demo/*.proto \
	  --gorm_out="${GOPATH}src" example/feature_demo/*.proto


	protoc -I . -I ${GOPATH}src -I ./vendor \
		-I=$(GOPATH)/src/github.com/google/protobuf/src/ \
    -I=$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway \
	 -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	 --go_out="plugins=grpc:${GOPATH}src" example/contacts/contacts.proto \
	  --gorm_out="${GOPATH}src" example/contacts/contacts.proto
