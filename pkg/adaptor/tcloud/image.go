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

package tcloud

import (
	"fmt"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/image"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// ListImage 查询公共镜像列表
// reference: https://cloud.tencent.com/document/api/213/15715
func (t *TCloudImpl) ListImage(kt *kit.Kit,
	opt *image.TCloudImageListOption) (*image.TCloudImageListResult, error) {

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return nil, err
	}

	images := make([]image.TCloudImage, 0)

	req := cvm.NewDescribeImagesRequest()

	if len(opt.CloudIDs) != 0 {
		req.ImageIds = common.StringPtrs(opt.CloudIDs)
		req.Limit = common.Uint64Ptr(uint64(core.TCloudQueryLimit))
	}

	if opt.Page != nil {
		req.Offset = common.Uint64Ptr(opt.Page.Offset)
		req.Limit = common.Uint64Ptr(opt.Page.Limit)
		req.Filters = []*cvm.Filter{
			{
				Name:   common.StringPtr("image-type"),
				Values: common.StringPtrs([]string{"PUBLIC_IMAGE"}),
			},
		}
	}

	resp, err := client.DescribeImagesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud images failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list tcloud images failed, err: %v", err)
	}

	for _, pImage := range resp.Response.ImageSet {
		images = append(images, image.TCloudImage{
			CloudID:      *pImage.ImageId,
			Name:         *pImage.ImageName,
			State:        *pImage.ImageState,
			Platform:     *pImage.Platform,
			Architecture: changeArchitecture(pImage.Architecture),
			Type:         "public",
			ImageSize:    *pImage.ImageSize,
			ImageSource:  *pImage.ImageSource,
			OsType:       image.GetOsTypeByPlatform(enumor.TCloud, *pImage.Platform),
		})
	}

	return &image.TCloudImageListResult{Details: images}, nil
}

func changeArchitecture(architecture *string) string {
	if architecture == nil {
		return constant.X86
	}

	if *architecture == "arm" {
		return constant.Arm64
	}

	return *architecture
}
