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

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	imagedao "hcm/pkg/dal/dao/cloud/image"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud/image"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

type imageSvc struct {
	dao.Set
	objectDao *imagedao.ImageDao
}

// Init 注册 imagedao.ImageDao 到 Capability.Dao, 并设置 objectDao
func (svc *imageSvc) Init() {
	d := &imagedao.ImageDao{}
	registeredDao := svc.GetObjectDao(d.Name())
	if registeredDao == nil {
		d.ObjectDaoManager = new(dao.ObjectDaoManager)
		svc.RegisterObjectDao(d)
	}

	svc.objectDao = svc.GetObjectDao(d.Name()).(*imagedao.ImageDao)
}

// BatchCreateImageExt ...
func (svc *imageSvc) BatchCreateImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchCreateImageExt[dataproto.TCloudImageExtensionCreateReq](cts, svc, vendor)
	case enumor.Aws:
		return batchCreateImageExt[dataproto.AwsImageExtensionCreateReq](cts, svc, vendor)
	case enumor.Gcp:
		return batchCreateImageExt[dataproto.GcpImageExtensionCreateReq](cts, svc, vendor)
	case enumor.HuaWei:
		return batchCreateImageExt[dataproto.HuaWeiImageExtensionCreateReq](cts, svc, vendor)
	case enumor.Azure:
		return batchCreateImageExt[dataproto.AzureImageExtensionCreateReq](cts, svc, vendor)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// RetrieveImageExt ...
func (svc *imageSvc) RetrieveImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageID := cts.PathParameter("id").String()
	opt := &types.ListOption{
		Filter: tools.EqualWithOpExpression(
			filter.And,
			map[string]interface{}{"id": imageID, "vendor": string(vendor)},
		),
		Page: &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if count := len(data.Details); count != 1 {
		return nil, fmt.Errorf("retrieve image failed: query id(%s) return total %d", imageID, count)
	}

	imageData := data.Details[0]
	switch vendor {
	case enumor.TCloud:
		return toProtoImageExtResult[dataproto.TCloudImageExtensionResult](imageData)
	case enumor.Aws:
		return toProtoImageExtResult[dataproto.AwsImageExtensionResult](imageData)
	case enumor.Gcp:
		return toProtoImageExtResult[dataproto.GcpImageExtensionResult](imageData)
	case enumor.Azure:
		return toProtoImageExtResult[dataproto.AzureImageExtensionResult](imageData)
	case enumor.HuaWei:
		return toProtoImageExtResult[dataproto.HuaWeiImageExtensionResult](imageData)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// ListImage ...
func (svc *imageSvc) ListImage(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.ImageListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*dataproto.ImageResult, len(data.Details))
	for indx, d := range data.Details {
		details[indx] = toProtoImageResult(d)
	}

	return &dataproto.ImageListResult{Details: details, Count: data.Count}, nil
}

// ListImageExt ...
func (svc *imageSvc) ListImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.ImageListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoImageExtListResult[dataproto.TCloudImageExtensionResult](data)
	case enumor.Aws:
		return toProtoImageExtListResult[dataproto.AwsImageExtensionResult](data)
	case enumor.Gcp:
		return toProtoImageExtListResult[dataproto.GcpImageExtensionResult](data)
	case enumor.HuaWei:
		return toProtoImageExtListResult[dataproto.HuaWeiImageExtensionResult](data)
	case enumor.Azure:
		return toProtoImageExtListResult[dataproto.AzureImageExtensionResult](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// BatchUpdateImageExt ...
func (svc *imageSvc) BatchUpdateImageExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchUpdateImageExt[dataproto.TCloudImageExtensionUpdateReq](cts, svc)
	case enumor.Aws:
		return batchUpdateImageExt[dataproto.AwsImageExtensionUpdateReq](cts, svc)
	case enumor.Gcp:
		return batchUpdateImageExt[dataproto.GcpImageExtensionUpdateReq](cts, svc)
	case enumor.HuaWei:
		return batchUpdateImageExt[dataproto.HuaWeiImageExtensionUpdateReq](cts, svc)
	case enumor.Azure:
		return batchUpdateImageExt[dataproto.AzureImageExtensionUpdateReq](cts, svc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// BatchDeleteImage ...
func (svc *imageSvc) BatchDeleteImage(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.ImageDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return nil, svc.objectDao.DeleteWithTx(cts.Kit, txn, req.Filter)
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// rawExtensions 根据条件查询原始的 extension 字段, 返回字典结构 {"镜像 ID": "原始的 extension 字段"}
// TODO 不同资源可以复用 rawExtensions 逻辑
func (svc *imageSvc) rawExtensions(
	cts *rest.Contexts,
	filterExp *filter.Expression,
) (map[string]tabletype.JsonField, error) {
	opt := &types.ListOption{
		Filter: filterExp,
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
		Fields: []string{"id", "extension"},
	}
	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	extensions := make(map[string]tabletype.JsonField)
	for _, d := range data.Details {
		extensions[d.ID] = d.Extension
	}

	return extensions, nil
}

func batchCreateImageExt[T dataproto.ImageExtensionCreateReq](
	cts *rest.Contexts,
	svc *imageSvc,
	vendor enumor.Vendor,
) (interface{}, error) {
	req := new(dataproto.ImageExtBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	imageIDs, err := svc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		images := make([]*tablecloud.ImageModel, len(*req))

		for indx, imageReq := range *req {
			extensionJson, err := json.MarshalToString(imageReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}
			images[indx] = &tablecloud.ImageModel{
				Vendor:       string(vendor),
				CloudID:      imageReq.CloudID,
				Name:         imageReq.Name,
				Architecture: imageReq.Architecture,
				Platform:     imageReq.Platform,
				State:        imageReq.State,
				Type:         imageReq.Type,
				Extension:    tabletype.JsonField(extensionJson),
				Creator:      cts.Kit.User,
				Reviser:      cts.Kit.User,
			}
		}
		return svc.objectDao.BatchCreateWithTx(cts.Kit, txn, images)
	})
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: imageIDs.([]string)}, nil
}

func batchUpdateImageExt[T dataproto.ImageExtensionUpdateReq](
	cts *rest.Contexts,
	svc *imageSvc,
) (interface{}, error) {
	req := new(dataproto.ImageExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	queryIDs := make([]string, len(*req))
	for indx, imageReq := range *req {
		queryIDs[indx] = imageReq.ID
	}
	rawExtensions, err := svc.rawExtensions(cts, tools.ContainersExpression("id", queryIDs))
	if err != nil {
		return nil, err
	}

	_, err = svc.Set.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, imageReq := range *req {
			updateData := &tablecloud.ImageModel{
				State: imageReq.State,
			}

			if imageReq.Extension != nil {
				rawExtension, exist := rawExtensions[imageReq.ID]
				if !exist {
					return nil, fmt.Errorf("image id (%s) not exit", imageReq.ID)
				}
				mergedExtension, err := json.UpdateMerge(imageReq.Extension, string(rawExtension))
				if err != nil {
					return nil, fmt.Errorf("image id (%s) merge extension failed, err: %v", imageReq.ID, err)
				}
				updateData.Extension = tabletype.JsonField(mergedExtension)
			}

			if err := svc.objectDao.UpdateByIDWithTx(cts.Kit, txn, imageReq.ID, updateData); err != nil {
				return nil, fmt.Errorf("update image failed, err: %v", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
