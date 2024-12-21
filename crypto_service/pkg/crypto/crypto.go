package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const keyLength = 32

// Encrypt зашифровывает последовательность байт алгоритмом AES-256.
func Encrypt(data, key []byte) ([]byte, error) {
	if len(key) != keyLength {
		return nil, errors.New("invalid key length has been transmitted")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherData := aesGCM.Seal(nonce, nonce, data, nil)

	return cipherData, nil
}

// Decrypt расшифровывает последовательность байт алгоритмом AES-256.
func Decrypt(data, key []byte) ([]byte, error) {
	if len(key) != keyLength {
		return nil, errors.New("invalid key length has been transmitted")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("the length of the ciphertext is less than the nonce size")
	}

	nonce, data := data[:nonceSize], data[nonceSize:]

	plainText, err := aesGCM.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}