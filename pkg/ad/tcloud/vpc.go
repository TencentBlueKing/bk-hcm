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
	"strconv"

	"hcm/pkg/ad/poller"
	"hcm/pkg/ad/provider"
	tcloudtypes "hcm/pkg/ad/tcloud/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// CreateVpc create vpc.
func (tcloud *TCloud) CreateVpc(kt *kit.Kit, meta *provider.Vpc, opt *provider.VpcCreateOption) (*provider.Vpc, error) {

	vpcClient, err := tcloud.clientSet.vpcClient(meta.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	data, err := tcloudtypes.ParseProviderVpc(meta)
	if err != nil {
		logs.Errorf("parse provider vpc failed, err: %v, vpc: %v, rid: %s", err, meta, kt.Rid)
		return nil, err
	}

	if err = data.CreateValidate(); err != nil {
		return nil, fmt.Errorf("vpc create validate failed, err: %v", err)
	}

	req := vpc.NewCreateVpcRequest()
	req.VpcName = converter.ValToPtr(data.Name)
	req.CidrBlock = converter.ValToPtr(data.IPv4Cidr)

	resp, err := vpcClient.CreateVpcWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	handler := &createVpcPollingHandler{}
	respPoller := poller.Poller[*TCloud, []provider.Vpc, []provider.Vpc]{Handler: handler}
	cloudVpc, err := respPoller.PollUntilDone(kt, tcloud, []*string{resp.Response.Vpc.VpcId},
		poller.NewBatchCreateVpcPollerOption())
	if err != nil {
		logs.Errorf("poll until done create vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(*cloudVpc) == 0 {
		return nil, fmt.Errorf("create vpc return vpc is empty")
	}

	vpcs := *cloudVpc
	return converter.ValToPtr(vpcs[0]), nil
}

type createVpcPollingHandler struct {
	region string
}

// Done ...
func (h *createVpcPollingHandler) Done(vpcs []provider.Vpc) (bool, *[]provider.Vpc) {
	return true, &vpcs
}

// Poll ...
func (h *createVpcPollingHandler) Poll(kt *kit.Kit, client *TCloud, cloudIDs []*string) ([]provider.Vpc, error) {
	cloudIDSplit := slice.Split(cloudIDs, tcloudtypes.QueryMaxLimit)

	items := make([]provider.Vpc, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		listOpt := &provider.VpcListOption{
			CloudIDs: converter.PtrToSlice(partIDs),
			Page: &provider.Page{
				Offset: converter.ValToPtr(int64(0)),
				Limit:  converter.ValToPtr(int64(tcloudtypes.QueryMaxLimit)),
			},
		}
		listResp, err := client.ListVpc(kt, listOpt)
		if err != nil {
			return nil, err
		}

		items = append(items, listResp.Items...)
	}

	if len(items) != len(cloudIDs) {
		return nil, fmt.Errorf("query vpc count: %d not equal return count: %d", len(cloudIDs), len(items))
	}

	return items, nil
}

// UpdateVpc update vpc.
func (tcloud *TCloud) UpdateVpc(kt *kit.Kit, meta *provider.Vpc,
	opt *provider.VpcCreateOption) (*provider.Vpc, error) {

	return nil, provider.ErrApiNotSupport
}

// DeleteVpc delete vpc.
func (tcloud *TCloud) DeleteVpc(kt *kit.Kit, meta *provider.Vpc, opt *provider.VpcDeleteOption) error {

	vpcClient, err := tcloud.clientSet.vpcClient(meta.Name)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteVpcRequest()
	req.VpcId = converter.ValToPtr(meta.CloudID)

	_, err = vpcClient.DeleteVpcWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListVpc list vpc.
func (tcloud *TCloud) ListVpc(kt *kit.Kit, opt *provider.VpcListOption) (*provider.VpcListResult, error) {

	vpcClient, err := tcloud.clientSet.vpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	page := tcloudtypes.ParsePage(opt.Page)
	if err = page.Validate(); err != nil {
		return nil, err
	}

	req := vpc.NewDescribeVpcsRequest()
	req.Offset = converter.ValToPtr(strconv.FormatInt(page.Offset, 10))
	req.Limit = converter.ValToPtr(strconv.FormatInt(page.Limit, 10))

	if len(opt.CloudIDs) != 0 {
		req.VpcIds = converter.SliceToPtr(opt.CloudIDs)
	}

	resp, err := vpcClient.DescribeVpcsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list vpc failed, err: %v", err)
	}

	nextPage := &tcloudtypes.Page{
		Offset: page.Offset + page.Limit,
		Limit:  page.Limit,
	}

	vpcs := make([]provider.Vpc, 0, len(resp.Response.VpcSet))
	for _, one := range resp.Response.VpcSet {
		tmp := tcloudtypes.ParseCloudVpc(one, opt.Region)
		providerVpc, err := tmp.ConvProviderVpc()
		if err != nil {
			logs.Errorf("conv vpc to provider vpc failed, err: %v, vpc: %s, rid: %s", err, tmp, kt.Rid)
			return nil, err
		}

		vpcs = append(vpcs, *providerVpc)
	}

	return &provider.VpcListResult{
		Items:    vpcs,
		NextPage: nextPage.ConvProviderPage(),
	}, nil
}
