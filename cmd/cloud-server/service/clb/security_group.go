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

package clb

import (
	proto "hcm/pkg/api/cloud-server/clb"
	"hcm/pkg/api/core"
	coreclb "hcm/pkg/api/core/cloud/clb"
	dsproto "hcm/pkg/api/data-service"
	protoaudit "hcm/pkg/api/data-service/audit"
	dataproto "hcm/pkg/api/data-service/cloud"
	protoclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// BatchBindSecurityGroupBizClb batch bind security group biz clb.
func (svc *clbSvc) BatchBindSecurityGroupBizClb(cts *rest.Contexts) (interface{}, error) {
	return svc.bindSecurityGroupBizClb(cts, handler.BizOperateAuth)
}

func (svc *clbSvc) bindSecurityGroupBizClb(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.BatchBindClbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ClbCloudResType,
		IDs:          []string{req.ClbID},
		Fields:       append(types.CommonBasicInfoFields, "region"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list clb resource basic info failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Clb,
		Action: meta.Associate, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResBaseOperationAudit(cts.Kit, enumor.ClbAuditResType, protoaudit.Associate,
		[]string{req.ClbID}); err != nil {
		logs.Errorf("create operation audit associate failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	// 根据ClbID查询负载均衡基本信息
	clbInfo, sgComList, err := svc.getClbInfoAndSGComRels(cts.Kit, req.ClbID)
	if err != nil {
		return nil, err
	}

	allSGIDs := make([]string, 0)
	sgComReq := &dataproto.SGCommonRelBatchCreateReq{
		Rels: make([]dataproto.SGCommonRelCreate, 0, len(req.SecurityGroupIDs)),
	}
	tmpPriority := int64(0)
	for _, sg := range sgComList.Details {
		allSGIDs = append(allSGIDs, sg.SecurityGroupID)
		tmpPriority = sg.Priority
	}

	for _, tmpSGID := range req.SecurityGroupIDs {
		tmpPriority++
		allSGIDs = append(allSGIDs, tmpSGID)
		sgComReq.Rels = append(sgComReq.Rels, dataproto.SGCommonRelCreate{
			SecurityGroupID: tmpSGID,
			ResID:           req.ClbID,
			ResType:         enumor.ClbCloudResType,
			Priority:        tmpPriority,
		})
	}

	opt := &protoclb.TCloudSetClbSecurityGroupReq{
		AccountID:      clbInfo.AccountID,
		LoadBalancerID: req.ClbID,
		SecurityGroups: allSGIDs,
	}
	err = svc.client.HCService().TCloud.Clb.BatchSetTCloudClbSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("call hcservice to set bind tcloud clb security group failed, clbID: %s, sgIDs: %v, "+
			"err: %v, rid: %s", req.ClbID, allSGIDs, err, cts.Kit.Rid)
		return nil, err
	}

	err = svc.client.DataService().Global.SGCommonRel.BatchCreate(cts.Kit, sgComReq)
	if err != nil {
		logs.Errorf("call dataservice to create tcloud clb security group failed, clbID: %s, sgIDs: %v, "+
			"err: %v, rid: %s", req.ClbID, allSGIDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *clbSvc) getClbInfoAndSGComRels(kt *kit.Kit, clbID string) (
	*coreclb.BaseClb, *dataproto.SGCommonRelListResult, error) {

	clbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", clbID),
		Page:   core.NewDefaultBasePage(),
	}
	clbList, err := svc.client.DataService().Global.LoadBalancer.ListClb(kt, clbReq)
	if err != nil {
		logs.Errorf("list load balancer by id failed, id: %s, err: %v, rid: %s", clbID, err, kt.Rid)
		return nil, nil, err
	}

	if len(clbList.Details) == 0 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "not found clb id: %s", clbID)
	}

	clbInfo := clbList.Details[0]
	// 查询目前绑定的安全组
	sgcomReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "res_id",
					Op:    filter.Equal.Factory(),
					Value: clbID,
				},
				&filter.AtomRule{
					Field: "res_type",
					Op:    filter.Equal.Factory(),
					Value: meta.Clb,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: clbInfo.Vendor,
				},
			},
		},
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "priority", Order: "ASC"},
	}
	sgComList, err := svc.client.DataService().Global.SGCommonRel.List(kt, sgcomReq)
	if err != nil {
		logs.Errorf("call dataserver to list sg common failed, clbID: %s, err: %v, rid: %s", clbID, err, kt.Rid)
		return nil, nil, err
	}

	return &clbInfo, sgComList, nil
}

// UnBindSecurityGroupBizClb unbind security group biz clb.
func (svc *clbSvc) UnBindSecurityGroupBizClb(cts *rest.Contexts) (interface{}, error) {
	return svc.unBindSecurityGroupBizClb(cts, handler.BizOperateAuth)
}

func (svc *clbSvc) unBindSecurityGroupBizClb(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.UnBindClbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ClbCloudResType,
		IDs:          []string{req.ClbID},
		Fields:       append(types.CommonBasicInfoFields, "region"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list clb resource basic info failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Clb,
		Action: meta.Disassociate, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResBaseOperationAudit(cts.Kit, enumor.ClbAuditResType, protoaudit.Disassociate,
		[]string{req.ClbID}); err != nil {
		logs.Errorf("create operation audit disassociate failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	// 根据ClbID查询负载均衡基本信息
	clbInfo, sgComList, err := svc.getClbInfoAndSGComRels(cts.Kit, req.ClbID)
	if err != nil {
		return nil, err
	}

	allSGIDs := make([]string, 0)
	existSG := false
	for _, sg := range sgComList.Details {
		if sg.SecurityGroupID == req.SecurityGroupID {
			existSG = true
			continue
		}
		allSGIDs = append(allSGIDs, sg.SecurityGroupID)
	}
	if !existSG {
		return nil, errf.Newf(errf.RecordNotFound, "not found sg id: %s", req.SecurityGroupID)
	}

	opt := &protoclb.TCloudSetClbSecurityGroupReq{
		AccountID:      clbInfo.AccountID,
		LoadBalancerID: req.ClbID,
		SecurityGroups: allSGIDs,
	}
	err = svc.client.HCService().TCloud.Clb.BatchSetTCloudClbSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("call hcservice to set unbind tcloud clb security group failed, clbID: %s, sgIDs: %v, "+
			"err: %v, rid: %s", req.ClbID, allSGIDs, err, cts.Kit.Rid)
		return nil, err
	}

	sgComDelReq := &dsproto.BatchDeleteReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "res_id",
					Op:    filter.Equal.Factory(),
					Value: req.ClbID,
				},
				&filter.AtomRule{
					Field: "security_group_id",
					Op:    filter.Equal.Factory(),
					Value: req.SecurityGroupID,
				},
			},
		},
	}
	err = svc.client.DataService().Global.SGCommonRel.BatchDelete(cts.Kit, sgComDelReq)
	if err != nil {
		logs.Errorf("call dataservice to delete tcloud clb security group failed, clbID: %s, sgID: %s, "+
			"err: %v, rid: %s", req.ClbID, req.SecurityGroupID, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
