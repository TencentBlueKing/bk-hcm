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
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	protoni "hcm/pkg/api/data-service/cloud/network-interface"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// CreateAzureSecurityGroup create azure security group.
func (g *securityGroup) CreateAzureSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureSecurityGroupCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureOption{
		ResourceGroupName: req.ResourceGroupName,
		Region:            req.Region,
		Name:              req.Name,
	}
	sg, err := client.CreateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to create azure security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AzureSecurityGroupExtension]{
			{
				CloudID:   *sg.ID,
				BkBizID:   req.BkBizID,
				Region:    req.Region,
				Name:      *sg.Name,
				Memo:      req.Memo,
				AccountID: req.AccountID,
				Extension: &corecloud.AzureSecurityGroupExtension{
					ResourceGroupName: req.ResourceGroupName,
					Etag:              sg.Etag,
					FlushConnection:   sg.FlushConnection,
					ResourceGUID:      sg.ResourceGUID,
				},
				// Tags:        core.NewTagMap(req.Tags...),
				Manager:     req.Manager,
				BakManager:  req.BakManager,
				UsageBizIds: req.UsageBizIds},
		},
	}
	result, err := g.dataCli.Azure.SecurityGroup.BatchCreateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: result.IDs[0]}, nil
}

// DeleteAzureSecurityGroup delete azure security group.
func (g *securityGroup) DeleteAzureSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get azure security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureOption{
		ResourceGroupName: sg.Extension.ResourceGroupName,
		Region:            sg.Region,
		Name:              sg.Name,
	}
	if err := client.DeleteSecurityGroup(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete azure security group failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	req := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err := g.dataCli.Global.SecurityGroup.BatchDeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req); err != nil {
		logs.Errorf("request dataservice delete azure security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UpdateAzureSecurityGroup update azure security group.
func (g *securityGroup) UpdateAzureSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.AzureSecurityGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.AzureSecurityGroupExtension]{
			{
				ID:   id,
				Memo: req.Memo,
			},
		},
	}
	if err := g.dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		updateReq); err != nil {

		logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AzureSecurityGroupAssociateSubnet ...
func (g *securityGroup) AzureSecurityGroupAssociateSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureSecurityGroupAssociateSubnetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req.SecurityGroupID)
	if err != nil {
		return nil, err
	}

	subnet, err := g.dataCli.Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), req.SubnetID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureAssociateSubnetOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
		CloudVpcID:           subnet.CloudVpcID,
		CloudSubnetID:        subnet.CloudID,
	}
	if err = client.SecurityGroupSubnetAssociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to azure security group associate subnet failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.SubnetBatchUpdateReq[protocloud.AzureSubnetUpdateExt]{
		Subnets: []protocloud.SubnetUpdateReq[protocloud.AzureSubnetUpdateExt]{
			{
				ID: req.SubnetID,
				Extension: &protocloud.AzureSubnetUpdateExt{
					CloudSecurityGroupID: converter.ValToPtr(opt.CloudSecurityGroupID),
					SecurityGroupID:      converter.ValToPtr(req.SecurityGroupID),
				},
			},
		},
	}
	if err = g.dataCli.Azure.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
		logs.Errorf("request dataservice update subnet failed, err: %v, req: %+v, rid: %s", err, updateReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AzureSecurityGroupDisassociateSubnet ...
func (g *securityGroup) AzureSecurityGroupDisassociateSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureSecurityGroupAssociateSubnetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req.SecurityGroupID)
	if err != nil {
		return nil, err
	}

	subnet, err := g.dataCli.Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), req.SubnetID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureAssociateSubnetOption{
		Region:               sg.Region,
		CloudSecurityGroupID: sg.CloudID,
		ResourceGroupName:    sg.Extension.ResourceGroupName,
		CloudVpcID:           subnet.CloudVpcID,
		CloudSubnetID:        subnet.CloudID,
	}
	if err = client.SecurityGroupSubnetDisassociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to azure security group disassociate subnet failed, err: %v, opt: %v, rid: %s",
			err, opt, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protocloud.SubnetBatchUpdateReq[protocloud.AzureSubnetUpdateExt]{
		Subnets: []protocloud.SubnetUpdateReq[protocloud.AzureSubnetUpdateExt]{
			{
				ID: req.SubnetID,
				Extension: &protocloud.AzureSubnetUpdateExt{
					CloudSecurityGroupID: nil,
					SecurityGroupID:      nil,
				},
			},
		},
	}
	if err = g.dataCli.Azure.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
		logs.Errorf("request dataservice update subnet failed, err: %v, req: %+v, rid: %s", err, updateReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AzureSecurityGroupAssociateNI ...
func (g *securityGroup) AzureSecurityGroupAssociateNI(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureSecurityGroupAssociateNIReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req.SecurityGroupID)
	if err != nil {
		return nil, err
	}

	ni, err := g.dataCli.Azure.NetworkInterface.Get(cts.Kit.Ctx, cts.Kit.Header(), req.NetworkInterfaceID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureAssociateNetworkInterfaceOption{
		Region:                  sg.Region,
		CloudSecurityGroupID:    sg.CloudID,
		ResourceGroupName:       sg.Extension.ResourceGroupName,
		CloudNetworkInterfaceID: ni.CloudID,
	}
	if err = client.SecurityGroupNetworkInterfaceAssociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to azure security group associate network interface failed,"+
			" err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protoni.NetworkInterfaceBatchUpdateReq[protoni.AzureNICreateExt]{
		NetworkInterfaces: []protoni.NetworkInterfaceUpdateReq[protoni.AzureNICreateExt]{
			{
				ID: req.NetworkInterfaceID,
				Extension: &protoni.AzureNICreateExt{
					CloudSecurityGroupID: converter.ValToPtr(opt.CloudSecurityGroupID),
					SecurityGroupID:      converter.ValToPtr(req.SecurityGroupID),
				},
			},
		},
	}
	if err = g.dataCli.Azure.NetworkInterface.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
		logs.Errorf("request dataservice update network interface failed, err: %v, req: %+v, rid: %s",
			err, updateReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AzureSecurityGroupDisassociateNI ...
func (g *securityGroup) AzureSecurityGroupDisassociateNI(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureSecurityGroupAssociateNIReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := g.dataCli.Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), req.SecurityGroupID)
	if err != nil {
		return nil, err
	}

	ni, err := g.dataCli.Azure.NetworkInterface.Get(cts.Kit.Ctx, cts.Kit.Header(), req.NetworkInterfaceID)
	if err != nil {
		return nil, err
	}

	client, err := g.ad.Azure(cts.Kit, sg.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &securitygroup.AzureAssociateNetworkInterfaceOption{
		Region:                  sg.Region,
		CloudSecurityGroupID:    sg.CloudID,
		ResourceGroupName:       sg.Extension.ResourceGroupName,
		CloudNetworkInterfaceID: ni.CloudID,
	}
	if err = client.SecurityGroupNetworkInterfaceDisassociate(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to azure security group disassociate network interface failed,"+
			" err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &protoni.NetworkInterfaceBatchUpdateReq[protoni.AzureNICreateExt]{
		NetworkInterfaces: []protoni.NetworkInterfaceUpdateReq[protoni.AzureNICreateExt]{
			{
				ID: req.NetworkInterfaceID,
				Extension: &protoni.AzureNICreateExt{
					CloudSecurityGroupID: nil,
					SecurityGroupID:      nil,
				},
			},
		},
	}
	if err = g.dataCli.Azure.NetworkInterface.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
		logs.Errorf("request dataservice update network interface failed, err: %v, req: %+v, rid: %s",
			err, updateReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
