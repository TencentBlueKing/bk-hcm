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
	"hcm/pkg/adaptor/types"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateHuaWeiSecurityGroup create huawei security group.
func (g *securityGroup) CreateHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupCreateReq[proto.BaseSecurityGroupAttachment])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.HuaWei(cts.Kit, req.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSecurityGroupCreateOption{
		Region:      req.Spec.Region,
		Name:        req.Spec.Name,
		Description: req.Spec.Memo,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create huawei security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SecurityGroupCreateReq[corecloud.HuaWeiSecurityGroupExtension]{
		Spec: &corecloud.SecurityGroupSpec{
			CloudID:   sg.Id,
			Assigned:  false,
			Region:    req.Spec.Region,
			Name:      sg.Name,
			Memo:      &sg.Description,
			AccountID: req.Spec.AccountID,
		},
		Extension: &corecloud.HuaWeiSecurityGroupExtension{
			CloudProjectID:           sg.ProjectId,
			CloudEnterpriseProjectID: sg.EnterpriseProjectId,
		},
	}
	result, err := g.dataCli.SecurityGroup().CreateHuaWeiSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// DeleteHuaWeiSecurityGroup delete huawei security group.
func (g *securityGroup) DeleteHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.SecurityGroup().GetHuaWeiSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.SecurityGroupDeleteOption{
		Region:  sg.Spec.Region,
		CloudID: sg.Spec.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete huawei security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := g.dataCli.SecurityGroup().Delete(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateHuaWeiSecurityGroup update huawei security group.
func (g *securityGroup) UpdateHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.SecurityGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.SecurityGroup().GetHuaWeiSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.HuaWei(cts.Kit, sg.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSecurityGroupUpdateOption{
		CloudID:     sg.Spec.CloudID,
		Region:      sg.Spec.Region,
		Name:        req.Spec.Name,
		Description: req.Spec.Memo,
	}
	if err := client.UpdateSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update huawei security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.SecurityGroupUpdateReq[corecloud.HuaWeiSecurityGroupExtension]{
		Spec: &protocloud.SecurityGroupSpecUpdate{
			Name: req.Spec.Name,
			Memo: req.Spec.Memo,
		},
	}
	if err := g.dataCli.SecurityGroup().UpdateHuaWeiSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id,
		updateReq); err != nil {

		logs.Errorf("request dataservice update huawei security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
