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
	"errors"
	"fmt"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// ListEip ...
// reference: https://cloud.tencent.com/document/api/215/16702
func (t *TCloudImpl) ListEip(kt *kit.Kit, opt *eip.TCloudEipListOption) (*eip.TCloudEipListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := vpc.NewDescribeAddressesRequest()

	if len(opt.CloudIDs) > 0 {
		req.Filters = []*vpc.Filter{
			{
				Name:   converter.ValToPtr("address-id"),
				Values: converter.SliceToPtr(opt.CloudIDs),
			},
		}
		req.Limit = common.Int64Ptr(int64(core.TCloudQueryLimit))
	}

	if len(opt.Ips) > 0 {
		req.Filters = []*vpc.Filter{
			{
				Name:   converter.ValToPtr("address-ip"),
				Values: converter.SliceToPtr(opt.Ips),
			},
		}
		req.Limit = common.Int64Ptr(int64(core.TCloudQueryLimit))
	}

	if opt.Page != nil {
		req.Offset = common.Int64Ptr(int64(opt.Page.Offset))
		req.Limit = common.Int64Ptr(int64(opt.Page.Limit))
	}

	resp, err := client.DescribeAddressesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tencent cloud eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list tencent cloud eip failed, err: %v", err)
	}

	eips := make([]*eip.TCloudEip, len(resp.Response.AddressSet))
	for idx, address := range resp.Response.AddressSet {
		eips[idx] = &eip.TCloudEip{
			CloudID:                 *address.AddressId,
			Name:                    address.AddressName,
			Region:                  opt.Region,
			InstanceId:              address.InstanceId,
			Status:                  address.AddressStatus,
			PublicIp:                address.AddressIp,
			PrivateIp:               address.PrivateAddressIp,
			Bandwidth:               address.Bandwidth,
			InternetChargeType:      address.InternetChargeType,
			InternetServiceProvider: address.InternetServiceProvider,
		}
	}

	count := uint64(*resp.Response.TotalCount)
	return &eip.TCloudEipListResult{Details: eips, Count: &count}, nil
}

// CountEip 基于 DescribeAddressesWithContext
// reference: https://cloud.tencent.com/document/api/215/16702
func (t *TCloudImpl) CountEip(kt *kit.Kit, region string) (int32, error) {

	client, err := t.clientSet.VpcClient(region)
	if err != nil {
		return 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeAddressesRequest()
	req.Limit = converter.ValToPtr(int64(1))
	resp, err := client.DescribeAddressesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud eip failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return 0, err
	}
	return int32(*resp.Response.TotalCount), nil
}

// DeleteEip ...
// reference: https://cloud.tencent.com/document/api/215/16700
func (t *TCloudImpl) DeleteEip(kt *kit.Kit, opt *eip.TCloudEipDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud eip delete option is required")
	}

	req, err := opt.ToReleaseAddressesRequest()
	if err != nil {
		return err
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.ReleaseAddressesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud release eip failed, err: %v, rid: %s, resp rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// AssociateEip ...
// reference: https://cloud.tencent.com/document/api/215/16700
func (t *TCloudImpl) AssociateEip(kt *kit.Kit, opt *eip.TCloudEipAssociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud eip associate option is required")
	}

	req, err := opt.ToAssociateAddressRequest()
	if err != nil {
		return err
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.AssociateAddressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud associate eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []*eip.TCloudEip,
		poller.BaseDoneResult]{Handler: &associateEipPollingHandler{region: opt.Region}}
	_, err = respPoller.PollUntilDone(t, kt, []*string{&opt.CloudEipID}, nil)
	if err != nil {
		return err
	}

	return nil
}

// DisassociateEip ...
// reference: https://cloud.tencent.com/document/api/215/16703
func (t *TCloudImpl) DisassociateEip(kt *kit.Kit, opt *eip.TCloudEipDisassociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud eip disassociate option is required")
	}

	req, err := opt.ToDisassociateAddressRequest()
	if err != nil {
		return err
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.DisassociateAddressWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud disassociate eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []*eip.TCloudEip,
		poller.BaseDoneResult]{Handler: &disassociateEipPollingHandler{region: opt.Region}}
	_, err = respPoller.PollUntilDone(t, kt, []*string{&opt.CloudEipID}, nil)
	if err != nil {
		return err
	}

	return nil
}

