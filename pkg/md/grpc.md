# gRPC简介

gRPC**是一个高性能、开源的 RPC（远程过程调用）框架**。
特点：
- **高性能**：基于 QUIC 协议，还利用了 HTTP2 的双向流特性。此外，gRPC支持流控和压缩，进一步提高了性能。
- **跨语言**：基本上主流的编程语言都有对应的 gRPC 实现，所以是异构系统的第一选择。
- **开源**：强大的开源社区。

## gRPC 与 IDL
gRPC**使用IDL来定义客户端和服务端之间的通信格式。**

IDL 全名接口描述语言，是一种用来描述软件组件接口的计算机语言，是跨平台的基础。

**gRPC 的 IDL 是和语言平台无关的，通过一个编译过程，生成各个语言的代码。**


# Protobuf 入门
Protobuf 是一种由 Google 开发的数据序列化协议，用于高效地序列化和反序列化结构化数据，它被广泛应用于跨平台通信、数据存储和 RPC（远程过程调用）等领域。

**gRPC 使用了 Protobuf 来作为自己的 IDL 语言。**

怎么理解 gRPC IDL 和 Protobuf？
- gRPC 先规定了 IDL。
- 而后 gRPC 需要一门编程语言来作为 IDL 落地的形式，因此选择了 Protobuf。


## Protobuf 的优势
- **高效性**：Protobuf 序列化和反序列化的速度非常快，压缩效率高，可以大大降低网络传输和数据存储的开销，在所有的序列化协议和反序列化协议里面名列前矛。
- **跨平台和语言无关性**：Protobuf 支持多种编程语言，包括 C、C++、Java、Python等，使得不同平台和语言的应用程序可以方便地进行数据交换。
- **强大的扩展性**：Protobuf 具有灵活的消息格式定义方式，可以方便地扩展和修改数据结构，而无需修改使用该数据的代码。
- **丰富的 API 支持**：Protobuf 提供了丰富的 API 和工具，包括编译器、代码生成器、调试工具等，方便开发人员进行使用和管理。

## Protobuf 基本原理
- Protobuf 使用**二进制格式**进行序列化和反序列化，与之对应的就是 JSON 这种文本格式。
- 它定义了一种标准的消息格式，即消息类型，**用于表示结构化数据**，举例来说，一个 User 这种对象，究竟该怎么表达。
- 消息类型由字段组成，**每个字段都有唯一的标签和类型。**
```protobuf
syntax = "proto3";

message SearchRequest {
  required string query = 1;
  required int32 page_number = 2;
  required int32 result_per_page = 3;
}
```

### protoc 命令
- **--proto_path=**:指定 .proto 文件的路径，填写 . 表示当前目录下。
- **--go_out=**:表示编译后的文件存放路径，如果编译的是C#，则使用 --csharp_out。
- **--go_opt**：用于设置 Go 编译选项。
- **--grpc_out**：指定 gRPC 代码生成输出目录。
- **--plugin**：指定代码生成插件。

```protobuf
syntax = "proto3";

// protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative user.proto
option go_package = "grpc-live/rpc";

message User {
    // 不能从零开始
    // protobuf 对前几个字段有性能优化
    int64 id = 1;
    string name = 2;
    // 编号可以不连续
    string avatar = 4;
    // int64 age = 5;

    map<string, string> attributes = 6;
    // 可选类型，创建为指针类型
    optional int32 age = 7;
    Address address = 8;
    // repeated 表示数组
    repeated string nicknames = 9;

    // oneof 表示只能选择一个
    oneof contacts {
        string email = 10;
        string phone = 11;
    }
    Gender gender = 12;
}

message Address {
    string city = 1;
    string street = 2;
}

// 枚举类型，从0开始
enum Gender {
    UnKnown = 0;
    Male = 1;
    Female = 2;
}

// 服务
service UserService {
    rpc GetById (GetByIdRequest) returns (GetByIdResponse);
}

// 请求体
message GetByIdRequest {
    int64 id = 1;
}

// 响应体
message GetByIdResponse {
    User user = 1;
}
```