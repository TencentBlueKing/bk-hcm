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

package image

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// RetrieveImageExt ...
func (svc *imageSvc) RetrieveImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageID := cts.PathParameter("id").String()
	opt := &types.ListOption{
		Filter: tools.EqualWithOpExpression(
			filter.And,
			map[string]interface{}{"id": imageID, "vendor": string(vendor)},
		),
		Page: &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	data, err := svc.dao.Image().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if count := len(data.Details); count != 1 {
		return nil, fmt.Errorf("retrieve image failed: query id(%s) return total %d", imageID, count)
	}

	imageData := data.Details[0]
	switch vendor {
	case enumor.TCloud:
		return toProtoImageExtResult[dataproto.TCloudImageExtensionResult](imageData)
	case enumor.Aws:
		return toProtoImageExtResult[dataproto.AwsImageExtensionResult](imageData)
	case enumor.Gcp:
		return toProtoImageExtResult[dataproto.GcpImageExtensionResult](imageData)
	case enumor.Azure:
		return toProtoImageExtResult[dataproto.AzureImageExtensionResult](imageData)
	case enumor.HuaWei:
		return toProtoImageExtResult[dataproto.HuaWeiImageExtensionResult](imageData)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// ListImage ...
func (svc *imageSvc) ListImage(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.ImageListReq)
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

	data, err := svc.dao.Image().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*dataproto.ImageResult, len(data.Details))
	for indx, d := range data.Details {
		details[indx] = toProtoImageResult(d)
	}

	return &dataproto.ImageListResult{Details: details, Count: data.Count}, nil
}

// ListImageExt ...
func (svc *imageSvc) ListImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.ImageListReq)
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

	data, err := svc.dao.Image().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoImageExtListResult[dataproto.TCloudImageExtensionResult](data)
	case enumor.Aws:
		return toProtoImageExtListResult[dataproto.AwsImageExtensionResult](data)
	case enumor.Gcp:
		return toProtoImageExtListResult[dataproto.GcpImageExtensionResult](data)
	case enumor.HuaWei:
		return toProtoImageExtListResult[dataproto.HuaWeiImageExtensionResult](data)
	case enumor.Azure:
		return toProtoImageExtListResult[dataproto.AzureImageExtensionResult](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}
