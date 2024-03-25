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

// Package firewall ...
package firewall

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// NewFirewall new firewall.
func NewFirewall(dao dao.Set) *Firewall {
	return &Firewall{
		dao: dao,
	}
}

// Firewall define firewall audit.
type Firewall struct {
	dao dao.Set
}

// FirewallRuleUpdateAuditBuild firewall rule update audit build.
func (f *Firewall) FirewallRuleUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idMap, err := f.listFirewallRule(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		rule, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: rule.CloudID,
			ResName:    rule.Name,
			ResType:    enumor.GcpFirewallRuleAuditResType,
			Action:     enumor.Update,
			BkBizID:    rule.BkBizID,
			Vendor:     enumor.Gcp,
			AccountID:  rule.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    rule,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// FirewallRuleDeleteAuditBuild firewall rule delete audit build.
func (f *Firewall) FirewallRuleDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idMap, err := f.listFirewallRule(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		rule, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: rule.CloudID,
			ResName:    rule.Name,
			ResType:    enumor.GcpFirewallRuleAuditResType,
			Action:     enumor.Delete,
			BkBizID:    rule.BkBizID,
			Vendor:     enumor.Gcp,
			AccountID:  rule.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: rule,
			},
		})
	}

	return audits, nil
}

// FirewallRuleAssignAuditBuild firewall rule assign audit build.
func (f *Firewall) FirewallRuleAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idMap, err := f.listFirewallRule(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		rule, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: rule.CloudID,
			ResName:    rule.Name,
			ResType:    enumor.GcpFirewallRuleAuditResType,
			Action:     enumor.Assign,
			BkBizID:    rule.BkBizID,
			Vendor:     enumor.Gcp,
			AccountID:  rule.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Changed: map[string]interface{}{
					"bk_biz_id": one.AssignedResID,
				},
			},
		})
	}

	return audits, nil
}

func (f *Firewall) listFirewallRule(kt *kit.Kit, ids []string) (map[string]tablecloud.GcpFirewallRuleTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := f.dao.GcpFirewallRule().List(kt, opt)
	if err != nil {
		logs.Errorf("list gcp firewall rule failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecloud.GcpFirewallRuleTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
