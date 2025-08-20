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

package aws

import (
	"fmt"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// ListEip ...
// reference: https://docs.amazonaws.cn/en_us/AWSEC2/latest/APIReference/API_DescribeAddresses.html
func (a *Aws) ListEip(kt *kit.Kit, opt *eip.AwsEipListOption) (*eip.AwsEipListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.DescribeAddressesInput{}

	if len(opt.Ips) > 0 {
		req.Filters = []*ec2.Filter{
			{
				Name:   converter.ValToPtr("public-ip"),
				Values: converter.SliceToPtr(opt.Ips),
			},
		}
	}

	if len(opt.CloudIDs) > 0 {
		req.Filters = []*ec2.Filter{
			{
				Name:   converter.ValToPtr("allocation-id"),
				Values: converter.SliceToPtr(opt.CloudIDs),
			},
		}
	}

	resp, err := client.DescribeAddressesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list aws cloud eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list aws cloud eip failed, err: %v", err)
	}

	eips := make([]*eip.AwsEip, len(resp.Addresses))
	for idx, address := range resp.Addresses {
		// aws的eip没有状态信息，通过判断构造状态值
		status := enumor.EipUnBind
		if address.InstanceId != nil || address.NetworkInterfaceId != nil {
			status = enumor.EipBind
		}
		eips[idx] = &eip.AwsEip{
			CloudID:            *address.AllocationId,
			InstanceId:         address.InstanceId,
			Region:             opt.Region,
			Status:             converter.ValToPtr(string(status)),
			PublicIp:           address.PublicIp,
			PrivateIp:          address.PrivateIpAddress,
			PublicIpv4Pool:     address.PublicIpv4Pool,
			Domain:             address.Domain,
			AssociationId:      address.AssociationId,
			PrivateIpAddress:   address.PrivateIpAddress,
			NetworkBorderGroup: address.NetworkBorderGroup,
			NetworkInterfaceId: address.NetworkInterfaceId,
		}
	}

	return &eip.AwsEipListResult{Details: eips}, nil
}

// CountEip 返回给定地域下所有EIP数量，基于DescribeAddresses接口
// reference: https://docs.amazonaws.cn/en_us/AWSEC2/latest/APIReference/API_DescribeAddresses.html
func (a *Aws) CountEip(kt *kit.Kit, region string) (int32, error) {
	client, err := a.clientSet.ec2Client(region)
	if err != nil {
		return 0, err
	}

	req := new(ec2.DescribeAddressesInput)
	resp, err := client.DescribeAddressesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count aws eip failed, err: %v, region:%s, rid: %s", err, region, kt.Rid)
		return 0, err
	}
	return int32(len(resp.Addresses)), nil
}

// DeleteEip ...
// reference: https://docs.amazonaws.cn/en_us/AWSEC2/latest/APIReference/API_ReleaseAddress.html
func (a *Aws) DeleteEip(kt *kit.Kit, opt *eip.AwsEipDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "aws eip delete option is required")
	}

	req, err := opt.ToReleaseAddressInput()
	if err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.ReleaseAddressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("release aws eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// AssociateEip ...
// reference: https://docs.amazonaws.cn/en_us/AWSEC2/latest/APIReference/API_AssociateAddress.html
func (a *Aws) AssociateEip(kt *kit.Kit, opt *eip.AwsEipAssociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "aws eip associate option is required")
	}

	req, err := opt.ToAssociateAddressInput()
	if err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.AssociateAddressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("associate aws eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*Aws, []*eip.AwsEip,
		poller.BaseDoneResult]{Handler: &associateEipPollingHandler{region: opt.Region}}
	_, err = respPoller.PollUntilDone(a, kt, []*string{&opt.PublicIp}, nil)
	if err != nil {
		return err
	}

	return nil
}

// DisassociateEip ...
// reference: https://docs.amazonaws.cn/en_us/AWSEC2/latest/APIReference/API_DisassociateAddress.html
func (a *Aws) DisassociateEip(kt *kit.Kit, opt *eip.AwsEipDisassociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "aws eip disassociate option is required")
	}

	req, err := opt.ToDisassociateAddressInput()
	if err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	result, err := a.ListEip(kt, &eip.AwsEipListOption{Region: opt.Region, Ips: []string{opt.PublicIp}})
	if err != nil {
		return err
	}

	req.AssociationId = result.Details[0].AssociationId

	_, err = client.DisassociateAddressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("disassociate aws eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*Aws, []*eip.AwsEip,
		poller.BaseDoneResult]{Handler: &disassociateEipPollingHandler{region: opt.Region}}
	_, err = respPoller.PollUntilDone(a, kt, []*string{&opt.PublicIp}, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateEip ...
// reference: https://docs.amazonaws.cn/en_us/AWSEC2/latest/APIReference/API_AllocateAddress.html
func (a *Aws) CreateEip(kt *kit.Kit, opt *eip.AwsEipCreateOption) (*string, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "aws eip create option is required")
	}

	req, err := opt.ToAllocateAddressInput()
	if err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.AllocateAddressWithContext(kt.Ctx, req)
	if err != nil {
		return nil, err
	}

	respPoller := poller.Poller[*Aws, []*eip.AwsEip,
		poller.BaseDoneResult]{Handler: &createEipPollingHandler{region: opt.Region}}
	_, err = respPoller.PollUntilDone(a, kt, []*string{resp.PublicIp}, nil)
	if err != nil {
		return nil, err
	}

	return resp.AllocationId, err
}

type createEipPollingHandler struct {
	region string
}

// Done ...
func (h *createEipPollingHandler) Done(pollResult []*eip.AwsEip) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if r.CloudID == "" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *createEipPollingHandler) Poll(client *Aws, kt *kit.Kit, ips []*string) ([]*eip.AwsEip, error) {
	if len(ips) != 1 {
		return nil, fmt.Errorf("poll only support one ip param, but get %v. rid: %s", ips, kt.Rid)
	}

	result, err := client.ListEip(kt, &eip.AwsEipListOption{Region: h.region, Ips: converter.PtrToSlice(ips)})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

type associateEipPollingHandler struct {
	region string
}

// Done ...
func (h *associateEipPollingHandler) Done(pollResult []*eip.AwsEip) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if converter.PtrToVal(r.InstanceId) == "" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *associateEipPollingHandler) Poll(client *Aws, kt *kit.Kit, ips []*string) ([]*eip.AwsEip, error) {
	if len(ips) != 1 {
		return nil, fmt.Errorf("poll only support one ip param, but get %v. rid: %s", ips, kt.Rid)
	}

	result, err := client.ListEip(kt, &eip.AwsEipListOption{Region: h.region, Ips: converter.PtrToSlice(ips)})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

type disassociateEipPollingHandler struct {
	region string
}

// Done ...
func (h *disassociateEipPollingHandler) Done(pollResult []*eip.AwsEip) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if converter.PtrToVal(r.InstanceId) != "" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *disassociateEipPollingHandler) Poll(client *Aws, kt *kit.Kit, ips []*string) ([]*eip.AwsEip, error) {
	if len(ips) != 1 {
		return nil, fmt.Errorf("poll only support one ip param, but get %v. rid: %s", ips, kt.Rid)
	}

	result, err := client.ListEip(kt, &eip.AwsEipListOption{Region: h.region, Ips: converter.PtrToSlice(ips)})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}
