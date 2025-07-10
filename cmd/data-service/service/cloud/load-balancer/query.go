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
	rawjson "encoding/json"
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types/load-balancer"
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
	"hcm/pkg/tools/cidr"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
)

// ListLoadBalancer list load balancer.
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
		logs.Errorf("list lb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list lb failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.LbListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseLoadBalancer, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseLB(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.LbListResult{Details: details}, nil
}

// ListLoadBalancerRaw ...
func (svc *lbSvc) ListLoadBalancerRaw(cts *rest.Contexts) (any, error) {
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
		logs.Errorf("list lb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list lb failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.LbRawListResult{Count: result.Count}, nil
	}

	details := make([]corelb.LoadBalancerRaw, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseLB(&one)
		details = append(details, corelb.LoadBalancerRaw{
			BaseLoadBalancer: *tmpOne,
			Extension:        rawjson.RawMessage(one.Extension),
		})
	}

	return &protocloud.LbRawListResult{Details: details}, nil
}

func convTableToBaseLB(one *tablelb.LoadBalancerTable) *corelb.BaseLoadBalancer {

	return &corelb.BaseLoadBalancer{
		ID:                   one.ID,
		CloudID:              one.CloudID,
		Name:                 one.Name,
		Vendor:               one.Vendor,
		AccountID:            one.AccountID,
		LoadBalancerType:     one.LBType,
		IPVersion:            enumor.IPAddressType(one.IPVersion),
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
		SyncTime:             one.SyncTime,
		Tags:                 core.TagMap(one.Tags),
		Memo:                 one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
		BandWidth: one.BandWidth,
		Isp:       one.Isp,
	}
}

