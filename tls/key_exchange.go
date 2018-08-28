// tls 1.3 key keyexchange, no use clinet/server's random
// https://medium.com/@oyrxx/tls-key-exchange-aed230aa114e
// https://blog.csdn.net/mrpre/article/details/80056618
//
// HKDF是基于HMAC的HKDF(密钥派生函数).
package main

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func init() {
	crypto.RegisterHash(crypto.SHA256, sha256.New)
}

func main() {
	TLS_AES_128_GCM_SHA256()
}

func TLS_AES_128_GCM_SHA256() {
	hash := crypto.SHA256

	//  HKDF-Extract = Early Secret
	var earlySecret []byte
	var usingPSK bool
	if usingPSK {
		pskSecret := make([]byte, 32)
		earlySecret = hkdfExtract(hash, pskSecret, nil)
	} else {
		earlySecret = hkdfExtract(hash, nil, nil)
	}

	//HKDF-Extract = Handshake Secret
	handshakeCtx := hash.New().Sum(nil)
	preHandshakeSecret := deriveSecret(hash, earlySecret, handshakeCtx, labelDerived)

	ecdheSecret := bytes.Repeat([]byte{0}, 32)
	handshakeSecret := hkdfExtract(hash, ecdheSecret, preHandshakeSecret)

	//HKDF-Extract = Master Secret
	preMasterSecret := deriveSecret(hash, handshakeSecret, handshakeCtx, labelDerived)
	masterSecret := hkdfExtract(hash, nil, preMasterSecret)

	// traffic_secret
	trafficSecret := deriveSecret(hash, masterSecret, handshakeCtx, labelServerApplicationTrafficSecret)

	// key
	key := hkdfExpandLabel(hash, trafficSecret, nil, "key", 16)
	iv := hkdfExpandLabel(hash, trafficSecret, nil, "iv", 12)

	fmt.Println(key, iv)
	_, _ = key, iv // aeadAES128GCM's key and nonce
}

// https://github.com/cloudflare/tls-tris/blob/master/13.go
const (
	labelExternalBinder                 = "ext binder"
	labelResumptionBinder               = "res binder"
	labelEarlyTrafficSecret             = "c e traffic"
	labelEarlyExporterSecret            = "e exp master"
	labelClientHandshakeTrafficSecret   = "c hs traffic"
	labelServerHandshakeTrafficSecret   = "s hs traffic"
	labelClientApplicationTrafficSecret = "c ap traffic"
	labelServerApplicationTrafficSecret = "s ap traffic"
	labelExporterSecret                 = "exp master"
	labelResumptionSecret               = "res master"
	labelDerived                        = "derived"
	labelFinished                       = "finished"
	labelResumption                     = "resumption"
)

// https://github.com/bifurcation/mint/blob/master/crypto.go
func deriveSecret(hash crypto.Hash, secret, messageHash []byte, label string) []byte {
	return hkdfExpandLabel(hash, secret, messageHash, label, hash.Size())
}

// https://github.com/cloudflare/tls-tris/blob/master/13.go
func hkdfExpandLabel(hash crypto.Hash, secret, hashValue []byte, label string, L int) []byte {
	prefix := "tls13 "
	hkdfLabel := make([]byte, 4+len(prefix)+len(label)+len(hashValue))
	hkdfLabel[0] = byte(L >> 8)
	hkdfLabel[1] = byte(L)
	hkdfLabel[2] = byte(len(prefix) + len(label))
	copy(hkdfLabel[3:], prefix)
	z := hkdfLabel[3+len(prefix):]
	copy(z, label)
	z = z[len(label):]
	z[0] = byte(len(hashValue))
	copy(z[1:], hashValue)

	return hkdfExpand(hash, secret, hkdfLabel, L)
}

// 通过一系列的哈希运算将密钥扩展到我们需要的长度
// https://github.com/cloudflare/tls-tris/blob/master/hkdf.go
func hkdfExpand(hash crypto.Hash, prk, info []byte, l int) []byte {
	var (
		expander = hmac.New(hash.New, prk)
		res      = make([]byte, l)
		counter  = byte(1)
		prev     []byte
	)

	if l > 255*expander.Size() {
		panic("hkdf: requested too much output")
	}

	p := res
	for len(p) > 0 {
		expander.Reset()
		expander.Write(prev)
		expander.Write(info)
		expander.Write([]byte{counter})
		prev = expander.Sum(prev[:0])
		counter++
		n := copy(p, prev)
		p = p[n:]
	}

	return res
}

// 将用户输入的密钥尽量的伪随机化
func hkdfExtract(hash crypto.Hash, secret, salt []byte) []byte {
	if salt == nil {
		salt = make([]byte, hash.Size())
	}
	if secret == nil {
		secret = make([]byte, hash.Size())
	}
	extractor := hmac.New(hash.New, salt)
	extractor.Write(secret)
	return extractor.Sum(nil)
}
