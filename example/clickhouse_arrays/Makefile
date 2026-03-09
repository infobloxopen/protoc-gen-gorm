all: protos

protos:
	@protoc \
		-I /usr/local/include \
		-I ${GOPATH}/src \
		--go_out=${GOPATH}/src \
		--gorm_out=engine=clickhouse:${GOPATH}/src \
		--proto_path=. \
		clickhouse_arrays.proto

protos_without_clickhouse:
	@protoc \
		-I /usr/local/include \
		-I ${GOPATH}/src \
		--go_out=${GOPATH}/src \
		--gorm_out=${GOPATH}/src \
		--proto_path=. \
		clickhouse_arrays.proto
