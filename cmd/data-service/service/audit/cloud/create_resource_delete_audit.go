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

// CloudResourceDeleteAudit cloud resource delete audit.
func (ad Audit) CloudResourceDeleteAudit(cts *rest.Contexts) (interface{}, error) {
	req := new(protoaudit.CloudResourceDeleteAuditReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 按云资源类型进行分类
	deleteMap := make(map[enumor.AuditResourceType][]protoaudit.CloudResourceDeleteInfo, 0)
	for _, one := range req.Deletes {
		if _, exist := deleteMap[one.ResType]; !exist {
			deleteMap[one.ResType] = make([]protoaudit.CloudResourceDeleteInfo, 0)
		}

		deleteMap[one.ResType] = append(deleteMap[one.ResType], one)
	}

	// 根据分类后的删除信息，对所需要记录的审计信息进行查询
	auditAll := make([]*tableaudit.AuditTable, 0, len(req.Deletes))
	for resType, deletes := range deleteMap {
		audits, err := ad.buildDeleteAuditInfo(cts.Kit, resType, req.ParentID, deletes)
		if err != nil {
			logs.Errorf("query delete audit info failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

func (ad Audit) buildDeleteAuditInfo(kt *kit.Kit, resType enumor.AuditResourceType, parentID string,
	deletes []protoaudit.CloudResourceDeleteInfo,
) ([]*tableaudit.AuditTable, error) {
	var audits []*tableaudit.AuditTable
	var err error
	switch resType {
	case enumor.SecurityGroupAuditResType:
		audits, err = ad.securityGroup.SecurityGroupDeleteAuditBuild(kt, deletes)
	case enumor.SecurityGroupRuleAuditResType:
		audits, err = ad.securityGroup.SecurityGroupRuleDeleteAuditBuild(kt, parentID, deletes)
	case enumor.GcpFirewallRuleAuditResType:
		audits, err = ad.firewall.FirewallRuleDeleteAuditBuild(kt, deletes)
	case enumor.VpcCloudAuditResType:
		audits, err = ad.vpcDeleteAuditBuild(kt, deletes)
	case enumor.SubnetAuditResType:
		audits, err = ad.subnet.SubnetDeleteAuditBuild(kt, deletes)
	case enumor.CvmAuditResType:
		audits, err = ad.cvm.CvmDeleteAuditBuild(kt, deletes)
	case enumor.EipAuditResType:
		audits, err = ad.eipDeleteAuditBuild(kt, deletes)
	case enumor.DiskAuditResType:
		audits, err = ad.diskDeleteAuditBuild(kt, deletes)
	case enumor.ArgumentTemplateAuditResType:
		audits, err = ad.argsTplDeleteAuditBuild(kt, deletes)
	case enumor.SslCertAuditResType:
		audits, err = ad.certDeleteAuditBuild(kt, deletes)
	case enumor.TargetGroupAuditResType:
		audits, err = ad.targetGroupDeleteAuditBuild(kt, deletes)
	case enumor.UrlRuleAuditResType:
		audits, err = ad.loadBalancer.UrlRuleDeleteAuditBuild(kt, parentID, deletes)
	case enumor.UrlRuleDomainAuditResType:
		audits, err = ad.loadBalancer.UrlRuleDeleteByDomainAuditBuild(kt, parentID, deletes)
	case enumor.ListenerAuditResType:
		audits, err = ad.listenerDeleteAuditBuild(kt, deletes)
	case enumor.LoadBalancerAuditResType:
		audits, err = ad.loadBalancer.LoadBalancerDeleteAuditBuild(kt, deletes)

	default:
		return nil, fmt.Errorf("build delete audit cloud resource type: %s not support", resType)
	}
	if err != nil {
		return nil, err
	}

	return audits, nil
}
