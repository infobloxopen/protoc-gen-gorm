module github.com/infobloxopen/protoc-gen-gorm/example

go 1.16

replace (
	github.com/infobloxopen/protoc-gen-gorm => ..
	github.com/infobloxopen/protoc-gen-gorm/types => ../types
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.5
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.4.0 // indirect
	github.com/infobloxopen/atlas-app-toolkit v0.24.1-0.20210416193901-4c7518b07e08
	github.com/infobloxopen/protoc-gen-gorm v0.0.0-00010101000000-000000000000
	github.com/infobloxopen/protoc-gen-gorm/types v0.0.0-00010101000000-000000000000
	github.com/jinzhu/gorm v1.9.16
	github.com/lib/pq v1.10.9
	github.com/satori/go.uuid v1.2.0
	go.opencensus.io v0.22.6
	google.golang.org/genproto v0.0.0-20210426193834-eac7f76ac494
	google.golang.org/grpc v1.37.0
	google.golang.org/grpc/examples v0.0.0-20210601155443-8bdcb4c9ab8d // indirect
	google.golang.org/protobuf v1.26.0
)
