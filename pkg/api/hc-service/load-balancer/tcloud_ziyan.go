/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package hclb

import (
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/validator"
)

// TCloudZiyanLoadBalancerCreateReq tcloud batch create req.
type TCloudZiyanLoadBalancerCreateReq struct {
	TCloudLoadBalancerCreateReq `json:",inline"`
	ZhiTong                     *bool    `json:"zhi_tong"`
	TgwGroupName                *string  `json:"tgw_group_name"`
	ClusterIDs                  []string `json:"cluster_ids"`
}

// TCloudZiyanLoadBalancerInquiryReq 自研云CLB询价请求
type TCloudZiyanLoadBalancerInquiryReq struct {
	TCloudLoadBalancerInquiryReq `json:",inline" validate:"required"`
}

// Validate inquiry request.
func (req *TCloudZiyanLoadBalancerInquiryReq) Validate() error {
	return req.TCloudLoadBalancerInquiryReq.Validate()
}

// TCloudDescribeSlaCapacityOption ...
type TCloudDescribeSlaCapacityOption struct {
	AccountID                               string `json:"account_id" validate:"required"`
	*typelb.TCloudDescribeSlaCapacityOption `json:",inline" validate:"required"`
}

// Validate tcloud clb list option.
func (opt TCloudDescribeSlaCapacityOption) Validate() error {
	return validator.Validate.Struct(opt)
}
