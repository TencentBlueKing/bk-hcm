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

package bwpkg

import (
	"net/http"

	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	hcbwpkg "hcm/pkg/api/hc-service/bandwidth-packages"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitBwPkgService initial argument template service.
func InitBwPkgService(cap *capability.Capability) {
	svc := &bwPkgSvc{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	svc.initTCloudBwPkgService(cap)
}

type bwPkgSvc struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
	client  *client.ClientSet
}

func (svc bwPkgSvc) initTCloudBwPkgService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("ListBandwidthPackage", http.MethodPost,
		"/vendors/tcloud/bandwidth_packages/list", svc.ListTCloudBandwidthPackage)

	h.Load(cap.WebService)
}

// ListTCloudBandwidthPackage 查询腾讯云带宽包
func (svc bwPkgSvc) ListTCloudBandwidthPackage(cts *rest.Contexts) (any, error) {
	req := new(hcbwpkg.ListTCloudBwPkgOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Page == nil {
		req.Page = &core.TCloudPage{Offset: 0, Limit: core.TCloudQueryLimit}
	}

	cli, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("tcloud request adaptor client err, err: %+v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	if req.Page.Limit > core.TCloudQueryLimit {
		req.Page.Limit = core.TCloudQueryLimit
	}
	opt := &types.TCloudListBwPkgOption{
		Region:        req.Region,
		Page:          req.Page,
		PkgCloudIds:   req.PkgCloudIds,
		PkgNames:      req.PkgNames,
		NetworkTypes:  req.NetworkTypes,
		ChargeTypes:   req.ChargeTypes,
		ResourceTypes: req.ResourceTypes,
		ResourceIds:   req.ResourceIds,
		ResAddressIps: req.ResAddressIps,
	}

	resp, err := cli.ListBandwidthPackage(cts.Kit, opt)
	if err != nil {
		logs.Errorf("tcloud request adaptor list bandwidth package failed, err: %v, req: %v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	return resp, nil
}
