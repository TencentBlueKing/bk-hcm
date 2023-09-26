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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// RetrieveImage ...
func (svc *imageSvc) RetrieveImage(cts *rest.Contexts) (interface{}, error) {
	imageID := cts.PathParameter("id").String()

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.GetImage(cts.Kit, imageID)
	case enumor.Aws:
		return svc.client.DataService().Aws.GetImage(cts.Kit, imageID)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.GetImage(cts.Kit, imageID)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.GetImage(cts.Kit, imageID)
	case enumor.Azure:
		return svc.client.DataService().Azure.GetImage(cts.Kit, imageID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// ListImage ...
func (svc *imageSvc) ListImage(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Filter == nil {
		req.Filter = tools.AllExpression()
	}
	return svc.client.DataService().Global.ListImage(cts.Kit, req)
}

// ListImageExt ...
func (svc *imageSvc) ListImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Filter == nil {
		req.Filter = tools.AllExpression()
	}

	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.ListImage(cts.Kit, req)
	case enumor.Aws:
		return svc.client.DataService().Aws.ListImage(cts.Kit, req)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.ListImage(cts.Kit, req)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.ListImage(cts.Kit, req)
	case enumor.Azure:
		return svc.client.DataService().Azure.ListImage(cts.Kit, req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}
