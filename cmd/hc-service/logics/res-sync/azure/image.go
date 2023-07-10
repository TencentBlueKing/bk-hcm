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
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesimage "hcm/pkg/adaptor/types/image"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	dateimage "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncImageOption ...
type SyncImageOption struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Publisher string `json:"publisher" validate:"required"`
	Offer     string `json:"offer" validate:"required"`
}

// Validate ...
func (opt SyncImageOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Image ...
func (cli *client) Image(kt *kit.Kit, opt *SyncImageOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageFromCloud, err := cli.listImageFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	imageFromDB, err := cli.listImageFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(imageFromCloud) == 0 && len(imageFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesimage.AzureImage, dateimage.ImageExtResult[dateimage.AzureImageExtensionResult]](
		imageFromCloud, imageFromDB, isImageChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteImage(kt, opt, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createImage(kt, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateImage(kt, opt, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) updateImage(kt *kit.Kit, opt *SyncImageOption,
	updateMap map[string]typesimage.AzureImage) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("image updateMap is <= 0, not update")
	}

	var updateReq dataproto.ImageExtBatchUpdateReq[dataproto.AzureImageExtensionUpdateReq]

	for id, one := range updateMap {
		image := &dataproto.ImageExtUpdateReq[dataproto.AzureImageExtensionUpdateReq]{
			ID:    id,
			State: one.State,
		}
		updateReq = append(updateReq, image)
	}

	if _, err := cli.dbCli.Azure.BatchUpdateImage(kt.Ctx, kt.Header(), &updateReq); err != nil {
		return err
	}

	logs.Infof("[%s] sync image to update image success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createImage(kt *kit.Kit, opt *SyncImageOption,
	addSlice []typesimage.AzureImage) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("cvm addSlice is <= 0, not create")
	}

	var createReq dataproto.ImageExtBatchCreateReq[dataproto.AzureImageExtensionCreateReq]

	for _, one := range addSlice {
		image := &dataproto.ImageExtCreateReq[dataproto.AzureImageExtensionCreateReq]{
			CloudID:      one.CloudID,
			Name:         one.Name,
			Architecture: one.Architecture,
			Platform:     one.Platform,
			State:        one.State,
			Type:         one.Type,
			Extension: &dataproto.AzureImageExtensionCreateReq{
				Region:    opt.Region,
				Publisher: opt.Publisher,
				Offer:     opt.Offer,
				Sku:       one.Sku,
			},
		}
		createReq = append(createReq, image)
	}

	_, err := cli.dbCli.Azure.BatchCreateImage(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		return err
	}

	logs.Infof("[%s] sync image to create image success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) deleteImage(kt *kit.Kit, opt *SyncImageOption, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("image delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncImageOption{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		Publisher: opt.Publisher,
		Offer:     opt.Offer,
	}
	delImageFromCloud, err := cli.listImageFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	canNotDelete := false
	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range delImageFromCloud {
		if _, exsit := delCloudMap[one.CloudID]; exsit {
			canNotDelete = true
			break
		}
	}

	if canNotDelete {
		logs.Errorf("[%s] validate image not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delImageFromCloud), kt.Rid)
		return fmt.Errorf("validate image not exist failed, before delete")
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		batchDeleteReq := &dataproto.ImageDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", parts),
		}
		if _, err := cli.dbCli.Global.DeleteImage(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
			logs.Errorf("request dataservice delete azure image failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync image to delete image success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listImageFromCloud(kt *kit.Kit, opt *SyncImageOption) ([]typesimage.AzureImage, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageOpt := &typesimage.AzureImageListOption{
		Region:    opt.Region,
		Publisher: opt.Publisher,
		Offer:     opt.Offer,
	}
	result, err := cli.cloudCli.ListImage(kt, imageOpt)
	if err != nil {
		logs.Errorf("[%s] list image from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listImageFromDB(kt *kit.Kit, opt *SyncImageOption) (
	[]dateimage.ImageExtResult[dateimage.AzureImageExtensionResult], error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &dataproto.ImageListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.Azure,
				},
				&filter.AtomRule{
					Field: "extension.region",
					Op:    filter.JSONEqual.Factory(),
					Value: opt.Region,
				},
				&filter.AtomRule{
					Field: "extension.publisher",
					Op:    filter.JSONEqual.Factory(),
					Value: opt.Publisher,
				},
				&filter.AtomRule{
					Field: "extension.offer",
					Op:    filter.JSONEqual.Factory(),
					Value: opt.Offer,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	start := uint32(0)
	results := make([]dateimage.ImageExtResult[dateimage.AzureImageExtensionResult], 0)
	for {
		req.Page.Start = start
		images, err := cli.dbCli.Azure.ListImage(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list image from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
				opt.AccountID, req, kt.Rid)
			return nil, err
		}

		results = append(results, converter.PtrToSlice(images.Details)...)

		if len(images.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return results, nil
}

func (cli *client) listImageFromDBForCvm(kt *kit.Kit, params *SyncBaseParams) (
	[]*dataproto.ImageResult, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &dataproto.ImageListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: params.CloudIDs},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Global.ListImage(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list image from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isImageChange(cloud typesimage.AzureImage, db dateimage.ImageExtResult[dateimage.AzureImageExtensionResult]) bool {

	if cloud.State != db.State {
		return true
	}

	return false
}
