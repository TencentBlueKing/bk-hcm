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

	"hcm/pkg/adaptor/types"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateAwsSecurityGroup create aws security group.
func (g *securityGroup) CreateAwsSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupCreateReq[proto.AwsSecurityGroupAttachment])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.Aws(cts.Kit, req.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AwsSecurityGroupCreateOption{
		Region:      req.Spec.Region,
		Name:        req.Spec.Name,
		Description: req.Spec.Memo,
		CloudVpcID:  req.Attachment.VpcID,
	}
	cloudID, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create aws security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	listOpt := &types.AwsSecurityGroupListOption{
		Region:   req.Spec.Region,
		CloudIDs: []string{cloudID},
	}
	result, err := client.ListSecurityGroup(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list aws security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	if len(result.SecurityGroups) != 1 {
		logs.Errorf("create aws security group succeeds, but query failed, cloud_id: %s, rid: %s", cloudID, cts.Kit.Rid)
		return nil, fmt.Errorf("create aws security group succeeds, but query failed")
	}

	createReq := &protocloud.SecurityGroupCreateReq[corecloud.AwsSecurityGroupExtension]{
		Spec: &corecloud.SecurityGroupSpec{
			CloudID:   *result.SecurityGroups[0].GroupId,
			Assigned:  false,
			Region:    req.Spec.Region,
			Name:      *result.SecurityGroups[0].GroupName,
			Memo:      result.SecurityGroups[0].Description,
			AccountID: req.Spec.AccountID,
		},
		Extension: &corecloud.AwsSecurityGroupExtension{
			CloudVpcID:   result.SecurityGroups[0].VpcId,
			CloudOwnerID: result.SecurityGroups[0].OwnerId,
		},
	}
	createResp, err := g.dataCli.SecurityGroup().CreateAwsSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return createResp, nil
}

// DeleteAwsSecurityGroup delete aws security group.
func (g *securityGroup) DeleteAwsSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.SecurityGroup().GetAwsSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get aws security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.Aws(cts.Kit, sg.Spec.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.SecurityGroupDeleteOption{
		Region:  sg.Spec.Region,
		CloudID: sg.Spec.CloudID,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete aws security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := g.dataCli.SecurityGroup().Delete(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete aws security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
