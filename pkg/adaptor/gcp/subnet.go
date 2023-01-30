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

// UpdateSubnet update subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/patch
func (g *Gcp) UpdateSubnet(kt *kit.Kit, opt *types.GcpSubnetUpdateOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	req := &compute.Subnetwork{
		Description: converter.PtrToVal(opt.Data.Memo),
		// make sure AutoCreateSubnetworks field is included in request
		// gcp has a bug with this api, if this is not specified, the request will fail
		// TODO 测试一下是否不需要这个了
		ForceSendFields: []string{"AutoCreateSubnetworks"},
		NullFields:      nil,
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID
<<<<<<< HEAD
	_, err = client.Subnetworks.Patch(cloudProjectID, opt.Region, opt.ResourceID, req).Context(kt.Ctx).
=======
	region := parseSelfLinkToName(opt.Region)
	_, err = client.Subnetworks.Patch(cloudProjectID, region, opt.ResourceID, req).Context(kt.Ctx).
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
		RequestId(kt.Rid).Do()
	if err != nil {
		logs.Errorf("create subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteSubnet delete subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/delete
func (g *Gcp) DeleteSubnet(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID
<<<<<<< HEAD
	_, err = client.Subnetworks.Delete(cloudProjectID, opt.Region, opt.ResourceID).Context(kt.Ctx).
=======
	region := parseSelfLinkToName(opt.Region)
	_, err = client.Subnetworks.Delete(cloudProjectID, region, opt.ResourceID).Context(kt.Ctx).
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
		RequestId(kt.Rid).Do()
	if err != nil {
		logs.Errorf("delete subnet failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSubnet list subnet.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/subnetworks/list
func (g *Gcp) ListSubnet(kt *kit.Kit, opt *types.GcpSubnetListOption) (*types.GcpSubnetListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID

	listCall := client.Subnetworks.List(cloudProjectID, opt.Region).Context(kt.Ctx)

	if len(opt.ResourceIDs) > 0 {
		listCall.Filter(generateResourceIDsFilter(opt.ResourceIDs))
	}

	if opt.Page != nil {
		listCall.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := listCall.Do()
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]types.GcpSubnet, 0, len(resp.Items))
	for _, item := range resp.Items {
		details = append(details, converter.PtrToVal(convertSubnet(item)))
	}

	return &types.GcpSubnetListResult{NextPageToken: resp.NextPageToken, Details: details}, nil
}

func convertSubnet(data *compute.Subnetwork) *types.GcpSubnet {
	if data == nil {
		return nil
	}

	subnet := &types.GcpSubnet{
		CloudVpcID: data.Network,
		CloudID:    strconv.FormatUint(data.Id, 10),
		Name:       data.Name,
		Memo:       &data.Description,
		Extension: &cloud.GcpSubnetExtension{
<<<<<<< HEAD
=======
			SelfLink:              data.SelfLink,
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
			Region:                data.Region,
			StackType:             data.StackType,
			Ipv6AccessType:        data.Ipv6AccessType,
			GatewayAddress:        data.GatewayAddress,
			PrivateIpGoogleAccess: data.PrivateIpGoogleAccess,
			EnableFlowLogs:        data.EnableFlowLogs,
		},
	}

	if len(data.IpCidrRange) != 0 {
		subnet.Ipv4Cidr = []string{data.IpCidrRange}
	}

	if len(data.Ipv6CidrRange) != 0 {
		subnet.Ipv6Cidr = []string{data.Ipv6CidrRange}
	}

	return subnet
}
