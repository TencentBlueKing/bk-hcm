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
	"fmt"

	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"
)

// SecurityGroupRuleUpdateAuditBuild build security group rule update audit.
func (s *SecurityGroup) SecurityGroupRuleUpdateAuditBuild(kt *kit.Kit, sgID string,
	updates []protoaudit.CloudResourceUpdateInfo) ([]*tableaudit.AuditTable, error) {

	idSgMap, err := s.listSecurityGroup(kt, []string{sgID})
	if err != nil {
		return nil, err
	}

	sg, exist := idSgMap[sgID]
	if !exist {
		return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", sgID)
	}

	switch sg.Vendor {
	case enumor.TCloud:
		return s.tcloudSGRuleUpdateAuditBuild(kt, sg, updates)
	case enumor.Aws:
		return s.awsSGRuleUpdateAuditBuild(kt, sg, updates)
	case enumor.HuaWei:
		return s.huaWeiSGRuleUpdateAuditBuild(kt, sg, updates)
	case enumor.Azure:
		return s.azureSGRuleUpdateAuditBuild(kt, sg, updates)
	default:
		return nil, fmt.Errorf("vendor: %s not support", sg.Vendor)
	}
}

// SecurityGroupRuleDeleteAuditBuild build security group rule delete audit.
func (s *SecurityGroup) SecurityGroupRuleDeleteAuditBuild(kt *kit.Kit, sgID string,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	idSgMap, err := s.listSecurityGroup(kt, []string{sgID})
	if err != nil {
		return nil, err
	}

	sg, exist := idSgMap[sgID]
	if !exist {
		return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", sgID)
	}

	switch sg.Vendor {
	case enumor.TCloud:
		return s.tcloudSGRuleDeleteAuditBuild(kt, sg, deletes)
	case enumor.Aws:
		return s.awsSGRuleDeleteAuditBuild(kt, sg, deletes)
	case enumor.HuaWei:
		return s.huaWeiSGRuleDeleteAuditBuild(kt, sg, deletes)
	case enumor.Azure:
		return s.azureSGRuleDeleteAuditBuild(kt, sg, deletes)
	default:
		return nil, fmt.Errorf("vendor: %s not support", sg.Vendor)
	}
}
