/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListTargetGroup list target group.
func (svc *lbSvc) ListTargetGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetGroup(cts, handler.ListResourceAuthRes)
}

// ListBizTargetGroup list biz target group.
func (svc *lbSvc) ListBizTargetGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listTargetGroup(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) listTargetGroup(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.TargetGroup, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list target group auth failed, noPermFlag: %v, err: %v, rid: %s", noPermFlag, err, cts.Kit.Rid)
		return nil, err
	}

	resList := &cslb.ListTargetGroupResult{Count: 0, Details: make([]cslb.ListTargetGroupSummary, 0)}
	if noPermFlag {
		logs.Errorf("list target group no perm auth, noPermFlag: %v, expr: %+v, rid: %s", noPermFlag, expr, cts.Kit.Rid)
		return resList, nil
	}

	tgReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
		Fields: req.Fields,
	}
	targetGroupList, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(cts.Kit, tgReq)
	if err != nil {
		logs.Errorf("list target group db failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if req.Page.Count {
		resList.Count = targetGroupList.Count
		return resList, nil
	}
	if len(targetGroupList.Details) == 0 {
		return resList, nil
	}

	resList.Count = targetGroupList.Count
	targetGroupIDs := make([]string, 0)
	for _, item := range targetGroupList.Details {
		targetGroupIDs = append(targetGroupIDs, item.ID)
		resList.Details = append(resList.Details, cslb.ListTargetGroupSummary{
			BaseTargetGroup: item,
		})
	}

	lbMap, tgLbMap, tgLblMap, err := svc.getLbAndLblMapByTgIDs(cts.Kit, targetGroupIDs)
	if err != nil {
		logs.Errorf("get lb and lbl map by tgids failed, tgIDs: %v, err: %v, rid: %s", targetGroupIDs, err, cts.Kit.Rid)
		return nil, err
	}

	for idx, tgItem := range resList.Details {
		resList.Details[idx].ListenerNum = tgLblMap[tgItem.ID]
		tmpLbID := tgLbMap[tgItem.ID]
		lbInfo, ok := lbMap[tmpLbID]
		if !ok {
			continue
		}
		resList.Details[idx].LbID = lbInfo.ID
		resList.Details[idx].LbName = lbInfo.Name
		resList.Details[idx].PrivateIPv4Addresses = lbInfo.PrivateIPv4Addresses
		resList.Details[idx].PrivateIPv6Addresses = lbInfo.PrivateIPv6Addresses
		resList.Details[idx].PublicIPv4Addresses = lbInfo.PublicIPv4Addresses
		resList.Details[idx].PublicIPv6Addresses = lbInfo.PublicIPv6Addresses
	}

	return resList, nil
}

func (svc *lbSvc) getLbAndLblMapByTgIDs(kt *kit.Kit, targetGroupIDs []string) (map[string]corelb.BaseLoadBalancer,
	map[string]string, map[string]int64, error) {

	// 根据目标组ID数组，批量查询负载均衡ID、监听器ID等信息
	tgListenerRelList, err := svc.listTGListenerRuleRelMapByTGIDs(kt, targetGroupIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	lbIDs := make([]string, 0)
	tgLbMap := make(map[string]string, len(tgListenerRelList))
	tgLblMap := make(map[string]int64)
	existLbl := make(map[string]struct{}, 0)
	for _, rel := range tgListenerRelList {
		lbIDs = append(lbIDs, rel.LbID)
		tgLbMap[rel.TargetGroupID] = rel.LbID
		if _, ok := existLbl[rel.TargetGroupID+rel.LblID]; ok {
			continue
		}
		existLbl[rel.TargetGroupID+rel.LblID] = struct{}{}
		tgLblMap[rel.TargetGroupID]++
	}

	// 根据负载均衡ID数组，批量查询负载均衡基本信息
	lbMap, err := lblogic.ListLoadBalancerMap(kt, svc.client.DataService(), lbIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return lbMap, tgLbMap, tgLblMap, nil
}

func (svc *lbSvc) listTGListenerRuleRelMapByTGIDs(kt *kit.Kit, tgIDs []string) (
	[]corelb.BaseTargetListenerRuleRel, error) {

	req := &core.ListReq{
		Filter: tools.ContainersExpression("target_group_id", tgIDs),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, req)
	if err != nil {
		logs.Errorf("list target group listener rel failed, tgIDs: %v, err: %v, rid: %s", tgIDs, err, kt.Rid)
		return nil, err
	}

	return list.Details, nil
}

// GetTargetGroup get target group.
func (svc *lbSvc) GetTargetGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.getTargetGroup(cts, handler.ListResourceAuthRes)
}

// GetBizTargetGroup get biz target group.
func (svc *lbSvc) GetBizTargetGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.getTargetGroup(cts, handler.ListBizAuthRes)
}

