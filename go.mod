module github.com/infobloxopen/protoc-gen-gorm

go 1.16

require (
	github.com/bufbuild/buf v0.37.0 // indirect
	github.com/denisenkom/go-mssqldb v0.9.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/infobloxopen/atlas-app-toolkit v0.24.1-0.20210416193901-4c7518b07e08
	github.com/jinzhu/inflection v1.0.0
	github.com/lib/pq v1.10.2
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/satori/go.uuid v1.2.0
	go.opencensus.io v0.22.6
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220118154757-00ab72f36ad5
	google.golang.org/grpc v1.43.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.0.0 // indirect
	google.golang.org/protobuf v1.27.1
	gorm.io/gorm v1.22.5
)

// let's use gorm v2!
// see https://github.com/infobloxopen/atlas-app-toolkit/pull/304
replace github.com/infobloxopen/atlas-app-toolkit v0.24.1-0.20210416193901-4c7518b07e08 => github.com/wk8/atlas-app-toolkit v1.1.2-0.20220215035731-0e577a49fa9d