// ListLoadBalancerExt list load balancer ext.
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
		return convLbListResult[corelb.TCloudClbExtension](data.Details)
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
		return nil, errf.New(errf.InvalidParameter, "lb id is required")
	}

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list lb(%s) failed, err: %v, rid: %s", err, id, cts.Kit.Rid)
		return nil, fmt.Errorf("list lb failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "load balancer not found")
	}

	lbTable := result.Details[0]
	switch lbTable.Vendor {
	case enumor.TCloud:
		return convLoadBalancerWithExt(&lbTable)
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

func convLbListResult[T corelb.Extension](tables []tablelb.LoadBalancerTable) (
	*protocloud.LbExtListResult[T], error) {

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

	return &protocloud.LbExtListResult[T]{
		Details: details,
	}, nil
}

// ListListener list listener.
func (svc *lbSvc) ListListener(cts *rest.Contexts) (interface{}, error) {
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

// ListListenerExt list listener with extension.
func (svc *lbSvc) ListListenerExt(cts *rest.Contexts) (any, error) {
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
	result, err := svc.dao.LoadBalancerListener().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list listener failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list listener failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ListenerListResult{Count: result.Count}, nil
	}

	details := make([]corelb.Listener[corelb.TCloudListenerExtension], 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, err := convTableToListener[corelb.TCloudListenerExtension](&one)
		if err != nil {
			logs.Errorf("fail to conv listener with extension, err: %v, rid: %s", err, cts.Kit.Rid)
		}
		details = append(details, *tmpOne)
	}

	return &protocloud.TCloudListenerListResult{Details: details}, nil
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
		Region:        one.Region,
		Zones:         one.Zones,
		Memo:          one.Memo,
		SniSwitch:     one.SniSwitch,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

func convTableToListener[T corelb.ListenerExtension](table *tablelb.LoadBalancerListenerTable) (
	*corelb.Listener[T], error) {
	base := convTableToBaseListener(table)
	extension := new(T)
	if table.Extension != "" {
		if err := json.UnmarshalFromString(string(table.Extension), extension); err != nil {
			return nil, fmt.Errorf("fail unmarshal listener extension, err: %v", err)
		}
	}
	return &corelb.Listener[T]{
		BaseListener: base,
		Extension:    extension,
	}, nil
}

// ListTCloudUrlRule list tcloud url rule.
func (svc *lbSvc) ListTCloudUrlRule(cts *rest.Contexts) (any, error) {
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
	result, err := svc.dao.LoadBalancerTCloudUrlRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud lb url rule failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud lb url rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TCloudURLRuleListResult{Count: result.Count}, nil
	}

	details := make([]corelb.TCloudLbUrlRule, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, err := convTableToBaseTCloudLbURLRule(cts.Kit, &one)
		if err != nil {
			continue
		}
		details = append(details, *tmpOne)
	}

	return &protocloud.TCloudURLRuleListResult{Details: details}, nil
}

func convTableToBaseTCloudLbURLRule(kt *kit.Kit, one *tablelb.TCloudLbUrlRuleTable) (
	*corelb.TCloudLbUrlRule, error) {

	var healthCheck *corelb.TCloudHealthCheckInfo
	err := json.UnmarshalFromString(string(one.HealthCheck), &healthCheck)
	if err != nil {
		logs.Errorf("unmarshal healthCheck failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
		return nil, err
	}

	var certInfo *corelb.TCloudCertificateInfo
	err = json.UnmarshalFromString(string(one.Certificate), &certInfo)
	if err != nil {
		logs.Errorf("unmarshal certificate failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
		return nil, err
	}

	return &corelb.TCloudLbUrlRule{
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
		Region:             one.Region,
		Domain:             one.Domain,
		URL:                one.URL,
		Scheduler:          one.Scheduler,
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

// GetListener ...
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
		newLblInfo, err := convTableToListener[corelb.TCloudListenerExtension](&lblInfo)
		if err != nil {
			logs.Errorf("fail to conv listener with extension, lblID: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
			return nil, err
		}
		return newLblInfo, nil
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", vendor)
	}
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

// ListResFlowLock list res flow lock.
func (svc *lbSvc) ListResFlowLock(cts *rest.Contexts) (interface{}, error) {
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
	result, err := svc.dao.ResourceFlowLock().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list res flow lock failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list res flow lock failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ResFlowLockListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseResFlowLock, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corelb.BaseResFlowLock{
			ResID:   one.ResID,
			ResType: one.ResType,
			Owner:   one.Owner,
			Revision: &core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.ResFlowLockListResult{Details: details}, nil
}

// ListResFlowRel list res flow rel.
func (svc *lbSvc) ListResFlowRel(cts *rest.Contexts) (interface{}, error) {
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
	result, err := svc.dao.ResourceFlowRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list res flow rel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list res flow rel failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ResFlowRelListResult{Count: result.Count}, nil
	}

	details := make([]corelb.BaseResFlowRel, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corelb.BaseResFlowRel{
			ID:       one.ID,
			ResID:    one.ResID,
			FlowID:   one.FlowID,
			TaskType: one.TaskType,
			Status:   one.Status,
			Revision: &core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.ResFlowRelListResult{Details: details}, nil
}

// CountListenerByLbIDs count listener by lbIDs.
func (svc *lbSvc) CountListenerByLbIDs(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ListListenerCountByLbIDsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.dao.LoadBalancerListener().CountListenerByLbIDs(cts.Kit, req.LbIDs)
}

// ListListenerWithTargets list listener with target.
func (svc *lbSvc) ListListenerWithTargets(cts *rest.Contexts) (any, error) {
	req := new(protocloud.ListListenerWithTargetsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listenerList := &protocloud.ListListenerWithTargetsResp{}
	for _, item := range req.ListenerQueryList {
		lblRsIPList, err := svc.queryListenerWithTargets(cts.Kit, req, item)
		if err != nil {
			return nil, err
		}
		listenerList.Details = append(listenerList.Details, lblRsIPList...)
	}
	return listenerList, nil
}

func (svc *lbSvc) queryListenerWithTargets(kt *kit.Kit, req *protocloud.ListListenerWithTargetsReq,
	lblReq protocloud.ListenerQueryItem) ([]*protocloud.ListBatchListenerResult, error) {

	// 查询符合条件的负载均衡列表
	cloudClbIDs, lbMap, err := svc.listLoadBalancerListCheckVip(kt, req, lblReq)
	if err != nil {
		return nil, err
	}
	// 未查询到符合条件的负载均衡列表
	if len(cloudClbIDs) == 0 {
		logs.Errorf("check list load balancer with targets empty, req: %+v, rid: %s", cvt.PtrToVal(req), kt.Rid)
		return nil, nil
	}

	// 查询符合条件的监听器列表
	lblMap, cloudLblIDs, _, err := svc.listBizListenerByLbIDs(kt, req, lblReq, cloudClbIDs)
	if err != nil {
		return nil, err
	}
	// 未查询到符合的监听器列表
	if len(cloudLblIDs) == 0 {
		logs.Errorf("list biz listener with targets empty, req: %+v, rid: %s", cvt.PtrToVal(req), kt.Rid)
		return nil, nil
	}

	// 获取监听器绑定的目标组ID列表
	cloudTargetGroupIDs, err := svc.listTargetGroupIDsByRelCond(kt, req, lblReq, cloudLblIDs)
	if err != nil {
		return nil, err
	}

	// 根据RSIP获取绑定的目标组ID列表
	targetGroupRsList, targetGroupIDs, err := svc.listListenerWithTarget(kt, req, lblReq, cloudTargetGroupIDs)
	if err != nil {
		return nil, err
	}
	// 未查询到符合的监听器列表
	if len(targetGroupIDs) == 0 {
		logs.Errorf("list load balancer target with targets empty, req: %+v, rid: %s", cvt.PtrToVal(req), kt.Rid)
		return nil, nil
	}

	// 根据负载均衡ID、监听器ID、目标组ID，获取监听器与目标组的绑定关系列表
	lblUrlRuleList := make([]protocloud.LoadBalancerUrlRuleResult, 0)
	switch req.Vendor {
	case enumor.TCloud:
		lblUrlRuleList, err = svc.listTCloudLBUrlRuleByTgIDs(kt, lblReq, cloudClbIDs,
			cloudLblIDs, targetGroupIDs)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "batch query listener with targets failed, invalid vendor: %s",
			req.Vendor)
	}
	if err != nil {
		return nil, err
	}
	// 未查询到符合的监听器与目标组绑定关系的列表
	if len(lblUrlRuleList) == 0 {
		logs.Errorf("[%s]list load balancer url rule empty, req: %+v, cloudClbIDs: %v, cloudLblIDs: %v, "+
			"targetGroupIDs: %v, rid: %s", req.Vendor, cvt.PtrToVal(req), cloudClbIDs,
			cloudLblIDs, targetGroupIDs, kt.Rid)
		return nil, nil
	}

	return svc.convertListListenerWithTargets(lbMap, lblUrlRuleList, lblMap, targetGroupRsList)
}

func (svc *lbSvc) convertListListenerWithTargets(lbMap map[string]tablelb.LoadBalancerTable,
	lblUrlRuleList []protocloud.LoadBalancerUrlRuleResult, lblMap map[string]tablelb.LoadBalancerListenerTable,
	targetGroupRsList map[string][]protocloud.LoadBalancerTargetRsList) (
	[]*protocloud.ListBatchListenerResult, error) {

	lblResult := make([]*protocloud.ListBatchListenerResult, 0)
	lblRsMap := make(map[string]*protocloud.ListBatchListenerResult)
	lblExist := make(map[string]struct{})
	for _, item := range lblUrlRuleList {
		// 遍历UrlRule列表，如果有多个监听器需要根据目标组ID，汇总RS列表
		if _, ok := lblExist[item.CloudLblID]; ok {
			lblRsMap = svc.getRsListByTargetGroupIDs(item, targetGroupRsList, lblRsMap)
			continue
		}
		lblExist[item.CloudLblID] = struct{}{}
		// 检查监听器是否存在
		lblInfo, ok := lblMap[item.CloudLblID]
		if !ok {
			continue
		}
		// 检查负载均衡是否存在
		lbInfo, ok := lbMap[item.CloudClbID]
		if !ok {
			continue
		}
		// 获取VIP/域名
		vipDomain, err := svc.getClbVipDomain(lbInfo)
		if err != nil {
			return nil, err
		}
		lblRsMap[item.CloudLblID] = &protocloud.ListBatchListenerResult{
			ClbID:        lbInfo.ID,
			CloudClbID:   item.CloudClbID,
			ClbVipDomain: strings.Join(vipDomain, ","),
			BkBizID:      lblInfo.BkBizID,
			Region:       lbInfo.Region,
			Vendor:       lbInfo.Vendor,
			LblID:        lblInfo.ID,
			CloudLblID:   item.CloudLblID,
			Protocol:     lblInfo.Protocol,
			Port:         lblInfo.Port,
			RsList:       make([]*protocloud.LoadBalancerTargetRsList, 0),
		}
		lblRsMap = svc.getRsListByTargetGroupIDs(item, targetGroupRsList, lblRsMap)
	}

	for _, item := range lblRsMap {
		lblResult = append(lblResult, &protocloud.ListBatchListenerResult{
			ClbID:        item.ClbID,
			CloudClbID:   item.CloudClbID,
			ClbVipDomain: item.ClbVipDomain,
			BkBizID:      item.BkBizID,
			Region:       item.Region,
			Vendor:       item.Vendor,
			LblID:        item.LblID,
			CloudLblID:   item.CloudLblID,
			Protocol:     item.Protocol,
			Port:         item.Port,
			RsList:       item.RsList,
		})
	}

	return lblResult, nil
}

func (svc *lbSvc) getRsListByTargetGroupIDs(item protocloud.LoadBalancerUrlRuleResult,
	targetGroupRsList map[string][]protocloud.LoadBalancerTargetRsList,
	lblRsMap map[string]*protocloud.ListBatchListenerResult) map[string]*protocloud.ListBatchListenerResult {

	if len(item.TargetGroupIDs) == 0 {
		return nil
	}

	for _, targetGroupID := range item.TargetGroupIDs {
		for _, targetGroupItem := range targetGroupRsList[targetGroupID] {
			lblRsMap[item.CloudLblID].RsList = append(lblRsMap[item.CloudLblID].RsList,
				&protocloud.LoadBalancerTargetRsList{
					BaseTarget:  targetGroupItem.BaseTarget,
					RuleID:      item.TargetGrouRuleMap[targetGroupID].RuleID,
					CloudRuleID: item.TargetGrouRuleMap[targetGroupID].CloudRuleID,
					RuleType:    item.TargetGrouRuleMap[targetGroupID].RuleType,
					Domain:      item.TargetGrouRuleMap[targetGroupID].Domain,
					Url:         item.TargetGrouRuleMap[targetGroupID].Url,
				})
		}
	}
	return lblRsMap
}

func (svc *lbSvc) getClbVipDomain(lbInfo tablelb.LoadBalancerTable) ([]string, error) {
	vipDomains := make([]string, 0)
	switch loadbalancer.TCloudLoadBalancerType(lbInfo.LBType) {
	case loadbalancer.InternalLoadBalancerType:
		if lbInfo.IPVersion == string(enumor.Ipv4) {
			vipDomains = append(vipDomains, lbInfo.PrivateIPv4Addresses...)
		} else {
			vipDomains = append(vipDomains, lbInfo.PrivateIPv6Addresses...)
		}
	case loadbalancer.OpenLoadBalancerType:
		if lbInfo.IPVersion == string(enumor.Ipv4) {
			vipDomains = append(vipDomains, lbInfo.PublicIPv4Addresses...)
		} else {
			vipDomains = append(vipDomains, lbInfo.PublicIPv6Addresses...)
		}
	default:
		return nil, fmt.Errorf("unsupported lb_type: %s(%s)", lbInfo.LBType, lbInfo.CloudID)
	}

	// 如果IP为空则获取负载均衡域名
	if len(vipDomains) == 0 && len(lbInfo.Domain) > 0 {
		vipDomains = append(vipDomains, lbInfo.Domain)
	}

	return vipDomains, nil
}

// listLoadBalancerListCheckVip 获取负载均衡列表并检查VIP、域名是否匹配
func (svc *lbSvc) listLoadBalancerListCheckVip(kt *kit.Kit, req *protocloud.ListListenerWithTargetsReq,
	lblReq protocloud.ListenerQueryItem) ([]string, map[string]tablelb.LoadBalancerTable, error) {

	lbOpt := &types.ListOption{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", req.Vendor),
			tools.RuleEqual("bk_biz_id", req.BkBizID),
			tools.RuleEqual("account_id", req.AccountID),
			tools.RuleEqual("region", lblReq.Region),
			tools.RuleIn("cloud_id", lblReq.CloudLbIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	lbAllList := make([]tablelb.LoadBalancerTable, 0)
	for {
		lbList, err := svc.dao.LoadBalancer().List(kt, lbOpt)
		if err != nil {
			logs.Errorf("check list load balancer failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, nil, fmt.Errorf("list load balancer failed, err: %v", err)
		}

		lbAllList = append(lbAllList, lbList.Details...)
		if len(lbList.Details) <= int(core.DefaultMaxPageLimit) {
			break
		}
		lbOpt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	// 检查ip地址/域名，是否在负载均衡的ip地址列表中
	cloudClbIDs, lbMap, err := checkClbVipAndDomain(lbAllList, lblReq.CloudLbIDs, lblReq.ClbVipDomains)
	if err != nil {
		logs.Errorf("check list load balancer and ip domain match failed, err: %v, req: %+v, rid: %s",
			err, cvt.PtrToVal(req), kt.Rid)
		return nil, nil, err
	}

	return cloudClbIDs, lbMap, nil
}

func checkClbVipAndDomain(list []tablelb.LoadBalancerTable, paramClbIDs, clbVipDomains []string) (
	[]string, map[string]tablelb.LoadBalancerTable, error) {

	cloudClbIDs := make([]string, 0)
	lbMap := cvt.SliceToMap(list, func(item tablelb.LoadBalancerTable) (string, tablelb.LoadBalancerTable) {
		return item.CloudID, item
	})

	for idx, cloudID := range paramClbIDs {
		lbInfo, ok := lbMap[cloudID]
		if !ok {
			return nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] is not found", cloudID)
		}

		// 检查对应的负载均衡VIP/域名是否匹配
		vipDomain := clbVipDomains[idx]
		if cidr.IsDomainName(vipDomain) && lbInfo.Domain != vipDomain {
			return nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] domain is not match, "+
				"paramDomain: %s, clbDomain: %s", cloudID, vipDomain, lbInfo.Domain)
		}

		switch lbInfo.LBType {
		case string(loadbalancer.InternalLoadBalancerType): // 内网
			if cidr.IsIPv4(vipDomain) && !slice.IsItemInSlice(lbInfo.PrivateIPv4Addresses, vipDomain) {
				return nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] privateIPv4 is not match, "+
					"paramIPv4: %s, clbPrivateIPv4: %v", cloudID, vipDomain, lbInfo.PrivateIPv4Addresses)
			}
			if cidr.IsIPv6(vipDomain) && !slice.IsItemInSlice(lbInfo.PrivateIPv6Addresses, vipDomain) {
				return nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] privateIPv6 is not match, "+
					"paramIPv6: %s, clbPrivateIPv6: %v", cloudID, vipDomain, lbInfo.PrivateIPv6Addresses)
			}
		case string(loadbalancer.OpenLoadBalancerType): // 公网
			if cidr.IsIPv4(vipDomain) && !slice.IsItemInSlice(lbInfo.PublicIPv4Addresses, vipDomain) {
				return nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] publicIPv4 is not match, "+
					"paramIPv4: %s, clbPublicIPv4: %v", cloudID, vipDomain, lbInfo.PublicIPv4Addresses)
			}
			if cidr.IsIPv6(vipDomain) && !slice.IsItemInSlice(lbInfo.PublicIPv6Addresses, vipDomain) {
				return nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] publicIPv6 is not match, "+
					"paramIPv6: %s, clbPublicIPv6: %v", cloudID, vipDomain, lbInfo.PublicIPv6Addresses)
			}
		default:
			return nil, nil, errf.Newf(errf.InvalidParameter, "unsupported hcm lb type: %s", lbInfo.LBType)
		}
		cloudClbIDs = append(cloudClbIDs, cloudID)
	}

	return slice.Unique(cloudClbIDs), lbMap, nil
}