// DetermineIPv6Type 判断ipv6地址是否是公网ip
func (t *TCloudImpl) DetermineIPv6Type(kt *kit.Kit, region string, ipv6Addresses []*string) ([]*string,
	[]*string, error,
) {
	if len(region) == 0 || len(ipv6Addresses) == 0 {
		return nil, nil, errors.New("region and ipv6Addresses is required")
	}

	client, err := t.clientSet.VpcClient(region)
	if err != nil {
		return nil, nil, err
	}

	req := vpc.NewDescribeIp6AddressesRequest()
	req.Filters = []*vpc.Filter{
		{
			Name:   converter.ValToPtr("address-ip"),
			Values: ipv6Addresses,
		},
	}

	resp, err := client.DescribeIp6AddressesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud ipv6 address failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, fmt.Errorf("list tencent cloud eip failed, err: %v", err)
	}

	if len(resp.Response.AddressSet) != len(ipv6Addresses) {
		return nil, nil, fmt.Errorf("list ipv6Address return count not right, ipv6Address: %v, count: %d",
			ipv6Addresses, len(resp.Response.AddressSet))
	}

	publicIPv6Address := make([]*string, 0)
	privateIPv6Address := make([]*string, 0)
	for _, one := range resp.Response.AddressSet {
		if one.Bandwidth == nil || *one.Bandwidth == 0 {
			privateIPv6Address = append(privateIPv6Address, one.AddressIp)
		} else {
			publicIPv6Address = append(publicIPv6Address, one.AddressIp)
		}
	}

	return publicIPv6Address, privateIPv6Address, nil
}

// CreateEip ...
// reference: https://cloud.tencent.com/document/api/215/16699
func (t *TCloudImpl) CreateEip(kt *kit.Kit, opt *eip.TCloudEipCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud eip create option is required")
	}

	req, err := opt.ToAllocateAddressesRequest()
	if err != nil {
		return nil, err
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.AllocateAddressesWithContext(kt.Ctx, req)
	if err != nil {
		return nil, err
	}

	respPoller := poller.Poller[*TCloudImpl, []*eip.TCloudEip,
		poller.BaseDoneResult]{Handler: &createEipPollingHandler{region: opt.Region}}
	return respPoller.PollUntilDone(t, kt, resp.Response.AddressSet, nil)
}

// DescribeNetworkAccountType 查询用户网络类型
// https://cloud.tencent.com/document/product/215/86509
func (t *TCloudImpl) DescribeNetworkAccountType(kt *kit.Kit) (
	*vpc.DescribeNetworkAccountTypeResponseParams, error) {

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v",
			constant.TCloudDefaultRegion, err)
	}
	req := vpc.NewDescribeNetworkAccountTypeRequest()
	resp, err := client.DescribeNetworkAccountTypeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf(" tencent cloud describe account type failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}
	return resp.Response, nil
}

type createEipPollingHandler struct {
	region string
}

// Done ...
func (h *createEipPollingHandler) Done(pollResult []*eip.TCloudEip) (bool, *poller.BaseDoneResult) {
	successCloudIDs := make([]string, 0)
	unknownCloudIDs := make([]string, 0)

	for _, r := range pollResult {
		if converter.PtrToVal(r.Status) == "CREATING" {
			unknownCloudIDs = append(unknownCloudIDs, r.CloudID)
		} else {
			successCloudIDs = append(successCloudIDs, r.CloudID)
		}
	}

	isDone := false
	if len(successCloudIDs) != 0 && len(successCloudIDs) == len(pollResult) {
		isDone = true
	}

	return isDone, &poller.BaseDoneResult{
		SuccessCloudIDs: successCloudIDs,
		UnknownCloudIDs: unknownCloudIDs,
	}
}

// Poll ...
func (h *createEipPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]*eip.TCloudEip, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.TCloudEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

var _ poller.PollingHandler[*TCloudImpl, []*eip.TCloudEip, poller.BaseDoneResult] = new(createEipPollingHandler)

type associateEipPollingHandler struct {
	region string
}

// Poll ...
func (h *associateEipPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]*eip.TCloudEip,
	error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.TCloudEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

// Done ...
func (h *associateEipPollingHandler) Done(pollResult []*eip.TCloudEip) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]

	if converter.PtrToVal(r.Status) != "BIND" || r.InstanceId == nil {
		return false, nil
	}
	return true, nil
}

type disassociateEipPollingHandler struct {
	region string
}

// Poll ...
func (h *disassociateEipPollingHandler) Poll(
	client *TCloudImpl,
	kt *kit.Kit,
	cloudIDs []*string,
) ([]*eip.TCloudEip, error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.TCloudEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

// Done ...
func (h *disassociateEipPollingHandler) Done(pollResult []*eip.TCloudEip) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]

	if converter.PtrToVal(r.Status) != "UNBIND" {
		return false, nil
	}
	return true, nil
}
