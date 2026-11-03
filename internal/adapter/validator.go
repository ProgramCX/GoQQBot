// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

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
// Validate 处理 Webhook 验证请求（op=13）。
func (v *Validator) Validate(data []byte) (*ValidationResponse, error) {
	var req struct {
		PlainToken string `json:"plain_token"`
		EventTs    string `json:"event_ts"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	// Generate private key
	// 生成私钥。
	seed := v.generateSeed()
	_, privateKey, err := ed25519.GenerateKey(strings.NewReader(seed))
	if err != nil {
		return nil, err
	}

	// Signate
	// 签名。
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
// VerifyEventSignature 验证普通事件回调的 Ed25519 签名。
// According to QQ Bot API v2 docs: https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/sign.html
// 依据 QQ 机器人 API v2 文档：https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/interface-framework/sign.html
func (v *Validator) VerifyEventSignature(timestamp string, body []byte, signature string) bool {
	// Decode signature from hex
	// 将十六进制签名解码。
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		return false
	}

	// Generate public key from seed
	// 由种子生成公钥。
	seed := v.generateSeed()
	publicKey, _, err := ed25519.GenerateKey(strings.NewReader(seed))
	if err != nil {
		return false
	}

	// Construct message: timestamp + body
	// 构造待验签消息：时间戳加请求正文。
	var msg bytes.Buffer
	msg.WriteString(timestamp)
	msg.Write(body)

	// Verify signature
	// 验证签名。
	return ed25519.Verify(publicKey, msg.Bytes(), sig)
}

// VerifyEventSignatureFromRequest is a helper to verify signature from HTTP request
// VerifyEventSignatureFromRequest 是从 HTTP 请求验证签名的便捷方法。
func (v *Validator) VerifyEventSignatureFromRequest(r *http.Request, body []byte) bool {
	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	if signature == "" || timestamp == "" {
		return false
	}

	return v.VerifyEventSignature(timestamp, body, signature)
}

// generateSeed generates seed from secret for Ed25519 key generation
// generateSeed 从密钥生成用于 Ed25519 密钥生成的种子。
// According to QQ Bot API v2 docs: use strings.Repeat to double the seed
// 依据 QQ 机器人 API v2 文档，使用 strings.Repeat 将种子重复一遍。
func (v *Validator) generateSeed() string {
	seed := v.secret
	for len(seed) < ed25519.SeedSize {
		seed = strings.Repeat(seed, 2)
	}
	return seed[:ed25519.SeedSize]
}
