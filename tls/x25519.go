// see https://github.com/cloudflare/tls-tris/blob/ad86d61c4229dd9ecb5ddf89e4b4a4794b31e415/13.go
package main

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
)

// CurveID is the type of a TLS identifier for an elliptic curve. See
// http://www.iana.org/assignments/tls-parameters/tls-parameters.xml#tls-parameters-8
//
// TLS 1.3 refers to these as Groups, but this library implements only
// curve-based ones anyway. See https://tools.ietf.org/html/draft-ietf-tls-tls13-18#section-4.2.4.
type CurveID uint16

const (
	X25519 CurveID = 29
)

// TLS 1.3 Key Share
// See https://tools.ietf.org/html/draft-ietf-tls-tls13-18#section-4.2.5
type keyShare struct {
	group CurveID
	data  []byte
}

func main() {
	clientKS := keyShare{group: X25519, data: []byte{0x2e, 0x3c, 0xc9, 0xa6, 0xe9, 0x44, 0x0f, 0x58, 0x94, 0x2b, 0x90, 0x73, 0x2e, 0x48, 0xde, 0x20, 0xda, 0xed, 0x7c, 0x6f, 0xe8, 0xac, 0x78, 0x8a, 0xa6, 0x20, 0x25, 0xc4, 0x99, 0xa6, 0x42, 0x4e}}
	privateKey, serverKS, err := generateKeyShare(X25519)
	CheckErr(err)

	ecdheSecret := deriveECDHESecret(clientKS, privateKey)
	if ecdheSecret == nil {
		panic("tls: bad ECDHE client share")
	}

	fmt.Println(serverKS, ecdheSecret)
}

// GenerateKey 基于curve25519椭圆曲线算法生成密钥对
func generateKeyShare(curveID CurveID) ([]byte, keyShare, error) {
	var scalar, public [32]byte
	if _, err := io.ReadFull(rand.Reader, scalar[:]); err != nil {
		return nil, keyShare{}, err
	}

	curve25519.ScalarBaseMult(&public, &scalar)
	return scalar[:], keyShare{group: curveID, data: public[:]}, nil
}

// 密钥协商
func deriveECDHESecret(ks keyShare, secretKey []byte) []byte {
	if len(ks.data) != 32 {
		return nil
	}

	var theirPublic, sharedKey, scalar [32]byte
	copy(theirPublic[:], ks.data)
	copy(scalar[:], secretKey)
	curve25519.ScalarMult(&sharedKey, &scalar, &theirPublic)
	return sharedKey[:]
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
