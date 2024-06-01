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
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// NewLoadBalancer new clb.
func NewLoadBalancer(dao dao.Set) *LoadBalancer {
	return &LoadBalancer{
		dao: dao,
	}
}

// LoadBalancer define clb audit.
type LoadBalancer struct {
	dao dao.Set
}

// LoadBalancerUpdateAuditBuild clb update audit build.
func (c *LoadBalancer) LoadBalancerUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListLoadBalancer(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		clbInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clbInfo.CloudID,
			ResName:    clbInfo.Name,
			ResType:    enumor.LoadBalancerAuditResType,
			Action:     enumor.Update,
			BkBizID:    clbInfo.BkBizID,
			Vendor:     clbInfo.Vendor,
			AccountID:  clbInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    clbInfo,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// LoadBalancerDeleteAuditBuild clb delete audit build.
func (c *LoadBalancer) LoadBalancerDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListLoadBalancer(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		clbInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clbInfo.CloudID,
			ResName:    clbInfo.Name,
			ResType:    enumor.LoadBalancerAuditResType,
			Action:     enumor.Delete,
			BkBizID:    clbInfo.BkBizID,
			Vendor:     clbInfo.Vendor,
			AccountID:  clbInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: clbInfo,
			},
		})
	}

	return audits, nil
}

// LoadBalancerAssignAuditBuild clb assign audit build.
func (c *LoadBalancer) LoadBalancerAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListLoadBalancer(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		clbInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audit := &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clbInfo.CloudID,
			ResName:    clbInfo.Name,
			ResType:    enumor.LoadBalancerAuditResType,
			Action:     enumor.Assign,
			BkBizID:    clbInfo.BkBizID,
			Vendor:     clbInfo.Vendor,
			AccountID:  clbInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Changed: map[string]interface{}{
					"bk_biz_id": one.AssignedResID,
				},
			},
		}

		switch one.AssignedResType {
		case enumor.BizAuditAssignedResType:
			audit.Action = enumor.Assign
		case enumor.DeliverAssignedResType:
			audit.Action = enumor.Deliver
		default:
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audits = append(audits, audit)
	}

	return audits, nil
}

// UrlRuleUpdateAuditBuild url 规则更新
func (c *LoadBalancer) UrlRuleUpdateAuditBuild(kt *kit.Kit, lblID string,
	updates []protoaudit.CloudResourceUpdateInfo) ([]*tableaudit.AuditTable, error) {

	idListenerMap, err := ListListener(kt, c.dao, []string{lblID})
	if err != nil {
		return nil, err
	}

	lbl, exist := idListenerMap[lblID]
	if !exist {
		return nil, errf.Newf(errf.RecordNotFound, "listener: %s not found", lblID)
	}

	switch lbl.Vendor {
	case enumor.TCloud:
		return c.tcloudUrlRuleUpdateAuditBuild(kt, lbl, updates)
	default:
		return nil, fmt.Errorf("vendor: %s not support", lbl.Vendor)
	}
}

func (c *LoadBalancer) tcloudUrlRuleUpdateAuditBuild(kt *kit.Kit, lbl tablelb.LoadBalancerListenerTable,
	updates []protoaudit.CloudResourceUpdateInfo) ([]*tableaudit.AuditTable, error) {

	ids := slice.Map(updates, func(u protoaudit.CloudResourceUpdateInfo) string { return u.ResID })

	idListenerRuleMap, err := ListTCloudUrlRule(kt, c.dao, lbl.ID, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		rule, exist := idListenerRuleMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: lbl.CloudID,
			ResName:    lbl.Name,
			ResType:    enumor.ListenerAuditResType,
			Action:     enumor.Update,
			BkBizID:    lbl.BkBizID,
			Vendor:     lbl.Vendor,
			AccountID:  lbl.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.ChildResAuditData{
					ChildResType: enumor.UrlRuleAuditResType,
					Action:       enumor.Update,
					ChildRes:     rule,
				},
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil

}

// UrlRuleDeleteAuditBuild 删除规则审计
func (c *LoadBalancer) UrlRuleDeleteAuditBuild(kt *kit.Kit, lblID string,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	idListenerMap, err := ListListener(kt, c.dao, []string{lblID})
	if err != nil {
		return nil, err
	}

	lbl, exist := idListenerMap[lblID]
	if !exist {
		return nil, errf.Newf(errf.RecordNotFound, "listener: %s not found", lblID)
	}

	switch lbl.Vendor {
	case enumor.TCloud:
		return c.tcloudUrlRuleDeleteAuditBuild(kt, lbl, deletes)
	default:
		return nil, fmt.Errorf("vendor: %s not support", lbl.Vendor)
	}
}
func (c *LoadBalancer) tcloudUrlRuleDeleteAuditBuild(kt *kit.Kit, lbl tablelb.LoadBalancerListenerTable,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	ids := slice.Map(deletes, func(u protoaudit.CloudResourceDeleteInfo) string { return u.ResID })

	idRuleMap, err := ListTCloudUrlRule(kt, c.dao, lbl.ID, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		ruleInfo, exist := idRuleMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: ruleInfo.CloudID,
			ResName:    ruleInfo.Name,
			ResType:    enumor.UrlRuleAuditResType,
			Action:     enumor.Delete,
			BkBizID:    lbl.BkBizID,
			Vendor:     lbl.Vendor,
			AccountID:  lbl.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: ruleInfo,
			},
		})
	}

	return audits, nil
}

