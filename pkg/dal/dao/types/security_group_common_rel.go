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
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/cloud"
)

// ListSecurityGroupCommonRelDetails list security group and common relation details.
type ListSecurityGroupCommonRelDetails struct {
	Count   uint64                              `json:"count,omitempty"`
	Details []cloud.SecurityGroupCommonRelTable `json:"details,omitempty"`
}

// ListSGCommonRelsJoinSGDetails list security group common relation join security group details.
type ListSGCommonRelsJoinSGDetails struct {
	Count   uint64                      `json:"count,omitempty"`
	Details []SecurityGroupWithCommonID `json:"details,omitempty"`
}

// SecurityGroupWithCommonID security group with common id.
type SecurityGroupWithCommonID struct {
	cloud.SecurityGroupTable `db:",inline" json:",inline"`
	ResID                    string                   `db:"res_id" json:"res_id"`
	ResType                  enumor.CloudResourceType `db:"res_type" json:"res_type"`
	Priority                 int64                    `db:"priority" json:"priority"`
	RelCreator               string                   `db:"rel_creator" json:"rel_creator"`
	RelCreatedAt             *time.Time               `db:"rel_created_at" json:"rel_created_at"`
}
