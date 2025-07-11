# README

## 证书生成工具
- [cfssl](https://github.com/cloudflare/cfssl)
- [使用cfssl配置nginx https](https://blog.csdn.net/calm0406/article/details/127838421)
- [CFSSL 生成证书](https://juejin.cn/post/6939380451484106759)

## issues
- [tls.RequireAndVerifyClientCert不起作用](https://github.com/lucas-clemente/quic-go/issues/1366)

## cfssl
ref:
- [How to use cfssl to create self signed certificates](https://rob-blackbourn.medium.com/how-to-use-cfssl-to-create-self-signed-certificates-d55f76ba5781)

    包含Intermediate CA

1. 下载[cfssl](https://github.com/cloudflare/cfssl/releases)

    - cfssl_1.6.5_linux_amd64 -> /usr/bin/cfssl
    - cfssljson_1.6.5_linux_amd64 -> /usr/bin/cfssljson
    - cfssl-certinfo_1.6.5_linux_amd64 -> /usr/bin/cfssl-certinfo
    - mkbundle_1.6.5_linux_amd64 -> /usr/bin/mkbundle

1. 生成config.json

    ```bash
    $ cfssl print-defaults config > config.json
    {
        "signing": {
            "default": {
                "expiry": "168h"
            },
            "profiles": {
                "www": {
                    "expiry": "8760h",
                    "usages": [
                        "signing",
                        "key encipherment",
                        "server auth"
                    ]
                },
                "client": {
                    "expiry": "8760h",
                    "usages": [
                        "signing",
                        "key encipherment",
                        "client auth"
                    ]
                }
            }
        }
    }

    ```
1. 生成ca证书
    ```bash
    $ cfssl print-defaults csr > ca-csr.json # 下面的输出已按需修改
    {
        "CN": "example.net",
        "key": {
            "algo": "ecdsa",
            "size": 256
        },
        "names": [
            {
                "C": "US",
                "ST": "CA",
                "L": "San Francisco"
            }
        ]
    }
    $ cfssl gencert -initca ca-csr.json | cfssljson -bare ca # 执行结束后得到三个文件：ca-key.pem、ca.csr、ca.pem. 使用现有私钥: cfssl gencert -initca -ca-key key.pem ca-csr.json | cfssljson -bare ca
    $ cfssl-certinfo -cert ca.pem # 查看ca.pem, 也可使用`cfssl certinfo -cert ca.pem`/`openssl x509 -noout -text -in server.pem`
    ```

    names字段:
    - "CN"：Common Name，kube-apiserver 从证书中提取该字段作为请求的用户名 (User Name)
    - "O"：Organization，kube-apiserver从证书中提取该字段作为请求用户所属的组 (Group)
    - C: Country， 国家
    - L: Locality，地区，城市
    - O: Organization Name，组织名称，公司名称
    - OU: Organization Unit Name，组织单位名称，公司部门
    - ST: State，州，省
1. 生成server证书

    ```bash
    $ cfssl print-defaults csr > server-csr.json # 下面的输出已按需修改
    {
        "CN": "example.net",
        "hosts": [
            "example.net",
            "www.example.net"
        ],
        "key": {
            "algo": "ecdsa",
            "size": 256
        },
        "names": [
            {
                "C": "US",
                "ST": "CA",
                "L": "San Francisco"
            }
        ]
    }
    $ cfssl gencert -ca=ca.pem -ca-key=ca-key.pem --config=config.json -profile=www server-csr.json | cfssljson -bare server
    $ cfssl-certinfo -cert server.pem
    $ mkbundle -f server-bundle.pem server.pem ca.pem  # mkbundle：将证书链和私钥打包成一个文件
    ```

    > **`hosts`不用包含端口**

    分开生成key和pem:
    ```bash
    $ cfssl genkey server-csr.json  |cfssljson -bare server
    $ cfssl sign -ca=ca.pem -ca-key=ca-key.pem -csr=./server.csr  |cfssljson -bare server
    ```

    server-bundle.pem和server-key.pem即nginx使用的证书及其私钥.

    > openssl x509 -in cert.pem -out cert.crt -outform DER: pem转crt