// listBizListenerByLbIDs 获取业务下指定账号、负载均衡ID列表下的监听器列表
func (svc *lbSvc) listBizListenerByLbIDs(kt *kit.Kit, req *protocloud.ListListenerWithTargetsReq,
	lblReq protocloud.ListenerQueryItem, cloudClbIDs []string) (
	map[string]tablelb.LoadBalancerListenerTable, []string, []tablelb.LoadBalancerListenerTable, error) {

	lblFilter := make([]*filter.AtomRule, 0)
	lblFilter = append(lblFilter, tools.RuleEqual("vendor", req.Vendor))
	lblFilter = append(lblFilter, tools.RuleEqual("bk_biz_id", req.BkBizID))
	lblFilter = append(lblFilter, tools.RuleEqual("account_id", req.AccountID))
	lblFilter = append(lblFilter, tools.RuleIn("cloud_lb_id", cloudClbIDs))
	lblFilter = append(lblFilter, tools.RuleEqual("protocol", lblReq.Protocol))
	if len(lblReq.Ports) > 0 {
		lblFilter = append(lblFilter, tools.RuleIn("port", lblReq.Ports))
	}

	lblList := make([]tablelb.LoadBalancerListenerTable, 0)
	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(lblFilter...),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		loopLblList, err := svc.dao.LoadBalancerListener().List(kt, opt)
		if err != nil {
			logs.Errorf("list biz listener by clbIDs failed, err: %v, req: %+v, rid: %s",
				err, cvt.PtrToVal(req), kt.Rid)
			return nil, nil, nil, fmt.Errorf("list biz listener by clbIDs failed, err: %v", err)
		}

		lblList = append(lblList, loopLblList.Details...)
		if uint(len(loopLblList.Details)) < core.DefaultMaxPageLimit {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	lblProtocolPortMap := make(map[string]tablelb.LoadBalancerListenerTable, len(lblList))
	lblMap := make(map[string]tablelb.LoadBalancerListenerTable, len(lblList))
	cloudLblIDs := make([]string, 0)
	for _, item := range lblList {
		cloudLblIDs = append(cloudLblIDs, item.CloudID)
		lblMap[item.CloudID] = item
		lblProtocolPortMap[fmt.Sprintf("%s_%d", item.Protocol, item.Port)] = item
	}

	// 如果传入了监听器端口，则需要进行校验
	if len(lblReq.Ports) > 0 {
		for _, port := range lblReq.Ports {
			if _, ok := lblProtocolPortMap[fmt.Sprintf("%s_%d", lblReq.Protocol, port)]; !ok {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "listener protocol[%s] port[%d] is not found",
					lblReq.Protocol, port)
			}
		}
	}

	return lblMap, cloudLblIDs, lblList, nil
}

