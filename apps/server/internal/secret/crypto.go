package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

// Encrypt encrypts plaintext using envelope encryption with AES-256-GCM.
//  1. Generate random 32-byte DEK (data encryption key)
//  2. Encrypt plaintext with DEK -> ciphertext + valueNonce
//  3. Encrypt DEK with masterKey -> encryptedDEK + dekNonce
func Encrypt(masterKey [32]byte, plaintext string) (encryptedDEK, dekNonce, ciphertext, valueNonce []byte, err error) {
	// Generate random DEK
	dek := make([]byte, 32)
	if _, err = rand.Read(dek); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("generate DEK: %w", err)
	}

	// Encrypt plaintext with DEK
	ciphertext, valueNonce, err = aesGCMEncrypt(dek, []byte(plaintext))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("encrypt value: %w", err)
	}

	// Encrypt DEK with master key
	encryptedDEK, dekNonce, err = aesGCMEncrypt(masterKey[:], dek)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("encrypt DEK: %w", err)
	}

	return encryptedDEK, dekNonce, ciphertext, valueNonce, nil
}

// Decrypt reverses envelope encryption.
func Decrypt(masterKey [32]byte, encryptedDEK, dekNonce, ciphertext, valueNonce []byte) (string, error) {
	// Decrypt DEK with master key
	dek, err := aesGCMDecrypt(masterKey[:], encryptedDEK, dekNonce)
	if err != nil {
		return "", fmt.Errorf("decrypt DEK: %w", err)
	}

	// Decrypt plaintext with DEK
	plaintext, err := aesGCMDecrypt(dek, ciphertext, valueNonce)
	if err != nil {
		return "", fmt.Errorf("decrypt value: %w", err)
	}

	return string(plaintext), nil
}

// CanDecrypt tests whether the master key can decrypt an encrypted DEK.
func CanDecrypt(masterKey [32]byte, encryptedDEK, dekNonce []byte) bool {
	_, err := aesGCMDecrypt(masterKey[:], encryptedDEK, dekNonce)
	return err == nil
}

func aesGCMEncrypt(key, plaintext []byte) (ciphertextOut, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ciphertextOut = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertextOut, nonce, nil
}

func aesGCMDecrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}
