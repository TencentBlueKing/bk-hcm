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

package mainaccount

import "hcm/pkg/thirdparty/api-gateway/itsm"

// PrepareReq 预处理申请单数据
func (a *ApplicationOfUpdateMainAccount) PrepareReq() error {
	// 二级账号变更不包含敏感信息，无需处理
	return nil
}

// GenerateApplicationContent 生成存储到DB的申请单content的内容，Interface格式，便于统一处理
func (a *ApplicationOfUpdateMainAccount) GenerateApplicationContent() interface{} {
	return a.req
}

// PrepareReqFromContent 申请单内容从DB里获取后可以进行预处理，便于资源交付时资源请求
func (a *ApplicationOfUpdateMainAccount) PrepareReqFromContent() error {
	return nil
}

// GetItsmApprover 获取itsm审批人信息
func (a *ApplicationOfUpdateMainAccount) GetItsmApprover(managers []string) []itsm.VariableApprover {
	return []itsm.VariableApprover{
		{
			Variable:  "platform_manager",
			Approvers: managers,
		},
	}
}
