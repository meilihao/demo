// https://github.com/SimonWaldherr/golang-examples/blob/master/advanced/aesgcm.go
// https://go.dev/src/crypto/cipher/example_test.go
// [Ciphertext and tag size and IV transmission with AES in GCM mode](https://crypto.stackexchange.com/questions/26783/ciphertext-and-tag-size-and-iv-transmission-with-aes-in-gcm-mode)
// [Why AES-256 with GCM adds 16 bytes to the ciphertext size?](https://stackoverflow.com/questions/67028762/why-aes-256-with-gcm-adds-16-bytes-to-the-ciphertext-size)
/*
aes-gcm:
- plaintext len limit see `aesgcm.Seal()`
- Output size = input size
- len(nonce)=12, Any value other than 12-byte is processed with GHASH
- len(tag)=16,  GCM is defined for the tag sizes 128, 120, 112, 104, or 96, 64 and 32. Note that the security of GCM is strongly dependent on the tag size.
- output=Nonce|ciphertext|tag
*/
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// AesGcmEncrypt takes an encryption key and a plaintext string and encrypts it with AES256 in GCM mode, which provides authenticated encryption. Returns the ciphertext and the used nonce.
// openssl_encrypt
func AesGcmEncrypt(key []byte, plaintext string) (ciphertext, nonce []byte) {
	plaintextBytes := []byte(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce = make([]byte, aesgcm.NonceSize()) // 12
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintextBytes, nil)
	fmt.Printf("Ciphertext: %x, %d -> %d\n", ciphertext, len(plaintextBytes), len(ciphertext))
	fmt.Printf("Nonce: %x\n", nonce)

	return
}

// AesGcmDecrypt takes an decryption key, a ciphertext and the corresponding nonce and decrypts it with AES256 in GCM mode. Returns the plaintext string.
// openssl_decrypt
func AesGcmDecrypt(key, ciphertext, nonce []byte) (plaintext string) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintextBytes, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	plaintext = string(plaintextBytes)
	fmt.Printf("%s\n", plaintext)

	return
}

func main() {
	// Generate an encryption key. 16 bytes = AES-128, 24 bytes = AES-192, 32 bytes = AES-256.
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err.Error())
	}

	// Specify the plaintext input
	plaintext := "Lorem Ipsum"
	ciphertext, nonce := AesGcmEncrypt(key, plaintext)

	// For decryption you need to provide the nonce which was used for the encryption
	plaintext = AesGcmDecrypt(key, ciphertext, nonce)
}
