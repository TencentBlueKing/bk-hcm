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
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

type tcloudLBUpdate = dataproto.LoadBalancerExtUpdateReq[corelb.TCloudClbExtension]

// fillTCloudLBBiz 同步建好本地记录后按请求补齐业务：只把"未分配"的记录改为请求中的业务，
// 业务已是其它值的记录不覆盖。失败直接返回错误，不重试。
func (svc *clbSvc) fillTCloudLBBiz(kt *kit.Kit, req *protolb.TCloudLoadBalancerCreateReq, cloudIDs []string) error {
	if len(cloudIDs) == 0 || req.BkBizID <= 0 {
		return nil
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloud),
			tools.RuleEqual("account_id", req.AccountID),
			tools.RuleEqual("region", req.Region),
			tools.RuleIn("cloud_id", cloudIDs),
		),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "cloud_id", "bk_biz_id"},
	}
	lbResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(kt, listReq)
	if err != nil {
		logs.Errorf("list lb for biz fill failed, cloud_ids: %v, err: %v, rid: %s", cloudIDs, err, kt.Rid)
		return err
	}

	updates, otherBizLBs := buildTCloudLBBizUpdates(lbResp.Details, req.BkBizID)
	if len(otherBizLBs) > 0 {
		logs.Warnf("lb already assigned to other biz, skip biz fill, req_biz: %d, cloud_id_to_biz: %v, rid: %s",
			req.BkBizID, otherBizLBs, kt.Rid)
	}
	if len(updates) == 0 {
		return nil
	}

	updateReq := &dataproto.TCloudClbBatchUpdateReq{Lbs: updates}
	if err = svc.dataCli.TCloud.LoadBalancer.BatchUpdate(kt, updateReq); err != nil {
		logs.Errorf("update lb biz failed, biz: %d, cloud_ids: %v, err: %v, rid: %s", req.BkBizID, cloudIDs, err,
			kt.Rid)
		return err
	}
	return nil
}

// buildTCloudLBBizUpdates 为未分配业务的记录生成补业务的更新项，并返回业务已是其它值的记录（云 ID -> 业务）。
func buildTCloudLBBizUpdates(lbs []corelb.BaseLoadBalancer, bizID int64) (
	updates []*tcloudLBUpdate, otherBizLBs map[string]int64) {

	updates = make([]*tcloudLBUpdate, 0, len(lbs))
	otherBizLBs = make(map[string]int64)
	for _, lb := range lbs {
		switch lb.BkBizID {
		case constant.UnassignedBiz:
			updates = append(updates, &tcloudLBUpdate{ID: lb.ID, BkBizID: bizID})
		case bizID:
		default:
			otherBizLBs[lb.CloudID] = lb.BkBizID
		}
	}
	return updates, otherBizLBs
}
