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
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest"
)

// InitDiskService initial the disk service
func InitDiskService(cap *capability.Capability) {
	d := &service{
		Adaptor: cap.CloudAdaptor,
		DataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()

	// 硬盘创建
	h.Add("CreateTCloudDisk", http.MethodPost, "/vendors/tcloud/disks/create", d.CreateTCloudDisk)
	h.Add("CreateGcpDisk", http.MethodPost, "/vendors/gcp/disks/create", d.CreateGcpDisk)
	h.Add("CreateAzureDisk", http.MethodPost, "/vendors/azure/disks/create", d.CreateAzureDisk)
	h.Add("CreateHuaWeiDisk", http.MethodPost, "/vendors/huawei/disks/create", d.CreateHuaWeiDisk)
	h.Add("CreateAwsDisk", http.MethodPost, "/vendors/aws/disks/create", d.CreateAwsDisk)

	// 删除云盘
	h.Add("DeleteTCloudDisk", http.MethodDelete, "/vendors/tcloud/disks", d.DeleteTCloudDisk)
	h.Add("DeleteGcpDisk", http.MethodDelete, "/vendors/gcp/disks", d.DeleteGcpDisk)
	h.Add("DeleteAzureDisk", http.MethodDelete, "/vendors/azure/disks", d.DeleteAzureDisk)
	h.Add("DeleteHuaWeiDisk", http.MethodDelete, "/vendors/huawei/disks", d.DeleteHuaWeiDisk)
	h.Add("DeleteAwsDisk", http.MethodDelete, "/vendors/aws/disks", d.DeleteAwsDisk)

	// 挂载云盘
	h.Add("AttachTCloudDisk", http.MethodPost, "/vendors/tcloud/disks/attach", d.AttachTCloudDisk)
	h.Add("AttachGcpDisk", http.MethodPost, "/vendors/gcp/disks/attach", d.AttachGcpDisk)
	h.Add("AttachAzureDisk", http.MethodPost, "/vendors/azure/disks/attach", d.AttachAzureDisk)
	h.Add("AttachHuaWeiDisk", http.MethodPost, "/vendors/huawei/disks/attach", d.AttachHuaWeiDisk)
	h.Add("AttachAwsDisk", http.MethodPost, "/vendors/aws/disks/attach", d.AttachAwsDisk)

	// 卸载云盘
	h.Add("DetachTCloudDisk", http.MethodPost, "/vendors/tcloud/disks/detach", d.DetachTCloudDisk)
	h.Add("DetachGcpDisk", http.MethodPost, "/vendors/gcp/disks/detach", d.DetachGcpDisk)
	h.Add("DetachAzureDisk", http.MethodPost, "/vendors/azure/disks/detach", d.DetachAzureDisk)
	h.Add("DetachHuaWeiDisk", http.MethodPost, "/vendors/huawei/disks/detach", d.DetachHuaWeiDisk)
	h.Add("DetachAwsDisk", http.MethodPost, "/vendors/aws/disks/detach", d.DetachAwsDisk)

	// 询价
	h.Add("InquiryPriceTCloudDisk", http.MethodPost, "/vendors/tcloud/disks/prices/inquiry", d.InquiryPriceTCloudDisk)
	h.Add("InquiryPriceHuaWeiDisk", http.MethodPost, "/vendors/huawei/disks/prices/inquiry", d.InquiryPriceHuaWeiDisk)

	h.Load(cap.WebService)
}

type service struct {
	DataCli *dataservice.Client
	Adaptor *cloudclient.CloudAdaptorClient
}
