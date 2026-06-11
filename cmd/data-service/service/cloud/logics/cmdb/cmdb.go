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

// Package cmdb ...
package cmdb

import (
	"strings"
	"time"

	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
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
func AddCloudHostToBiz[T cvm.Extension](c *CmdbLogics, kt *kit.Kit, req *AddCloudHostToBizReq[T]) ([]int64, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	vendorCmdbHostStatusMap, exists := cmdb.HcmCmdbHostStatusMap[req.Vendor]
	if !exists {
		return nil, errf.Newf(errf.InvalidParameter, "vendor %s is invalid", req.Vendor)
	}
	hosts := make([]cmdb.HostCreateParam, 0, len(req.Hosts))
	for _, host := range req.Hosts {
		if host.Vendor != "" && req.Vendor != host.Vendor {
			return nil, errf.Newf(errf.InvalidParameter, "host vendor %s not match req vendor %s", host.Vendor,
				req.Vendor)
		}
		if host.Vendor == "" {
			host.Vendor = req.Vendor
		}

		status, exists := vendorCmdbHostStatusMap[host.Status]
		if !exists {
			status = "1"
		}
		onShelfDate, err := getOnShelfDate(kt, host.BaseCvm)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, cmdb.HostCreateParam{
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
			IsGPU:             host.IsGPU,
			OnShelfDate:       onShelfDate,
			BkCloudRegion:     host.Region,
			BkCloudZone:       host.Zone,
			InstanceType:      host.MachineType,
			// Operator 仅对首次新增到 cmdb 的主机下发（由上层 buildCmdbOperators 推导后放入 req.Operators）。
			// 未命中 map 时取到的是 nil 指针，配合 omitempty 不会下发该字段，从而保留 cmdb 侧已有的 operator。
			Operator: req.Operators[host.ID],
		})
	}

	params := &cmdb.AddCloudHostToBizParams{
		BizID:    req.BizID,
		HostInfo: hosts,
	}
	logs.Infof("add cmdb cloud host to biz, vendor: %s, bizID: %d, hostCount: %d, operatorCount: %d, rid: %s",
		req.Vendor, req.BizID, len(hosts), len(req.Operators), kt.Rid)
	result, err := c.client.AddCloudHostToBiz(kt, params)
	if err != nil {
		logs.Errorf("add cmdb cloud host to biz failed, err: %v, vendor: %s, bizID: %d, rid: %s",
			err, req.Vendor, req.BizID, kt.Rid)
		return nil, err
	}

	return result.IDs, nil
}

// getOnShelfDate 返回用作 cmdb on_shelf_date 的主机上架时间。
func getOnShelfDate(kt *kit.Kit, host cvm.BaseCvm) (string, error) {
	if host.Vendor == enumor.Other {
		return "", nil
	}

	var shelfTime string
	shelfTime = host.CloudCreatedTime
	if host.Vendor == enumor.Aws {
		// AWS 是只有CloudLaunchedTime，代表购买时间
		shelfTime = host.CloudLaunchedTime
	}

	formDate, err := time.Parse(constant.TimeStdFormat, shelfTime)
	if err != nil {
		logs.Errorf("parse shelf time failed, err: %v, shelfTime: %s, rid: %s", err, shelfTime, kt.Rid)
		return "", err
	}

	return formDate.Format(constant.DateLayout), nil
}

// AddBaseCloudHostToBiz add cmdb cloud host basic info to biz, update cmdb host if exists.
func AddBaseCloudHostToBiz(c *CmdbLogics, kt *kit.Kit, req *AddBaseCloudHostToBizReq) ([]int64, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	hosts := make([]cmdb.HostCreateParam, 0, len(req.Hosts))
	for _, host := range req.Hosts {
		if err := host.Vendor.Validate(); err != nil {
			return nil, err
		}

		status, exists := cmdb.HcmCmdbHostStatusMap[host.Vendor][host.Status]
		if !exists {
			status = "1"
		}
		onShelfDate, err := getOnShelfDate(kt, host)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, cmdb.HostCreateParam{
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
			IsGPU:             host.IsGPU,
			BkCloudRegion:     host.Region,
			BkCloudZone:       host.Zone,
			InstanceType:      host.MachineType,
			OnShelfDate:       onShelfDate,
			// Operator 仅对首次新增到 cmdb 的主机下发（由上层 buildCmdbOperators 推导后放入 req.Operators）。
			// 未命中 map 时取到的是 nil 指针，配合 omitempty 不会下发该字段，从而保留 cmdb 侧已有的 operator。
			Operator: req.Operators[host.ID],
		})
	}

	params := &cmdb.AddCloudHostToBizParams{
		BizID:    req.BizID,
		HostInfo: hosts,
	}
	logs.Infof("add cmdb base cloud host to biz, bizID: %d, hostCount: %d, operatorCount: %d, rid: %s",
		req.BizID, len(hosts), len(req.Operators), kt.Rid)
	result, err := c.client.AddCloudHostToBiz(kt, params)
	if err != nil {
		logs.Errorf("add cmdb base cloud host to biz failed, err: %v, bizID: %d, rid: %s",
			err, req.BizID, kt.Rid)
		return nil, err
	}

	return result.IDs, nil
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
		Page:   &cmdb.BasePage{Limit: 500},
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
