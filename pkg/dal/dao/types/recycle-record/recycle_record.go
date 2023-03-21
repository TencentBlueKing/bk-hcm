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

package recyclerecord

import (
	"hcm/pkg/criteria/enumor"
	rr "hcm/pkg/dal/table/recycle-record"
)

// RecycleRecordListResult list recycle record result.
type RecycleRecordListResult struct {
	Count   uint64                  `json:"count"`
	Details []rr.RecycleRecordTable `json:"details"`
}

// RecycleResourceInfo define recycle resource info.
type RecycleResourceInfo struct {
	Vendor    enumor.Vendor `db:"vendor" json:"vendor"`
	ID        string        `db:"id" json:"id"`
	CloudID   string        `db:"cloud_id" json:"cloud_id"`
	Name      string        `db:"name" json:"name"`
	BkBizID   int64         `db:"bk_biz_id" json:"bk_biz_id"`
	AccountID string        `db:"account_id" json:"account_id"`
	Region    string        `db:"region" json:"region"`
}

// ResourceUpdateOptions define resource update for recycle/recover options.
type ResourceUpdateOptions struct {
	ResType enumor.CloudResourceType
	IDs     []string
	Status  string
	BkBizID int64
}
