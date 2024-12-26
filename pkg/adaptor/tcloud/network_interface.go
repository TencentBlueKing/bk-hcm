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

package tcloud

import (
	"fmt"

	networkinterface "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// DescribeNetworkInterfaces 查询弹性网卡列表
// reference: https://cloud.tencent.com/document/product/215/15817
func (t *TCloudImpl) DescribeNetworkInterfaces(kt *kit.Kit,
	opt *networkinterface.TCloudNetworkInterfaceListOption) (
	*networkinterface.TCloudNetworkInterfaceWithCountResp, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	request := vpc.NewDescribeNetworkInterfacesRequest()
	request.Filters = opt.Filters
	request.Offset = common.Uint64Ptr(opt.Page.Offset)
	request.Limit = common.Uint64Ptr(opt.Page.Limit)

	response, err := client.DescribeNetworkInterfacesWithContext(kt.Ctx, request)
	if err != nil {
		logs.Errorf("describe network interfaces failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("describe network interfaces failed, err: %v", err)
	}

	details := make([]networkinterface.TCloudNetworkInterface, 0)
	for _, v := range response.Response.NetworkInterfaceSet {
		details = append(details, networkinterface.TCloudNetworkInterface{NetworkInterface: v})
	}

	return &networkinterface.TCloudNetworkInterfaceWithCountResp{
		TotalCount: converter.PtrToVal(response.Response.TotalCount),
		Details:    details,
	}, nil
}
