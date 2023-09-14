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

// Package region ...
package region

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	protoregion "hcm/pkg/api/cloud-server/region"
	"hcm/pkg/api/core"
	dataregion "hcm/pkg/api/data-service/cloud/region"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitRegionService initialize the region service.
func InitRegionService(c *capability.Capability) {
	svc := &RegionSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("ListRegion", http.MethodPost, "/vendors/{vendor}/regions/list", svc.ListRegion)

	h.Load(c.WebService)
}

// RegionSvc region svc
type RegionSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// ListRegion ...
func (svc *RegionSvc) ListRegion(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(protoregion.RegionListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	reqPage := &core.BasePage{
		Count: req.Page.Count,
		Start: req.Page.Start,
		Limit: req.Page.Limit,
	}

	switch vendor {
	case enumor.TCloud:
		listReq := &dataregion.TCloudRegionListReq{
			Filter: req.Filter,
			Page:   reqPage,
		}
		return svc.client.DataService().TCloud.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)

	case enumor.Aws:
		listReq := &dataregion.AwsRegionListReq{
			Filter: req.Filter,
			Page:   reqPage,
		}
		return svc.client.DataService().Aws.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)

	case enumor.HuaWei:
		listReq := &dataregion.HuaWeiRegionListReq{
			Filter: req.Filter,
			Page:   reqPage,
		}
		return svc.client.DataService().HuaWei.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)

	case enumor.Azure:
		listReq := &dataregion.AzureRegionListReq{
			Filter: req.Filter,
			Page:   reqPage,
		}
		return svc.client.DataService().Azure.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)

	case enumor.Gcp:
		listReq := &dataregion.GcpRegionListReq{
			Filter: req.Filter,
			Page:   reqPage,
		}
		return svc.client.DataService().Gcp.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), listReq)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}
