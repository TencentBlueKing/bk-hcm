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

package accountset

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/cryptography"
)

// BaseMainAccount 云主账号/云二级账号
type BaseMainAccount struct {
	ID                string                         `json:"id"`
	Name              string                         `json:"name"`
	Vendor            enumor.Vendor                  `json:"vendor"`
	CloudID           string                         `json:"cloud_id"`
	Email             string                         `json:"email"`
	Managers          []string                       `json:"managers"`
	BakManagers       []string                       `json:"bak_managers"`
	Site              enumor.MainAccountSiteType     `json:"site"`
	BusinessType      enumor.MainAccountBusinessType `json:"business_type"`
	Status            enumor.MainAccountStatus       `json:"status"`
	ParentAccountName string                         `json:"parent_account_name"`
	ParentAccountID   string                         `json:"parent_account_id"`
	DeptID            int64                          `json:"dept_id"`
	BkBizID           int64                          `json:"bk_biz_id"`
	OpProductID       int64                          `json:"op_product_id"`
	Memo              *string                        `json:"memo"`
	core.Revision     `json:",inline"`
}

// AwsMainAccountExtension 云主账号/云二级账号扩展字段
type AwsMainAccountExtension struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// DecryptSecretKey ...
func (e *AwsMainAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudInitPassword != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudInitPassword)
		if err != nil {
			return err
		}
		e.CloudInitPassword = plainSecretKey
	}
	return nil
}

// GcpMainAccountExtension 云主账号/云二级账号扩展字段
type GcpMainAccountExtension struct {
	CloudProjectID   string `json:"cloud_project_id"`
	CloudProjectName string `json:"cloud_project_name"`
	// gcp为账号邀请制，没有初始号密码
}

// DecryptSecretKey ...
func (e *GcpMainAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	return nil
}

// HuaWeiMainAccountExtension 云主账号/云二级账号扩展字段
type HuaWeiMainAccountExtension struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// DecryptSecretKey ...
func (e *HuaWeiMainAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudInitPassword != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudInitPassword)
		if err != nil {
			return err
		}
		e.CloudInitPassword = plainSecretKey
	}
	return nil
}

// AzureMainAccountExtension 云主账号/云二级账号扩展字段
type AzureMainAccountExtension struct {
	CloudSubscriptionID   string `json:"cloud_subscription_id"`
	CloudSubscriptionName string `json:"cloud_subscription_name"`
	CloudInitPassword     string `json:"cloud_init_password"`
}

// DecryptSecretKey ...
func (e *AzureMainAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudInitPassword != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudInitPassword)
		if err != nil {
			return err
		}
		e.CloudInitPassword = plainSecretKey
	}
	return nil
}

// ZenlayerMainAccountExtension 云主账号/云二级账号扩展字段
type ZenlayerMainAccountExtension struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// DecryptSecretKey ...
func (e *ZenlayerMainAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudInitPassword != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudInitPassword)
		if err != nil {
			return err
		}
		e.CloudInitPassword = plainSecretKey
	}
	return nil
}

// KaopuMainAccountExtension 云主账号/云二级账号扩展字段
type KaopuMainAccountExtension struct {
	CloudMainAccountID   string `json:"cloud_main_account_id"`
	CloudMainAccountName string `json:"cloud_main_account_name"`
	CloudInitPassword    string `json:"cloud_init_password"`
}

// DecryptSecretKey ...
func (e *KaopuMainAccountExtension) DecryptSecretKey(cipher cryptography.Crypto) error {
	if e.CloudInitPassword != "" {
		plainSecretKey, err := cipher.DecryptFromBase64(e.CloudInitPassword)
		if err != nil {
			return err
		}
		e.CloudInitPassword = plainSecretKey
	}
	return nil
}
