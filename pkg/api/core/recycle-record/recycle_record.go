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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// RecycleRecord defines recycle record info.
type RecycleRecord struct {
	BaseRecycleRecord `json:",inline"`
	Detail            interface{} `json:"detail"`
}

// BaseRecycleRecord defines recycle record basic info.
type BaseRecycleRecord struct {
	ID            string                     `json:"id"`
	TaskID        string                     `json:"task_id"`
	RecycleType   enumor.RecycleType         `json:"recycle_type"`
	Vendor        enumor.Vendor              `json:"vendor"`
	ResType       enumor.CloudResourceType   `json:"res_type"`
	ResID         string                     `json:"res_id"`
	CloudResID    string                     `json:"cloud_res_id"`
	ResName       string                     `json:"res_name"`
	BkBizID       int64                      `json:"bk_biz_id"`
	AccountID     string                     `json:"account_id"`
	Region        string                     `json:"region"`
	Status        enumor.RecycleRecordStatus `json:"status"`
	core.Revision `json:",inline"`
}
