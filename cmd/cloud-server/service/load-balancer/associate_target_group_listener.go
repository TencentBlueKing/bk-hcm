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

package loadbalancer

import (
	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// AssociateTargetGroupListenerRel associate target group listener rel.
func (svc *lbSvc) AssociateTargetGroupListenerRel(cts *rest.Contexts) (interface{}, error) {
	return svc.associateTargetGroupListenerRel(cts, handler.ResOperateAuth)
}

// AssociateBizTargetGroupListenerRel associate biz target group listener rel.
func (svc *lbSvc) AssociateBizTargetGroupListenerRel(cts *rest.Contexts) (interface{}, error) {
	return svc.associateTargetGroupListenerRel(cts, handler.BizOperateAuth)
}

func (svc *lbSvc) associateTargetGroupListenerRel(cts *rest.Contexts,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	req := new(cslb.TargetGroupListenerRelAssociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 鉴权和校验资源分配状态
	basicReq := &dataproto.BatchListResourceBasicInfoReq{
		Items: []dataproto.ListResourceBasicInfoReq{
			{ResourceType: enumor.TargetGroupCloudResType, IDs: []string{req.TargetGroupID},
				Fields: types.CommonBasicInfoFields},
			{ResourceType: enumor.ListenerCloudResType, IDs: []string{req.ListenerID},
				Fields: types.CommonBasicInfoFields},
		},
	}
	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list listener or target group resource basic info failed, err: %v, req: %+v, rid: %s",
			err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Associate, BasicInfos: basicInfos})
	if err != nil {
		logs.Errorf("valid lb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	var vendor enumor.Vendor
	for _, info := range basicInfos {
		vendor = info.Vendor
		break
	}

	// create operation audit.
	audit := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.TargetGroupAuditResType,
		ResID:             req.TargetGroupID,
		Action:            protoaudit.Associate,
		AssociatedResType: enumor.ListenerAuditResType,
		AssociatedResID:   req.ListenerID,
	}
	if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
		logs.Errorf("create target group listener rel operation audit failed, req: %+v, err: %v, rid: %s",
			req, err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return svc.tcloudTargetGroupListenerRel(cts.Kit, req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

func (svc *lbSvc) tcloudTargetGroupListenerRel(kt *kit.Kit, req *cslb.TargetGroupListenerRelAssociateReq) (
	interface{}, error) {

	lblInfo, err := lblogic.GetListenerByID(kt, svc.client.DataService(), req.ListenerID)
	if err != nil {
		return nil, err
	}

	// 查询目标组基本信息
	tg, err := svc.getTargetGroupByID(kt, req.TargetGroupID)
	if err != nil {
		return nil, err
	}
	if tg == nil {
		return nil, errf.Newf(errf.RecordNotFound, "target_group_id: %s not found", req.TargetGroupID)
	}
	// 查询目标组下，是否有RS信息
	targetList, err := svc.getTargetByTGIDs(kt, []string{req.TargetGroupID})
	if err != nil {
		return nil, err
	}
	if len(targetList) > 0 {
		return nil, errf.Newf(errf.InvalidParameter, "target_group_id: %s has bound rs", req.TargetGroupID)
	}

	// 根据ruleID，查询规则详情信息
	ruleReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", req.ListenerRuleID),
			tools.RuleEqual("lbl_id", req.ListenerID),
		),
		Page: core.NewDefaultBasePage(),
	}
	ruleList, err := svc.listRuleWithCondition(kt, ruleReq)
	if err != nil {
		return nil, err
	}
	if len(ruleList.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "listener_rule_id: %s not found or not belong to %s",
			req.ListenerRuleID, req.ListenerID)
	}

	// 检查监听器ID、目标组ID是否已经关联过了
	tgLblRelReq := &core.ListReq{
		Filter: tools.ExpressionOr(
			tools.RuleEqual("target_group_id", req.TargetGroupID),
			tools.RuleEqual("lbl_id", req.ListenerID),
		),
		Page: core.NewDefaultBasePage(),
	}
	tgLblRelList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, tgLblRelReq)
	if err != nil {
		return nil, err
	}
	if len(tgLblRelList.Details) > 0 {
		return nil, errf.Newf(errf.RecordDuplicated, "target group %s or listener %s rel already exists",
			req.TargetGroupID, req.ListenerID)
	}

	relReq := &dataproto.TargetGroupListenerRelCreateReq{
		ListenerRuleID:      req.ListenerRuleID,
		CloudListenerRuleID: ruleList.Details[0].CloudID,
		ListenerRuleType:    ruleList.Details[0].RuleType,
		TargetGroupID:       req.TargetGroupID,
		CloudTargetGroupID:  tg.CloudID,
		LbID:                lblInfo.LbID,
		CloudLbID:           lblInfo.CloudLbID,
		LblID:               req.ListenerID,
		CloudLblID:          lblInfo.CloudID,
		BindingStatus:       enumor.SuccessBindingStatus,
	}
	return svc.client.DataService().Global.LoadBalancer.CreateTargetGroupListenerRel(kt, relReq)
}
