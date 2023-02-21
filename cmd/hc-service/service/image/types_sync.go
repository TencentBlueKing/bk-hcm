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
	"hcm/pkg/adaptor/types/image"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCloudImageSync
type TCloudImageSync struct {
	IsUpdate bool
	Image    image.TCloudImage
}

// TCloudDSImageSync
type TCloudDSImageSync struct {
	Image *dataproto.ImageExtResult[dataproto.TCloudImageExtensionResult]
}

// AwsImageSync
type AwsImageSync struct {
	IsUpdate bool
	Image    image.AwsImage
}

// AwsDSImageSync
type AwsDSImageSync struct {
	Image *dataproto.ImageExtResult[dataproto.AwsImageExtensionResult]
}

// HuaWeiImageSync
type HuaWeiImageSync struct {
	IsUpdate bool
	Image    image.HuaWeiImage
}

// HuaWeiDSImageSync
type HuaWeiDSImageSync struct {
	Image *dataproto.ImageExtResult[dataproto.HuaWeiImageExtensionResult]
}

// GcpImageSync
type GcpImageSync struct {
	IsUpdate bool
	Image    image.GcpImage
}

// GcpDSImageSync
type GcpDSImageSync struct {
	Image *dataproto.ImageExtResult[dataproto.GcpImageExtensionResult]
}

// AzureImageSync
type AzureImageSync struct {
	IsUpdate bool
	Image    image.AzureImage
}

// AzureDSImageSync
type AzureDSImageSync struct {
	Image *dataproto.ImageExtResult[dataproto.AzureImageExtensionResult]
}

// AzurePublisherAndOffer
type AzurePublisherAndOffer struct {
	Publisher string
	Offer     string
}

// syncImageDelete for delete
func (da *imageAdaptor) syncImageDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	batchDeleteReq := &dataproto.ImageDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}
	if _, err := da.dataCli.Global.DeleteImage(cts.Kit.Ctx, cts.Kit.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud image failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}
