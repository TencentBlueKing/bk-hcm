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
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesimage "hcm/pkg/adaptor/types/image"
	"hcm/pkg/api/core"
	coreimage "hcm/pkg/api/core/cloud/image"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/model"
)

// SyncImageOption ...
type SyncImageOption struct {
	Platform model.ListImagesRequestPlatform `json:"platform" validate:"required"`
}

// Validate ...
func (opt SyncImageOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Image ...
func (cli *client) Image(kt *kit.Kit, params *SyncBaseParams, opt *SyncImageOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageFromCloud, err := cli.listImageFromCloud(kt, params, opt.Platform)
	if err != nil {
		return nil, err
	}

	imageFromDB, err := cli.listImageFromDB(kt, params, opt.Platform)
	if err != nil {
		return nil, err
	}

	if len(imageFromCloud) == 0 && len(imageFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesimage.HuaWeiImage, coreimage.Image[coreimage.HuaWeiExtension]](
		imageFromCloud, imageFromDB, isImageChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteImage(kt, params.AccountID, params.Region, delCloudIDs, opt.Platform); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createImage(kt, params.AccountID, params.Region, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateImage(kt, params.AccountID, params.Region, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// updateImage 更新镜像
func (cli *client) updateImage(kt *kit.Kit, accountID string, region string,
	updateMap map[string]typesimage.HuaWeiImage) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("image updateMap is <= 0, not update")
	}

	items := make([]dataproto.ImageUpdate[coreimage.HuaWeiExtension], 0, len(updateMap))

	for id, one := range updateMap {
		image := dataproto.ImageUpdate[coreimage.HuaWeiExtension]{
			ID:     id,
			State:  one.State,
			OsType: one.OsType,
		}
		items = append(items, image)
	}

	updateReq := &dataproto.BatchUpdateReq[coreimage.HuaWeiExtension]{
		Items: items,
	}
	if _, err := cli.dbCli.HuaWei.BatchUpdateImage(kt, updateReq); err != nil {
		return err
	}

	logs.Infof("[%s] sync image to update image success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// createImage 创建镜像
func (cli *client) createImage(kt *kit.Kit, accountID string, region string,
	addSlice []typesimage.HuaWeiImage) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	items := make([]dataproto.ImageCreate[coreimage.HuaWeiExtension], 0, len(addSlice))

	for _, one := range addSlice {
		image := dataproto.ImageCreate[coreimage.HuaWeiExtension]{
			CloudID:      one.CloudID,
			Name:         one.Name,
			Architecture: one.Architecture,
			Platform:     one.Platform,
			State:        one.State,
			Type:         one.Type,
			OsType:       one.OsType,
			Extension: &coreimage.HuaWeiExtension{
				Region: region,
			},
		}
		items = append(items, image)
	}

	createReq := &dataproto.BatchCreateReq[coreimage.HuaWeiExtension]{
		Items: items,
	}
	_, err := cli.dbCli.HuaWei.BatchCreateImage(kt, createReq)
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync image to create image success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(addSlice), kt.Rid)

	return nil
}

// deleteImage delete image in db and validate image not exist in cloud
func (cli *client) deleteImage(kt *kit.Kit, accountID string, region string, delCloudIDs []string,
	platform model.ListImagesRequestPlatform) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("image delCloudIDs is <= 0, not delete")
	}

	// validate image not exist in cloud before delete
	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delImageFromCloud, err := cli.listImageFromCloud(kt, checkParams, platform)
	if err != nil {
		return err
	}

	if len(delImageFromCloud) > 0 {
		logs.Errorf("[%s] validate image not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.HuaWei, checkParams, len(delImageFromCloud), kt.Rid)
		return fmt.Errorf("validate image not exist failed, before delete")
	}

	batchDeleteReq := &dataproto.DeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.DeleteImage(kt, batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete huawei image failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync image to delete image success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// listImageFromCloud lists images from cloud
func (cli *client) listImageFromCloud(kt *kit.Kit, params *SyncBaseParams,
	platform model.ListImagesRequestPlatform) ([]typesimage.HuaWeiImage, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	results := make([]typesimage.HuaWeiImage, 0)

	for _, id := range params.CloudIDs {
		opt := &typesimage.HuaWeiImageListOption{
			Region:   params.Region,
			Platform: platform,
			CloudID:  id,
		}
		image, err := cli.cloudCli.ListImage(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list image from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
				err, params.AccountID, opt, kt.Rid)
			return nil, err
		}
		results = append(results, image.Details...)
	}

	return results, nil
}

// listImageFromDB lists images from database
func (cli *client) listImageFromDB(kt *kit.Kit, params *SyncBaseParams, platform model.ListImagesRequestPlatform) (
	[]coreimage.Image[coreimage.HuaWeiExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.HuaWei,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "extension.region",
					Op:    filter.JSONEqual.Factory(),
					Value: params.Region,
				},
				&filter.AtomRule{
					Field: "platform",
					Op:    filter.Equal.Factory(),
					Value: platform.Value(),
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	images, err := cli.dbCli.HuaWei.ListImage(kt, req)
	if err != nil {
		logs.Errorf("[%s] list image from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	results := make([]coreimage.Image[coreimage.HuaWeiExtension], 0, len(images.Details))
	for _, one := range images.Details {
		results = append(results, converter.PtrToVal(one))
	}

	return results, nil
}

// RemoveImageDeleteFromCloud remove image from cloud
func (cli *client) RemoveImageDeleteFromCloud(kt *kit.Kit, accountID, region string,
	platform model.ListImagesRequestPlatform) error {

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: enumor.HuaWei},
				&filter.AtomRule{Field: "platform", Op: filter.Equal.Factory(), Value: platform.Value()},
				&filter.AtomRule{Field: "extension.region", Op: filter.JSONEqual.Factory(), Value: region},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.HuaWei.ListImage(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list image failed, err: %v, req: %v, rid: %s", enumor.HuaWei,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listImageFromCloud(kt, params, platform)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.CloudID)
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteImage(kt, accountID, region, cloudIDs, platform); err != nil {
					return err
				}
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

// isImageChange checks if the image has changed based on its state and os type
func isImageChange(cloud typesimage.HuaWeiImage, db coreimage.Image[coreimage.HuaWeiExtension]) bool {

	if cloud.State != db.State {
		return true
	}

	if cloud.OsType != db.OsType {
		return true
	}

	return false
}

// listImageFromDBForCvm lists images from database for CVM
func (cli *client) listImageFromDBForCvm(kt *kit.Kit, params *SyncBaseParams) (
	[]*coreimage.BaseImage, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "extension.region", Op: filter.JSONEqual.Factory(), Value: params.Region},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: params.CloudIDs},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Global.ListImage(kt, req)
	if err != nil {
		logs.Errorf("[%s] list image from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}
