# protoc-gen-gorm

### 目的

一个protobuf（https://developers.google.com/protocol-buffers/）编译器插件

设计用于简化执行简单对象持久性所需的API调用任务。当前，这是通过创建辅助.pb.gorm.go文件来完成的

其中包含标准.pb.go文件中生成的姐妹对象

满足这些条件:

- Go字段装饰器/标签可以通过.proto文件中的选项定义，用于GORM / SQL
- 有一些选项可用于从PB对象中删除字段，或添加其他字段（一般不建议使用，因为这会降低.proto文件的清晰度）
- 包括对象的PB版本和ORM版本之间的转换器
- 挂钩接口允许自定义应用程序逻辑

### 先决条件

#### 1. Protobuf 编译器

需要protobuf编译器（protoc）

[官方指示](https://github.com/google/protobuf#protocol-compiler-installation)

[Abbreviated version](https://github.com/grpc-ecosystem/grpc-gateway#installation)

#### 2. Golang Protobuf代码生成器

获取golang protobuf代码生成器：

```
go get -u github.com/golang/protobuf/protoc-gen-go
```

#### 3. 供应商依赖性

检索并安装此项目的供应商依赖项 [dep](https://github.com/golang/dep):

```
dep ensure
```

### 安装

- 方式一

```go
go get github.com/infobloxopen/protoc-gen-gorm
```

- 方式二

```go
git clone https://github.com/infobloxopen/protoc-gen-gorm.git
make install || go install
```

### 用法

安装后，`--gorm_out=.` 或者 `--gorm_out=${GOPATH}src`可以在protoc命令中指定option来生成.pb.gorm.go文件。

任何带有`option (gorm.opts).ormable = true`选项的消息类型都将具有

以下是自动生成的：

- 具有 `ORM` 兼容类型和后缀 `ORM` 的结构

- GORM[tags](https://gorm.io/zh_CN/docs/models.html)通过字段选项构建 `[(gorm.field).tag = {..., tag: value, ...}]`.
- 一个 `{PbType}.ToORM` 和  `{TypeORM}.ToPB` 函数
- 从选项 `option (gorm.opts) = {include: []}` 中添加的其他未暴露字段，它们是内置类型，
  - 内置类型, 例如
    - `{type: "int32", name: "secret_key"}` 
  - 导入的类型, 例如
    - `{type: "StringArray", name: "array", package:"github.com/lib/pq"}`
- 接受protobuf版本(例如来自API调用)和 `context` (与multiaccount选项一起使用，用于[operators](https://github.com/infobloxopen/atlas-app-toolkit#collection-operators)以及用于集合运算符)的准系统的C/U/R/D/L处理程序, 然后gorm.DB使用对象在数据库上执行基本操作
- 每次转换之前和之后的接口挂钩，可以实现添加自定义处理。

任何带有 `option (gorm.server).autogen = true` 选项的服务都将生成基本的grpc服务器：

- 对于名称以`Create|Read|Update|Delete`开头的服务方法，生成的实现将调用基本CRUD处理程序。
- 对于其他方法 `return &MethodResponse{}, nil ` 将生成nil stub。

为了正确生成CRUD方法，您需要遵循特定的约定:

- 用于Create和Update方法的请求消息在名为 `payload` 的字段中应具有Ormable Type，对于Read和Delete方法，则需要一个 `id` 字段。列表请求中不需要任何内容。
- 用于创建，读取和更新的响应消息在名为 `results` 的字段中需要一个Ormable Type，对于在List中列出一个名为 `results` 的重复Ormable Type的响应消息。
- 删除方法需要使用 `(gorm.method).object_type`  选项来指示应删除的Ormable Type，并且没有响应类型要求。

要自定义生成的服务器，请将其嵌入到新类型中并覆盖任何所需的功能。

如果不符合约定，则为CRUD方法生成存根。你可以看这个例子
[feature_demo/demo_service](example/feature_demo/demo_service.proto)

要利用数据库的特定功能，请在生成过程中使用 `--gorm_out="engine={postgres,...}:{path}"`. 当前只有Postgres支持特殊类型，其他任何选择都将作为默认类型。

生成的代码还可以与在以下代码中提供的grpc服务器gorm交易中间件集成: [atlas-app-toolkit](https://github.com/infobloxopen/atlas-app-toolkit#middlewares), 使用服务级别选项 `option (gorm.server).txn_middleware = true`.

### 例子

示例 .proto文件和生成的.pb.gorm.go文件都包含在 `example` 目录中。

用户文件包含来自[GORM](https://gorm.io/zh_CN/docs)中文文档的模型示例，[feature_demo/demo_types](feature_demo/demo_types.proto)演示类型处理和multi_account函数，[feature_demo/demo_service](feature_demo/demo_service.proto) 显示服务自动生成。

如果要测试更改选项和字段的效果，运行  `make example`  将重新编译所有这些测试原型文件。

### 支持的类型

在原始文件中，支持以下类型：

- 标准基本类型`uint32`，`uint64`，`int32`，`int64`，`float`，在ORM级别将`double`，`bool`，`string`映射到相同类型

- [google wrapper types](https://github.com/golang/protobuf/blob/master/ptypes/wrappers/wrappers.proto)

	`google.protobuf.StringValue`，`.BoolValue` ，`.UInt32Value`，`.FloatValue`等。

	在ORM级别上映射到内部类型的指针，例如`*string`, `*bool`, `*uint32`, `*float`

- [google timestamp type](https://github.com/golang/protobuf/blob/master/ptypes/timestamp/timestamp.proto)
 `google.protobuf.Timestamp` 对应的ORM级别上是 `time.Time` 
 
- ‧自定义包装类型 `gorm.types.UUID` 和 `gorm.types.UUIDValue` ，它们包装在ORM级别将字符串转换为 `uuid.UUID` 和 `*uuid.UUID`, 来自https://github.com/satori/go.uuid的空值或缺失`gorm.types.UUID` 将成为ZeroUUID(`00000000-0000-0000-0000-000000000000`) 

- 自定义包装器类型 `gorm.types.JSONValue` ，它将字符串包装在protobuf中

   包含任意JSON并转换为 `postgres.Jsonb`  GORM类型

   (https://github.com/jinzhu/gorm/blob/master/dialects/postgres/postgres.go#L59)

   如果Postgres是选定的数据库引擎，否则当前已删除

- 自定义包装器类型 `gorm.types.InetValue` ，它包装字符串并转换为类型。ORM级别的 `types.Inet` ，使用golang `net.IPNet` 类型来保存与扫描兼容的ip地址和掩码，IPv4和IPv6兼容以及写入数据库所需的值函数。与JSONValue一样，如果数据库引擎不是Postgres，则当前删除.

- 可以从同一程序包内的其他.proto文件(协议调用)或程序包之间导入其他类型。可以在同一程序包中正确生成所有关联，但是交叉包仅 `belongs-to ` 和 `many-to-many ` 将起作用.

- github.com/lib/pq可以为Postgres自动处理一些重复的类型，并且只要将引擎设置为postgres，就会创建往返于映射(你可以通过查看[example/postgres_arrays/postgres_arrays.proto](example/postgres_arrays/postgres_arrays.proto)这个例子)

  - []bool: pq.BoolArray
  - []float64: pq.Float64Array
  - []int64: pq.Int64Array
  - []string: pq.StringArray

### 关联

该插件支持以下[GORM](https://gorm.io/zh_CN/docs/)

- Belongs-To
- Has-One
- Has-Many
- Many-To-Many

注意：==目前不支持多态关联==。

关联是通过添加一些可配置消息类型（单个或重复）的字段来定义的。

```golang
message Contact {
    option (gorm.opts).ormable = true;
    uint64 id = 1;
    string name = 2;
    repeated Email emails = 3;
    Address home_address = 4;
}
```

Has-One是单个消息类型的默认关联。

```golang
message Contact {
    option (gorm.opts).ormable = true;
    uint64 id = 1;
    string first_name = 2;
    string middle_name = 3;
    string last_name = 4;
    Address address = 5;
}

message Address {
    option (gorm.opts).ormable = true;
    string address = 1;
    string city = 2;
    string state = 3;
    string zip = 4;
    string country = 5;
}
```

在字段上设置`(gorm.field).belongs_to` 选项以定义Belongs-To。

```golang
message Contact {
    option (gorm.opts).ormable = true;
    uint64 id = 1;
    string first_name = 2;
    string middle_name = 3;
    string last_name = 4;
    Profile profile = 5 [(gorm.field).belongs_to = {}];
}

message Profile {
    option (gorm.opts).ormable = true;
    uint64 id = 1;
    string name = 2;
    string notes = 3;
}
```

Has-Many是重复消息类型的默认关联。

```golang
message Contact {
    option (gorm.opts).ormable = true;
    uint64 id = 1;
    string first_name = 2;
    string middle_name = 3;
    string last_name = 4;
    repeated Email emails = 5;
}

message Email {
    option (gorm.opts).ormable = true;
    string address = 1;
    bool is_primary = 2;
}
```

在字段上设置 `(gorm.field).many_to_many` 选项，以定义“Many-To-Many”。

```golang
message Contact {
    option (gorm.opts).ormable = true;
    uint64 id = 1;
    string first_name = 2;
    string middle_name = 3;
    string last_name = 4;
    repeated Group groups = 5 [(gorm.field).many_to_many = {}];
}

message Group {
    option (gorm.opts).ormable = true;
    uint64 id = 1;
    string name = 2;
    string notes = 3;
}
```

对于每种关联类型，除了Many-To-Many外，会自动创建指向主键（关联键）的外键

如果它们不存在于原始消息中，则其名称对应于GORM[默认](https://gorm.io/zh_CN/docs/belongs_to.html)外键名称。

GORM关联标签也会自动插入。

#### 客制化

- 对于每种关联类型，您可以通过设置 `Foreignkey` 和 `association_foreignkey` 选项来覆盖默认外键和关联键。

- 对于每种关联类型，您都可以覆盖创建/更新记录的默认行为。可以根据以下情况创建/更新其引用`association_autoupdate`, `association_autocreate` and `association_save_reference` 选项, 有关更多信息，请参阅官方文档[GORM](https://gorm.io/zh_CN/docs/associations.html)。
- 您可以为每种关联类型设置`preload`选项, 有关更多信息，请参阅官方文档[GORM](https://gorm.io/zh_CN/docs/preload.html#Auto-Preloading)。
- 默认情况下，删除和替换更新子关联时。可以切换此功能，使其与gorm处理该功能的方式相同，请参见[GORM](https://gorm.io/docs/associations.html)。这可以通过添加其中一个gorm关联处理程序选项来完成，这些选项是
  - 添加关联 : [添加关联](https://gorm.io/zh_CN/docs/associations.html#添加关联)
  - 删除关联 :[删除关联](https://gorm.io/zh_CN/docs/associations.html#删除关联)
  - 替换关联: [替换关联](https://gorm.io/zh_CN/docs/associations.html#替换关联)
- 对于Has-Many，您可以设置`position_field`，以便如果原始消息中不存在其他字段来创建其他字段以维持关联顺序。相应的CRUDL处理程序会执行所有必要的工作以维持顺序。
- 对于自动创建的外键和位置字段，您可以通过设置 `foreignkey_tag` 和 `position_field_tag` 选项来分配GORM标签。
- 对于多对多，您可以通过设置 `jointable` , `jointable_foreignkey` , `association_jointable_foreignkey` 来覆盖默认的联接表名称和列名称。

查看关联用法的真实示例--> [user](example/user/user.proto)

### 局限性

目前仅支持proto3。

该项目目前正在开发中，预计将进行“重大”更改（和修复）

## 贡献

欢迎Pull requests！

任何新功能都应包括对该功能的测试。

在进行更改之前，请使用以下命令运行测试：

```
make gentool-test
```
这将在具有特定已知依赖版本的docker容器中运行测试。

在测试运行之前，它们会生成代码。作为拉取请求的一部分，提交任何新的和修改的生成的代码。