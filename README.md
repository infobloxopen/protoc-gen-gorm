# protoc-gen-gorm

### Purpose

A protobuf (https://developers.google.com/protocol-buffers/) compiler plugin
designed to simplify the API calls needed to perform simple object persistence
tasks. This is currently accomplished by creating a secondary .pb.orm.go file
which contains sister objects to those generated in the standard .pb.go file
satisfying these criteria:

- These objects are GoLint compatible
- Go field decorators/tags can be defined by option in the .proto file for
GORM/SQL
- Converters between pb version and ORM version of the objects
- Field options for dropping fields from the pb object, or adding additional
field (not recommended, as it reduces clarity of .proto file)

### Usage

The protobuf compiler (protoc) is required.

You can install protoc, from code, with
```
mkdir tmp
cd tmp
git clone https://github.com/google/protobuf
cd protobuf
./autogen.sh
./configure
make
make check
sudo make install
```
This may require the installation of additional dependencies.
Then you will want the golang code generator
```
go get -u github.com/golang/protobuf/protoc-gen-go
```

You can run `make install` or `go install` directly.

Once installed, the `--gorm_out=.` option can be specified in a protoc
command to write the files to a given directory. Specifying inside the proto
file the option `option (orm.package) = "<package-name>";` will generate a
subdirectory with the given package name.
Not specifying a different package for the orm objects may cause naming
collisions.

Pending completion of a grpc server interceptor, ORMable proto objects can be
passed into rpc calls directly or as components of the request, the ConvertTo\*
function can be used to quickly convert to a GORM ready object, and the
ConvertFrom\* function can be used on an object read by GORM to be put into an
rpc ready type.

Running `make example` will compile two test proto files into the pb and
pb/orm directories and demonstrates most of the current capabilities of this
plugin.

**Known Issue:** The import statements will probably require alteration until
some issues are resolved.

### Limitations

Currently only proto3 is supported.

This project is currently in pre-alpha, and is expected to undergo "breaking"
(and fixing) changes
