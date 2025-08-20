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
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// ListTarget list target.
func (svc *lbSvc) ListTarget(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancerTarget().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list lb target failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list lb target failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TargetListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseTarget, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseTarget(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.TargetListResult{Details: details}, nil
}

func convTableToBaseTarget(one *tablelb.LoadBalancerTargetTable) *corelb.BaseTarget {
	return &corelb.BaseTarget{
		ID:                 one.ID,
		IP:                 one.IP,
		AccountID:          one.AccountID,
		InstType:           one.InstType,
		InstID:             one.InstID,
		CloudInstID:        one.CloudInstID,
		InstName:           one.InstName,
		TargetGroupRegion:  one.TargetGroupRegion,
		TargetGroupID:      one.TargetGroupID,
		CloudTargetGroupID: one.CloudTargetGroupID,
		Port:               one.Port,
		Weight:             one.Weight,
		PrivateIPAddress:   one.PrivateIPAddress,
		PublicIPAddress:    one.PublicIPAddress,
		CloudVpcIDs:        one.CloudVpcIDs,
		Zone:               one.Zone,
		Memo:               one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// listTargetByCond 根据账号ID、RsIP查询绑定的目标组列表
func (svc *lbSvc) listTargetByCond(kt *kit.Kit, req *protocloud.ListListenerWithTargetsReq,
	lblReq protocloud.ListenerQueryItem, cloudTargetGroupIDs []string) ([]corelb.BaseTarget, error) {

	targetList := make([]corelb.BaseTarget, 0)
	for _, partCloudTargetGroupIDs := range slice.Split(cloudTargetGroupIDs, int(filter.DefaultMaxInLimit)) {
		targetFilter := make([]*filter.AtomRule, 0)
		targetFilter = append(targetFilter, tools.RuleEqual("account_id", req.AccountID))
		targetFilter = append(targetFilter, tools.RuleEqual("inst_type", lblReq.InstType))
		targetFilter = append(targetFilter, tools.RuleIn("cloud_target_group_id", partCloudTargetGroupIDs))
		if len(lblReq.RsIPs) > 0 {
			targetFilter = append(targetFilter, tools.RuleIn("ip", lblReq.RsIPs))
		}
		if len(lblReq.RsPorts) > 0 {
			targetFilter = append(targetFilter, tools.RuleIn("port", lblReq.RsPorts))
		}
		if len(lblReq.RsWeights) > 0 {
			targetFilter = append(targetFilter, tools.RuleIn("weight", lblReq.RsWeights))
		}
		opt := &types.ListOption{
			Filter: tools.ExpressionAnd(targetFilter...),
			Page:   core.NewDefaultBasePage(),
		}
		loopTargetList, err := svc.dao.LoadBalancerTarget().List(kt, opt)
		if err != nil {
			logs.Errorf("list load balancer target failed, err: %v, req: %+v, rid: %s",
				err, cvt.PtrToVal(req), kt.Rid)
			return nil, fmt.Errorf("list load balancer target failed, err: %v", err)
		}

		for _, item := range loopTargetList.Details {
			targetList = append(targetList, corelb.BaseTarget{
				ID:                 item.ID,
				AccountID:          item.AccountID,
				IP:                 item.IP,
				Port:               item.Port,
				Weight:             item.Weight,
				InstType:           item.InstType,
				InstID:             item.InstID,
				CloudInstID:        item.CloudInstID,
				InstName:           item.InstName,
				TargetGroupRegion:  item.TargetGroupRegion,
				TargetGroupID:      item.TargetGroupID,
				CloudTargetGroupID: item.CloudTargetGroupID,
				PrivateIPAddress:   item.PrivateIPAddress,
				PublicIPAddress:    item.PublicIPAddress,
				CloudVpcIDs:        item.CloudVpcIDs,
				Zone:               item.Zone,
				Memo:               item.Memo,
				Revision: &core.Revision{
					Creator:   item.Creator,
					Reviser:   item.Reviser,
					CreatedAt: item.CreatedAt.String(),
					UpdatedAt: item.UpdatedAt.String(),
				},
			})
		}
	}
	return targetList, nil
}
