# README

```bash
# make
# insmod netlink_test.ko
# dmesg |grep test_netlink # 检查ko是否载入
[505819.566429] test_netlink_init
# gcc client.c -o client # 编译client
# ./client
# dmesg
...
[505944.039425] kernel recv from user: hello netlink!! # 显示client发送的消息
```