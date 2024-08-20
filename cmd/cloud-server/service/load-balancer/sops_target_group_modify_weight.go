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
	"hcm/pkg/tools/slice"
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
		return svc.buildModifyWeightTCloudTarget(cts.Kit, req.Data, accountInfo.AccountID, enumor.TCloud)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildModifyWeightTCloudTarget(kt *kit.Kit, body json.RawMessage, accountID string, vendor enumor.Vendor) (any, error) {
	req := new(cslb.TCloudSopsTargetBatchModifyWeightReq)
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

	// 查询每一行筛选出的目标组对应的目标，按照当前行填写的条件进行筛选
	tgTargetsMap := make(map[string][]string)
	for index, tgIDs := range tgIDsMap {
		targetList, err := svc.getTargetByTGIDs(kt, tgIDs)
		if err != nil {
			logs.Errorf("get target by target group ids failed, err: %v, tgIDs: %v, rid: %s", err, tgIDs, kt.Rid)
			return nil, err
		}

		rsIPs := req.RuleQueryList[index].RsIP
		rsType := req.RuleQueryList[index].RsType
		for _, target := range targetList {
			// 筛选rsType
			if string(target.InstType) != rsType {
				continue
			}
			// 筛选rsIp
			for _, rsIp := range target.PrivateIPAddress {
				if slice.IsItemInSlice(rsIPs, rsIp) {
					if _, ok := tgTargetsMap[target.TargetGroupID]; !ok {
						tgTargetsMap[target.TargetGroupID] = make([]string, 0)
					}
					tgTargetsMap[target.TargetGroupID] = append(tgTargetsMap[target.TargetGroupID], target.ID)
					continue
				}
			}
		}
	}

	flowStateResults, err := svc.buildBatchModifyTCloudTargetWeight(kt, tgTargetsMap, &req.RsWeight, accountID)
	if err != nil {
		logs.Errorf("build batch modify tcloud target weight failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return flowStateResults, nil
}
