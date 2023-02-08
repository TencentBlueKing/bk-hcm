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

// Package region defines region service.
package region

import (
	"fmt"

	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// InitRegionService initial the region service
func InitRegionService(cap *capability.Capability) {
	v := &region{
		ad: cap.CloudAdaptor,
		cs: cap.ClientSet,
	}

	h := rest.NewHandler()
	h.Add("BatchSyncRegion", "POST", "/vendors/{vendor}/regions/sync", v.BatchSyncRegion)

	h.Load(cap.WebService)
}

type region struct {
	ad *cloudadaptor.CloudAdaptorClient
	cs *client.ClientSet
}

// BatchSyncRegion batch sync region.
func (r *region) BatchSyncRegion(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	err := vendor.Validate()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		err = r.TCloudSyncRegion(cts, vendor)
	case enumor.Aws:
		err = r.AwsSyncRegion(cts, vendor)
	case enumor.Gcp:
		err = r.GcpSyncRegion(cts, vendor)
	default:
		err = fmt.Errorf("%s does not support the creation of cloud region", vendor)
	}
	return nil, err
}
