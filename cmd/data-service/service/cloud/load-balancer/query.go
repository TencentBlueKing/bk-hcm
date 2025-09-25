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

// convTableToBaseLB convert LoadBalancerTable to BaseLoadBalancer.
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

// convLbListResult convert load balancer list result to extended type.
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
	result, err := svc.dao.LoadBalancerTCloudUrlRule().ListJoinListener(cts.Kit, opt)
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

// convTableToBaseTCloudLbURLRule convert TCloudLbUrlRuleTable to TCloudLbUrlRule.
func convTableToBaseTCloudLbURLRule(kt *kit.Kit, one *tablelb.TCloudLbUrlRuleWithListener) (
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
		BkBizID:            one.BkBizID,
		AccountID:          one.AccountID,
		LblName:            one.LblName,
		Protocol:           enumor.ProtocolType(one.Protocol),
		Port:               int64(one.Port),
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}, nil
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

// listLoadBalancerListCheckVip 获取负载均衡列表并检查VIP、域名是否匹配
func (svc *lbSvc) listLoadBalancerListCheckVip(kt *kit.Kit, lblReq protocloud.ListListenerQueryReq) (
	[]string, []string, map[string]tablelb.LoadBalancerTable, error) {

	lbOpt := &types.ListOption{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", lblReq.Vendor),
			tools.RuleEqual("bk_biz_id", lblReq.BkBizID),
			tools.RuleEqual("account_id", lblReq.AccountID),
			tools.RuleEqual("region", lblReq.ListenerQueryItem.Region),
			tools.RuleIn("cloud_id", lblReq.ListenerQueryItem.CloudLbIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	lbAllList := make([]tablelb.LoadBalancerTable, 0)
	for {
		lbList, err := svc.dao.LoadBalancer().List(kt, lbOpt)
		if err != nil {
			logs.Errorf("check list load balancer failed, err: %v, req: %+v, rid: %s", err, lblReq, kt.Rid)
			return nil, nil, nil, fmt.Errorf("list load balancer failed, err: %v", err)
		}

		lbAllList = append(lbAllList, lbList.Details...)
		if len(lbList.Details) <= int(core.DefaultMaxPageLimit) {
			break
		}
		lbOpt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	// 检查ip地址/域名，是否在负载均衡的ip地址列表中
	cloudClbIDs, clbIDs, lbMap, err := checkClbVipAndDomain(lbAllList, lblReq.ListenerQueryItem.CloudLbIDs,
		lblReq.ListenerQueryItem.ClbVipDomains)
	if err != nil {
		logs.Errorf("check list load balancer and ip domain match failed, err: %v, req: %+v, rid: %s",
			err, lblReq, kt.Rid)
		return nil, nil, nil, err
	}

	return cloudClbIDs, clbIDs, lbMap, nil
}

func checkClbVipAndDomain(list []tablelb.LoadBalancerTable, paramClbIDs, clbVipDomains []string) (
	[]string, []string, map[string]tablelb.LoadBalancerTable, error) {

	lbMap := cvt.SliceToMap(list, func(item tablelb.LoadBalancerTable) (string, tablelb.LoadBalancerTable) {
		return item.CloudID, item
	})

	cloudClbIDs := make([]string, 0)
	clbIDs := make([]string, 0)
	for idx, cloudID := range paramClbIDs {
		lbInfo, ok := lbMap[cloudID]
		if !ok {
			return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] is not found", cloudID)
		}

		// 检查对应的负载均衡VIP/域名是否匹配
		vipDomain := clbVipDomains[idx]
		if cidr.IsDomainName(vipDomain) && lbInfo.Domain != vipDomain {
			return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] domain is not match, "+
				"paramDomain: %s, clbDomain: %s", cloudID, vipDomain, lbInfo.Domain)
		}

		switch lbInfo.LBType {
		case string(loadbalancer.InternalLoadBalancerType): // 内网
			if cidr.IsIPv4(vipDomain) && !slice.IsItemInSlice(lbInfo.PrivateIPv4Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] privateIPv4 is not match, "+
					"paramIPv4: %s, clbPrivateIPv4: %v", cloudID, vipDomain, lbInfo.PrivateIPv4Addresses)
			}
			if cidr.IsIPv6(vipDomain) && !slice.IsItemInSlice(lbInfo.PrivateIPv6Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] privateIPv6 is not match, "+
					"paramIPv6: %s, clbPrivateIPv6: %v", cloudID, vipDomain, lbInfo.PrivateIPv6Addresses)
			}
		case string(loadbalancer.OpenLoadBalancerType): // 公网
			if cidr.IsIPv4(vipDomain) && !slice.IsItemInSlice(lbInfo.PublicIPv4Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] publicIPv4 is not match, "+
					"paramIPv4: %s, clbPublicIPv4: %v", cloudID, vipDomain, lbInfo.PublicIPv4Addresses)
			}
			if cidr.IsIPv6(vipDomain) && !slice.IsItemInSlice(lbInfo.PublicIPv6Addresses, vipDomain) {
				return nil, nil, nil, errf.Newf(errf.InvalidParameter, "load balancer[%s] publicIPv6 is not match, "+
					"paramIPv6: %s, clbPublicIPv6: %v", cloudID, vipDomain, lbInfo.PublicIPv6Addresses)
			}
		default:
			return nil, nil, nil, errf.Newf(errf.InvalidParameter, "unsupported hcm lb type: %s", lbInfo.LBType)
		}
		cloudClbIDs = append(cloudClbIDs, cloudID)
		clbIDs = append(clbIDs, lbInfo.ID)
	}

	return slice.Unique(cloudClbIDs), clbIDs, lbMap, nil
}
