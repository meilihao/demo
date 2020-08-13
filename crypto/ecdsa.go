package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"crypto/sha256"
	"fmt"
	"math/big"
)

func main()  {

    message := []byte("hello")

    //设置生成的私钥为256位
    privatekey,_ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

    //创建公钥
    publicKey := privatekey.PublicKey

    //hash散列明文
    digest := sha256.Sum256(message)


    //用私钥签名
    r,s,_:= ecdsa.Sign(rand.Reader,privatekey, digest[:])

    //设置私钥的参数类型
    param := privatekey.Curve.Params()

    //获取私钥的长度（字节）
    curveOrderBytes:=param.P.BitLen()/8

    //获得签名返回的字节
    rByte,sByte := r.Bytes(), s.Bytes()

    //创建数组合并字节
    signature := make([]byte,curveOrderBytes*2)
    copy(signature[:len(rByte)], rByte)
    copy(signature[len(sByte):], sByte)

    //现在signature中就存放了完整的签名的结果

    //验签
    digest = sha256.Sum256(message)
    //获得公钥的字节长度
    curveOrderBytes= publicKey.Curve.Params().P.BitLen()/8

    //创建大整数类型保存rbyte,sbyte
    r,s = new(big.Int),new(big.Int)

    r.SetBytes(signature[:curveOrderBytes])
    s.SetBytes(signature[curveOrderBytes:])


    //开始认证
    e:=ecdsa.Verify(&publicKey,digest[:],r,s)
    if e== true {
        fmt.Println("验签成功")
    }
}