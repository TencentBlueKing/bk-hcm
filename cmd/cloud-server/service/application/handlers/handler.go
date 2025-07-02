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

package handlers

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/thirdparty/api-gateway/itsm"
)

// ApplicationHandler 定义了申请单的表单校验，与itsm对接、审批通过后的资源交付函数
// 创建申请单：CheckReq -> PrepareReq -> CreateITSMTicket -> GenerateApplicationContent -> "SaveToDB"
// 审批通过交付："LoadApplicationFromDB" -> PrepareReqForContent-> CheckReq -> Deliver -> "UpdateStatusToDB"
// Note: 这里创建申请单的请求数据和交付资源的请求数据结构是一样的，这是一种"偷懒"行为，
// 更好的方式是Handler拆分成两种抽象：申请单创建者Creator、申请单交付者Deliverer，然后定义各自的数据结构
type ApplicationHandler interface {
	GetType() enumor.ApplicationType

	// GetItsmApprover 获取itsm审批人信息
	GetItsmApprover(managers []string) []itsm.VariableApprover

	// CheckReq 申请单的表单校验
	CheckReq() error
	// PrepareReq 预处理申请单数据
	PrepareReq() error

	// RenderItsmTitle 渲染ITSM单据标题
	RenderItsmTitle() (string, error)
	// RenderItsmForm 渲染ITSM表单
	RenderItsmForm() (string, error)

	// GenerateApplicationContent 生成存储到DB的申请单内容，Interface格式，便于统一处理
	GenerateApplicationContent() interface{}

	// PrepareReqFromContent 申请单内容从DB里获取后可以进行预处理，便于资源交付时资源请求
	PrepareReqFromContent() error
	// Deliver  审批通过后资源的交付
	Deliver() (status enumor.ApplicationStatus, deliverDetail map[string]interface{}, err error)

	// Complete 审批通过后资源的交付如未完成，需要进一步手动完善申请流程
	Complete() (status enumor.ApplicationStatus, deliverDetail map[string]interface{}, err error)

	// GetBkBizIDs 获取当前的业务IDs
	GetBkBizIDs() []int64
}
