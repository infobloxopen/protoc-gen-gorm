module github.com/infobloxopen/protoc-gen-gorm/example

go 1.17

replace (
	github.com/infobloxopen/protoc-gen-gorm => ..
	github.com/infobloxopen/protoc-gen-gorm/types => ../types
)

require (
	github.com/golang/protobuf v1.5.4
	github.com/google/go-cmp v0.5.9
	github.com/infobloxopen/atlas-app-toolkit/v2 v2.2.1-0.20240313220428-5449c0c2a27f
	github.com/infobloxopen/protoc-gen-gorm v0.0.0-00010101000000-000000000000
	github.com/infobloxopen/protoc-gen-gorm/types v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.2
	github.com/satori/go.uuid v1.2.0
	go.opencensus.io v0.23.0
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1
	google.golang.org/grpc v1.56.3
	google.golang.org/protobuf v1.33.0
	gorm.io/gorm v1.23.2
)

require (
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	gorm.io/datatypes v1.0.6 // indirect
	gorm.io/driver/mysql v1.3.2 // indirect
)
