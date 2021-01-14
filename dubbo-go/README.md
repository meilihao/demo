# readme(未完成, 依赖难弄, 待dubbo-go v3.0发布后再弄)
dubbo-go 1.5.5

example base [dubbo-samples/golang](https://github.com/dubbogo/dubbo-samples/blob/master/golang):
- general/grpc
- registry/etcd

## prepare
```bash
# --- generate pb from [protoc-gen-dubbo](https://github.com/apache/dubbo-go/blob/master/protocol/grpc/protoc-gen-dubbo/examples/Makefile)
# go get -u github.com/apache/dubbo-go/protocol/grpc/protoc-gen-dubbo
# cd pkg/pb
# protoc -I . *.proto --dubbo_out=plugins=grpc+dubbo:.
```

## FAQ
### undefined: grpc.SupportPackageIsVersion6 和 undefined: grpc.ClientConnInterface 解决办法
参考:
- [使用 etcd 和 grpc 遇到的版本冲突的那些事儿](https://learnku.com/articles/43758)

降级protoc-gen-go的版本

> 注意：使用命令 go get -u github.com/golang/protobuf/protoc-gen-go 的效果是安装最新版的protoc-gen-go

降低protoc-gen-go的具体办法，在终端运行如下命令，这里降低到版本 v1.2.0
```bash
GIT_TAG="v1.2.0"
go get -d -u github.com/golang/protobuf/protoc-gen-go
git -C "$(go env GOPATH)"/src/github.com/golang/protobuf checkout $GIT_TAG
go install github.com/golang/protobuf/protoc-gen-go
```

### undefined: balancer.PickOptions 和 undefined: balancer.PickOptions
参考:
- [使用 etcd 和 grpc 遇到的版本冲突的那些事儿](https://learnku.com/articles/43758)

将google.golang.org/grpc为v1.26.0