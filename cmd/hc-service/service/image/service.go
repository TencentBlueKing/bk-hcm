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

	"hcm/cmd/hc-service/service/capability"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// InitImageService initial the disk service
func InitImageService(cap *capability.Capability) {
	i := &imageAdaptor{
		idaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	// 公共镜像每个周期只同步一次
	h.Add("SyncImages", "POST", "/vendors/{vendor}/images/sync", i.SyncImages)

	h.Load(cap.WebService)
}

type imageAdaptor struct {
	idaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}

var SyncFuncs = map[enumor.Vendor]func(da *imageAdaptor, cts *rest.Contexts) (interface{}, error){
	enumor.TCloud: TCloudSyncImage,
	enumor.HuaWei: HuaWeiSyncImage,
	enumor.Aws:    AwsSyncImage,
	enumor.Azure:  AzureSyncImage,
	enumor.Gcp:    GcpSyncImage,
}

// SyncImages sync images
func (da *imageAdaptor) SyncImages(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if cfunc, ok := SyncFuncs[vendor]; !ok {
		return nil, fmt.Errorf("%s does not support the sync of cloud image", vendor)
	} else {
		return cfunc(da, cts)
	}
}
