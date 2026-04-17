package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

func VerifyGitHubWebhookSignature(secret string) func(http.Handler) http.Handler { // returns a middleware function
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// if the secret is empty, we can't verify the signature, so we should return an error
			if secret == "" {
				http.Error(w, "Webhook secret not configured", http.StatusInternalServerError)
				return
			}

			signatureHeader := r.Header.Get("X-Hub-Signature-256")
			if signatureHeader == "" {
				http.Error(w, "Missing X-Hub-Signature-256 header", http.StatusBadRequest)
				return
			}

			// read the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}

			// restore the request body for the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body)) //NopCloser is used to create a ReadCloser from a Buffer, allowing us to read the body again in the next handler

			// compute the HMAC SHA256 signature
			mac := hmac.New(sha256.New, []byte(secret)) //we create a new HMAC using the SHA256 hash function and the secret key
			mac.Write(body) 
			expectedMAC := mac.Sum(nil) // Sum returns the HMAC signature as a byte slice

			// since the signature header is in the format "sha256=signature" we need to extract the actual signature
			const prefix = "sha256="
			if !strings.HasPrefix(signatureHeader, prefix) {
				http.Error(w, "Invalid signature format", http.StatusBadRequest)
				return
			}

			signature := strings.TrimPrefix(signatureHeader, prefix) // we remove the "sha256=" prefix to get the actual signature

			// hex to bytes
			receivedMAC, err := hex.DecodeString(signature)
			if err != nil {
				http.Error(w, "Invalid signature encoding", http.StatusBadRequest)
				return
			}

			// compare the expected MAC with the received MAC
			if !hmac.Equal(expectedMAC, receivedMAC) { // hmac.Equal is a constant-time comparison function that compares two byte slices and returns true if they are equal
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}
			
			// if the signature is valid, call the next handler
			next.ServeHTTP(w, r)
		})
	}

}

