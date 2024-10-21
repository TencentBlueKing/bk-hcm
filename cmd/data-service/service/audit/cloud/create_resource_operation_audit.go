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
	"fmt"

	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CloudResourceOperationAudit cloud resource operation audit.
func (ad Audit) CloudResourceOperationAudit(cts *rest.Contexts) (interface{}, error) {
	req := new(protoaudit.CloudResourceOperationAuditReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 按云资源类型进行分类
	operationMap := make(map[enumor.AuditResourceType][]protoaudit.CloudResourceOperationInfo, 0)
	for _, one := range req.Operations {
		if _, exist := operationMap[one.ResType]; !exist {
			operationMap[one.ResType] = make([]protoaudit.CloudResourceOperationInfo, 0)
		}

		operationMap[one.ResType] = append(operationMap[one.ResType], one)
	}

	// 根据分类后的删除信息，对所需要记录的审计信息进行查询
	auditAll := make([]*tableaudit.AuditTable, 0, len(req.Operations))
	for resType, operations := range operationMap {
		audits, err := ad.buildOperationAuditInfo(cts.Kit, resType, operations)
		if err != nil {
			logs.Errorf("query operation audit info failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		auditAll = append(auditAll, audits...)
	}

	// 审计信息保存
	if err := ad.dao.Audit().BatchCreate(cts.Kit, auditAll); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (ad Audit) buildOperationAuditInfo(kt *kit.Kit, resType enumor.AuditResourceType,
	operations []protoaudit.CloudResourceOperationInfo,
) ([]*tableaudit.AuditTable, error) {
	var audits []*tableaudit.AuditTable
	var err error
	switch resType {
	case enumor.CvmAuditResType:
		audits, err = ad.cvm.CvmOperationAuditBuild(kt, operations)
	case enumor.SecurityGroupRuleAuditResType:
		audits, err = ad.securityGroup.OperationAuditBuild(kt, operations)
	case enumor.SecurityGroupAuditResType:
		audits, err = ad.securityGroup.OperationAuditBuild(kt, operations)
	case enumor.EipAuditResType:
		audits, err = ad.eipOperationAuditBuild(kt, operations)
	case enumor.DiskAuditResType:
		audits, err = ad.diskOperationAuditBuild(kt, operations)
	case enumor.TargetGroupAuditResType:
		audits, err = ad.loadBalancer.TargetGroupOperationAuditBuild(kt, operations)
	default:
		return nil, fmt.Errorf("cloud resource type: %s not support", resType)
	}
	if err != nil {
		return nil, err
	}

	return audits, nil
}
