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

package application

import (
	"fmt"

	"hcm/cmd/cloud-server/service/application/handlers"
	accounthandler "hcm/cmd/cloud-server/service/application/handlers/account"
	tcloudcvmhandler "hcm/cmd/cloud-server/service/application/handlers/cvm/tcloud"
	proto "hcm/pkg/api/cloud-server/application"
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// create 创建申请单的通用逻辑
func (a *applicationSvc) create(cts *rest.Contexts, handler handlers.ApplicationHandler) (interface{}, error) {
	// 校验数据正确性
	if err := handler.CheckReq(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 预处理数据
	if err := handler.PrepareReq(); err != nil {
		return nil, err
	}

	// 查询审批流程服务ID
	applicationType := handler.GetType()
	serviceID, err := a.getApprovalProcessServiceID(cts, applicationType)
	if err != nil {
		return "", fmt.Errorf("get approval process service id failed, err: %v", err)
	}

	// 生成ITSM的回调地址
	callback := a.getCallbackUrl()

	// 调用ITSM
	sn, err := handler.CreateITSMTicket(serviceID, callback)
	if err != nil {
		return nil, err
	}

	// 调用DB创建单据
	content, err := json.MarshalToString(handler.GenerateApplicationContent())
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("json marshal request data failed, err: %w", err))
	}

	result, err := a.client.DataService().Global.Application.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.ApplicationCreateReq{
			SN:             sn,
			Type:           applicationType,
			Status:         enumor.Pending,
			Applicant:      cts.Kit.User,
			Content:        content,
			DeliveryDetail: "{}",
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parseReqFromRequestBody[T any](cts *rest.Contexts) (*T, error) {
	req := new(T)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	return req, nil
}

// CreateForAddAccount ...
func (a *applicationSvc) CreateForAddAccount(cts *rest.Contexts) (interface{}, error) {
	req, err := parseReqFromRequestBody[proto.AccountAddReq](cts)
	if err != nil {
		return nil, err
	}
	handler := accounthandler.NewApplicationOfAddAccount(a.getHandlerOption(cts), req, a.platformManagers)

	return a.create(cts, handler)
}

// CreateForCreateCvm ...
func (a *applicationSvc) CreateForCreateCvm(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := a.getHandlerOption(cts)

	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromRequestBody[proto.TCloudCvmCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := tcloudcvmhandler.NewApplicationOfCreateTCloudCvm(opt, req, a.platformManagers, false)
		return a.create(cts, handler)
	}

	return nil, nil
}
