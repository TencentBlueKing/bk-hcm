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

package account

import (
	"errors"
	"fmt"
	"regexp"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/json"
)

var (
	validAccountNameRegex   = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9-_]{1,62}[a-zA-Z0-9]$")
	accountNameInvalidError = errors.New("invalid account name: name should begin with a letter(a-zA-Z), " +
		"contains letters, numbers(0-9) or hyphen(-), underline(_),end with a letter or number, " +
		"length should be 3 to 64 letters")
	secretEmptyError = errors.New("SecretID/SecretKey can not be empty")
	allBizError      = errors.New("can't choose specific biz when choose all biz")
)

// -------------------------- 一些通用的校验 ------------------------

func validateAccountName(name string) error {
	if name != "" && !validAccountNameRegex.MatchString(name) {
		return accountNameInvalidError
	}
	return nil
}

func validateBkBizIDs(bkBizIDs []int64) error {
	for _, bizID := range bkBizIDs {
		// 非全业务时，校验是否非法业务ID
		if bizID != constant.AttachedAllBiz && bizID <= 0 {
			return fmt.Errorf("invalid biz id: %d", bizID)
		}
		// 选择全业务时，不可选择其他具体业务，即全业务时业务数量只能是1
		if bizID == constant.AttachedAllBiz && len(bkBizIDs) != 1 {
			return allBizError
		}
	}
	return nil
}

func validateResAccountBkBizIDs(bkBizIDs []int64) error {
	if len(bkBizIDs) <= 0 {
		return fmt.Errorf("invalid res account have no bizIDs")
	}

	if len(bkBizIDs) == 1 && bkBizIDs[0] == -1 {
		return fmt.Errorf("invalid res account not assigned bizIDs")
	}

	return nil
}

// gcpAccountCloudServiceSecretKey 由于gcp密钥非普通字符串，而是一个map 字符串，用户容易出错，所以定义结构进行校验，避免透传给gcp api
type gcpAccountCloudServiceSecretKey struct {
	Type                    string `json:"type" validate:"required"`
	ProjectID               string `json:"project_id" validate:"required"`
	PrivateKeyID            string `json:"private_key_id" validate:"required"`
	PrivateKey              string `json:"private_key" validate:"required"`
	ClientEmail             string `json:"client_email" validate:"required"`
	ClientID                string `json:"client_id" validate:"required"`
	AuthURI                 string `json:"auth_uri" validate:"required"`
	TokenURI                string `json:"token_uri" validate:"required"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url" validate:"required"`
	ClientX509CertURL       string `json:"client_x509_cert_url" validate:"required"`
}

// validateGcpCloudServiceSK 检查GCP秘钥是否合法
func validateGcpCloudServiceSK(cloudServiceSecretKey string) error {
	if cloudServiceSecretKey != "" {
		secretKey, err := DecodeGcpSecretKey(cloudServiceSecretKey)
		if err != nil {
			return err
		}
		if err := validator.Validate.Struct(secretKey); err != nil {
			return fmt.Errorf("secret key of service account is invalid data, err: %v", err)
		}
	}

	return nil
}

// DecodeGcpSecretKey 解析GCP秘钥JSON字符串为结构体
func DecodeGcpSecretKey(cloudServiceSecretKey string) (*gcpAccountCloudServiceSecretKey, error) {
	secretKey := new(gcpAccountCloudServiceSecretKey)
	if err := json.UnmarshalFromString(cloudServiceSecretKey, &secretKey); err != nil {
		return nil, fmt.Errorf("the secret key format of service account is invalid , err: %v", err)
	}
	return secretKey, nil
}
