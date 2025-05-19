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

	sglogic "hcm/cmd/cloud-server/logics/security-group"
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
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// BatchAssignBiz batch assign biz.
func (svc *securityGroupSvc) BatchAssignBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.BatchAssignBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	securityGroups, err := svc.listSecurityGroupsByID(cts.Kit, req.IDs)
	if err != nil {
		logs.Errorf("listSecurityGroups failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}
	if err := svc.checkAssignPermission(cts, securityGroups); err != nil {
		logs.Errorf("checkAssignPermission failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	previewResult, err := svc.checkSecurityGroupAssignable(cts.Kit, securityGroups)
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
		err := svc.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupCommonInfo(kt.Ctx, kt.Header(),
			update)
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

	securityGroups, err := svc.listSecurityGroupsByID(cts.Kit, req.IDs)
	if err != nil {
		logs.Errorf("listSecurityGroups failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}
	if err := svc.checkAssignPermission(cts, securityGroups); err != nil {
		logs.Errorf("check assign permission failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.checkSecurityGroupAssignable(cts.Kit, securityGroups)
}

func (svc *securityGroupSvc) checkAssignPermission(cts *rest.Contexts, sgInfos []cloud.BaseSecurityGroup) error {

	attrList := make([]meta.ResourceAttribute, 0, len(sgInfos))
	for i := range sgInfos {
		sg := sgInfos[i]
		attr := meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:   meta.SecurityGroup,
				Action: meta.Assign, ResourceID: sg.AccountID,
			},
			BizID: sg.MgmtBizID,
		}
		attrList = append(attrList, attr)
	}

	decisions, _, err := svc.authorizer.Authorize(cts.Kit, attrList...)
	if err != nil {
		return err
	}
	for i := range decisions {
		sg := sgInfos[i]
		if !decisions[i].Authorized {
			return errf.Newf(errf.PermissionDenied, "permission denied: %s(%s)", sg.CloudID, sg.ID)
		}
	}

	return nil
}

func (svc *securityGroupSvc) checkSecurityGroupAssignable(kt *kit.Kit, sgInfos []cloud.BaseSecurityGroup) (
	[]*proto.AssignBizPreviewResp, error) {

	resultList := make([]*proto.AssignBizPreviewResp, 0, len(sgInfos))
	for _, sg := range sgInfos {
		item := &proto.AssignBizPreviewResp{
			ID:            sg.ID,
			Assignable:    true,
			AssignedBizID: sg.MgmtBizID,
		}
		if err := svc.validateSecurityGroupRuleRel(kt, sg, item); err != nil {
			logs.Errorf("validateSecurityGroupRuleRel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		validateSecurityGroupManagerAndBakManager(sg, item)
		validateSecurityGroupManagementTypeAndBizID(sg, item)
		resultList = append(resultList, item)
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

	cloudSGToSgRulesMap, err := sglogic.ListSecurityGroupRulesByCloudTargetSGID(kt,
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