func (svc *lbSvc) getTargetGroup(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.TargetGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get target group basic info failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		logs.Errorf("get target group no perm auth, noPerm: %v, rid: %s", noPerm, cts.Kit.Rid)
		return nil, errf.New(errf.PermissionDenied, "permission denied for get target group")
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.getTCloudTargetGroup(cts.Kit, id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}

func (svc *lbSvc) getTCloudTargetGroup(kt *kit.Kit, tgID string) (*cslb.GetTargetGroupDetail, error) {
	targetGroupInfo, err := svc.client.DataService().TCloud.LoadBalancer.GetTargetGroup(kt, tgID)
	if err != nil {
		logs.Errorf("get tcloud target group detail failed, tgID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
		return nil, err
	}

	targetList, err := svc.getTargetByTGIDs(kt, []string{tgID})
	if err != nil {
		logs.Errorf("list target db failed, tgID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
		return nil, err
	}

	result := &cslb.GetTargetGroupDetail{
		BaseTargetGroup: targetGroupInfo.BaseTargetGroup,
		TargetList:      targetList,
	}

	return result, nil
}

// 查询目标组，查不到时返回nil
func (svc *lbSvc) getTargetGroupByID(kt *kit.Kit, targetGroupID string) (*corelb.BaseTargetGroup, error) {

	tgReq := &core.ListReq{
		Filter: tools.EqualExpression("id", targetGroupID),
		Page:   core.NewDefaultBasePage(),
	}
	targetGroupInfo, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroup(kt, tgReq)
	if err != nil {
		logs.Errorf("list target group failed, tgID: %s, err: %v, rid: %s", targetGroupID, err, kt.Rid)
		return nil, err
	}
	if len(targetGroupInfo.Details) == 0 {
		return nil, nil
	}
	return cvt.ValToPtr(targetGroupInfo.Details[0]), nil
}

func (svc *lbSvc) getTargetByTGIDs(kt *kit.Kit, targetGroupIDs []string) ([]corelb.BaseTarget, error) {
	tgReq := &core.ListReq{
		Filter: tools.ContainersExpression("target_group_id", targetGroupIDs),
		Page:   core.NewDefaultBasePage(),
	}
	targetResult, err := svc.client.DataService().Global.LoadBalancer.ListTarget(kt, tgReq)
	if err != nil {
		logs.Errorf("list target by tgIDs failed, tgIDs: %v, err: %v, rid: %s", targetGroupIDs, err, kt.Rid)
		return nil, err
	}

	return targetResult.Details, nil
}

// StatBizTargetWeight 统计目标组下的RS权重情况
func (svc *lbSvc) StatBizTargetWeight(cts *rest.Contexts) (any, error) {
	return svc.statTargetWeight(cts, handler.BizOperateAuth)
}

// ListTargetWeightNumMap ...
func (svc *lbSvc) statTargetWeight(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (any, error) {

	req := new(cslb.ListTargetWeightNumReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.TargetGroupCloudResType,
		IDs:          req.TargetGroupIDs,
	}

	tgInfos, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	// list authorized instances
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.TargetGroup,
		Action:     meta.Find,
		BasicInfos: tgInfos,
	})
	if err != nil {
		return nil, err
	}
	// 获取rs
	targetResp, err := svc.getTargetByTGIDs(cts.Kit, req.TargetGroupIDs)
	if err != nil {
		return nil, err
	}

	targetWeightMap := make(map[string]cslb.TargetGroupRsWeightNum, 0)
	for _, item := range targetResp {
		tmpTarget := targetWeightMap[item.TargetGroupID]
		if cvt.PtrToVal(item.Weight) == 0 {
			tmpTarget.RsWeightZeroNum++
		} else {
			tmpTarget.RsWeightNonZeroNum++
		}
		targetWeightMap[item.TargetGroupID] = tmpTarget
	}

	result := make([]cslb.TargetGroupRsWeightNum, 0, len(req.TargetGroupIDs))
	for _, tgID := range req.TargetGroupIDs {
		info := targetWeightMap[tgID]
		info.TargetGroupID = tgID
		result = append(result, info)
	}
	return result, nil
}
