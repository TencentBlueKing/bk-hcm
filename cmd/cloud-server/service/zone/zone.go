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

package zone

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	cloudproto "hcm/pkg/api/cloud-server/zone"
	"hcm/pkg/api/core/cloud/zone"
	dataproto "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

const (
	Azure = "azure"
	Zone1 = "1"
	Zone2 = "2"
	Zone3 = "3"
)

// InitZoneService initialize the zone service.
func InitZoneService(c *capability.Capability) {
	svc := &ZoneSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("ListZone", http.MethodPost, "/vendors/{vendor}/regions/{region}/zones/list", svc.ListZone)

	h.Load(c.WebService)
}

// ZoneSvc zone svc
type ZoneSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// ListZone ...
func (dSvc *ZoneSvc) ListZone(cts *rest.Contexts) (interface{}, error) {
	vendor := cts.PathParameter("vendor").String()
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	region := cts.PathParameter("region").String()
	if len(region) == 0 {
		return nil, errf.New(errf.InvalidParameter, "region is required")
	}

	if vendor == Azure {
		resp := new(dataproto.ZoneListResult)
		resp.Count = 3
		resp.Details = []zone.BaseZone{
			{
				ID:      "",
				Vendor:  Azure,
				CloudID: "",
				Name:    Zone1,
				Region:  region,
			},
			{
				ID:      "",
				Vendor:  Azure,
				CloudID: "",
				Name:    Zone2,
				Region:  region,
			},
			{
				ID:      "",
				Vendor:  Azure,
				CloudID: "",
				Name:    Zone3,
				Region:  region,
			},
		}

		return resp, nil
	}

	req := new(cloudproto.ZoneListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return dSvc.client.DataService().Global.Zone.ListZone(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.ZoneListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: vendor,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: region,
					},
				},
			},
			Page: req.Page,
		},
	)
}
