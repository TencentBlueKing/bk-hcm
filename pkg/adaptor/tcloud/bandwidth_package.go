/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// ListBandwidthPackage 查询带宽包资源
// reference: https://cloud.tencent.com/document/product/215/19209
func (t *TCloudImpl) ListBandwidthPackage(kt *kit.Kit, opt *types.TCloudListBwPkgOption) (
	*types.TCloudListBwPkgResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, region: %s, err: %v", opt.Region, err)
	}

	req := convListBwPkgRequest(opt)
	resp, err := client.DescribeBandwidthPackagesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud bandwidth packages failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	bandwidthPackages := make([]types.TCloudBandwidthPackage, 0, len(resp.Response.BandwidthPackageSet))
	for _, one := range resp.Response.BandwidthPackageSet {
		bandwidthPackages = append(bandwidthPackages, types.TCloudBandwidthPackage{
			ID:          cvt.PtrToVal(one.BandwidthPackageId),
			Name:        cvt.PtrToVal(one.BandwidthPackageName),
			NetworkType: types.BwPkgNetworkType(cvt.PtrToVal(one.NetworkType)),
			ChargeType:  types.BwPkgChargeType(cvt.PtrToVal(one.ChargeType)),
			Status:      cvt.PtrToVal(one.Status),
			Bandwidth:   cvt.PtrToVal(one.Bandwidth),
			Egress:      cvt.PtrToVal(one.Egress),
			CreateTime:  cvt.PtrToVal(one.CreatedTime),
			Deadline:    cvt.PtrToVal(one.Deadline),
			ResourceSet: slice.Map(one.ResourceSet, convResource),
		})
	}

	return &types.TCloudListBwPkgResult{
		TotalCount: cvt.PtrToVal(resp.Response.TotalCount),
		Packages:   bandwidthPackages,
	}, nil
}

func convListBwPkgRequest(opt *types.TCloudListBwPkgOption) *vpc.DescribeBandwidthPackagesRequest {

	req := vpc.NewDescribeBandwidthPackagesRequest()
	if opt.Page != nil {
		req.Offset = cvt.ValToPtr(opt.Page.Offset)
		if opt.Page.Limit > 0 {
			req.Limit = cvt.ValToPtr(opt.Page.Limit)
		}
	}
	if len(opt.PkgCloudIds) != 0 {
		req.BandwidthPackageIds = common.StringPtrs(opt.PkgCloudIds)
	}

	if len(opt.PkgNames) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name:   cvt.ValToPtr("bandwidth-package-name"),
			Values: cvt.SliceToPtr(opt.PkgNames),
		})
	}
	if len(opt.NetworkTypes) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name:   cvt.ValToPtr("network-type"),
			Values: convToStringPtrSlice(opt.NetworkTypes),
		})
	}
	if len(opt.ChargeTypes) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name:   cvt.ValToPtr("charge-type"),
			Values: convToStringPtrSlice(opt.ChargeTypes),
		})
	}
	if len(opt.ResourceTypes) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name:   cvt.ValToPtr("resource.resource-type"),
			Values: cvt.SliceToPtr(opt.ResourceTypes),
		})
	}
	if len(opt.ResourceIds) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name:   cvt.ValToPtr("resource.resource-id"),
			Values: cvt.SliceToPtr(opt.ResourceIds),
		})
	}
	if len(opt.ResAddressIps) != 0 {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name:   cvt.ValToPtr("resource.address-ip"),
			Values: cvt.SliceToPtr(opt.ResAddressIps),
		})
	}

	return req
}

func convResource(res *vpc.Resource) types.Resource {
	return types.Resource{
		ResourceType: cvt.PtrToVal(res.ResourceType),
		ResourceID:   cvt.PtrToVal(res.ResourceId),
		AddressIP:    cvt.PtrToVal(res.AddressIp),
	}
}

// convToStringPtrSlice convert to string pointer slice
func convToStringPtrSlice[T ~string](vals []T) []*string {
	s := make([]*string, len(vals))
	for i, v := range vals {
		s[i] = cvt.ValToPtr((string)(v))
	}
	return s
}
