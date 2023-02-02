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
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// NewSecurityGroup new firewall.
func NewSecurityGroup(dao dao.Set) *SecurityGroup {
	return &SecurityGroup{
		dao: dao,
	}
}

// SecurityGroup define firewall audit.
type SecurityGroup struct {
	dao dao.Set
}

// SecurityGroupUpdateAuditBuild security group update audit.
func (s *SecurityGroup) SecurityGroupUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idSgMap, err := s.listSecurityGroup(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		sg, exist := idSgMap[one.ResID]
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
				Data:    sg,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// SecurityGroupDeleteAuditBuild security group delete audit.
func (s *SecurityGroup) SecurityGroupDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idSgMap, err := s.listSecurityGroup(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		sg, exist := idSgMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Delete,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: sg,
			},
		})
	}

	return audits, nil
}

// SecurityGroupAssignAuditBuild security group assign audit.
func (s *SecurityGroup) SecurityGroupAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idSgMap, err := s.listSecurityGroup(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		sg, exist := idSgMap[one.ResID]
		if !exist {
			continue
		}

		if one.AssignedResType != enumor.BizAuditAssignedResType {
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Assign,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
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

func (s *SecurityGroup) listSecurityGroup(kt *kit.Kit, ids []string) (map[string]tablecloud.SecurityGroupTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.DefaultBasePage,
	}
	list, err := s.dao.SecurityGroup().List(kt, opt)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecloud.SecurityGroupTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
