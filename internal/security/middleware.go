package security

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrCouldntReadBody = errors.New("couldn't read body")
)

// BodyDecrypt decrypts request body
func BodyDecrypt(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			switch {
			case errors.Is(err, io.EOF):
				next.ServeHTTP(w, r)
				return
			case err != nil:
				http.Error(w, fmt.Sprintf("Cannot read provided data: %q", err), http.StatusInternalServerError)
				return
			}
			decryptedBody, err := RSADecrypt(body, privateKey)
			if err != nil {
				http.Error(w, fmt.Sprintf("Cannot decrypt provided data: %q", err), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
			r.ContentLength = int64(len(decryptedBody))
			r.Header.Set("Content-Type", "application/json")

			next.ServeHTTP(w, r)
		})
	}
}

// EncryptionMiddleware encrypts HTTP client requests and decrypts HTTP server responses
type EncryptionMiddleware struct {
	Proxied   http.RoundTripper
	PublicKey *rsa.PublicKey
}

func (ert EncryptionMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	if ert.PublicKey == nil || req.Body == nil {
		return ert.Proxied.RoundTrip(req)
	}
	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()

	switch {
	case errors.Is(err, io.EOF):
		return ert.Proxied.RoundTrip(req)
	case err != nil:
		return nil, ErrCouldntReadBody
	}

	encryptedBody, err := RSAEncrypt(body, ert.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("couldn't encrypt body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(encryptedBody))
	req.ContentLength = int64(len(encryptedBody))

	return ert.Proxied.RoundTrip(req)
}
