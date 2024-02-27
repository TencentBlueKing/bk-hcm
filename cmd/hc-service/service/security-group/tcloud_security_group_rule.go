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

	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	securitygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// BatchCreateTCloudSGRule batch create tcloud security group rule.
// 腾讯云安全组规则索引是一个动态的，所以每次创建需要将云上安全组规则计算一遍。
func (g *securityGroup) BatchCreateTCloudSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(hcservice.TCloudSGRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, sgID,
			cts.Kit.Rid)
		return nil, err
	}
	if sg.AccountID != req.AccountID {
		return nil, fmt.Errorf("'%s' security group does not belong to '%s' account", sgID, req.AccountID)
	}

	client, err := g.ad.TCloud(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	syncParam := &synctcloud.SyncBaseParams{AccountID: sg.AccountID, Region: sg.Region, CloudIDs: []string{sg.ID}}
	opt := &securitygrouprule.TCloudCreateOption{Region: sg.Region, CloudSecurityGroupID: sg.CloudID}
	if req.EgressRuleSet != nil {
		opt.EgressRuleSet = make([]securitygrouprule.TCloud, 0, len(req.EgressRuleSet))
		for _, rule := range req.EgressRuleSet {
			opt.EgressRuleSet = append(opt.EgressRuleSet, securitygrouprule.TCloud{
				Protocol:                   rule.Protocol,
				Port:                       rule.Port,
				CloudServiceID:             rule.CloudServiceID,
				CloudServiceGroupID:        rule.CloudServiceGroupID,
				IPv4Cidr:                   rule.IPv4Cidr,
				IPv6Cidr:                   rule.IPv6Cidr,
				CloudAddressID:             rule.CloudAddressID,
				CloudAddressGroupID:        rule.CloudAddressGroupID,
				CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
				Action:                     rule.Action,
				Description:                rule.Memo,
			})
		}
	}

	if req.IngressRuleSet != nil {
		opt.IngressRuleSet = make([]securitygrouprule.TCloud, 0, len(req.IngressRuleSet))
		for _, rule := range req.IngressRuleSet {
			opt.IngressRuleSet = append(opt.IngressRuleSet, securitygrouprule.TCloud{
				Protocol:                   rule.Protocol,
				Port:                       rule.Port,
				CloudServiceID:             rule.CloudServiceID,
				CloudServiceGroupID:        rule.CloudServiceGroupID,
				IPv4Cidr:                   rule.IPv4Cidr,
				IPv6Cidr:                   rule.IPv6Cidr,
				CloudAddressID:             rule.CloudAddressID,
				CloudAddressGroupID:        rule.CloudAddressGroupID,
				CloudTargetSecurityGroupID: rule.CloudTargetSecurityGroupID,
				Action:                     rule.Action,
				Description:                rule.Memo,
			})
		}
	}
	if err = client.CreateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to create tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		// 里面已经有日志了，不处理
		_, _ = g.syncSGRule(cts.Kit, syncParam)
		return nil, err
	}

	createdIds, syncErr := g.syncSGRule(cts.Kit, syncParam)
	if syncErr != nil {
		return nil, syncErr
	}
	return &core.BatchCreateResult{IDs: createdIds}, nil
}

// syncSGRule 调用同步客户端同步云上规则，返回新增的id
func (g *securityGroup) syncSGRule(kt *kit.Kit, syncParams *synctcloud.SyncBaseParams) ([]string, error) {

	syncCli, err := g.syncCli.TCloud(kt, syncParams.AccountID)
	if err != nil {
		return nil, err
	}

	syncResult, syncErr := syncCli.SecurityGroupRule(kt, syncParams, new(synctcloud.SyncSGRuleOption))
	if syncErr != nil {
		logs.Errorf("sync tcloud security group failed, err: %v, params: %+v, rid: %s", err, syncParams, kt.Rid)
		return nil, syncErr
	}
	return syncResult.CreatedIds, nil
}

