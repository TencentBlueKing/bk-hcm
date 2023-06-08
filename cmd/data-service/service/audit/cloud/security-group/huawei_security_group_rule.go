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

package securitygroup

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (s *SecurityGroup) huaWeiSGRuleUpdateAuditBuild(kt *kit.Kit, sg tablecloud.SecurityGroupTable,
	updates []protoaudit.CloudResourceUpdateInfo) ([]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}

	idSgRuleMap, err := s.listHuaWeiSGRule(kt, sg.ID, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		rule, exist := idSgRuleMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Update,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.ChildResAuditData{
					ChildResType: enumor.SecurityGroupRuleAuditResType,
					Action:       enumor.Update,
					ChildRes:     rule,
				},
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

func (s *SecurityGroup) huaWeiSGRuleDeleteAuditBuild(kt *kit.Kit, sg tablecloud.SecurityGroupTable,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}

	idSgRuleMap, err := s.listHuaWeiSGRule(kt, sg.ID, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		rule, exist := idSgRuleMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Update,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.ChildResAuditData{
					ChildResType: enumor.SecurityGroupRuleAuditResType,
					Action:       enumor.Delete,
					ChildRes:     rule,
				},
			},
		})
	}

	return audits, nil
}

func (s *SecurityGroup) listHuaWeiSGRule(kt *kit.Kit, sgID string, ids []string) (
	map[string]tablecloud.HuaWeiSecurityGroupRuleTable, error) {

	opt := &types.SGRuleListOption{
		SecurityGroupID: sgID,
		Filter:          tools.ContainersExpression("id", ids),
		Page:            core.NewDefaultBasePage(),
	}
	list, err := s.dao.HuaWeiSGRule().List(kt, opt)
	if err != nil {
		logs.Errorf("list huawei security group rule failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecloud.HuaWeiSecurityGroupRuleTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
