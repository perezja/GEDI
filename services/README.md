# Protoc

Using `protoc` to compile gRPC protobuf files that have dependencies must import relative to to the path of the directory shared with the source file.

For example, if `common/common.pb` and `discovery/discovery.pb` are both present in `github.com/gedilabs/services`

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative common/common.pb discovery/discovery.pb
```
