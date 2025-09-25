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

package loadbalancer

import (
	"hcm/pkg/criteria/enumor"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/types"
)

// ListLoadBalancerTargetDetails list load balancer target details.
type ListLoadBalancerTargetDetails struct {
	Count   uint64                            `json:"count,omitempty"`
	Details []tablelb.LoadBalancerTargetTable `json:"details,omitempty"`
}

// ListInstInfoDetails list instance info details.
type ListInstInfoDetails struct {
	Count   uint64         `json:"count,omitempty"`
	Details []ListInstInfo `json:"details,omitempty"`
}

// ListInstInfo list instance info.
type ListInstInfo struct {
	InstID      string            `db:"inst_id" json:"inst_id"`
	InstType    enumor.InstType   `db:"inst_type" json:"inst_type"`
	InstName    string            `db:"inst_name" json:"inst_name"`
	IP          string            `db:"ip" json:"ip"`
	Zone        string            `db:"zone" json:"zone"`
	CloudVpcIDs types.StringArray `db:"cloud_vpc_ids" json:"cloud_vpc_ids"`
}
