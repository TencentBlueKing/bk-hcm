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

// CloudResourceAssignAudit cloud resource assign audit.
func (ad Audit) CloudResourceAssignAudit(cts *rest.Contexts) (interface{}, error) {
	req := new(protoaudit.CloudResourceAssignAuditReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	auditAll, err := ad.GenCloudResAssignAudit(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	// 审计信息保存
	if err := ad.dao.Audit().BatchCreate(cts.Kit, auditAll); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GenCloudResAssignAudit generate cloud resource assign audit.
func (ad Audit) GenCloudResAssignAudit(kt *kit.Kit, req *protoaudit.CloudResourceAssignAuditReq) (
	[]*tableaudit.AuditTable, error) {

	if req == nil {
		return nil, errf.New(errf.InvalidParameter, "cloud resource assign audit request cannot be empty")
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 按云资源类型进行分类
	assignMap := make(map[enumor.AuditResourceType][]protoaudit.CloudResourceAssignInfo, 0)
	for _, one := range req.Assigns {
		if _, exist := assignMap[one.ResType]; !exist {
			assignMap[one.ResType] = make([]protoaudit.CloudResourceAssignInfo, 0)
		}

		assignMap[one.ResType] = append(assignMap[one.ResType], one)
	}

	// 根据分类后的分配信息，对所需要记录的审计信息进行查询
	auditAll := make([]*tableaudit.AuditTable, 0, len(req.Assigns))
	for resType, assigns := range assignMap {
		audits, err := ad.buildAssignAuditInfo(kt, resType, assigns)
		if err != nil {
			logs.Errorf("query assign audit info failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		auditAll = append(auditAll, audits...)
	}

	return auditAll, nil
}

func (ad Audit) buildAssignAuditInfo(kt *kit.Kit, resType enumor.AuditResourceType,
	assigns []protoaudit.CloudResourceAssignInfo) ([]*tableaudit.AuditTable, error) {
	var audits []*tableaudit.AuditTable
	var err error
	switch resType {
	case enumor.SecurityGroupAuditResType:
		audits, err = ad.securityGroup.SecurityGroupAssignAuditBuild(kt, assigns)
	case enumor.GcpFirewallRuleAuditResType:
		audits, err = ad.firewall.FirewallRuleAssignAuditBuild(kt, assigns)
	case enumor.VpcCloudAuditResType:
		audits, err = ad.vpcAssignAuditBuild(kt, assigns)
	case enumor.SubnetAuditResType:
		audits, err = ad.subnet.SubnetAssignAuditBuild(kt, assigns)
	case enumor.EipAuditResType:
		audits, err = ad.eipAssignAuditBuild(kt, assigns)
	case enumor.DiskAuditResType:
		audits, err = ad.diskAssignAuditBuild(kt, assigns)
	case enumor.CvmAuditResType:
		audits, err = ad.cvm.CvmAssignAuditBuild(kt, assigns)
	case enumor.NetworkInterfaceAuditResType:
		audits, err = ad.networkInterface.NetworkInterfaceAssignAuditBuild(kt, assigns)
	case enumor.RouteTableAuditResType:
		audits, err = ad.routeTable.RouteTableAssignAuditBuild(kt, assigns)
	case enumor.ArgumentTemplateAuditResType:
		audits, err = ad.argsTplAssignAuditBuild(kt, assigns)
	default:
		return nil, fmt.Errorf("cloud resource type: %s not support", resType)
	}
	if err != nil {
		return nil, err
	}

	return audits, nil
}
