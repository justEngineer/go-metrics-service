package encryption

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
	"io"
	"net/http"
	"os"
)

const (
	RSAKeySize = 2048
	ChunkSize  = (RSAKeySize / 8) - 2*sha256.Size - 2
)

var ErrBadPublicKeyFormat = errors.New("wrong public key format, ensure -----BEGIN PUBLIC KEY----- header")

// GetPrivateKey reads private key from file
func GetPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	if privateKeyPath == "" {
		return nil, nil
	}

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(privateKeyBytes)
	key, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// GetPublicKey reads public key from file
func GetPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	if publicKeyPath == "" {
		return nil, nil
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(publicKeyBytes)
	untypedKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := untypedKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrBadPublicKeyFormat
	}

	return key, nil
}

func EncryptWithPublicKey(data []byte, key *rsa.PublicKey) ([]byte, error) {
	numChunks := (len(data) + ChunkSize - 1) / ChunkSize
	encryptedChunks := make([]byte, 0, numChunks*RSAKeySize/8)

	for start := 0; start < len(data); start += ChunkSize {
		end := start + ChunkSize

		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		encryptedChunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, chunk, nil)

		if err != nil {
			return nil, err
		}

		encryptedChunks = append(encryptedChunks, encryptedChunk...)
	}

	return encryptedChunks, nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemFile, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemFile)

	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func DecryptWithPrivateKey(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	numChunks := len(data) / ChunkSize
	decryptedMessage := make([]byte, 0, numChunks*ChunkSize)

	chunkSize := RSAKeySize / 8

	for start := 0; start < len(data); start += chunkSize {
		end := start + chunkSize

		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, chunk, nil)

		if err != nil {
			return nil, err
		}

		decryptedMessage = append(decryptedMessage, decryptedChunk...)
	}

	return decryptedMessage, nil
}

func DecryptMiddleware(privateKey *rsa.PrivateKey) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			decryptedMessage, err := DecryptWithPrivateKey(body, privateKey)

			if err != nil {
				http.Error(w, "Failed to decrypt message", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(decryptedMessage))

			next.ServeHTTP(w, r)
		}
	}
}

// RSAEncrypt encrypts data with public key
func RSAEncrypt(msg []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return msg, nil
	}

	ciphertext, err := EncryptOAEP(sha256.New(), rand.Reader, publicKey, msg, nil)

	return ciphertext, err
}

// RSADecrypt decrypts data with private key
func RSADecrypt(ciphertext []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return ciphertext, nil
	}

	plaintext, err := DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)

	return plaintext, err
}

// EncryptOAEP is a chanked format of rsa func
func EncryptOAEP(hash hash.Hash, random io.Reader, public *rsa.PublicKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := public.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, public, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}

// DecryptOAEP is a chanked format of rsa func
func DecryptOAEP(hash hash.Hash, random io.Reader, private *rsa.PrivateKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := private.PublicKey.Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, random, private, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}
