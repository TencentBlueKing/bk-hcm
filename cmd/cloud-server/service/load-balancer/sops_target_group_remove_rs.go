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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
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
		return svc.buildDeleteTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildDeleteTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string) (any, error) {
	req := new(cslb.TCloudSopsTargetBatchRemoveReq)
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

	params := &cslb.TCloudTargetBatchRemoveReq{
		TargetGroups: []*cslb.TCloudRemoveTargetReq{},
	}
	for _, tmpTgID := range tgIDs {
		tmpTargetReq := &cslb.TCloudRemoveTargetReq{
			TargetGroupID: tmpTgID,
			TargetIDs:     []string{},
		}
		if _, ok := targetGroupMap[tmpTgID]; !ok {
			continue
		}
		tmpTargetReq.TargetIDs = targetGroupMap[tmpTgID]
		params.TargetGroups = append(params.TargetGroups, tmpTargetReq)
	}

	if len(params.TargetGroups) == 0 {
		logs.Errorf("build sops tcloud remove target params parse failed, err: %v, accountID: %s, tgIDs: %v, rid: %s",
			err, accountID, tgIDs, kt.Rid)
		return nil, errf.NewFromErr(errf.RecordNotFound, fmt.Errorf("build add target param parse failed"))
	}

	removeTargetJSON, err := json.Marshal(params)
	if err != nil {
		logs.Errorf("build sops tcloud remove target params marshal failed, err: %v, params: %+v, rid: %s",
			err, params, kt.Rid)
		return nil, err
	}

	// 记录标准运维参数转换后的数据，方便排查问题
	logs.Infof("build sops tcloud remove target params jsonmarshal success, tgIDs: %v, removeTargetJSON: %s, rid: %s",
		tgIDs, removeTargetJSON, kt.Rid)

	return svc.buildRemoveTCloudTarget(kt, removeTargetJSON, accountID)
}
