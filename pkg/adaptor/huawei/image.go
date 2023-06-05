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

package huawei

import (
	"hcm/pkg/adaptor/types/image"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/region"
)

// PublicImagePlatforms 公有镜像平台类型
var PublicImagePlatforms = []model.ListImagesRequestPlatform{model.GetListImagesRequestPlatformEnum().WINDOWS,
	model.GetListImagesRequestPlatformEnum().CENT_OS}

// ListImage 查询公共镜像列表
// reference: https://support.huaweicloud.com/api-ims/ims_03_0602.html
func (h *HuaWei) ListImage(kt *kit.Kit, opt *image.HuaWeiImageListOption) (*image.HuaWeiImageListResult, error) {

	client, err := h.clientSet.imsClientV2(region.ValueOf(opt.Region))
	if err != nil {
		return nil, err
	}

	publicType := model.GetListImagesRequestImagetypeEnum().GOLD
	status := model.GetListImagesRequestStatusEnum().ACTIVE
	req := &model.ListImagesRequest{
		Imagetype: &publicType,
		Platform:  &opt.Platform,
		Status:    &status,
	}

	if opt.CloudID != "" {
		req.Id = converter.ValToPtr(opt.CloudID)
	}

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	resp, err := client.ListImages(req)
	if err != nil {
		return nil, err
	}

	images := make([]image.HuaWeiImage, 0)
	for _, pImage := range *resp.Images {
		images = append(images, image.HuaWeiImage{
			CloudID:      pImage.Id,
			Name:         pImage.Name,
			Architecture: changeArchitecture(pImage.OsBit),
			Platform:     pImage.Platform.Value(),
			State:        model.GetListImagesRequestStatusEnum().ACTIVE.Value(),
			Type:         "public",
		})
	}
	return &image.HuaWeiImageListResult{Details: images}, nil
}

func changeArchitecture(osBit *model.ImageInfoOsBit) string {
	if osBit == nil {
		return constant.X86
	}

	if osBit.Value() == "64" {
		return constant.X86
	}

	return osBit.Value()
}
