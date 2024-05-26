package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

const (
	HashHeader = "HashSHA256"
)

type securedResponseWriter struct {
	http.ResponseWriter
	securityKey string
}

func AddSign(data []byte, securityKey string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(securityKey))
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (w securedResponseWriter) Write(data []byte) (int, error) {
	if w.securityKey != "" {
		data, err := AddSign(data, w.securityKey)
		if err != nil {
			return 0, fmt.Errorf("failed to sign data: %w", err)
		}
		w.ResponseWriter.Header().Set(HashHeader, hex.EncodeToString(data))
		return w.ResponseWriter.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func New(key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentHashHeader := r.Header.Get(HashHeader)
			if contentHashHeader == "" {
				h.ServeHTTP(w, r)
				return
			}
			response, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(response))
			_, err = AddSign(response, key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = hex.DecodeString(contentHashHeader)
			if err != nil {
				http.Error(w, "Header with security sign is not found", http.StatusInternalServerError)
				return
			}
			// if contentHashHeader != "" && !hmac.Equal(decodedHash, signedResponse) {
			// 	http.Error(w, "Wrong security sign", http.StatusBadRequest)
			// 	return
			// }
			h.ServeHTTP(securedResponseWriter{
				ResponseWriter: w,
				securityKey:    key,
			}, r)
		})
	}
}
