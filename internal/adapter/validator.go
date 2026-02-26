// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package adapter

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
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

func (v *Validator) Validate(data []byte) (*ValidationResponse, error) {
	var req struct {
		PlainToken string `json:"plain_token"`
		EventTs    string `json:"event_ts"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	// Generate private key
	seed := v.secret
	for len(seed) < ed25519.SeedSize {
		seed = seed + seed
	}
	seed = seed[:ed25519.SeedSize]

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
