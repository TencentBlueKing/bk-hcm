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

// Package mainaccount Package service defines service.
package mainaccount

import (
	proto "hcm/pkg/api/hc-service/main-account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// GcpCreateMainAccount 创建gcp账号
func (s *service) GcpCreateMainAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.CreateGcpMainAccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 1、获取一级账号Gcp Client
	client, err := s.ad.GcpRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		return nil, err
	}

	// 2、创建项目
	projectId, err := client.CreateProject(cts.Kit, req.ProjectName, req.CloudOrganization)
	if err != nil {
		return nil, err
	}

	// 3、为项目设置账单信息, 如果没有后台设置结算账号，则使用默认的结算账号设置
	if req.CloudBillingAccount != "" {
		err = client.UpdateBillingInfo(cts.Kit, projectId, req.CloudBillingAccount)
		if err != nil {
			return nil, err
		}
	}

	// 4、把申请账号加入项目中，授予权限
	err = client.BindingProjectEditor(cts.Kit, projectId, req.Email)
	if err != nil {
		return nil, err
	}

	result := proto.CreateGcpMainAccountResp{
		ProjectName: req.ProjectName,
		ProjectID:   projectId,
	}
	return result, nil
}
