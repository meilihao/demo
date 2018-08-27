package main

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/secretbox"
)

type cipherSuite struct {
	Name         string
	GenerateAEAD func([]byte) (cipher.AEAD, error)
	NonceSize    int
}

func Chacha20poly1305() {
	var cipherHandlers = []cipherSuite{
		{
			Name:         "Chacha20poly1305",
			GenerateAEAD: chacha20poly1305.New,
			NonceSize:    chacha20poly1305.NonceSize,
		}, {
			Name:         "XChacha20poly1305",
			GenerateAEAD: chacha20poly1305.NewX,
			NonceSize:    chacha20poly1305.NonceSizeX,
		},
	}

	msg := "hello"

	for _, suite := range cipherHandlers {
		aead, err := suite.GenerateAEAD(generateKey(chacha20poly1305.KeySize))
		if err != nil {
			log.Fatalf("Failed to instantiate %v: %v", suite, err)
		}

		// Encryption.
		nonce := make([]byte, suite.NonceSize)
		if _, err := rand.Read(nonce); err != nil {
			panic(err)
		}
		ciphertext := aead.Seal(nil, nonce, []byte(msg), nil)

		// Decryption.
		plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			log.Fatalln("Failed to decrypt or authenticate message:", err)
		}

		fmt.Printf("%s\n", plaintext)
	}
}

func XSalsa20poly1305() {
	key := generateKey(32)
	var secretKey [32]byte

	copy(secretKey[:], key)

	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	// This encrypts "hello world" and appends the result to the nonce.
	encrypted := secretbox.Seal(nonce[:], []byte("hello world"), &nonce, &secretKey)

	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &secretKey)
	if !ok {
		panic("decryption error")
	}

	fmt.Println(string(decrypted))
}

func main() {
	Chacha20poly1305()
	XSalsa20poly1305()
}

func generateKey(length int) []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err.Error())
	}

	return key
}
