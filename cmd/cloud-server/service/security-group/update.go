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
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// UpdateSecurityGroup update security group.
func (svc *securityGroupSvc) UpdateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.updateSecurityGroup(cts, handler.ResOperateAuth)
}

// UpdateBizSecurityGroup update biz security group.
func (svc *securityGroupSvc) UpdateBizSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.updateSecurityGroup(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) updateSecurityGroup(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.SecurityGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get resource vendor failed, id: %s, err: %s, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Update, BasicInfo: baseInfo})
	if err != nil {
		return nil, err
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = svc.audit.ResUpdateAudit(cts.Kit, enumor.SecurityGroupAuditResType, id, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		updateReq := &hcproto.SecurityGroupUpdateReq{
			Name: req.Name,
			Memo: req.Memo,
		}
		err = svc.client.HCService().TCloud.SecurityGroup.UpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			id, updateReq)

	case enumor.HuaWei:
		updateReq := &hcproto.SecurityGroupUpdateReq{
			Name: req.Name,
			Memo: req.Memo,
		}
		err = svc.client.HCService().HuaWei.SecurityGroup.UpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			id, updateReq)

	case enumor.Azure:
		if len(req.Name) != 0 {
			return nil, errf.Newf(errf.InvalidParameter, "azure resource name not support update")
		}

		updateReq := &hcproto.AzureSecurityGroupUpdateReq{
			Memo: req.Memo,
		}
		err = svc.client.HCService().Azure.SecurityGroup.UpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			id, updateReq)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, baseInfo.Vendor)
	}
	if err != nil {
		logs.Errorf("update security group failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateSGMgmtAttr update security group management attribute.
func (svc *securityGroupSvc) UpdateSGMgmtAttr(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupUpdateMgmtAttrReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.updateSGMgmtAttr(cts, req, handler.ResOperateAuth)
}

// UpdateBizSGMgmtAttr update biz security group management attribute.
func (svc *securityGroupSvc) UpdateBizSGMgmtAttr(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupUpdateMgmtAttrReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.MgmtBizID != 0 {
		return nil, errf.Newf(errf.InvalidParameter, "mgmt_biz_id not support update")
	}

	return svc.updateSGMgmtAttr(cts, req, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) updateSGMgmtAttr(cts *rest.Contexts, req *proto.SecurityGroupUpdateMgmtAttrReq,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get resource sg basic info failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Update, BasicInfo: baseInfo})
	if err != nil {
		return nil, err
	}

	if err := svc.sgLogic.UpdateSGMgmtAttr(cts.Kit, req, id); err != nil {
		logs.Errorf("update security group management attribute failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchUpdateSGMgmtAttr batch update security group management attribute.
func (svc *securityGroupSvc) BatchUpdateSGMgmtAttr(cts *rest.Contexts) (interface{}, error) {
	return svc.batchUpdateSGMgmtAttr(cts, handler.ResOperateAuth)
}

func (svc *securityGroupSvc) batchUpdateSGMgmtAttr(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.BatchUpdateSecurityGroupMgmtAttrReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgIDs := make([]string, len(req.SecurityGroups))
	for i, sg := range req.SecurityGroups {
		sgIDs[i] = sg.ID
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.SecurityGroupCloudResType,
		IDs:          sgIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Update, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err := svc.sgLogic.BatchUpdateSGMgmtAttr(cts.Kit, req.SecurityGroups); err != nil {
		logs.Errorf("batch update security group management attribute failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
