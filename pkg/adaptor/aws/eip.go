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

	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
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

	req, err := opt.ToDescribeAddressesInput()
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeAddressesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list aws cloud eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list aws cloud eip failed, err: %v", err)
	}

	eips := make([]*eip.AwsEip, len(resp.Addresses))
	for idx, address := range resp.Addresses {
		eips[idx] = &eip.AwsEip{
			CloudID:        *address.AllocationId,
			InstanceId:     address.InstanceId,
			Region:         opt.Region,
			PublicIp:       address.PublicIp,
			PrivateIp:      address.PrivateIpAddress,
			PublicIpv4Pool: address.PublicIpv4Pool,
			Domain:         address.Domain,
		}
	}

	return &eip.AwsEipListResult{Details: eips}, nil
}