// UpdateTCloudSGRule update tcloud security group rule.
func (g *securityGroup) UpdateTCloudSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hcservice.TCloudSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rule, err := g.getTCloudSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	syncParam := &synctcloud.SyncBaseParams{AccountID: rule.AccountID, Region: rule.Region,
		CloudIDs: []string{rule.SecurityGroupID},
	}
	opt := &securitygrouprule.TCloudUpdateOption{Region: rule.Region, CloudSecurityGroupID: rule.CloudSecurityGroupID,
		Version: rule.Version,
	}
	switch rule.Type {
	case enumor.Egress:
		opt.EgressRuleSet = []securitygrouprule.TCloudUpdateSpec{{
			CloudPolicyIndex:           rule.CloudPolicyIndex,
			Protocol:                   req.Protocol,
			Port:                       req.Port,
			CloudServiceID:             req.CloudServiceID,
			CloudServiceGroupID:        req.CloudServiceGroupID,
			IPv4Cidr:                   req.IPv4Cidr,
			IPv6Cidr:                   req.IPv6Cidr,
			CloudAddressID:             req.CloudAddressID,
			CloudAddressGroupID:        req.CloudAddressGroupID,
			CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
			Action:                     req.Action,
			Description:                req.Memo,
		}}
	case enumor.Ingress:
		opt.IngressRuleSet = []securitygrouprule.TCloudUpdateSpec{{
			CloudPolicyIndex:           rule.CloudPolicyIndex,
			Protocol:                   req.Protocol,
			Port:                       req.Port,
			CloudServiceID:             req.CloudServiceID,
			CloudServiceGroupID:        req.CloudServiceGroupID,
			IPv4Cidr:                   req.IPv4Cidr,
			IPv6Cidr:                   req.IPv6Cidr,
			CloudAddressID:             req.CloudAddressID,
			CloudAddressGroupID:        req.CloudAddressGroupID,
			CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
			Action:                     req.Action,
			Description:                req.Memo,
		}}
	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Type)
	}

	if err = client.UpdateSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to update tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		_, _ = g.syncSGRule(cts.Kit, syncParam)
		return nil, err
	}
	if _, syncErr := g.syncSGRule(cts.Kit, syncParam); syncErr != nil {
		return nil, syncErr
	}
	return nil, nil
}

func (g *securityGroup) getTCloudSGRuleByID(cts *rest.Contexts, id string, sgID string) (*corecloud.
	TCloudSecurityGroupRule, error) {

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := g.dataCli.TCloud.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(), listReq, sgID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "security group rule: %s not found", id)
	}

	return &listResp.Details[0], nil
}

// DeleteTCloudSGRule delete tcloud security group rule.
func (g *securityGroup) DeleteTCloudSGRule(cts *rest.Contexts) (interface{}, error) {
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	rule, err := g.getTCloudSGRuleByID(cts, id, sgID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.TCloud(cts.Kit, rule.AccountID)
	if err != nil {
		return nil, err
	}

	syncParam := &synctcloud.SyncBaseParams{
		AccountID: rule.AccountID,
		Region:    rule.Region,
		CloudIDs:  []string{rule.SecurityGroupID},
	}
	opt := &securitygrouprule.TCloudDeleteOption{
		Region:               rule.Region,
		CloudSecurityGroupID: rule.CloudSecurityGroupID,
		Version:              rule.Version,
	}
	switch rule.Type {
	case enumor.Egress:
		opt.EgressRuleIndexes = []int64{rule.CloudPolicyIndex}

	case enumor.Ingress:
		opt.IngressRuleIndexes = []int64{rule.CloudPolicyIndex}

	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", rule.Type)
	}
	if err := client.DeleteSecurityGroupRule(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete tcloud security group rule failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)

		_, _ = g.syncSGRule(cts.Kit, syncParam)
		return nil, err
	}

	if _, syncErr := g.syncSGRule(cts.Kit, syncParam); syncErr != nil {
		return nil, syncErr
	}

	return nil, nil
}
