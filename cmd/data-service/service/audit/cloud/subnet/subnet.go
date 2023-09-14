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

// Package subnet ...
package subnet

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
	"hcm/pkg/tools/converter"
)

// NewSubnet new subnet.
func NewSubnet(dao dao.Set) *Subnet {
	return &Subnet{
		dao: dao,
	}
}

// Subnet define subnet audit.
type Subnet struct {
	dao dao.Set
}

// SubnetUpdateAuditBuild ...
func (ad *Subnet) SubnetUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idSubnetMap, err := ListSubnet(kt, ad.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		subnet, exist := idSubnetMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: subnet.CloudID,
			ResName:    converter.PtrToVal(subnet.Name),
			ResType:    enumor.SubnetAuditResType,
			Action:     enumor.Update,
			BkBizID:    subnet.BkBizID,
			Vendor:     subnet.Vendor,
			AccountID:  subnet.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    subnet,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// SubnetDeleteAuditBuild ...
func (ad *Subnet) SubnetDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idSubnetMap, err := ListSubnet(kt, ad.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		subnet, exist := idSubnetMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: subnet.CloudID,
			ResName:    converter.PtrToVal(subnet.Name),
			ResType:    enumor.SubnetAuditResType,
			Action:     enumor.Delete,
			BkBizID:    subnet.BkBizID,
			Vendor:     subnet.Vendor,
			AccountID:  subnet.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: subnet,
			},
		})
	}

	return audits, nil
}

// SubnetAssignAuditBuild ...
func (ad *Subnet) SubnetAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idSubnetMap, err := ListSubnet(kt, ad.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		subnet, exist := idSubnetMap[one.ResID]
		if !exist {
			continue
		}

		var action enumor.AuditAction
		switch one.AssignedResType {
		case enumor.BizAuditAssignedResType:
			action = enumor.Assign
		case enumor.DeliverAssignedResType:
			action = enumor.Deliver
		default:
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: subnet.CloudID,
			ResName:    converter.PtrToVal(subnet.Name),
			ResType:    enumor.SubnetAuditResType,
			Action:     action,
			BkBizID:    subnet.BkBizID,
			Vendor:     subnet.Vendor,
			AccountID:  subnet.AccountID,
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

// ListSubnet list subnet.
func ListSubnet(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablecloud.SubnetTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.Subnet().List(kt, opt)
	if err != nil {
		logs.Errorf("list subnets failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecloud.SubnetTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
