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

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	v2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// CreateVpc create vpc.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_api01_0001.html
// TODO returns created vpc after sdk supports (API docs returns created vpc info)
func (h *HuaWei) CreateVpc(kt *kit.Kit, opt *types.HuaWeiVpcCreateOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	vpcClient, err := h.clientSet.vpcClientV2(opt.Extension.Region)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &v2.CreateVpcRequest{
		Body: &v2.CreateVpcRequestBody{
			Vpc: &v2.CreateVpcOption{
				Name:                &opt.Name,
				Description:         opt.Memo,
				Cidr:                &opt.Extension.IPv4Cidr,
				EnterpriseProjectId: opt.Extension.EnterpriseProjectID,
			},
		},
	}

	resp, err := vpcClient.CreateVpc(req)
	if err != nil {
		logs.Errorf("create huawei vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	handler := &createVpcPollingHandler{
		opt.Extension.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.Vpc, []model.Vpc]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, []*string{converter.ValToPtr(resp.Vpc.Id)},
		types.NewBatchCreateVpcPollerOption())
	if err != nil {
		return err
	}

	return nil
}

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

// ListVpcRaw list vpc raw.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_apiv3_0003.html
func (h *HuaWei) ListVpcRaw(kt *kit.Kit, opt *types.HuaWeiVpcListOption) (*model.ListVpcsResponse, error) {
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
			return nil, nil
		}
		logs.Errorf("list huawei vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list huawei vpc failed, err: %v", err)
	}

	return resp, nil
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

type createVpcPollingHandler struct {
	region string
}

// Done ...
func (h *createVpcPollingHandler) Done(vpcs []model.Vpc) (bool, *[]model.Vpc) {
	results := make([]model.Vpc, 0)

	flag := true
	for _, vpc := range vpcs {
		if vpc.Status == "PENDING" {
			flag = false
			continue
		}

		results = append(results, vpc)
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createVpcPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.Vpc, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.HuaWeiQueryLimit)

	vpcs := make([]model.Vpc, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(model.ListVpcsRequest)
		tmpCloudIDs := converter.PtrToSlice(partIDs)
		req.Id = &tmpCloudIDs
		req.Limit = converter.ValToPtr(int32(core.HuaWeiQueryLimit))

		vpcClient, err := client.clientSet.vpcClient(h.region)
		if err != nil {
			return nil, fmt.Errorf("new vpc client failed, err: %v", err)
		}

		resp, err := vpcClient.ListVpcs(req)
		if err != nil {
			if strings.Contains(err.Error(), ErrDataNotFound) {
				return make([]model.Vpc, 0), nil
			}
			return nil, err
		}

		vpcs = append(vpcs, *resp.Vpcs...)
	}

	if len(vpcs) != len(cloudIDs) {
		return nil, fmt.Errorf("query vpc count: %d not equal return count: %d", len(cloudIDs), len(vpcs))
	}

	return vpcs, nil
}

// ListPorts list ports.
// reference: https://support.huaweicloud.com/api-vpc/vpc_port01_0003.html
// https://support.huaweicloud.com/vpc_faq/faq_security_0007.html
func (h *HuaWei) ListPorts(kt *kit.Kit, opt *types.HuaweiListPortOption) ([]v2.Port, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list port option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client v2 failed, err: %v", err)
	}

	req := new(v2.ListPortsRequest)
	req.SecurityGroups = converter.ValToPtr(opt.SecurityGroupIDs)
	if opt.Marker != "" {
		req.Marker = converter.ValToPtr(opt.Marker)
	}

	resp, err := client.ListPorts(req)
	if err != nil {
		logs.Errorf("list huawei ports failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return converter.PtrToVal(resp.Ports), nil
}
