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

package types

import "hcm/pkg/criteria/errf"

// TCloudAccountInfo define tencent cloud account info that is used to validate account.
type TCloudAccountInfo struct {
	CloudMainAccountID string `json:"cloud_main_account_id"`
	CloudSubAccountID  string `json:"cloud_sub_account_id"`
}

// Validate TCloudAccountInfo.
func (t *TCloudAccountInfo) Validate() error {
	if len(t.CloudMainAccountID) == 0 {
		return errf.New(errf.InvalidParameter, "main account id is required")
	}

	if len(t.CloudSubAccountID) == 0 {
		return errf.New(errf.InvalidParameter, "account id is required")
	}

	return nil
}

// AwsAccountInfo define aws account info that used to check account.
type AwsAccountInfo struct {
	CloudAccountID   string `json:"cloud_account_id"`
	CloudIamUsername string `json:"cloud_iam_username"`
}

// Validate AwsAccountInfo
func (a *AwsAccountInfo) Validate() error {
	if len(a.CloudAccountID) == 0 {
		return errf.New(errf.InvalidParameter, "account id is required")
	}

	if len(a.CloudIamUsername) == 0 {
		return errf.New(errf.InvalidParameter, "iam user name is required")
	}

	return nil
}

// HuaWeiAccountInfo define huawei account info that used to check account.
type HuaWeiAccountInfo struct {
	CloudSubAccountID   string `json:"cloud_sub_account_id"`
	CloudSubAccountName string `json:"cloud_sub_account_name"`
	CloudIamUserID      string `json:"cloud_iam_user_id"`
	CloudIamUsername    string `json:"cloud_iam_username"`
}
