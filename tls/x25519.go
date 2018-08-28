// see https://github.com/cloudflare/tls-tris/blob/ad86d61c4229dd9ecb5ddf89e4b4a4794b31e415/13.go
package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"reflect"

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
	clientPrivateKey, clientKS, err := generateKeyShare(X25519)
	CheckErr(err)

	serverPrivateKey, serverKS, err := generateKeyShare(X25519)
	CheckErr(err)

	ecdheSecret := deriveECDHESecret(clientKS, serverPrivateKey)
	if ecdheSecret == nil {
		panic("tls: bad ECDHE server share")
	}

	ecdheSecret2 := deriveECDHESecret(serverKS, clientPrivateKey)
	if ecdheSecret2 == nil {
		panic("tls: bad ECDHE client share")
	}

	fmt.Println(reflect.DeepEqual(ecdheSecret, ecdheSecret2))
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
