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
	proto "hcm/pkg/api/cloud-server"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// OverwriteSecurityGroupRule overwrite security group rule.
func (svc *securityGroupSvc) OverwriteSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
	return svc.overwriteSGRule(cts, handler.ResOperateAuth)
}

// OverwriteBizSGRule overwrite biz security group rule.
func (svc *securityGroupSvc) OverwriteBizSGRule(cts *rest.Contexts) (interface{}, error) {
	return svc.overwriteSGRule(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) overwriteSGRule(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	sgBaseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		logs.Errorf("get security group basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroupRule,
		Action: meta.Update, BasicInfo: sgBaseInfo})
	if err != nil {
		logs.Errorf("validate biz and authorize failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if sgBaseInfo.Vendor != vendor {
		logs.Errorf("security group vendor: %s not match, request vendor: %s, rid: %s",
			sgBaseInfo.Vendor, vendor, cts.Kit.Rid)
		return nil, errf.Newf(errf.InvalidParameter, "security group vendor: %s not match, request vendor: %s",
			sgBaseInfo.Vendor, vendor)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.overwriteTCloudSGRule(cts, sgBaseInfo)
	default:
		return nil, errf.Newf(errf.Aborted, "vendor: %s not support", vendor)
	}
}

func (svc *securityGroupSvc) overwriteTCloudSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
	interface{}, error) {

	req := new(proto.TCloudSGRuleOverwriteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode overwrite tcloud security group rule request failed, err: %s, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("validate overwrite tcloud security group rule request failed, err: %s, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = svc.audit.ResUpdateAudit(cts.Kit, enumor.SecurityGroupAuditResType, sgBaseInfo.ID, updateFields)
	if err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	overwriteReq := &hcproto.TCloudSGRuleOverwriteReq{
		AccountID:      sgBaseInfo.AccountID,
		EgressRuleSet:  convertToHCCreateSpec(req.EgressRuleSet),
		IngressRuleSet: convertToHCCreateSpec(req.IngressRuleSet),
	}
	if err := svc.client.HCService().TCloud.SecurityGroup.OverwriteSecurityGroupRule(
		cts.Kit, sgBaseInfo.ID, overwriteReq); err != nil {
		logs.Errorf("request hc to overwrite tcloud security group rule failed, err: %s, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func convertToHCCreateSpec(rules []proto.TCloudSecurityGroupRule) []hcproto.TCloudSGRuleCreate {
	result := make([]hcproto.TCloudSGRuleCreate, 0, len(rules))
	for _, r := range rules {
		result = append(result, hcproto.TCloudSGRuleCreate{
			Protocol:                   r.Protocol,
			Port:                       r.Port,
			CloudServiceID:             r.CloudServiceID,
			CloudServiceGroupID:        r.CloudServiceGroupID,
			IPv4Cidr:                   r.IPv4Cidr,
			IPv6Cidr:                   r.IPv6Cidr,
			CloudAddressID:             r.CloudAddressID,
			CloudAddressGroupID:        r.CloudAddressGroupID,
			CloudTargetSecurityGroupID: r.CloudTargetSecurityGroupID,
			Action:                     r.Action,
			Memo:                       r.Memo,
		})
	}
	return result
}
