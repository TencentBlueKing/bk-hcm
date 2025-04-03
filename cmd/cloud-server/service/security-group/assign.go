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
	"fmt"

	security_group "hcm/cmd/cloud-server/logics/security-group"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// AssignSecurityGroupToBiz assign security group to biz.
// Deprecated: use BatchAssignSecurityGroupToBiz instead.
func (svc *securityGroupSvc) AssignSecurityGroupToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignSecurityGroupToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.SecurityGroupCloudResType,
		IDs:          req.SecurityGroupIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.SecurityGroup,
			Action: meta.Assign, ResourceID: info.AccountID}, BizID: info.BkBizID})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	if err = svc.checkSGUnAssign(cts.Kit, req); err != nil {
		logs.Errorf("check security group unAssign failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// create assign audit.
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.SecurityGroupAuditResType, req.SecurityGroupIDs, req.BkBizID)
	if err != nil {
		logs.Errorf("create assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	update := &dataproto.SecurityGroupCommonInfoBatchUpdateReq{
		IDs:     req.SecurityGroupIDs,
		BkBizID: req.BkBizID,
	}
	if err := svc.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupCommonInfo(cts.Kit.Ctx,
		cts.Kit.Header(), update); err != nil {

		logs.Errorf("BatchUpdateSecurityGroupCommonInfo failed, err: %v, req: %v, rid: %s", err, update,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *securityGroupSvc) checkSGUnAssign(kt *kit.Kit, req *proto.AssignSecurityGroupToBizReq) error {
	listReq := &dataproto.SecurityGroupListReq{
		Field: []string{"id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "id",
					Op:    filter.In.Factory(),
					Value: req.SecurityGroupIDs,
				},
				&filter.AtomRule{
					Field: "bk_biz_id",
					Op:    filter.NotEqual.Factory(),
					Value: constant.UnassignedBiz,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("ListSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(result.Details) != 0 {
		ids := make([]string, len(result.Details))
		for index, one := range result.Details {
			ids[index] = one.ID
		}
		return fmt.Errorf("security group%v already assigned", ids)
	}

	return nil
}

// BatchAssignBiz batch assign biz.
func (svc *securityGroupSvc) BatchAssignBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.BatchAssignBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authorizeAssignSecurityGroupPermission(cts, req.IDs); err != nil {
		logs.Errorf("authorizeAssignSecurityGroupPermission failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	previewResult, err := svc.checkSecurityGroupAssignable(cts.Kit, req.IDs)
	if err != nil {
		logs.Errorf("checkSecurityGroupAssignable failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	for _, one := range previewResult {
		if !one.Assignable {
			logs.Errorf("security group %s not assignable, reason: %s, rid: %s", one.ID, one.Reason, cts.Kit.Rid)
			return nil, fmt.Errorf("security group %s not assignable, reason: %s", one.ID, one.Reason)
		}
	}

	if err = svc.batchAssignBiz(cts.Kit, previewResult); err != nil {
		logs.Errorf("batchAssignBiz failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

func (svc *securityGroupSvc) batchAssignBiz(kt *kit.Kit, items []*proto.AssignBizPreviewResp) error {
	itemsByBizID := classifier.ClassifySlice(items, func(item *proto.AssignBizPreviewResp) int64 {
		return item.AssignedBizID
	})
	for bizID, securityGroups := range itemsByBizID {
		ids := make([]string, 0, len(securityGroups))
		for _, one := range securityGroups {
			ids = append(ids, one.ID)
		}
		update := &dataproto.SecurityGroupCommonInfoBatchUpdateReq{
			IDs:     ids,
			BkBizID: bizID,
		}
		err := svc.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupCommonInfo(kt.Ctx, kt.Header(), update)
		if err != nil {
			logs.Errorf("BatchUpdateSecurityGroupCommonInfo failed, err: %v, req: %v, rid: %s", err, update,
				kt.Rid)
			return err
		}
	}
	return nil
}

// AssignBizPreview batch assign biz preview.
func (svc *securityGroupSvc) AssignBizPreview(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.BatchAssignBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authorizeAssignSecurityGroupPermission(cts, req.IDs); err != nil {
		logs.Errorf("authorizeAssignSecurityGroupPermission failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.checkSecurityGroupAssignable(cts.Kit, req.IDs)
}

func (svc *securityGroupSvc) authorizeAssignSecurityGroupPermission(cts *rest.Contexts, sgIDs []string) error {
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.SecurityGroupCloudResType,
		IDs:          sgIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return err
	}

	err = handler.ResOperateAuth(cts,
		&handler.ValidWithAuthOption{
			Authorizer: svc.authorizer,
			ResType:    meta.SecurityGroup,
			Action:     meta.Assign,
			BasicInfos: basicInfoMap},
	)
	if err != nil {
		return err
	}
	return nil
}

func (svc *securityGroupSvc) checkSecurityGroupAssignable(kt *kit.Kit, ids []string) (
	[]*proto.AssignBizPreviewResp, error) {

	securityGroups, err := svc.listSecurityGroupsByID(kt, ids)
	if err != nil {
		logs.Errorf("listSecurityGroups failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	resultMap := make(map[string]*proto.AssignBizPreviewResp, len(ids))
	resultList := make([]*proto.AssignBizPreviewResp, 0, len(ids))
	for _, id := range ids {
		item := &proto.AssignBizPreviewResp{
			ID:            id,
			Assignable:    true,
			AssignedBizID: constant.UnassignedBiz,
		}
		resultMap[id] = item
		resultList = append(resultList, item)
	}

	for _, securityGroup := range securityGroups {
		resultMap[securityGroup.ID].AssignedBizID = securityGroup.MgmtBizID
		if err = svc.validateSecurityGroupRuleRel(kt, securityGroup, resultMap[securityGroup.ID]); err != nil {
			logs.Errorf("validateSecurityGroupRuleRel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		validateSecurityGroupManagerAndBakManager(securityGroup, resultMap[securityGroup.ID])
		validateSecurityGroupManagementTypeAndBizID(securityGroup, resultMap[securityGroup.ID])
	}

	return resultList, nil
}

func (svc *securityGroupSvc) listSecurityGroupsByID(kt *kit.Kit, ids []string) ([]cloud.BaseSecurityGroup, error) {
	resultMap := make(map[string]cloud.BaseSecurityGroup, len(ids))
	for _, sgIDs := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listReq := &dataproto.SecurityGroupListReq{
			Filter: tools.ContainersExpression("id", sgIDs),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("ListSecurityGroup failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			resultMap[detail.ID] = detail
		}
	}
	result := make([]cloud.BaseSecurityGroup, 0, len(ids))
	for _, id := range ids {
		item, ok := resultMap[id]
		if !ok {
			logs.Errorf("security group %s not found, rid: %s", id, kt.Rid)
			return nil, fmt.Errorf("security group %s not found", id)
		}
		result = append(result, item)
	}
	return result, nil
}

func (svc *securityGroupSvc) listSecurityGroupsByCloudID(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string) (
	[]cloud.BaseSecurityGroup, error) {

	resultMap := make(map[string]cloud.BaseSecurityGroup, len(cloudIDs))
	for _, ids := range slice.Split(cloudIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &dataproto.SecurityGroupListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", ids),
				tools.RuleEqual("vendor", vendor),
			),
			Page: core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("ListSecurityGroup failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			resultMap[detail.CloudID] = detail
		}
	}
	result := make([]cloud.BaseSecurityGroup, 0, len(cloudIDs))
	for _, id := range cloudIDs {
		item, ok := resultMap[id]
		if !ok {
			logs.Errorf("security group %s not found, rid: %s", id, kt.Rid)
			return nil, fmt.Errorf("security group %s not found", id)
		}
		result = append(result, item)
	}
	return result, nil
}

func validateSecurityGroupManagementTypeAndBizID(securityGroup cloud.BaseSecurityGroup,
	preview *proto.AssignBizPreviewResp) {

	if securityGroup.MgmtType != enumor.MgmtTypeBiz {
		preview.Assignable = false
		preview.Reason = "安全组管理类型为 [未确认] 或 [平台管理], 不可分配"
	}
	if securityGroup.MgmtBizID == constant.UnassignedBiz {
		preview.Assignable = false
		preview.Reason = "安全组管理业务未指定, 不可分配"
	}
	if securityGroup.BkBizID != constant.UnassignedBiz {
		preview.Assignable = false
		preview.Reason = "安全组业务已分配, 不可重复分配"
	}
}

func validateSecurityGroupManagerAndBakManager(securityGroup cloud.BaseSecurityGroup,
	preview *proto.AssignBizPreviewResp) {

	if len(securityGroup.Manager) == 0 || len(securityGroup.BakManager) == 0 {
		preview.Assignable = false
		preview.Reason = "安全组负责人或备份负责人为空, 不可分配"
	}
}

func (svc *securityGroupSvc) validateSecurityGroupRuleRel(kt *kit.Kit, sg cloud.BaseSecurityGroup,
	preview *proto.AssignBizPreviewResp) error {

	cloudSGToSgRulesMap, err := security_group.ListSecurityGroupRulesByCloudTargetSGID(kt,
		svc.client.DataService(), sg.Vendor, sg.ID)
	if err != nil {
		logs.Errorf("listSecurityGroupRulesByCloudTargetSGID failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	securityGroups, err := svc.listSecurityGroupsByCloudID(kt, sg.Vendor, converter.MapKeyToSlice(cloudSGToSgRulesMap))
	if err != nil {
		logs.Errorf("listSecurityGroupsByCloudID failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	for _, one := range securityGroups {
		if one.MgmtBizID != sg.MgmtBizID {
			preview.Assignable = false
			preview.Reason = fmt.Sprintf("安全组规则: %v, 引用的安全组(%s)不属于同一业务, 不可分配",
				cloudSGToSgRulesMap[one.CloudID], one.ID)
			return nil
		}
	}
	return nil
}
