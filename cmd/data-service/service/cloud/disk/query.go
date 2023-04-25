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

package disk

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// RetrieveDiskExt 获取云盘详情
func (dSvc *diskSvc) RetrieveDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskID := cts.PathParameter("id").String()
	opt := &types.ListOption{
		Filter: tools.EqualWithOpExpression(
			filter.And,
			map[string]interface{}{"id": diskID, "vendor": string(vendor)},
		),
		Page: &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	data, err := dSvc.dao.Disk().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if count := len(data.Details); count != 1 {
		return nil, fmt.Errorf("retrieve disk failed: query id(%s) return total %d", diskID, count)
	}

	diskData := data.Details[0]
	switch vendor {
	case enumor.TCloud:
		return toProtoDiskExtResult[dataproto.TCloudDiskExtensionResult](diskData)
	case enumor.Aws:
		return toProtoDiskExtResult[dataproto.AwsDiskExtensionResult](diskData)
	case enumor.Gcp:
		return toProtoDiskExtResult[dataproto.GcpDiskExtensionResult](diskData)
	case enumor.Azure:
		return toProtoDiskExtResult[dataproto.AzureDiskExtensionResult](diskData)
	case enumor.HuaWei:
		return toProtoDiskExtResult[dataproto.HuaWeiDiskExtensionResult](diskData)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// ListDisk 查询云盘列表
func (dSvc *diskSvc) ListDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.DiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := dSvc.dao.Disk().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*dataproto.DiskResult, len(data.Details))
	for indx, d := range data.Details {
		details[indx] = toProtoDiskResult(d)
	}

	return &dataproto.DiskListResult{Details: details, Count: data.Count}, nil
}

// ListDiskExt 获取云盘列表(带 extension 字段)
func (dSvc *diskSvc) ListDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.DiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := dSvc.dao.Disk().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoDiskExtListResult[dataproto.TCloudDiskExtensionResult](data)
	case enumor.Aws:
		return toProtoDiskExtListResult[dataproto.AwsDiskExtensionResult](data)
	case enumor.Gcp:
		return toProtoDiskExtListResult[dataproto.GcpDiskExtensionResult](data)
	case enumor.Azure:
		return toProtoDiskExtListResult[dataproto.AzureDiskExtensionResult](data)
	case enumor.HuaWei:
		return toProtoDiskExtListResult[dataproto.HuaWeiDiskExtensionResult](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// CountDisk 统计云盘数量
func (dSvc *diskSvc) CountDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.DiskCountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.CountOption{
		Filter: req.Filter,
	}

	data, err := dSvc.dao.Disk().Count(cts.Kit, opt)
	if err != nil {
		return nil, err
	}
	return &dataproto.DiskCountResult{Count: data.Count}, nil
}
