/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package cryptography

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

const (
	// MinRSAPublicKeyBits is the minimum required bit length for RSA public keys.
	MinRSAPublicKeyBits = 3072
	// spkiPublicKeyBlockType is the PEM block type for SPKI encoded public keys.
	spkiPublicKeyBlockType = "PUBLIC KEY"
	// rsaOaepSHA256Overhead is the byte overhead of RSA-OAEP with SHA-256 (2*32 + 2 = 66).
	rsaOaepSHA256Overhead = 2*sha256.Size + 2
)

// RSAOaep implements RSA-OAEP encryption with SHA-256 hash and MGF1-SHA-256.
type RSAOaep struct {
	// pubKey is the parsed RSA public key.
	pubKey *rsa.PublicKey
}

// ParseRSAPublicKey parses a PEM-encoded SPKI RSA public key and validates its bit length (>= 3072).
func ParseRSAPublicKey(pubKeyPEM string) (*rsa.PublicKey, error) {
	if strings.TrimSpace(pubKeyPEM) == "" {
		return nil, errors.New("public key PEM is empty")
	}

	block, rest := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block from public key")
	}
	if strings.TrimSpace(string(rest)) != "" {
		return nil, errors.New("unexpected trailing data after public key PEM block")
	}
	if block.Type != spkiPublicKeyBlockType {
		return nil, fmt.Errorf("invalid PEM block type %q, expected %q", block.Type, spkiPublicKeyBlockType)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse SPKI public key failed, err: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not an RSA public key")
	}
	if rsaPub.N.BitLen() < MinRSAPublicKeyBits {
		return nil, fmt.Errorf("RSA public key bit length %d is less than minimum required %d",
			rsaPub.N.BitLen(), MinRSAPublicKeyBits)
	}

	return rsaPub, nil
}

// NewRSAOaepFromPEM parses a PEM-encoded SPKI RSA public key and creates a new RSAOaep instance.
func NewRSAOaepFromPEM(pubKeyPEM string) (*RSAOaep, error) {
	pubKey, err := ParseRSAPublicKey(pubKeyPEM)
	if err != nil {
		return nil, err
	}

	return &RSAOaep{pubKey: pubKey}, nil
}

// MaxPlaintextBytes returns the maximum allowed plaintext byte length for the RSA public key.
func (r *RSAOaep) MaxPlaintextBytes() int {
	return r.pubKey.Size() - rsaOaepSHA256Overhead
}

// Encrypt encrypts plaintext using RSA-OAEP with SHA-256 (for both OAEP hash and MGF1) and empty label.
func (r *RSAOaep) Encrypt(plaintext []byte) ([]byte, error) {
	if r == nil || r.pubKey == nil {
		return nil, errors.New("RSAOaep is not initialized")
	}

	maxLen := r.MaxPlaintextBytes()
	if len(plaintext) > maxLen {
		return nil, fmt.Errorf("plaintext length %d exceeds max allowed %d bytes for %d-bit RSA key",
			len(plaintext), maxLen, r.pubKey.N.BitLen())
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, r.pubKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt with RSA-OAEP failed, err: %v", err)
	}

	return ciphertext, nil
}

// EncryptToBase64 encrypts plaintext using RSA-OAEP and encodes the ciphertext using standard Base64 (RFC 4648 §4).
func (r *RSAOaep) EncryptToBase64(plaintext []byte) (string, error) {
	ciphertext, err := r.Encrypt(plaintext)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
