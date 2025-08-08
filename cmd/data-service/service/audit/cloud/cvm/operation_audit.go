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
	"errors"
	"fmt"

	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"
)

// CvmOperationAuditBuild cvm operation audit build.
func (c *Cvm) CvmOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	baseOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	assOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	for _, operation := range operations {
		switch operation.Action {
		case protoaudit.Start, protoaudit.Stop, protoaudit.Reboot, protoaudit.ResetPwd, protoaudit.ResetSystem:
			baseOperations = append(baseOperations, operation)
		case protoaudit.Associate, protoaudit.Disassociate:
			assOperations = append(assOperations, operation)

		default:
			return nil, fmt.Errorf("audit action: %s not support", operation.Action)
		}
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	if len(baseOperations) != 0 {
		audit, err := c.baseOperationAuditBuild(kt, baseOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	if len(assOperations) != 0 {
		audit, err := c.assOperationAuditBuild(kt, assOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	return audits, nil
}

func (c *Cvm) baseOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(operations))
	for _, one := range operations {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListCvm(kt, c.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	for _, one := range operations {
		cvm, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: cvm.CloudID,
			ResName:    cvm.Name,
			ResType:    enumor.CvmAuditResType,
			Action:     action,
			BkBizID:    cvm.BkBizID,
			Vendor:     cvm.Vendor,
			AccountID:  cvm.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail:     &tableaudit.BasicDetail{},
		})
	}

	return audits, nil
}

func (c *Cvm) assOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {
	// TODO: 添加关联操作审计
	return nil, errors.New("暂不支持")
}
