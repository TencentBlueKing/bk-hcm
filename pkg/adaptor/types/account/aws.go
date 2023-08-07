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
	"time"

	"hcm/pkg/tools/converter"
)

// AwsAccount  define aws account.
type AwsAccount struct {
	Arn             *string    `json:"arn"`
	Email           *string    `json:"email"`
	ID              *string    `json:"id"`
	JoinedMethod    *string    `json:"joined_method"`
	JoinedTimestamp *time.Time `json:"joined_timestamp"`
	Name            *string    `json:"name"`
	Status          *string    `json:"status"`
}

// GetCloudID ...
func (account AwsAccount) GetCloudID() string {
	return converter.PtrToVal(account.ID)
}

// AwsListAccountResult defines aws list account result.
type AwsListAccountResult struct {
	NextToken *string      `json:"next_token,omitempty"`
	Details   []AwsAccount `json:"details"`
}
