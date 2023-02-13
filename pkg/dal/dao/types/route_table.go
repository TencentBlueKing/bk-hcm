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

import routetable "hcm/pkg/dal/table/cloud/route-table"

// ------------------------ route table ------------------------

// RouteTableListResult list route table result.
type RouteTableListResult struct {
	Count   uint64                       `json:"count"`
	Details []routetable.RouteTableTable `json:"details"`
}

// --------------------------- route ---------------------------

// TCloudRouteListResult list tcloud route result.
type TCloudRouteListResult struct {
	Count   uint64                        `json:"count"`
	Details []routetable.TCloudRouteTable `json:"details"`
}

// AwsRouteListResult list aws route result.
type AwsRouteListResult struct {
	Count   uint64                     `json:"count"`
	Details []routetable.AwsRouteTable `json:"details"`
}

// AzureRouteListResult list azure route result.
type AzureRouteListResult struct {
	Count   uint64                       `json:"count"`
	Details []routetable.AzureRouteTable `json:"details"`
}

// HuaWeiRouteListResult list huawei route result.
type HuaWeiRouteListResult struct {
	Count   uint64                        `json:"count"`
	Details []routetable.HuaWeiRouteTable `json:"details"`
}

// GcpRouteListResult list gcp route result.
type GcpRouteListResult struct {
	Count   uint64                     `json:"count"`
	Details []routetable.GcpRouteTable `json:"details"`
}
