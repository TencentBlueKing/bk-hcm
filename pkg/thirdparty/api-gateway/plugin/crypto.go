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

// Package plugin provides the southbound HTTP client and protocol models for HCM plugins.
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"hcm/pkg/cryptography"
)

// CryptoType defines the encryption algorithm type for plugin launch_config.
type CryptoType string

const (
	// CryptoTypeRSA represents RSA-OAEP (SHA-256) encryption for launch_config.
	CryptoTypeRSA CryptoType = "RSA"
)

// RSACryptoConfig defines the crypto_config parameters when crypto_type is RSA.
type RSACryptoConfig struct {
	// PublicKey is the PEM-encoded SPKI RSA public key (>= 3072 bits).
	PublicKey string `json:"public_key"`
}

// LaunchConfigEncryptor defines the interface for encrypting plugin launch_config parameters.
type LaunchConfigEncryptor interface {
	Encrypt(providerConfig map[string]any) (string, error)
	EncryptJSON(rawJSON []byte) (string, error)
}

type rsaLaunchConfigEncryptor struct {
	rsaOaep *cryptography.RSAOaep
}

// ValidateCryptoConfig validates the plugin crypto_type and its corresponding crypto_config.
func ValidateCryptoConfig(cryptoType CryptoType, cryptoConfig json.RawMessage) error {
	_, err := NewLaunchConfigEncryptor(cryptoType, cryptoConfig)
	return err
}

// NewLaunchConfigEncryptor validates crypto_type and crypto_config and constructs a LaunchConfigEncryptor.
func NewLaunchConfigEncryptor(cryptoType CryptoType, cryptoConfig json.RawMessage) (LaunchConfigEncryptor, error) {
	switch cryptoType {
	case CryptoTypeRSA:
		if len(cryptoConfig) == 0 {
			return nil, errors.New("crypto_config is required for RSA crypto_type")
		}
		var cfg RSACryptoConfig
		if err := json.Unmarshal(cryptoConfig, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal RSA crypto_config failed, err: %v", err)
		}
		if strings.TrimSpace(cfg.PublicKey) == "" {
			return nil, errors.New("crypto_config.public_key is required for RSA crypto_type")
		}
		return NewRSALaunchConfigEncryptor(cfg.PublicKey)
	default:
		return nil, fmt.Errorf("unsupported crypto_type %q", cryptoType)
	}
}

// NewRSALaunchConfigEncryptor creates a LaunchConfigEncryptor from a PEM-encoded SPKI RSA public key.
func NewRSALaunchConfigEncryptor(publicKeyPEM string) (LaunchConfigEncryptor, error) {
	rsaOaep, err := cryptography.NewRSAOaepFromPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	return &rsaLaunchConfigEncryptor{rsaOaep: rsaOaep}, nil
}

// Encrypt marshals providerConfig into UTF-8 JSON and encrypts it as a standard Base64 launch_config string.
func (e *rsaLaunchConfigEncryptor) Encrypt(providerConfig map[string]any) (string, error) {
	if providerConfig == nil {
		providerConfig = make(map[string]any)
	}

	rawJSON, err := json.Marshal(providerConfig)
	if err != nil {
		return "", fmt.Errorf("marshal provider config failed, err: %v", err)
	}

	return e.EncryptJSON(rawJSON)
}

// EncryptJSON encrypts UTF-8 JSON plaintext bytes into a standard Base64 launch_config string.
func (e *rsaLaunchConfigEncryptor) EncryptJSON(rawJSON []byte) (string, error) {
	if e == nil || e.rsaOaep == nil {
		return "", errors.New("launch_config encryptor is not initialized")
	}

	return e.rsaOaep.EncryptToBase64(rawJSON)
}