// ListTCloudUrlRule ...
func ListTCloudUrlRule(kt *kit.Kit, dao dao.Set, lblID string,
	ruleIds []string) (map[string]tablelb.TCloudLbUrlRuleTable, error) {

	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(tools.RuleEqual("lbl_id", lblID), tools.RuleIn("id", ruleIds)),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancerTCloudUrlRule().List(kt, opt)
	if err != nil {
		logs.Errorf("list tcloud url rule of  listener(id=%s) failed, err: %v, ids: %v, rid: %s",
			lblID, err, ruleIds, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablelb.TCloudLbUrlRuleTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

// UrlRuleDeleteByDomainAuditBuild 按域名删除url规则
func (c *LoadBalancer) UrlRuleDeleteByDomainAuditBuild(kt *kit.Kit, lblID string,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	idListenerMap, err := ListListener(kt, c.dao, []string{lblID})
	if err != nil {
		return nil, err
	}

	lbl, exist := idListenerMap[lblID]
	if !exist {
		return nil, errf.Newf(errf.RecordNotFound, "listener: %s not found", lblID)
	}

	switch lbl.Vendor {
	case enumor.TCloud:
		return c.tcloudUrlRuleDeleteByDomainAuditBuild(kt, lbl, deletes)
	default:
		return nil, fmt.Errorf("vendor: %s not support", lbl.Vendor)
	}
}

func (c *LoadBalancer) tcloudUrlRuleDeleteByDomainAuditBuild(kt *kit.Kit, lbl tablelb.LoadBalancerListenerTable,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	domains := slice.Map(deletes, func(u protoaudit.CloudResourceDeleteInfo) string { return u.ResID })

	domainRuleMap, err := ListTCloudUrlRuleByDomain(kt, c.dao, lbl.ID, domains)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		rules, exist := domainRuleMap[one.ResID]
		if !exist {
			// 找不到与域名，返回错误
			return nil, fmt.Errorf("fail to find rule while delete url by domain: %s", one.ResID)
		}
		// add domain and each into audits
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: one.ResID,
			ResName:    one.ResID,
			ResType:    enumor.UrlRuleDomainAuditResType,
			Action:     enumor.Delete,
			BkBizID:    lbl.BkBizID,
			Vendor:     lbl.Vendor,
			AccountID:  lbl.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: one.ResID,
			},
		})
		for _, rule := range rules {
			audits = append(audits, &tableaudit.AuditTable{
				ResID:      rule.ID,
				CloudResID: rule.CloudID,
				ResName:    rule.Name,
				ResType:    enumor.UrlRuleAuditResType,
				Action:     enumor.Delete,
				BkBizID:    lbl.BkBizID,
				Vendor:     lbl.Vendor,
				AccountID:  lbl.AccountID,
				Operator:   kt.User,
				Source:     kt.GetRequestSource(),
				Rid:        kt.Rid,
				AppCode:    kt.AppCode,
				Detail: &tableaudit.BasicDetail{
					Data: rule,
				},
			})
		}
	}

	return audits, nil
}

// ListTCloudUrlRuleByDomain ...
func ListTCloudUrlRuleByDomain(kt *kit.Kit, dao dao.Set, lblID string,
	domains []string) (map[string][]tablelb.TCloudLbUrlRuleTable, error) {

	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(tools.RuleEqual("lbl_id", lblID), tools.RuleIn("domain", domains)),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancerTCloudUrlRule().List(kt, opt)
	if err != nil {
		logs.Errorf("list tcloud url rule of listener(id=%s) failed, err: %v, domains: %v, rid: %s",
			lblID, err, domains, kt.Rid)
		return nil, err
	}

	result := make(map[string][]tablelb.TCloudLbUrlRuleTable, len(list.Details))
	for _, one := range list.Details {
		result[one.Domain] = append(result[one.Domain], one)
	}

	return result, nil
}

// ListLoadBalancer list load balancer.
func ListLoadBalancer(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablelb.LoadBalancerTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancer().List(kt, opt)
	if err != nil {
		logs.Errorf("list load balancer failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablelb.LoadBalancerTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

// ListTargetGroup list target group.
func ListTargetGroup(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablelb.LoadBalancerTargetGroupTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancerTargetGroup().List(kt, opt)
	if err != nil {
		logs.Errorf("list target group failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablelb.LoadBalancerTargetGroupTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

// ListListener list listener.
func ListListener(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablelb.LoadBalancerListenerTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancerListener().List(kt, opt)
	if err != nil {
		logs.Errorf("list listener failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablelb.LoadBalancerListenerTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
