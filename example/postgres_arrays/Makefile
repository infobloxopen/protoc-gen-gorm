all: protos

protos:
	@protoc \
		-I /usr/local/include \
		-I ${GOPATH}/src \
		--go_out=${GOPATH}/src \
		--gorm_out=engine=postgres:${GOPATH}/src \
		--proto_path=. \
		postgres_arrays.proto

protos_without_postgres:
	@protoc \
		-I /usr/local/include \
		-I ${GOPATH}/src \
		--go_out=${GOPATH}/src \
		--gorm_out=${GOPATH}/src \
		--proto_path=. \
		postgres_arrays.proto
