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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
)

// ListTargetGroup list target group.
func (svc *lbSvc) ListTargetGroup(cts *rest.Contexts) (interface{}, error) {
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
	result, err := svc.dao.LoadBalancerTargetGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list target group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list target group failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TargetGroupListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseTargetGroup, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, err := convTableToBaseTargetGroup(cts.Kit, &one)
		if err != nil {
			continue
		}
		details = append(details, *tmpOne)
	}

	return &protocloud.TargetGroupListResult{Details: details}, nil
}

// GetTargetGroup ...
func (svc *lbSvc) GetTargetGroup(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "target group id is required")
	}

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.LoadBalancerTargetGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list target group failed, lblID: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, fmt.Errorf("get target group failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "target group is not found")
	}

	tgInfo := result.Details[0]
	switch tgInfo.Vendor {
	case enumor.TCloud:
		return convTableToBaseTargetGroup(cts.Kit, &tgInfo)
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", vendor)
	}
}

func convTableToBaseTargetGroup(kt *kit.Kit, one *tablelb.LoadBalancerTargetGroupTable) (
	*corelb.BaseTargetGroup, error) {

	var healthCheck *corelb.TCloudHealthCheckInfo
	// 支持不返回该字段
	if len(one.HealthCheck) != 0 {
		err := json.UnmarshalFromString(string(one.HealthCheck), &healthCheck)
		if err != nil {
			logs.Errorf("unmarshal healthCheck failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
			return nil, err
		}
	}

	return &corelb.BaseTargetGroup{
		ID:              one.ID,
		CloudID:         one.CloudID,
		Name:            one.Name,
		Vendor:          one.Vendor,
		AccountID:       one.AccountID,
		BkBizID:         one.BkBizID,
		TargetGroupType: one.TargetGroupType,
		VpcID:           one.VpcID,
		CloudVpcID:      one.CloudVpcID,
		Protocol:        one.Protocol,
		Region:          one.Region,
		Port:            one.Port,
		Weight:          cvt.PtrToVal(one.Weight),
		HealthCheck:     healthCheck,
		Memo:            one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}, nil
}

// ListTargetGroupListenerRel list target group listener rel.
func (svc *lbSvc) ListTargetGroupListenerRel(cts *rest.Contexts) (interface{}, error) {
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
	result, err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list target group listener rule rel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list target listener rule rel failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TargetListenerRuleRelListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseTargetListenerRuleRel, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseTargetListenerRuleRel(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.TargetListenerRuleRelListResult{Details: details}, nil
}

func convTableToBaseTargetListenerRuleRel(one *tablelb.TargetGroupListenerRuleRelTable) *corelb.
	BaseTargetListenerRuleRel {

	return &corelb.BaseTargetListenerRuleRel{
		ID:                  one.ID,
		Vendor:              one.Vendor,
		ListenerRuleID:      one.ListenerRuleID,
		ListenerRuleType:    one.ListenerRuleType,
		CloudListenerRuleID: one.CloudListenerRuleID,
		TargetGroupID:       one.TargetGroupID,
		CloudTargetGroupID:  one.CloudTargetGroupID,
		LbID:                one.LbID,
		CloudLbID:           one.CloudLbID,
		LblID:               one.LblID,
		CloudLblID:          one.CloudLblID,
		BindingStatus:       one.BindingStatus,
		Detail:              one.Detail,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// listTargetGroupIDsByRelCond 根据监听器查询条件，查询目标组ID列表
func (svc *lbSvc) listTargetGroupIDsByRelCond(kt *kit.Kit, req *protocloud.ListListenerWithTargetsReq,
	lblReq protocloud.ListenerQueryItem, cloudLblIDs []string) ([]string, error) {

	cloudTargetGroupIDs := make([]string, 0)
	for _, partCloudLblIDs := range slice.Split(cloudLblIDs, int(filter.DefaultMaxInLimit)) {
		ruleRelFilter := make([]*filter.AtomRule, 0)
		ruleRelFilter = append(ruleRelFilter, tools.RuleEqual("vendor", req.Vendor))
		ruleRelFilter = append(ruleRelFilter, tools.RuleIn("cloud_lb_id", lblReq.CloudLbIDs))
		ruleRelFilter = append(ruleRelFilter, tools.RuleIn("cloud_lbl_id", partCloudLblIDs))
		ruleRelFilter = append(ruleRelFilter, tools.RuleEqual("listener_rule_type", lblReq.RuleType))
		ruleRelFilter = append(ruleRelFilter, tools.RuleEqual("binding_status", enumor.SuccessBindingStatus))
		opt := &types.ListOption{
			Filter: tools.ExpressionAnd(ruleRelFilter...),
			Page:   core.NewDefaultBasePage(),
		}
		targetGroupRelList, err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().List(kt, opt)
		if err != nil {
			logs.Errorf("list target group listener rule rel failed, err: %v, req: %+v, lblReq: %+v, rid: %s",
				err, cvt.PtrToVal(req), lblReq, kt.Rid)
			return nil, fmt.Errorf("list target group listener rule rel failed, err: %v", err)
		}
		for _, item := range targetGroupRelList.Details {
			cloudTargetGroupIDs = append(cloudTargetGroupIDs, item.CloudTargetGroupID)
		}
	}
	return slice.Unique(cloudTargetGroupIDs), nil
}
