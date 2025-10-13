# README
ref:
- [打造高性能实时通信：Go语言SSE实现指南](https://zhuanlan.zhihu.com/p/1890020818458748395)

## FAQ
### chrome f12时sse请求未显示出"EventStream"标签页
"EventStream"标签页只有有数据时才显示

解决方法:
1. 具体逻辑前先返回一个时间戳

### sse连接返回409时, chrome上该请求还是显示200
使用了manager.HasClient(client), 检测到有client重复连接时返回409, 但chrome上该请求响应码是200, 也试过返回400, 浏览器依旧显示200

浏览器sse机制

### 刷新浏览器后sse容易连不上
时序:
1. 浏览器sse连接成功
1. 浏览器刷新

    1. js重新发起sse连接, 检测到client重复连接返回409
    1. 旧sse连接断开

解决: 直接使用随机client id