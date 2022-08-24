FROM golang:1.16.5 AS builder

LABEL stage=server-intermediate

WORKDIR /go/src/github.com/infobloxopen/protoc-gen-gorm
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/usr/bin/protoc-gen-gorm main.go

FROM infoblox/atlas-gentool:v25.1-6-g1ac2da0 AS runner

COPY --from=builder /out/usr/bin/protoc-gen-gorm /usr/bin/protoc-gen-gorm
COPY --from=builder /go/src/github.com/infobloxopen/protoc-gen-gorm/proto /go/src/github.com/infobloxopen/protoc-gen-gorm/proto
COPY --from=builder /go/src/github.com/infobloxopen/protoc-gen-gorm/third_party /go/src/github.com/infobloxopen/protoc-gen-gorm/third_party

WORKDIR /go/src
ENTRYPOINT ["protoc", "-I.", \
    # required import paths for protoc-gen-openapiv2 plugin
    "-Igithub.com/infobloxopen/grpc-gateway", \
    "-Igithub.com/infobloxopen/grpc-gateway/protoc-gen-openapiv2/options", \
    # required import paths for protoc-gen-validate plugin
    "-Igithub.com/envoyproxy/protoc-gen-validate/validate", \
    # required import paths for go-proto-validators plugin
    "-Igithub.com/mwitkow/go-proto-validators", \
    # required import paths for protoc-gen-gorm plugin, Should add /proto path once updated
    "-Igithub.com/infobloxopen/protoc-gen-gorm", \
    # required import paths for protoc-gen-atlas-query-validate plugin
    "-Igithub.com/infobloxopen/protoc-gen-atlas-query-validate", \
    # required import paths for protoc-gen-preprocess plugin
    "-Igithub.com/infobloxopen/protoc-gen-preprocess", \
    # required import paths for protoc-gen-atlas-validate plugin
    "-Igithub.com/infobloxopen/protoc-gen-atlas-validate" \
]
