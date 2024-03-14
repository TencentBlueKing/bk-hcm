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
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"
)

// ListLoadBalancer list clb.
func (svc *lbSvc) ListLoadBalancer(cts *rest.Contexts) (interface{}, error) {
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
	result, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list clb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list clb failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ClbListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseLoadBalancer, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseLB(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.ClbListResult{Details: details}, nil
}

func convTableToBaseLB(one *tablelb.LoadBalancerTable) *corelb.BaseLoadBalancer {
	return &corelb.BaseLoadBalancer{
		ID:                   one.ID,
		CloudID:              one.CloudID,
		Name:                 one.Name,
		Vendor:               one.Vendor,
		AccountID:            one.AccountID,
		BkBizID:              one.BkBizID,
		Region:               one.Region,
		Zones:                one.Zones,
		BackupZones:          one.BackupZones,
		VpcID:                one.VpcID,
		CloudVpcID:           one.CloudVpcID,
		SubnetID:             one.SubnetID,
		CloudSubnetID:        one.CloudSubnetID,
		PrivateIPv4Addresses: one.PrivateIPv4Addresses,
		PrivateIPv6Addresses: one.PrivateIPv6Addresses,
		PublicIPv4Addresses:  one.PublicIPv4Addresses,
		PublicIPv6Addresses:  one.PublicIPv6Addresses,
		Domain:               one.Domain,
		Status:               one.Status,
		CloudCreatedTime:     one.CloudCreatedTime,
		CloudStatusTime:      one.CloudStatusTime,
		CloudExpiredTime:     one.CloudExpiredTime,
		Memo:                 one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// ListLoadBalancerExt list clb ext.
func (svc *lbSvc) ListLoadBalancerExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.EipListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return convClbListResult[corelb.TCloudClbExtension](data.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// GetLoadBalancer ...
func (svc *lbSvc) GetLoadBalancer(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "clb id is required")
	}

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list clb(%s) failed, err: %v, rid: %s", err, id, cts.Kit.Rid)
		return nil, fmt.Errorf("get clb failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "load balancer not found")
	}

	clbTable := result.Details[0]
	switch clbTable.Vendor {
	case enumor.TCloud:
		return convLoadBalancerWithExt(&clbTable)
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", vendor)
	}
}
func convLoadBalancerWithExt[T corelb.Extension](tableLB *tablelb.LoadBalancerTable) (*corelb.LoadBalancer[T], error) {
	base := convTableToBaseLB(tableLB)
	extension := new(T)
	if tableLB.Extension != "" {
		if err := json.UnmarshalFromString(string(tableLB.Extension), extension); err != nil {
			return nil, fmt.Errorf("fail unmarshal load balancer extension, err: %v", err)
		}
	}
	return &corelb.LoadBalancer[T]{
		BaseLoadBalancer: *base,
		Extension:        extension,
	}, nil
}

func convClbListResult[T corelb.Extension](tables []tablelb.LoadBalancerTable) (
	*protocloud.ClbExtListResult[T], error) {

	details := make([]corelb.LoadBalancer[T], 0, len(tables))
	for _, tableLB := range tables {
		base := convTableToBaseLB(&tableLB)
		extension := new(T)
		if tableLB.Extension != "" {
			if err := json.UnmarshalFromString(string(tableLB.Extension), extension); err != nil {
				return nil, fmt.Errorf("fail unmarshal load balancer extension, err: %v", err)
			}
		}
		details = append(details, corelb.LoadBalancer[T]{
			BaseLoadBalancer: *base,
			Extension:        extension,
		})
	}

	return &protocloud.ClbExtListResult[T]{
		Details: details,
	}, nil
}

// ListListener list listener.
func (svc *lbSvc) ListListener(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ListListenerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	reqFilter := &filter.Expression{
		Op: filter.And,
	}
	if len(req.LbID) > 0 {
		reqFilter.Rules = append(reqFilter.Rules,
			filter.AtomRule{Field: "lb_id", Op: filter.Equal.Factory(), Value: req.LbID})
	}
	// 加上请求里过滤条件
	if req.Filter != nil && !req.Filter.IsEmpty() {
		reqFilter.Rules = append(reqFilter.Rules, req.Filter)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: reqFilter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancerListener().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list listener failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list listener failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ListenerListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseListener, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseListener(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.ListenerListResult{Details: details}, nil
}

func convTableToBaseListener(one *tablelb.LoadBalancerListenerTable) *corelb.BaseListener {
	return &corelb.BaseListener{
		ID:            one.ID,
		CloudID:       one.CloudID,
		Name:          one.Name,
		Vendor:        one.Vendor,
		AccountID:     one.AccountID,
		BkBizID:       one.BkBizID,
		LbID:          one.LBID,
		CloudLbID:     one.CloudLBID,
		Protocol:      one.Protocol,
		Port:          one.Port,
		DefaultDomain: one.DefaultDomain,
		Zones:         one.Zones,
		Memo:          one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// ListUrlRule list url rule.
func (svc *lbSvc) ListUrlRule(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ListTCloudURLRuleReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	reqFilter := &filter.Expression{
		Op: filter.And,
	}
	if len(req.TargetGroupID) > 0 {
		reqFilter.Rules = append(reqFilter.Rules,
			filter.AtomRule{Field: "target_group_id", Op: filter.Equal.Factory(), Value: req.TargetGroupID})
	}
	// 加上请求里过滤条件
	if req.Filter != nil && !req.Filter.IsEmpty() {
		reqFilter.Rules = append(reqFilter.Rules, req.Filter)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: reqFilter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancerTCloudUrlRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud clb url rule failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud clb url rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TCloudURLRuleListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseTCloudClbURLRule, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, err := convTableToBaseTCloudClbURLRule(cts.Kit, &one)
		if err != nil {
			continue
		}
		details = append(details, *tmpOne)
	}

	return &protocloud.TCloudURLRuleListResult{Details: details}, nil
}

func convTableToBaseTCloudClbURLRule(kt *kit.Kit, one *tablelb.TCloudClbUrlRuleTable) (
	*corelb.BaseTCloudClbURLRule, error) {

	var healthCheck *corelb.HealthCheckInfo
	err := json.UnmarshalFromString(string(one.HealthCheck), &healthCheck)
	if err != nil {
		logs.Errorf("unmarshal healthCheck failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
		return nil, err
	}

	var certInfo *corelb.CertificateInfo
	err = json.UnmarshalFromString(string(one.Certificate), &certInfo)
	if err != nil {
		logs.Errorf("unmarshal certificate failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
		return nil, err
	}

	return &corelb.BaseTCloudClbURLRule{
		ID:                 one.ID,
		CloudID:            one.CloudID,
		Name:               one.Name,
		RuleType:           one.RuleType,
		LbID:               one.LbID,
		CloudLbID:          one.CloudLbID,
		LblID:              one.LblID,
		CloudLBLID:         one.CloudLBLID,
		TargetGroupID:      one.TargetGroupID,
		CloudTargetGroupID: one.CloudTargetGroupID,
		Domain:             one.Domain,
		URL:                one.URL,
		Scheduler:          one.Scheduler,
		SniSwitch:          one.SniSwitch,
		SessionType:        one.SessionType,
		SessionExpire:      one.SessionExpire,
		HealthCheck:        healthCheck,
		Certificate:        certInfo,
		Memo:               one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}, nil
}

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
		logs.Errorf("list clb target failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list clb target failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ClbTargetListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseClbTarget, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseClbTarget(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.ClbTargetListResult{Details: details}, nil
}

func convTableToBaseClbTarget(one *tablelb.LoadBalancerTargetTable) *corelb.BaseClbTarget {
	return &corelb.BaseClbTarget{
		ID:                 one.ID,
		AccountID:          one.AccountID,
		InstType:           one.InstType,
		CloudInstID:        one.CloudInstID,
		InstName:           one.InstName,
		TargetGroupID:      one.TargetGroupID,
		CloudTargetGroupID: one.CloudTargetGroupID,
		Port:               one.Port,
		Weight:             one.Weight,
		PrivateIPAddress:   one.PrivateIPAddress,
		PublicIPAddress:    one.PublicIPAddress,
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
		logs.Errorf("list clb target group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list clb target group failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ClbTargetGroupListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseClbTargetGroup, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, err := convTableToBaseClbTargetGroup(cts.Kit, &one)
		if err != nil {
			continue
		}
		details = append(details, *tmpOne)
	}

	return &protocloud.ClbTargetGroupListResult{Details: details}, nil
}

func convTableToBaseClbTargetGroup(kt *kit.Kit, one *tablelb.LoadBalancerTargetGroupTable) (
	*corelb.BaseClbTargetGroup, error) {

	var healthCheck *corelb.HealthCheckInfo
	err := json.UnmarshalFromString(string(one.HealthCheck), &healthCheck)
	if err != nil {
		logs.Errorf("unmarshal healthCheck failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
		return nil, err
	}

	return &corelb.BaseClbTargetGroup{
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
		Weight:          one.Weight,
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

func (svc *lbSvc) GetListener(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "listener id is required")
	}

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.LoadBalancerListener().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list listener failed, lblID: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, fmt.Errorf("get listener failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "listener is not found")
	}

	lblInfo := result.Details[0]
	switch lblInfo.Vendor {
	case enumor.TCloud:
		return convTableToBaseListener(&lblInfo), nil
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", vendor)
	}
}
