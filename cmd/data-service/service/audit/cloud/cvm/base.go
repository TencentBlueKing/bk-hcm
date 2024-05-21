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

package cvm

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecvm "hcm/pkg/dal/table/cloud/cvm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// NewCvm new cvm.
func NewCvm(dao dao.Set) *Cvm {
	return &Cvm{
		dao: dao,
	}
}

// Cvm define cvm audit.
type Cvm struct {
	dao dao.Set
}

// CvmUpdateAuditBuild cvm update audit build.
func (c *Cvm) CvmUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListCvm(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		cvm, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: cvm.CloudID,
			ResName:    cvm.Name,
			ResType:    enumor.CvmAuditResType,
			Action:     enumor.Update,
			BkBizID:    cvm.BkBizID,
			Vendor:     cvm.Vendor,
			AccountID:  cvm.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    cvm,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// CvmDeleteAuditBuild cvm delete audit build.
func (c *Cvm) CvmDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListCvm(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		cvm, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: cvm.CloudID,
			ResName:    cvm.Name,
			ResType:    enumor.CvmAuditResType,
			Action:     enumor.Delete,
			BkBizID:    cvm.BkBizID,
			Vendor:     cvm.Vendor,
			AccountID:  cvm.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: cvm,
			},
		})
	}

	return audits, nil
}

// CvmAssignAuditBuild cvm assign audit build.
func (c *Cvm) CvmAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListCvm(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		cvm, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audit := &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: cvm.CloudID,
			ResName:    cvm.Name,
			ResType:    enumor.CvmAuditResType,
			Action:     enumor.Assign,
			BkBizID:    cvm.BkBizID,
			Vendor:     cvm.Vendor,
			AccountID:  cvm.AccountID,
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

// ListCvm list cvm.
func ListCvm(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablecvm.Table, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.Cvm().List(kt, opt)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecvm.Table, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
