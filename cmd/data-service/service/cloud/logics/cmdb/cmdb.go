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

package cmdb

import (
	"strings"

	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// CmdbLogics defines cmdb logics.
type CmdbLogics struct {
	client cmdb.Client
}

// NewCmdbLogics init cmdb logics.
func NewCmdbLogics(client cmdb.Client) *CmdbLogics {
	return &CmdbLogics{client: client}
}

// AddCloudHostToBiz add cmdb cloud host to biz, update cmdb host if exists.
func AddCloudHostToBiz[T cvm.Extension](c *CmdbLogics, kt *kit.Kit, req *AddCloudHostToBizReq[T]) error {
	if err := req.Validate(); err != nil {
		return err
	}

	vendorCmdbHostStatusMap, exists := cmdb.HcmCmdbHostStatusMap[req.Vendor]
	if !exists {
		return errf.Newf(errf.InvalidParameter, "vendor %s is invalid", req.Vendor)
	}

	hosts := make([]cmdb.Host, 0, len(req.Hosts))
	for _, host := range req.Hosts {
		if host.Vendor != "" && req.Vendor != host.Vendor {
			return errf.Newf(errf.InvalidParameter, "host vendor %s not match req vendor %s", host.Vendor, req.Vendor)
		}
		if host.Vendor == "" {
			host.Vendor = req.Vendor
		}

		status, exists := vendorCmdbHostStatusMap[host.Status]
		if !exists {
			status = "1"
		}

		hosts = append(hosts, cmdb.Host{
			BkCloudVendor:     cmdb.HcmCmdbVendorMap[req.Vendor],
			BkCloudInstID:     host.CloudID,
			BkCloudHostStatus: status,
			BkCloudID:         host.BkCloudID,
			BkHostInnerIP:     strings.Join(host.PrivateIPv4Addresses, ","),
			BkHostOuterIP:     strings.Join(host.PublicIPv4Addresses, ","),
			BkHostInnerIPv6:   strings.Join(host.PrivateIPv6Addresses, ","),
			BkHostOuterIPv6:   strings.Join(host.PublicIPv6Addresses, ","),
			BkHostName:        host.Name,
			BkComment:         host.Memo,
		})
	}

	params := &cmdb.AddCloudHostToBizParams{
		BizID:    req.BizID,
		HostInfo: hosts,
	}
	_, err := c.client.AddCloudHostToBiz(kt, params)
	if err != nil {
		return err
	}

	return nil
}

// AddBaseCloudHostToBiz add cmdb cloud host basic info to biz, update cmdb host if exists.
func AddBaseCloudHostToBiz(c *CmdbLogics, kt *kit.Kit, req *AddBaseCloudHostToBizReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	hosts := make([]cmdb.Host, 0, len(req.Hosts))
	for _, host := range req.Hosts {
		if err := host.Vendor.Validate(); err != nil {
			return err
		}

		status, exists := cmdb.HcmCmdbHostStatusMap[host.Vendor][host.Status]
		if !exists {
			status = "1"
		}

		hosts = append(hosts, cmdb.Host{
			BkCloudVendor:     cmdb.HcmCmdbVendorMap[host.Vendor],
			BkCloudInstID:     host.CloudID,
			BkCloudHostStatus: status,
			BkCloudID:         host.BkCloudID,
			BkHostInnerIP:     strings.Join(host.PrivateIPv4Addresses, ","),
			BkHostOuterIP:     strings.Join(host.PublicIPv4Addresses, ","),
			BkHostInnerIPv6:   strings.Join(host.PrivateIPv6Addresses, ","),
			BkHostOuterIPv6:   strings.Join(host.PublicIPv6Addresses, ","),
			BkHostName:        host.Name,
			BkComment:         host.Memo,
		})
	}

	params := &cmdb.AddCloudHostToBizParams{
		BizID:    req.BizID,
		HostInfo: hosts,
	}
	_, err := c.client.AddCloudHostToBiz(kt, params)
	if err != nil {
		return err
	}

	return nil
}

// DeleteCloudHostFromBiz delete cmdb cloud host from biz.
func (c *CmdbLogics) DeleteCloudHostFromBiz(kt *kit.Kit, req *DeleteCloudHostFromBizReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// get cmdb host ids
	rules := make([]cmdb.Rule, 0)
	for vendor, cloudIDs := range req.VendorCloudIDs {
		rules = append(rules, &cmdb.CombinedRule{
			Condition: "AND",
			Rules: []cmdb.Rule{
				&cmdb.AtomRule{
					Field:    "bk_cloud_vendor",
					Operator: cmdb.OperatorEqual,
					Value:    cmdb.HcmCmdbVendorMap[vendor],
				},
				&cmdb.AtomRule{
					Field:    "bk_cloud_inst_id",
					Operator: cmdb.OperatorIn,
					Value:    cloudIDs,
				},
			},
		})
	}

	listParams := &cmdb.ListBizHostParams{
		BizID:  req.BizID,
		Fields: []string{"bk_host_id"},
		Page:   cmdb.BasePage{Limit: 500},
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: &cmdb.CombinedRule{
				Condition: "OR",
				Rules:     rules,
			},
		},
	}
	hosts, err := c.client.ListBizHost(kt, listParams)
	if err != nil {
		return err
	}

	if len(hosts.Info) == 0 {
		return nil
	}

	hostIDs := make([]int64, len(hosts.Info))
	for i, host := range hosts.Info {
		hostIDs[i] = host.BkHostID
	}

	// delete cmdb host
	delParams := &cmdb.DeleteCloudHostFromBizParams{
		BizID:   req.BizID,
		HostIDs: hostIDs,
	}
	err = c.client.DeleteCloudHostFromBiz(kt, delParams)
	if err != nil {
		return err
	}

	return nil
}
