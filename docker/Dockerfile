FROM golang:1.10.0 AS builder

LABEL stage=server-intermediate

WORKDIR /go/src/github.com/infobloxopen/protoc-gen-gorm
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/usr/bin/protoc-gen-gorm main.go

FROM infoblox/atlas-gentool:v19.1 AS runner

COPY --from=builder /out/usr/bin/protoc-gen-gorm /usr/bin/protoc-gen-gorm
COPY --from=builder /go/src/github.com/infobloxopen/protoc-gen-gorm/options/*.proto /go/src/github.com/infobloxopen/protoc-gen-gorm/options
COPY --from=builder /go/src/github.com/infobloxopen/protoc-gen-gorm/types/*.proto /go/src/github.com/infobloxopen/protoc-gen-gorm/types

WORKDIR /go/src
