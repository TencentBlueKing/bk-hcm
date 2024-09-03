/*
 *
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
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchBizRemoveTargetGroupRS batch biz remove target group rs.
func (svc *lbSvc) BatchBizRemoveTargetGroupRS(cts *rest.Contexts) (any, error) {
	return svc.batchRemoveTargetGroupRS(cts, handler.BizOperateAuth)
}

// BatchRemoveTargetGroupRS batch remove target group rs.
func (svc *lbSvc) BatchRemoveTargetGroupRS(cts *rest.Contexts) (any, error) {
	return svc.batchRemoveTargetGroupRS(cts, handler.ResOperateAuth)
}

func (svc *lbSvc) batchRemoveTargetGroupRS(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch sops remove target group rs request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch sops remove target auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get sops account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.buildDeleteTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID, enumor.TCloud)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildDeleteTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string,
	vendor enumor.Vendor) (any, error) {

	req := new(cslb.TCloudSopsTargetBatchRemoveReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询规则列表，查出符合条件的目标组
	tgIDsMap, err := svc.parseSOpsTargetParams(kt, accountID, vendor, req.RuleQueryList)
	if err != nil {
		return nil, err
	}
	if len(tgIDsMap) == 0 {
		return nil, errf.New(errf.RecordNotFound, "no matching target groups were found")
	}
	lbTgsMap, err := svc.iterateTargetGroupGroupByCLB(kt, tgIDsMap)
	if err != nil {
		logs.Errorf("iterate target group group by clb failed, tgIDsMap: %v, err: %v, rid: %s",
			tgIDsMap, err, kt.Rid)
		return nil, err
	}

	tgTargetMap := make(map[string][]string)
	for index, tgIDs := range tgIDsMap {
		// 查询每一行筛选出的目标组对应的目标，按照当前行填写的条件进行筛选
		targetList, err := svc.getTargetByTGIDs(kt, tgIDs)
		if err != nil {
			logs.Errorf("get target by target group ids failed, err: %v, tgIDs: %v, rid: %s", err, tgIDs, kt.Rid)
			return nil, err
		}
		rsIPs := req.RuleQueryList[index-1].RsIP
		rsType := req.RuleQueryList[index-1].RsType
		for _, target := range targetList {
			// 筛选rsType
			if string(target.InstType) != rsType {
				continue
			}
			// 筛选rsIp
			for _, rsIp := range target.PrivateIPAddress {
				if slice.IsItemInSlice(rsIPs, rsIp) {
					if _, ok := tgTargetMap[target.TargetGroupID]; !ok {
						tgTargetMap[target.TargetGroupID] = make([]string, 0)
					}
					tgTargetMap[target.TargetGroupID] = append(tgTargetMap[target.TargetGroupID], target.ID)
					continue
				}
			}
		}
	}

	flowStateResults := make([]*core.FlowStateResult, 0)
	for lbID, lbTgIDs := range lbTgsMap {
		removeTargetJSON, err := buildTCloudTargetBatchRemoveReq(lbTgIDs, tgTargetMap)
		if err != nil {
			logs.Errorf("build sops tcloud remove target params parse failed, "+
				"err: %v, accountID: %s, tgIDs: %v, rid: %s",
				err, accountID, lbTgIDs, kt.Rid)
			return nil, err
		}
		logs.Infof(
			"build sops tcloud remove target params success,lbID: %s tgIDs: %v, removeTargetJSON: %s, rid: %s",
			lbID, lbTgIDs, removeTargetJSON, kt.Rid)
		result, err := svc.buildRemoveTCloudTarget(kt, removeTargetJSON, accountID)
		if err != nil {
			return nil, err
		}
		resultValue, ok := result.(*core.FlowStateResult)
		if !ok {
			return nil, fmt.Errorf("buildAddTCloudTarget failed, result: %v", resultValue)
		}
		flowStateResults = append(flowStateResults, resultValue)
	}

	return flowStateResults, nil
}

func buildTCloudTargetBatchRemoveReq(lbTgIDs []string, tgTargetMap map[string][]string) ([]byte, error) {
	params := &cslb.TCloudTargetBatchRemoveReq{
		TargetGroups: []*cslb.TCloudRemoveTargetReq{},
	}
	for _, tmpTgID := range lbTgIDs {
		tmpTargetReq := &cslb.TCloudRemoveTargetReq{
			TargetGroupID: tmpTgID,
			TargetIDs:     []string{},
		}
		if _, ok := tgTargetMap[tmpTgID]; !ok {
			continue
		}
		tmpTargetReq.TargetIDs = tgTargetMap[tmpTgID]
		params.TargetGroups = append(params.TargetGroups, tmpTargetReq)
	}

	if len(params.TargetGroups) == 0 {
		return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("build add target param parse failed"))
	}

	removeTargetJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return removeTargetJSON, nil
}

// 按照clb分组targetGroup
func (svc *lbSvc) iterateTargetGroupGroupByCLB(kt *kit.Kit,
	tgIDsMap map[int][]string) (map[string][]string, error) {

	lbTgsMap := make(map[string][]string)
	for _, tgIDs := range tgIDsMap {
		for _, tgID := range tgIDs {
			// 根据目标组ID，获取目标组绑定的监听器、规则列表
			ruleRelReq := &core.ListReq{
				Filter: tools.EqualExpression("target_group_id", tgID),
				Page:   core.NewDefaultBasePage(),
			}
			for {
				listRuleRelResult, err := svc.client.DataService().Global.LoadBalancer.ListTargetGroupListenerRel(kt, ruleRelReq)
				if err != nil {
					logs.Errorf("list tcloud listener url rule failed, tgID: %s, err: %v, rid: %s", tgID, err, kt.Rid)
					return nil, err
				}

				// 还未绑定监听器及规则
				lbID := "-1"
				if len(listRuleRelResult.Details) != 0 {
					// 已经绑定了监听器及规则，归属某一clb
					lbID = listRuleRelResult.Details[0].LbID
				}
				if _, exists := lbTgsMap[lbID]; !exists {
					lbTgsMap[lbID] = make([]string, 0)
				}
				lbTgsMap[lbID] = append(lbTgsMap[lbID], tgID)

				if uint(len(listRuleRelResult.Details)) < core.DefaultMaxPageLimit {
					break
				}
				ruleRelReq.Page.Start += uint32(core.DefaultMaxPageLimit)
			}
		}

	}

	return lbTgsMap, nil
}
