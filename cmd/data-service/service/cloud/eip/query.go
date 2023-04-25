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

package eip

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// RetrieveEipExt ...
func (svc *eipSvc) RetrieveEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	eipID := cts.PathParameter("id").String()
	opt := &types.ListOption{
		Filter: tools.EqualWithOpExpression(
			filter.And,
			map[string]interface{}{"id": eipID, "vendor": string(vendor)},
		),
		Page: &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	data, err := svc.dao.Eip().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if count := len(data.Details); count != 1 {
		return nil, fmt.Errorf("retrieve eip failed: query id(%s) return total %d", eipID, count)
	}

	eipData := data.Details[0]
	switch vendor {
	case enumor.TCloud:
		return toProtoEipExtResult[dataproto.TCloudEipExtensionResult](eipData)
	case enumor.Aws:
		return toProtoEipExtResult[dataproto.AwsEipExtensionResult](eipData)
	case enumor.Gcp:
		return toProtoEipExtResult[dataproto.GcpEipExtensionResult](eipData)
	case enumor.Azure:
		return toProtoEipExtResult[dataproto.AzureEipExtensionResult](eipData)
	case enumor.HuaWei:
		return toProtoEipExtResult[dataproto.HuaWeiEipExtensionResult](eipData)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// ListEip ...
func (svc *eipSvc) ListEip(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipListReq)
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

	data, err := svc.dao.Eip().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*dataproto.EipResult, len(data.Details))
	for indx, d := range data.Details {
		details[indx] = toProtoEipResult(d)
	}

	return &dataproto.EipListResult{Details: details, Count: data.Count}, nil
}

// ListEipExt ...
func (svc *eipSvc) ListEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.EipListReq)
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

	data, err := svc.dao.Eip().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoEipExtListResult[dataproto.TCloudEipExtensionResult](data)
	case enumor.Aws:
		return toProtoEipExtListResult[dataproto.AwsEipExtensionResult](data)
	case enumor.Gcp:
		return toProtoEipExtListResult[dataproto.GcpEipExtensionResult](data)
	case enumor.HuaWei:
		return toProtoEipExtListResult[dataproto.HuaWeiEipExtensionResult](data)
	case enumor.Azure:
		return toProtoEipExtListResult[dataproto.AzureEipExtensionResult](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}
