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
	"errors"
	"fmt"
	"strconv"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"google.golang.org/api/compute/v1"
)

// CreateVpc create vpc.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/networks/insert
func (g *Gcp) CreateVpc(kt *kit.Kit, opt *types.GcpVpcCreateOption) (uint64, error) {
	if err := opt.Validate(); err != nil {
		return 0, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return 0, err
	}

	req := &compute.Network{
		IPv4Range:                             "",
		AutoCreateSubnetworks:                 opt.Extension.AutoCreateSubnetworks,
		Description:                           converter.PtrToVal(opt.Memo),
		EnableUlaInternalIpv6:                 opt.Extension.EnableUlaInternalIpv6,
		InternalIpv6Range:                     opt.Extension.InternalIpv6Range,
		Mtu:                                   0,
		Name:                                  opt.Name,
		NetworkFirewallPolicyEnforcementOrder: "",
		RoutingConfig:                         nil,
		// make sure AutoCreateSubnetworks field is included in request
		// gcp has a bug with this api, if this is not specified, the request will fail
		ForceSendFields: []string{"AutoCreateSubnetworks"},
		NullFields:      nil,
	}

	if opt.Extension.RoutingMode != "" {
		req.RoutingConfig = &compute.NetworkRoutingConfig{
			RoutingMode:     opt.Extension.RoutingMode,
			ForceSendFields: nil,
			NullFields:      nil,
		}
	}

	resp, err := client.Networks.Insert(g.CloudProjectID(), req).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("create vpc failed, err: %v, rid: %s", err, kt.Rid)
		return 0, err
	}

	handler := &createVpcPollingHandler{
		parseSelfLinkToName(resp.Zone),
	}
	respPoller := poller.Poller[*Gcp, []*compute.Operation, []uint64]{Handler: handler}
	results, err := respPoller.PollUntilDone(g, kt, []*string{to.Ptr(resp.OperationGroupId)},
		types.NewBatchCreateVpcPollerOption())
	if err != nil {
		return 0, err
	}

	if len(converter.PtrToVal(results)) <= 0 {
		return 0, fmt.Errorf("create vpc failed")
	}

	return (converter.PtrToVal(results))[0], nil
}

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
	_, err = client.Networks.Delete(cloudProjectID, opt.ResourceID).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("delete vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// CountVpc count vpc.
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/networks/list
func (g *Gcp) CountVpc(kt *kit.Kit) (int32, error) {

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return 0, err
	}

	request := client.Networks.List(g.CloudProjectID()).Context(kt.Ctx)

	var count int32
	for {
		resp, err := request.Do()
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, kt.Rid)
			return 0, err
		}

		count += int32(len(resp.Items))

		if resp.NextPageToken == "" {
			break
		}
	}

	return count, nil
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
			InternalIpv6Range:     data.InternalIpv6Range,
			Mtu:                   data.Mtu,
			RoutingMode:           converter.PtrToVal(data.RoutingConfig).RoutingMode,
		},
	}

	return vpc
}

type createVpcPollingHandler struct {
	zone string
}

// Done ...
func (h *createVpcPollingHandler) Done(items []*compute.Operation) (bool, *[]uint64) {
	results := make([]uint64, 0)

	flag := true
	for _, item := range items {
		if item.OperationType == "insert" && item.Status == "DONE" {
			results = append(results, item.TargetId)
			continue
		}

		if item.OperationType == "insert" && item.Status == "PENDING" {
			flag = false
			continue
		}

		if item.OperationType == "insert" && item.Status == "RUNNING" {
			flag = false
			continue
		}
	}

	return flag, converter.ValToPtr(results)
}

// Poll ...
func (h *createVpcPollingHandler) Poll(client *Gcp, kt *kit.Kit, operGroupIDs []*string) ([]*compute.Operation, error) {
	if len(operGroupIDs) == 0 {
		return nil, errors.New("operation group id is required")
	}

	computeClient, err := client.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	operResp, err := computeClient.ZoneOperations.List(client.CloudProjectID(), h.zone).Context(kt.Ctx).
		Filter(fmt.Sprintf(`operationGroupId="%s"`, *operGroupIDs[0])).Do()
	if err != nil {
		logs.Errorf("list zone operations failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(operResp.Items) <= 1 {
		return nil, errors.New("operation has not been created yet, need to wait")
	}

	return operResp.Items, nil
}