// listListenerWithTarget 根据账号ID、RsIP查询绑定的目标组列表
func (svc *lbSvc) listListenerWithTarget(kt *kit.Kit, req *protocloud.ListListenerWithTargetsReq,
	lblReq protocloud.ListenerQueryItem, cloudTargetGroupIDs []string) (
	map[string][]protocloud.LoadBalancerTargetRsList, []string, error) {

	targetList, err := svc.listTargetByCond(kt, req, lblReq, cloudTargetGroupIDs)
	if err != nil {
		return nil, nil, err
	}

	// 如果传入了RSPORT，则进行校验
	var targetIPPortMap = make(map[string]struct{}, len(targetList))
	if len(lblReq.RsPorts) > 0 {
		for idx, ip := range lblReq.RsIPs {
			targetIPPortMap[fmt.Sprintf("%s_%s_%d", lblReq.InstType, ip, lblReq.RsPorts[idx])] = struct{}{}
		}
	}

	// 统计每个目标组有多少RS
	targetGroupRsList := make(map[string][]protocloud.LoadBalancerTargetRsList)
	targetGroupIDs := make([]string, 0)
	for _, item := range targetList {
		// 不符合的数据需要过滤掉
		if _, ok := targetIPPortMap[fmt.Sprintf("%s_%s_%d", item.InstType, item.IP, item.Port)]; !ok &&
			len(lblReq.RsPorts) > 0 {
			logs.Warnf("list load balancer target rsip[%s] port[%d] is not found, rid: %s", item.IP, item.Port, kt.Rid)
			continue
		}

		if _, ok := targetGroupRsList[item.TargetGroupID]; !ok {
			targetGroupRsList[item.TargetGroupID] = make([]protocloud.LoadBalancerTargetRsList, 0)
		}
		targetGroupIDs = append(targetGroupIDs, item.TargetGroupID)
		targetGroupRsList[item.TargetGroupID] = append(targetGroupRsList[item.TargetGroupID],
			protocloud.LoadBalancerTargetRsList{
				BaseTarget: item,
			})
	}
	return targetGroupRsList, slice.Unique(targetGroupIDs), nil
}

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

