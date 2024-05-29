// [HMAC Key Derivation Function (HKDF) in Golang](https://asecuritysite.com/golang/go_hkdf)
// [SubtleCrypto.deriveBits()](https://developer.mozilla.org/zh-CN/docs/Web/API/SubtleCrypto/deriveBits)
// [密码学系列之:1Password的加密基础PBKDF2](https://cloud.tencent.com/developer/article/1883405)
// [hkdf pbkdf2 算法](https://juejin.cn/post/6964547918736916487)
// //              算法       明文密码（原始二进制）     输出长度  应用程序/特定于上下文的信息字符串    salt值
// $hkdf1 = hash_hkdf('sha256', '123456', 32, 'aes-256-encryption', random_bytes(2));
// $hkdf2 = hash_hkdf('sha256', '123456', 32, 'sha-256-authentication', random_bytes(2));
// var_dump($hkdf1);
// var_dump($hkdf2);
// // string(32) "ԇ`q��X�l�
// //                      f�yð����}Ozb+�"
// // string(32) "%���]�+̀�\JdG��HL��GK��
// //                                   -"

// //              算法       明文密码     salt值        迭代次数  数据长度
// echo hash_pbkdf2("sha256", '123456', random_bytes(2), 1000, 20)
// // e27156f9a6e2c55f3b72
package main

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// func getSalt(n int) []byte {
// 	nonce := make([]byte, n)
// 	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
// 		panic(err.Error())
// 	}
// 	return (nonce)

// }
func main() {
	hash := sha256.New
	s := "The quick brown fox jumps over the lazy dog"
	salt := []byte(nil) //getSalt(hash().Size())

	info := []byte("")

	secret := []byte("test")

	kdf := hkdf.New(hash, secret, salt, info) // 相同hash, secret, salt, info, len(,key), io.ReadFull(kdf, key)相同

	key1 := make([]byte, 16)
	_, _ = io.ReadFull(kdf, key1)

	fmt.Printf("Secret: %s\n", s)
	fmt.Printf("HKDF 16 byte key: %x\n", key1)

	key2 := make([]byte, 32)
	_, _ = io.ReadFull(kdf, key2)

	fmt.Printf("HKDF 32 byte key: %x\n", key2)

	key3 := make([]byte, 1024)
	_, _ = io.ReadFull(kdf, key3)

	fmt.Printf("HKDF 1024 byte key: %x\n", key3)
}
