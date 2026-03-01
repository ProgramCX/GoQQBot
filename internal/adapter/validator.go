// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package adapter

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

type Validator struct {
	secret string
}

type ValidationResponse struct {
	PlainToken string `json:"plain_token"`
	Signature  string `json:"signature"`
}

func NewValidator(secret string) *Validator {
	return &Validator{secret: secret}
}

// Validate handles webhook validation request (op=13)
func (v *Validator) Validate(data []byte) (*ValidationResponse, error) {
	var req struct {
		PlainToken string `json:"plain_token"`
		EventTs    string `json:"event_ts"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	// Generate private key
	seed := v.generateSeed()
	_, privateKey, err := ed25519.GenerateKey(strings.NewReader(seed))
	if err != nil {
		return nil, err
	}

	// Signate
	var msg bytes.Buffer
	msg.WriteString(req.EventTs)
	msg.WriteString(req.PlainToken)
	sig := ed25519.Sign(privateKey, msg.Bytes())

	return &ValidationResponse{
		PlainToken: req.PlainToken,
		Signature:  hex.EncodeToString(sig),
	}, nil
}

// VerifyEventSignature verifies the Ed25519 signature for normal event callbacks
// According to QQ Bot API v2 docs: https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/sign.html
func (v *Validator) VerifyEventSignature(timestamp string, body []byte, signature string) bool {
	// Decode signature from hex
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		return false
	}

	// Generate public key from seed
	seed := v.generateSeed()
	publicKey, _, err := ed25519.GenerateKey(strings.NewReader(seed))
	if err != nil {
		return false
	}

	// Construct message: timestamp + body
	var msg bytes.Buffer
	msg.WriteString(timestamp)
	msg.Write(body)

	// Verify signature
	return ed25519.Verify(publicKey, msg.Bytes(), sig)
}

// VerifyEventSignatureFromRequest is a helper to verify signature from HTTP request
func (v *Validator) VerifyEventSignatureFromRequest(r *http.Request, body []byte) bool {
	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	if signature == "" || timestamp == "" {
		return false
	}

	return v.VerifyEventSignature(timestamp, body, signature)
}

// generateSeed generates seed from secret for Ed25519 key generation
// According to QQ Bot API v2 docs: use strings.Repeat to double the seed
func (v *Validator) generateSeed() string {
	seed := v.secret
	for len(seed) < ed25519.SeedSize {
		seed = strings.Repeat(seed, 2)
	}
	return seed[:ed25519.SeedSize]
}
