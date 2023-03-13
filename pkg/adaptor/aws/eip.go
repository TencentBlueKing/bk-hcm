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
	"strings"

	"hcm/pkg/adaptor/types/eip"
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
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return new(eip.AwsEipListResult), nil
		}
		logs.Errorf("list aws cloud eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list aws cloud eip failed, err: %v", err)
	}

	eips := make([]*eip.AwsEip, len(resp.Addresses))
	for idx, address := range resp.Addresses {
		eips[idx] = &eip.AwsEip{
			CloudID:        *address.AllocationId,
			InstanceId:     address.InstanceId,
			Region:         opt.Region,
			Status:         converter.ValToPtr("todo"),
			PublicIp:       address.PublicIp,
			PrivateIp:      address.PrivateIpAddress,
			PublicIpv4Pool: address.PublicIpv4Pool,
			Domain:         address.Domain,
		}
	}

	return &eip.AwsEipListResult{Details: eips}, nil
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

	_, err = client.DisassociateAddressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("disassociate aws eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
