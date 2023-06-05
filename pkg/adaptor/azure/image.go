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

package azure

import (
	"strings"

	"hcm/pkg/adaptor/types/image"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

// ListImage 查询公共镜像列表
// reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machine-images/list?tabs=HTTP
func (a *Azure) ListImage(kt *kit.Kit,
	opt *image.AzureImageListOption) (*image.AzureImageListResult, error) {

	client, err := a.clientSet.imageClient()
	if err != nil {
		return nil, err
	}

	resSKU, err := client.ListSKUs(kt.Ctx, opt.Region, opt.Publisher, opt.Offer, nil)
	if err != nil {
		logs.Errorf("failed to ListSKUs, err: %v, rid: %s", err, kt.Rid)
	}

	images := make([]image.AzureImage, 0)
	for _, sku := range resSKU.VirtualMachineImageResourceArray {

		res, err := client.List(kt.Ctx, opt.Region, opt.Publisher, opt.Offer, *sku.Name,
			&armcompute.VirtualMachineImagesClientListOptions{})
		if err != nil {
			logs.Errorf("failed to List, err: %v, rid: %s", err, kt.Rid)
		}

		for _, pImage := range res.VirtualMachineImageResourceArray {
			images = append(images, image.AzureImage{
				CloudID:      SPtrToLowerStr(pImage.ID),
				Name:         converter.PtrToVal(pImage.Name),
				Architecture: changeArchitecture(sku.Name),
				Platform:     opt.Offer,
				Sku:          converter.PtrToVal(sku.Name),
				State:        "available",
				Type:         "public",
			})
		}
	}

	return &image.AzureImageListResult{Details: images}, nil
}

func changeArchitecture(sku *string) string {
	if sku == nil {
		return constant.X86
	}

	if strings.Contains(*sku, "arm64") {
		return constant.Arm64
	}

	return constant.X86
}
