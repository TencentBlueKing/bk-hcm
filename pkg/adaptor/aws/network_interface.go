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
	types "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// DescribeNetworkInterfaces ...
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_DescribeNetworkInterfaces.html
func (a *Aws) DescribeNetworkInterfaces(kt *kit.Kit, opt *types.AwsNetworkInterfaceListOption) (
	*types.AwsNetworkInterfaceWithCountResp, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.DescribeNetworkInterfacesInput{
		Filters: opt.Filters,
	}
	if opt.Page != nil {
		req.MaxResults = opt.Page.MaxResults
		req.NextToken = opt.Page.NextToken
	}

	resp, err := client.DescribeNetworkInterfacesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("describe network interfaces failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	networkInterfaces := make([]types.AwsNetworkInterface, 0, len(resp.NetworkInterfaces))
	for _, one := range resp.NetworkInterfaces {
		networkInterfaces = append(networkInterfaces, types.AwsNetworkInterface{NetworkInterface: one})
	}

	return &types.AwsNetworkInterfaceWithCountResp{
		NextToken: resp.NextToken,
		Details:   networkInterfaces,
	}, nil
}
