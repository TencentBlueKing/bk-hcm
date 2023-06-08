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

package networkinterface

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tableni "hcm/pkg/dal/table/cloud/network-interface"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// NewNetworkInterface new network interface.
func NewNetworkInterface(dao dao.Set) *NetworkInterface {
	return &NetworkInterface{
		dao: dao,
	}
}

// NetworkInterface define network interface audit.
type NetworkInterface struct {
	dao dao.Set
}

// NetworkInterfaceUpdateAuditBuild network interface update audit build.
func (n *NetworkInterface) NetworkInterfaceUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListNetworkInterface(kt, n.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		networkInterface, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: networkInterface.CloudID,
			ResName:    networkInterface.Name,
			ResType:    enumor.NetworkInterfaceAuditResType,
			Action:     enumor.Update,
			BkBizID:    networkInterface.BkBizID,
			Vendor:     networkInterface.Vendor,
			AccountID:  networkInterface.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    networkInterface,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// NetworkInterfaceDeleteAuditBuild network interface delete audit build.
func (n *NetworkInterface) NetworkInterfaceDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListNetworkInterface(kt, n.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		networkInterface, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: networkInterface.CloudID,
			ResName:    networkInterface.Name,
			ResType:    enumor.NetworkInterfaceAuditResType,
			Action:     enumor.Delete,
			BkBizID:    networkInterface.BkBizID,
			Vendor:     networkInterface.Vendor,
			AccountID:  networkInterface.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: networkInterface,
			},
		})
	}

	return audits, nil
}

// NetworkInterfaceAssignAuditBuild network interface assign audit.
func (n *NetworkInterface) NetworkInterfaceAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListNetworkInterface(kt, n.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		networkInterface, exist := idMap[one.ResID]
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
			CloudResID: networkInterface.CloudID,
			ResName:    networkInterface.Name,
			ResType:    enumor.NetworkInterfaceAuditResType,
			Action:     action,
			BkBizID:    networkInterface.BkBizID,
			Vendor:     networkInterface.Vendor,
			AccountID:  networkInterface.AccountID,
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

// ListNetworkInterface list network interface.
func ListNetworkInterface(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tableni.NetworkInterfaceTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.NetworkInterface().List(kt, opt)
	if err != nil {
		logs.Errorf("list network interface failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tableni.NetworkInterfaceTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
