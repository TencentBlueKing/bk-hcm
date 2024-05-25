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
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// BatchBizModifyWeightTargetGroup batch biz modify weight target group.
func (svc *lbSvc) BatchBizModifyWeightTargetGroup(cts *rest.Contexts) (any, error) {
	return svc.batchModifyWeightTargetGroup(cts, handler.BizOperateAuth)
}

// BatchModifyWeightTargetGroup batch modify weight target group.
func (svc *lbSvc) BatchModifyWeightTargetGroup(cts *rest.Contexts) (any, error) {
	return svc.batchModifyWeightTargetGroup(cts, handler.ResOperateAuth)
}

func (svc *lbSvc) batchModifyWeightTargetGroup(cts *rest.Contexts,
	authHandler handler.ValidWithAuthHandler) (any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch sops modify weight target group request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch sops modify weight target group auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		return svc.buildModifyWeightTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildModifyWeightTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string) (any, error) {
	req := new(cslb.TCloudSopsTargetBatchModifyWeightReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询规则列表，查出符合条件的目标组
	tgIDs, err := svc.parseSOpsTargetParams(kt, accountID, req.RuleQueryList)
	if err != nil {
		return nil, err
	}
	if len(tgIDs) == 0 {
		return nil, errf.New(errf.RecordNotFound, "no matching target groups were found")
	}

	targetList, err := svc.getTargetByTGIDs(kt, tgIDs)
	if err != nil {
		return nil, err
	}

	targetGroupMap := make(map[string][]string)
	for _, item := range targetList {
		if _, ok := targetGroupMap[item.TargetGroupID]; !ok {
			targetGroupMap[item.TargetGroupID] = []string{item.ID}
			continue
		}
		targetGroupMap[item.TargetGroupID] = append(targetGroupMap[item.TargetGroupID], item.ID)
	}

	flowStateList := make([]*core.FlowStateResult, 0)
	for _, tmpTgID := range tgIDs {
		targetIDs, ok := targetGroupMap[tmpTgID]
		if !ok {
			logs.Errorf("build sops tcloud modify weight, target group not bind target, tgID: %s, rid: %s",
				tmpTgID, kt.Rid)
			continue
		}

		params := &cslb.TCloudBatchModifyTargetWeightReq{
			TargetIDs: targetIDs,
			NewWeight: cvt.ValToPtr(req.RsWeight),
		}
		targetJSON, err := json.Marshal(params)
		if err != nil {
			logs.Errorf("build sops tcloud modify weight target params marshal failed, err: %v, tgIDs: %v, "+
				"targetIDs: %v, params: %+v, rid: %s", err, tgIDs, targetIDs, params, kt.Rid)
			return nil, err
		}

		// 记录标准运维参数转换后的数据，方便排查问题
		logs.Infof("build sops tcloud modify weight target params jsonmarshal success, tgIDs: %v, targetIDs: %v, "+
			"targetJSON: %s, rid: %s", tgIDs, targetIDs, targetJSON, kt.Rid)

		flowState, err := svc.buildModifyTCloudTargetWeight(kt, targetJSON, tmpTgID, accountID)
		if err != nil {
			logs.Errorf("build sops tcloud modify weight target async call failed, err: %v, tgIDs: %v, targetIDs: %v, "+
				"targetJSON: %s, rid: %s", err, tgIDs, targetIDs, targetJSON, kt.Rid)
			return nil, err
		}
		flowStateList = append(flowStateList, flowState)
	}

	return flowStateList, nil
}
