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

package huawei

import (
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	v2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// UpdateVpc update vpc.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_api01_0004.html
func (h *HuaWei) UpdateVpc(kt *kit.Kit, opt *types.HuaWeiVpcUpdateOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &v2.UpdateVpcRequest{
		VpcId: opt.ResourceID,
		Body: &v2.UpdateVpcRequestBody{
			Vpc: &v2.UpdateVpcOption{
				Description: opt.Data.Memo,
			},
		},
	}

	_, err = vpcClient.UpdateVpc(req)
	if err != nil {
		logs.Errorf("update huawei vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteVpc delete vpc.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_api01_0005.html
func (h *HuaWei) DeleteVpc(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &v2.DeleteVpcRequest{
		VpcId: opt.ResourceID,
	}

	_, err = vpcClient.DeleteVpc(req)
	if err != nil {
		logs.Errorf("delete huawei vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListVpc list vpc.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_apiv3_0003.html
func (h *HuaWei) ListVpc(kt *kit.Kit, opt *types.HuaWeiVpcListOption) (*types.HuaWeiVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	vpcClient, err := h.clientSet.vpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := new(model.ListVpcsRequest)

	if len(opt.CloudIDs) != 0 {
		req.Id = &opt.CloudIDs
	}

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	if len(opt.Names) != 0 {
		req.Name = &opt.Names
	}

	resp, err := vpcClient.ListVpcs(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return new(types.HuaWeiVpcListResult), nil
		}
		logs.Errorf("list huawei vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list huawei vpc failed, err: %v", err)
	}

	vpcs := converter.PtrToVal(resp.Vpcs)
	details := make([]types.HuaWeiVpc, 0, len(vpcs))

	for _, data := range vpcs {
		details = append(details, converter.PtrToVal(convertVpc(&data, opt.Region)))
	}

	return &types.HuaWeiVpcListResult{NextMarker: converter.PtrToVal(resp.PageInfo).NextMarker, Details: details}, nil
}

func convertVpc(data *model.Vpc, region string) *types.HuaWeiVpc {
	if data == nil {
		return nil
	}

	v := &types.HuaWeiVpc{
		CloudID: data.Id,
		Name:    data.Name,
		Region:  region,
		Memo:    converter.ValToPtr(data.Description),
		Extension: &cloud.HuaWeiVpcExtension{
			Cidr:                nil,
			Status:              data.Status,
			EnterpriseProjectId: data.EnterpriseProjectId,
		},
	}

	if data.Cidr != "" {
		v.Extension.Cidr = append(v.Extension.Cidr, cloud.HuaWeiCidr{
			Type: enumor.Ipv4,
			Cidr: data.Cidr,
		})
	}

	for _, cidr := range data.ExtendCidrs {
		v.Extension.Cidr = append(v.Extension.Cidr, cloud.HuaWeiCidr{
			Type: enumor.Ipv4,
			Cidr: cidr,
		})
	}

	return v
}
