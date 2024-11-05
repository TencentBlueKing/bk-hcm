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

package securitygroup

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListResourceIdBySecurityGroup list resource id by security group
func (svc *securityGroupSvc) ListResourceIdBySecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listResourceIdBySecurityGroup(cts, handler.ResOperateAuth)
}

// ListBizResourceIDBySecurityGroup list biz resource id by security group
func (svc *securityGroupSvc) ListBizResourceIDBySecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listResourceIdBySecurityGroup(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listResourceIdBySecurityGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get security group vendor failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: baseInfo})
	if err != nil {
		return nil, err
	}

	return svc.client.DataService().Global.SGCommonRel.ListSgCommonRels(cts.Kit, req)
}

// ListCvmIdBySecurityGroup list cvm id by security group
func (svc *securityGroupSvc) ListCvmIdBySecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listCvmIDBySecurityGroup(cts, handler.ResOperateAuth)
}

// ListBizCvmIdBySecurityGroup list biz cvm id by security group
func (svc *securityGroupSvc) ListBizCvmIdBySecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listCvmIDBySecurityGroup(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listCvmIDBySecurityGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get security group vendor failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: baseInfo})
	if err != nil {
		return nil, err
	}

	return svc.client.DataService().Global.SGCvmRel.ListSgCvmRels(cts.Kit.Ctx, cts.Kit.Header(), req)
}
