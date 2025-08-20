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
	protocloud "hcm/pkg/api/data-service/cloud"
	protoni "hcm/pkg/api/data-service/cloud/network-interface"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

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

// AzureSGDisassociateSubnet ...
func (g *securityGroup) AzureSGDisassociateSubnet(cts *rest.Contexts) (interface{}, error) {
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
	if err = client.SGNetworkInterfaceAssociate(cts.Kit, opt); err != nil {
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
	if err = client.SGNetworkInterfaceDisassociate(cts.Kit, opt); err != nil {
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
