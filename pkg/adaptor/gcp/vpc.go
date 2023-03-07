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

package gcp

import (
	"strconv"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// UpdateVpc update vpc.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/networks/patch
// TODO right now only memo is supported to update, but gcp description can not be updated.
func (g *Gcp) UpdateVpc(kt *kit.Kit, opt *types.GcpVpcUpdateOption) error {
	return nil
}

// DeleteVpc delete vpc.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/networks/delete
func (g *Gcp) DeleteVpc(kt *kit.Kit, opt *core.BaseDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID
	_, err = client.Networks.Delete(cloudProjectID, opt.ResourceID).Context(kt.Ctx).RequestId(kt.Rid).Do()
	if err != nil {
		logs.Errorf("delete vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListVpc list vpc.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/networks/list
func (g *Gcp) ListVpc(kt *kit.Kit, opt *types.GcpListOption) (*types.GcpVpcListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID

	listCall := client.Networks.List(cloudProjectID).Context(kt.Ctx)

	if len(opt.CloudIDs) > 0 {
		listCall.Filter(generateResourceIDsFilter(opt.CloudIDs))
	}

	if len(opt.SelfLinks) > 0 {
		listCall.Filter(generateResourceFilter("selfLink", opt.SelfLinks))
	}

	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := listCall.Do()
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]types.GcpVpc, 0, len(resp.Items))
	for _, item := range resp.Items {
		details = append(details, converter.PtrToVal(convertVpc(item)))
	}

	return &types.GcpVpcListResult{NextPageToken: resp.NextPageToken, Details: details}, nil
}

func convertVpc(data *compute.Network) *types.GcpVpc {
	if data == nil {
		return nil
	}

	vpc := &types.GcpVpc{
		CloudID: strconv.FormatUint(data.Id, 10),
		Name:    data.Name,
		Memo:    &data.Description,
		Extension: &cloud.GcpVpcExtension{
			SelfLink:              data.SelfLink,
			AutoCreateSubnetworks: data.AutoCreateSubnetworks,
			EnableUlaInternalIpv6: data.EnableUlaInternalIpv6,
			Mtu:                   data.Mtu,
			RoutingMode:           converter.PtrToVal(data.RoutingConfig).RoutingMode,
		},
	}

	return vpc
}
