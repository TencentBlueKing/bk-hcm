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

import (
	"hcm/pkg/dal/table/cloud"
	tableaccount "hcm/pkg/dal/table/cloud/sub-account"
)

// ListAccountDetails list account details.
type ListAccountDetails struct {
	Count   uint64                `json:"count,omitempty"`
	Details []*cloud.AccountTable `json:"details,omitempty"`
}

// ListSubAccountDetails list sub account details.
type ListSubAccountDetails struct {
	Count   uint64               `json:"count,omitempty"`
	Details []tableaccount.Table `json:"details,omitempty"`
}

// Account ...
type Account struct {
	cloud.AccountTable `json:",inline"`
	BkBizIDs           []int64 `db:"bk_biz_ids" json:"bk_biz_ids"`
}
