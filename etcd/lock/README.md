# readme
参考:
- [Etcd Lock详解](https://tangxusc.github.io/blog/2019/05/etcd-lock%E8%AF%A6%E8%A7%A3/)
- [6.2 分布式锁](https://chai2010.cn/advanced-go-programming-book/ch6-cloud/ch6-02-lock.html)
- [etcd v3客户端用法](https://yuerblog.cc/2017/12/12/etcd-v3-sdk-usage/)

基于etcd实现分布式锁

## pre build
```bash
cd /home/chen/test/etcd-master # 855eeb7, 2021-02-01, 3.5-pre
rm -rf client/v2/*_test.go
rm -rf client/v3/*_test.go
rm -rf client/v3/concurrency/*_test.go
```

## todo
- rwlock : [etcd-lock](https://github.com/flaviostutz/etcd-lock)