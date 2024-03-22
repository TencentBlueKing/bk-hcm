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

	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"
)

// TargetGroupOperationAuditBuild target group operation audit build.
func (c *LoadBalancer) TargetGroupOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	lblAssOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	for _, operation := range operations {
		switch operation.Action {
		case protoaudit.Associate, protoaudit.Disassociate:
			switch operation.AssociatedResType {
			case enumor.ListenerAuditResType:
				lblAssOperations = append(lblAssOperations, operation)
			default:
				return nil, fmt.Errorf("audit associated resource type: %s not support", operation.AssociatedResType)
			}

		default:
			return nil, fmt.Errorf("audit action: %s not support", operation.Action)
		}
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	if len(lblAssOperations) != 0 {
		audit, err := c.listenerAssOperationAuditBuild(kt, lblAssOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	return audits, nil
}

func (c *LoadBalancer) listenerAssOperationAuditBuild(kt *kit.Kit,
	operations []protoaudit.CloudResourceOperationInfo) ([]*tableaudit.AuditTable, error) {

	tgIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	for _, one := range operations {
		tgIDs = append(tgIDs, one.ResID)
		lblIDs = append(lblIDs, one.AssociatedResID)
	}

	tgIDMap, err := ListTargetGroup(kt, c.dao, tgIDs)
	if err != nil {
		return nil, err
	}

	lblIDMap, err := ListListener(kt, c.dao, lblIDs)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	for _, one := range operations {
		tgInfo, exist := tgIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "target group: %s not found", one.ResID)
		}

		lblInfo, exist := lblIDMap[one.AssociatedResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "listener: %s not found", one.AssociatedResID)
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      tgInfo.ID,
			CloudResID: tgInfo.CloudID,
			ResName:    tgInfo.Name,
			ResType:    enumor.TargetGroupAuditResType,
			Action:     action,
			BkBizID:    tgInfo.BkBizID,
			Vendor:     tgInfo.Vendor,
			AccountID:  tgInfo.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.AssociatedOperationAudit{
					AssResType:    enumor.ListenerAuditResType,
					AssResID:      lblInfo.ID,
					AssResCloudID: lblInfo.CloudID,
					AssResName:    lblInfo.Name,
				},
			},
		})
	}

	return audits, nil
}