func (svc *lbSvc) listTCloudLBUrlRuleByTgIDs(kt *kit.Kit,
	lblReq protocloud.ListenerQueryItem, cloudClbIDs, cloudLblIDs, targetGroupIDs []string) (
	[]protocloud.LoadBalancerUrlRuleResult, error) {

	lblTargetFilter := make([]*filter.AtomRule, 0)
	lblTargetFilter = append(lblTargetFilter, tools.RuleIn("cloud_lb_id", cloudClbIDs))
	lblTargetFilter = append(lblTargetFilter, tools.RuleIn("cloud_lbl_id", cloudLblIDs))
	if len(targetGroupIDs) > 0 {
		lblTargetFilter = append(lblTargetFilter, tools.RuleIn("target_group_id", targetGroupIDs))
	}
	if len(lblReq.RuleType) > 0 {
		lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("rule_type", lblReq.RuleType))
		if lblReq.RuleType == enumor.Layer7RuleType {
			if len(lblReq.Domain) > 0 {
				lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("domain", lblReq.Domain))
			}
			if len(lblReq.Url) > 0 {
				lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("url", lblReq.Url))
			}
		}
	}
	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(lblTargetFilter...),
		Page:   core.NewDefaultBasePage(),
	}
	lblTargetList := make([]protocloud.LoadBalancerUrlRuleResult, 0)
	for {
		loopLblTargetList, err := svc.dao.LoadBalancerTCloudUrlRule().List(kt, opt)
		if err != nil {
			logs.Errorf("list load balancer tcloud url rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list load balancer tcloud url rule failed, err: %v", err)
		}

		for _, item := range loopLblTargetList.Details {
			urlRuleResult := protocloud.LoadBalancerUrlRuleResult{
				LbID:              item.LbID,
				CloudClbID:        item.CloudLbID,
				LblID:             item.LblID,
				CloudLblID:        item.CloudLBLID,
				TargetGrouRuleMap: make(map[string]protocloud.DomainUrlRuleInfo),
			}
			urlRuleResult.TargetGroupIDs = append(urlRuleResult.TargetGroupIDs, item.TargetGroupID)
			urlRuleResult.TargetGrouRuleMap[item.TargetGroupID] = protocloud.DomainUrlRuleInfo{
				RuleID:      item.ID,
				CloudRuleID: item.CloudID,
				RuleType:    item.RuleType,
				Domain:      item.Domain,
				Url:         item.URL,
			}
			lblTargetList = append(lblTargetList, urlRuleResult)
		}
		if uint(len(loopLblTargetList.Details)) < core.DefaultMaxPageLimit {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return lblTargetList, nil
}

// ListBatchListeners list batch listener.
func (svc *lbSvc) ListBatchListeners(cts *rest.Contexts) (any, error) {
	req := new(protocloud.BatchDeleteListenerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listenerList := &protocloud.BatchListListenerResp{}
	for _, item := range req.ListenerQueryList {
		lblList, err := svc.batchQueryListeners(cts.Kit, req, item)
		if err != nil {
			return nil, err
		}
		listenerList.Details = append(listenerList.Details, lblList...)
	}
	return listenerList, nil
}

func (svc *lbSvc) batchQueryListeners(kt *kit.Kit, req *protocloud.BatchDeleteListenerReq,
	lblReq *protocloud.ListenerDeleteReq) ([]*corelb.BaseListener, error) {

	// 查询符合条件的负载均衡列表
	lbReq := &protocloud.ListListenerWithTargetsReq{
		Vendor:    req.Vendor,
		AccountID: req.AccountID,
		BkBizID:   req.BkBizID,
	}
	listenerReq := protocloud.ListenerQueryItem{
		Region:        lblReq.Region,
		ClbVipDomains: lblReq.ClbVipDomains,
		CloudLbIDs:    lblReq.CloudLbIDs,
		Protocol:      lblReq.Protocol,
		Ports:         lblReq.Ports,
	}
	cloudClbIDs, _, err := svc.listLoadBalancerListCheckVip(kt, lbReq, listenerReq)
	if err != nil {
		return nil, err
	}

	// 未查询到符合条件的负载均衡列表
	if len(cloudClbIDs) == 0 {
		logs.Errorf("check list load balancer empty, req: %+v, lblReq: %+v, rid: %s", cvt.PtrToVal(req), lblReq, kt.Rid)
		return nil, nil
	}

	// 查询符合条件的监听器列表
	_, _, lblList, err := svc.listBizListenerByLbIDs(kt, lbReq, listenerReq, cloudClbIDs)
	if err != nil {
		return nil, err
	}

	// 未查询到符合的监听器列表
	if len(lblList) == 0 {
		logs.Errorf("list biz listener empty, req: %+v, lblReq: %+v, rid: %s", cvt.PtrToVal(req), lblReq, kt.Rid)
		return nil, nil
	}

	return svc.convertBatchListListener(lblList)
}

func (svc *lbSvc) convertBatchListListener(lblList []tablelb.LoadBalancerListenerTable) (
	[]*corelb.BaseListener, error) {

	lblResult := make([]*corelb.BaseListener, 0)
	for _, item := range lblList {
		lblResult = append(lblResult, &corelb.BaseListener{
			ID:            item.ID,
			CloudID:       item.CloudID,
			Name:          item.Name,
			Vendor:        item.Vendor,
			AccountID:     item.AccountID,
			BkBizID:       item.BkBizID,
			LbID:          item.LBID,
			CloudLbID:     item.CloudLBID,
			Protocol:      item.Protocol,
			Port:          item.Port,
			DefaultDomain: item.DefaultDomain,
			Region:        item.Region,
			Zones:         item.Zones,
			SniSwitch:     item.SniSwitch,
		})
	}
	return lblResult, nil
}

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
