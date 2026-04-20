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

package subaccountsecret

import (
	"hcm/pkg/criteria/enumor"
)

// TCloudSubAccountSecretExtension 腾讯云子账号密钥扩展
type TCloudSubAccountSecretExtension struct {
	CloudSecretID      string `json:"cloud_secret_id"`
	CloudMainAccountID string `json:"cloud_main_account_id"`
	CloudSubAccountID  string `json:"cloud_sub_account_id"`
}

// GetCloudSecretID 返回云侧密钥唯一标识
func (e TCloudSubAccountSecretExtension) GetCloudSecretID() string {
	return e.CloudSecretID
}

// TCloudSubAccountSecretJoinExtension 腾讯云子账号密钥关联子账号扩展
type TCloudSubAccountSecretJoinExtension struct {
	TCloudSubAccountSecretExtension `json:",inline"`
	// 来源为子账号扩展字段
	ConsoleLogin *enumor.SubAccountConsoleLogin `json:"console_login"`
}
