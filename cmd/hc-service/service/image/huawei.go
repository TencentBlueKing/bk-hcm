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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/image"
	apicore "hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	protoimage "hcm/pkg/api/hc-service/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2/model"
)

// HuaWeiSyncImage sync huawei to hcm
func HuaWeiSyncImage(da *imageAdaptor, cts *rest.Contexts) (interface{}, error) {
	req := new(protoimage.HuaWeiImageSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := da.idaptor.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, platform := range []model.ListImagesRequestPlatform{model.GetListImagesRequestPlatformEnum().WINDOWS,
		model.GetListImagesRequestPlatformEnum().CENT_OS} {
		cloudAllIDs := make(map[string]bool)
		limit := int32(core.HuaWeiQueryLimit)
		opt := &image.HuaWeiImageListOption{
			Region:   req.Region,
			Limit:    &limit,
			Marker:   nil,
			Platform: platform,
		}

		datas, err := client.ListImage(opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei Image failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(datas.Details) <= 0 {
			logs.Errorf("request adaptor to list huawei Image num <= 0, rid: %s", cts.Kit.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*HuaWeiImageSync)
		cloudIDs := make([]string, 0, len(datas.Details))

		for _, data := range datas.Details {
			imageSync := new(HuaWeiImageSync)
			imageSync.IsUpdate = false
			imageSync.Image = data
			cloudMap[data.CloudID] = imageSync
			cloudIDs = append(cloudIDs, data.CloudID)
			cloudAllIDs[data.CloudID] = true
		}

		updateIDs := make([]string, 0)
		dsMap := make(map[string]*HuaWeiDSImageSync)

		start := 0
		step := int(filter.DefaultMaxInLimit)
		for {
			var tmpCloudIDs []string
			if start+step > len(cloudIDs) {
				tmpCloudIDs = make([]string, len(cloudIDs)-start)
				copy(tmpCloudIDs, cloudIDs[start:])
			} else {
				tmpCloudIDs = make([]string, step)
				copy(tmpCloudIDs, cloudIDs[start:start+step])
			}

			if len(tmpCloudIDs) > 0 {
				tmpIDs, tmpMap, err := da.getHuaWeiImageDSSync(tmpCloudIDs, req, cts, platform.Value())
				if err != nil {
					logs.Errorf("request getHuaWeiImageDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
					return nil, err
				}

				updateIDs = append(updateIDs, tmpIDs...)
				for k, v := range tmpMap {
					dsMap[k] = v
				}
			}

			start = start + step
			if start > len(cloudIDs) {
				break
			}
		}

		if len(updateIDs) > 0 {
			err := da.syncHuaWeiImageUpdate(updateIDs, cloudMap, dsMap, cts)
			if err != nil {
				logs.Errorf("request syncHuaWeiImageUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}

		addIDs := make([]string, 0)
		for _, id := range updateIDs {
			if _, ok := cloudMap[id]; ok {
				cloudMap[id].IsUpdate = true
			}
		}

		for k, v := range cloudMap {
			if !v.IsUpdate {
				addIDs = append(addIDs, k)
			}
		}

		if len(addIDs) > 0 {
			err := da.syncHuaWeiImageAdd(addIDs, cts, req, cloudMap)
			if err != nil {
				logs.Errorf("request syncHuaWeiImageAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}

		dsIDs, err := da.getHuaWeiImageAllDS(req, cts, platform.Value())
		if err != nil {
			logs.Errorf("request getHuaWeiImageAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		deleteIDs := make([]string, 0)
		for _, id := range dsIDs {
			if _, ok := cloudAllIDs[id]; !ok {
				deleteIDs = append(deleteIDs, id)
			}
		}

		if len(deleteIDs) > 0 {
			realDeleteIDs := make([]string, 0)
			datas, err := client.ListImage(opt)
			if err != nil {
				logs.Errorf("request adaptor to list huawei Image failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas.Details {
					if data.CloudID == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			err = da.syncImageDelete(cts, realDeleteIDs)
			if err != nil {
				logs.Errorf("request syncImageDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func (da *imageAdaptor) syncHuaWeiImageAdd(addIDs []string, cts *rest.Contexts, req *protoimage.HuaWeiImageSyncReq,
	cloudMap map[string]*HuaWeiImageSync) error {
	var createReq dataproto.ImageExtBatchCreateReq[dataproto.HuaWeiImageExtensionCreateReq]

	for _, id := range addIDs {
		image := &dataproto.ImageExtCreateReq[dataproto.HuaWeiImageExtensionCreateReq]{
			CloudID:      id,
			Name:         cloudMap[id].Image.Name,
			Architecture: cloudMap[id].Image.Architecture,
			Platform:     cloudMap[id].Image.Platform,
			State:        cloudMap[id].Image.State,
			Type:         cloudMap[id].Image.Type,
			Extension: &dataproto.HuaWeiImageExtensionCreateReq{
				Region: req.Region,
			},
		}
		createReq = append(createReq, image)
	}

	if len(createReq) <= 0 {
		return nil
	}

	_, err := da.dataCli.HuaWei.BatchCreateImage(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create huawei image failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

func (da *imageAdaptor) syncHuaWeiImageUpdate(updateIDs []string, cloudMap map[string]*HuaWeiImageSync,
	dsMap map[string]*HuaWeiDSImageSync, cts *rest.Contexts) error {

	var updateReq dataproto.ImageExtBatchUpdateReq[dataproto.HuaWeiImageExtensionUpdateReq]

	for _, id := range updateIDs {
		if cloudMap[id].Image.State == dsMap[id].Image.State {
			continue
		}
		image := &dataproto.ImageExtUpdateReq[dataproto.HuaWeiImageExtensionUpdateReq]{
			ID:    dsMap[id].Image.ID,
			State: cloudMap[id].Image.State,
		}
		updateReq = append(updateReq, image)
	}

	if len(updateReq) > 0 {
		if _, err := da.dataCli.HuaWei.BatchUpdateImage(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateImage failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (da *imageAdaptor) getHuaWeiImageDSSync(cloudIDs []string, req *protoimage.HuaWeiImageSyncReq,
	cts *rest.Contexts, platform string) ([]string, map[string]*HuaWeiDSImageSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*HuaWeiDSImageSync)

	start := 0
	for {

		dataReq := &dataproto.ImageListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.HuaWei,
					},
					&filter.AtomRule{
						Field: "platform",
						Op:    filter.Equal.Factory(),
						Value: platform,
					},
					&filter.AtomRule{
						Field: "extension.region",
						Op:    filter.JSONEqual.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "cloud_id",
						Op:    filter.In.Factory(),
						Value: cloudIDs,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		results, err := da.dataCli.HuaWei.ListImage(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list image failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(HuaWeiDSImageSync)
				dsImageSync.Image = detail
				dsMap[detail.CloudID] = dsImageSync
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}

	}

	return updateIDs, dsMap, nil
}

func (da *imageAdaptor) getHuaWeiImageAllDS(req *protoimage.HuaWeiImageSyncReq,
	cts *rest.Contexts, platform string) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {

		dataReq := &dataproto.ImageListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.HuaWei,
					},
					&filter.AtomRule{
						Field: "platform",
						Op:    filter.Equal.Factory(),
						Value: platform,
					},
					&filter.AtomRule{
						Field: "extension.region",
						Op:    filter.JSONEqual.Factory(),
						Value: req.Region,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		results, err := da.dataCli.HuaWei.ListImage(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list image failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return dsIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsIDs = append(dsIDs, detail.CloudID)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsIDs, nil
}
