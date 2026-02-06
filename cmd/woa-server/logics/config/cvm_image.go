/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	coreimage "hcm/pkg/api/core/cloud/image"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// CvmImageIf provides management interface for operations of cvm image config
type CvmImageIf interface {
	// GetCvmImage get cvm image type config list
	GetCvmImage(kt *kit.Kit, param *types.GetCvmImageParam) (*types.GetCvmImageResult, error)

	// BatchEnableImageCvm enables CVM functionality for images in batch
	BatchEnableImageCvm(kt *kit.Kit, imageIDs []string) error
	// BatchDisableImageCvm disables CVM functionality for images in batch
	BatchDisableImageCvm(kt *kit.Kit, imageIDs []string) error
}

// NewCvmImageOp creates a cvm image interface
func NewCvmImageOp(client *client.ClientSet) CvmImageIf {
	return &cvmImage{
		client: client,
	}
}

type cvmImage struct {
	client *client.ClientSet
}

// GetCvmImage get cvm image type config list
func (i *cvmImage) GetCvmImage(kt *kit.Kit, param *types.GetCvmImageParam) (*types.GetCvmImageResult, error) {
	// 构建查询条件
	rules := []*filter.AtomRule{
		tools.RuleEqual("vendor", enumor.TCloudZiyan),
		tools.RuleJSONEqual("extension.enable_cvm", "true"),
	}

	// 如果指定了 region，添加 region 过滤
	if len(param.Region) > 0 {
		rules = append(rules, tools.RuleIn("region", param.Region))
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(rules...),
		Page:   core.NewDefaultBasePage(),
	}

	imageList := make([]*types.CvmImage, 0)
	for {
		images, err := i.client.DataService().TCloudZiyan.ListImage(kt, req)
		if err != nil {
			logs.Errorf("failed to list images from data-service, err: %v, param: %+v, rid: %s", err, param, kt.Rid)
			return nil, fmt.Errorf("list images failed, err: %v", err)
		}

		// 转换为返回格式
		for _, image := range images.Details {
			if image == nil {
				continue
			}

			imageItem := &types.CvmImage{
				Region:    image.Region,
				ImageId:   image.CloudID,
				ImageName: image.Name,
			}
			imageList = append(imageList, imageItem)
		}

		if len(images.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	rst := &types.GetCvmImageResult{
		Count: int64(len(imageList)),
		Info:  imageList,
	}

	return rst, nil
}

// BatchEnableImageCvm enables CVM functionality for images in batch
func (i *cvmImage) BatchEnableImageCvm(kt *kit.Kit, imageIDs []string) error {
	return i.batchUpdateImageCvm(kt, enumor.TCloudZiyan, imageIDs, true)
}

// BatchDisableImageCvm disables CVM functionality for images in batch
func (i *cvmImage) BatchDisableImageCvm(kt *kit.Kit, imageIDs []string) error {
	return i.batchUpdateImageCvm(kt, enumor.TCloudZiyan, imageIDs, false)
}

// batchUpdateImageCvm 批量更新镜像的 enable_cvm 字段
func (i *cvmImage) batchUpdateImageCvm(kt *kit.Kit, vendor enumor.Vendor, imageIDs []string, enable bool) error {
	if len(imageIDs) == 0 {
		return nil
	}

	if err := vendor.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 批量查询镜像现有extension
	imageMap := make(map[string]*coreimage.Image[coreimage.TCloudZiyanExtension])
	for _, batch := range slice.Split(imageIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleIn("id", batch),
			),
			Page: core.NewDefaultBasePage(),
		}

		images, err := i.client.DataService().TCloudZiyan.ListImage(kt, req)
		if err != nil {
			logs.Errorf("failed to list images, err: %v, imageIDs: %v, rid: %s", err, imageIDs, kt.Rid)
			return fmt.Errorf("list images failed, err: %v", err)
		}

		if len(images.Details) == 0 {
			return errf.Newf(errf.RecordNotFound, "no images found for ids: %v", imageIDs)
		}

		for _, image := range images.Details {
			if image != nil {
				imageMap[image.ID] = image
			}
		}
	}

	// 构建更新项
	updateItems := make([]dataproto.ImageUpdate[coreimage.TCloudZiyanExtension], 0, len(imageIDs))
	for _, imageID := range imageIDs {
		image, ok := imageMap[imageID]
		if !ok {
			logs.Errorf("image %s not found in db, rid: %s", imageID, kt.Rid)
			return errf.Newf(errf.RecordNotFound, "image %s not found", imageID)
		}

		// 解析当前的 extension
		var currentExt coreimage.TCloudZiyanExtension
		if image.Extension != nil {
			currentExt = *image.Extension
		}

		// 更新 enable_cvm 字段
		currentExt.EnableCvm = enable
		updateItems = append(updateItems, dataproto.ImageUpdate[coreimage.TCloudZiyanExtension]{
			ID:        imageID,
			Extension: &currentExt,
		})
	}

	// 调用 data-service 的批量 update 接口
	for _, batch := range slice.Split(updateItems, constant.BatchOperationMaxLimit) {
		updateReq := &dataproto.BatchUpdateReq[coreimage.TCloudZiyanExtension]{
			Items: batch,
		}
		if _, err := i.client.DataService().TCloudZiyan.BatchUpdateImage(kt, updateReq); err != nil {
			logs.Errorf("failed to batch update image enable_cvm, err: %v, imageIDs: %v, rid: %s", err, imageIDs,
				kt.Rid)
			return fmt.Errorf("batch update image enable_cvm failed, err: %v", err)
		}
	}
	return nil
}
