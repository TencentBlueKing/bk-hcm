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

package tcloud

import (
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/thirdparty/api-gateway/itsm"
)

// PrepareReq 预处理请求参数，比如敏感数据加密
func (a *ApplicationOfCreateTCloudLB) PrepareReq() error {

	return nil
}

// GenerateApplicationContent 获取预处理过的数据，以interface格式
func (a *ApplicationOfCreateTCloudLB) GenerateApplicationContent() interface{} {
	// 需要将Vendor也存储进去
	return &struct {
		*hclb.TCloudLoadBalancerCreateReq `json:",inline"`
		Vendor                            enumor.Vendor `json:"vendor"`
	}{
		TCloudLoadBalancerCreateReq: a.req,
		Vendor:                      a.Vendor(),
	}
}

// PrepareReqFromContent 预处理请求参数，对于申请内容来着DB，其实入库前是加密了的
func (a *ApplicationOfCreateTCloudLB) PrepareReqFromContent() error {
	return nil
}

// GetItsmApprover 获取itsm审批人
func (a *ApplicationOfCreateTCloudLB) GetItsmApprover(managers []string) []itsm.VariableApprover {
	return a.GetItsmPlatformAndAccountApprover(managers, a.req.AccountID)
}

// GetBkBizIDs 获取当前的业务IDs
func (a *ApplicationOfCreateTCloudLB) GetBkBizIDs() []int64 {
	return []int64{a.req.BkBizID}
}
