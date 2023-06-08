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

package cloud

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

func (ad Audit) vpcUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idVpcMap, err := ad.listVpc(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		vpc, exist := idVpcMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: vpc.CloudID,
			ResName:    converter.PtrToVal(vpc.Name),
			ResType:    enumor.VpcCloudAuditResType,
			Action:     enumor.Update,
			BkBizID:    vpc.BkBizID,
			Vendor:     vpc.Vendor,
			AccountID:  vpc.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    vpc,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

func (ad Audit) vpcDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idVpcMap, err := ad.listVpc(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		vpc, exist := idVpcMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: vpc.CloudID,
			ResName:    converter.PtrToVal(vpc.Name),
			ResType:    enumor.VpcCloudAuditResType,
			Action:     enumor.Delete,
			BkBizID:    vpc.BkBizID,
			Vendor:     vpc.Vendor,
			AccountID:  vpc.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: vpc,
			},
		})
	}

	return audits, nil
}

func (ad Audit) vpcAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idVpcMap, err := ad.listVpc(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		vpc, exist := idVpcMap[one.ResID]
		if !exist {
			continue
		}

		audit := &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: vpc.CloudID,
			ResName:    converter.PtrToVal(vpc.Name),
			ResType:    enumor.VpcCloudAuditResType,
			Action:     enumor.Assign,
			BkBizID:    vpc.BkBizID,
			Vendor:     vpc.Vendor,
			AccountID:  vpc.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
		}

		changed := make(map[string]interface{})
		switch one.AssignedResType {
		case enumor.BizAuditAssignedResType:
			changed["bk_biz_id"] = one.AssignedResID
		case enumor.CloudAreaAuditAssignedResType:
			changed["bk_cloud_id"] = one.AssignedResID
			audit.Action = enumor.Bind
		case enumor.DeliverAssignedResType:
			changed["bk_biz_id"] = one.AssignedResID
			audit.Action = enumor.Deliver
		default:
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audit.Detail = &tableaudit.BasicDetail{
			Changed: changed,
		}
		audits = append(audits, audit)
	}

	return audits, nil
}

func (ad Audit) listVpc(kt *kit.Kit, ids []string) (map[string]tablecloud.VpcTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := ad.dao.Vpc().List(kt, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecloud.VpcTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
