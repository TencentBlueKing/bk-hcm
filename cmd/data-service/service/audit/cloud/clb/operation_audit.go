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

package clb

import (
	"errors"
	"fmt"

	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"
)

// ClbOperationAuditBuild clb operation audit build.
func (c *LoadBalancer) ClbOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	assOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	for _, operation := range operations {
		switch operation.Action {
		case protoaudit.Associate, protoaudit.Disassociate:
			assOperations = append(assOperations, operation)

		default:
			return nil, fmt.Errorf("audit action: %s not support", operation.Action)
		}
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	if len(assOperations) != 0 {
		audit, err := c.assOperationAuditBuild(kt, assOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	return audits, nil
}

func (c *LoadBalancer) baseOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(operations))
	for _, one := range operations {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListLoadBalancer(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	for _, one := range operations {
		clbInfo, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: clbInfo.CloudID,
			ResName:    clbInfo.Name,
			ResType:    enumor.LoadBalancerAuditResType,
			Action:     action,
			BkBizID:    clbInfo.BkBizID,
			Vendor:     clbInfo.Vendor,
			AccountID:  clbInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail:     &tableaudit.BasicDetail{},
		})
	}

	return audits, nil
}

func (c *LoadBalancer) assOperationAuditBuild(_ *kit.Kit, _ []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	return nil, errors.New("not supported")
}
