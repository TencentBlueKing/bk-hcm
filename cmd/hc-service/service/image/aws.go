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
	"hcm/pkg/tools/converter"
)

// AwsSyncImage sync aws to hcm
func AwsSyncImage(da *imageAdaptor, cts *rest.Contexts) (interface{}, error) {
	req := new(protoimage.AwsImageSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get cloud data by page
	client, err := da.idaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	cloudAllIDs := make(map[string]bool)
	for {
		opt := &image.AwsImageListOption{
			Region: req.Region,
			Page:   &core.AwsPage{MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit))},
		}

		if nextToken != "" {
			opt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		datas, err := client.ListImage(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws Image failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(datas.Details) <= 0 {
			logs.Errorf("request adaptor to list aws Image num <= 0, rid: %s", cts.Kit.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*AwsImageSync)
		cloudIDs := make([]string, 0, len(datas.Details))
		for _, data := range datas.Details {
			imageSync := new(AwsImageSync)
			imageSync.IsUpdate = false
			imageSync.Image = data
			cloudMap[data.CloudID] = imageSync
			cloudIDs = append(cloudIDs, data.CloudID)
			cloudAllIDs[data.CloudID] = true
		}

		updateIDs, dsMap, err := da.getAwsImageDSSync(cloudIDs, req, cts)
		if err != nil {
			logs.Errorf("request getImageDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := da.syncAwsImageUpdate(updateIDs, cloudMap, dsMap, cts)
			if err != nil {
				logs.Errorf("request syncImageUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
			err := da.syncAwsImageAdd(addIDs, cts, req, cloudMap)
			if err != nil {
				logs.Errorf("request syncAwsImageAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}

		if datas.NextToken == nil {
			break
		}
		nextToken = *datas.NextToken
	}

	dsIDs, err := da.getAwsImageAllDS(req, cts)
	if err != nil {
		logs.Errorf("request getTAwsImageAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		nextToken := ""
		for {
			opt := &image.AwsImageListOption{
				Region: req.Region,
				Page:   &core.AwsPage{MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit))},
			}
			if nextToken != "" {
				opt.Page.NextToken = converter.ValToPtr(nextToken)
			}

			datas, err := client.ListImage(cts.Kit, opt)
			if err != nil {
				logs.Errorf("request adaptor to list aws Image failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

			if datas.NextToken == nil {
				break
			}
			nextToken = *datas.NextToken
		}

		err := da.syncImageDelete(cts, realDeleteIDs)
		if err != nil {
			logs.Errorf("request syncImageDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (da *imageAdaptor) syncAwsImageAdd(addIDs []string, cts *rest.Contexts, req *protoimage.AwsImageSyncReq,
	cloudMap map[string]*AwsImageSync) error {
	var createReq dataproto.ImageExtBatchCreateReq[dataproto.AwsImageExtensionCreateReq]

	for _, id := range addIDs {
		image := &dataproto.ImageExtCreateReq[dataproto.AwsImageExtensionCreateReq]{
			CloudID:      id,
			Name:         cloudMap[id].Image.Name,
			Architecture: cloudMap[id].Image.Architecture,
			Platform:     cloudMap[id].Image.Platform,
			State:        cloudMap[id].Image.State,
			Type:         cloudMap[id].Image.Type,
			Extension: &dataproto.AwsImageExtensionCreateReq{
				Region: req.Region,
			},
		}
		createReq = append(createReq, image)
	}

	if len(createReq) <= 0 {
		return nil
	}

	_, err := da.dataCli.Aws.BatchCreateImage(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create aws image failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

func (da *imageAdaptor) syncAwsImageUpdate(updateIDs []string, cloudMap map[string]*AwsImageSync,
	dsMap map[string]*AwsDSImageSync, cts *rest.Contexts) error {

	var updateReq dataproto.ImageExtBatchUpdateReq[dataproto.AwsImageExtensionUpdateReq]

	for _, id := range updateIDs {
		if !isAwsImageChange(cloudMap[id].Image, dsMap[id].Image) {
			continue
		}

		// TODO: 添加其他字段更新能力
		image := &dataproto.ImageExtUpdateReq[dataproto.AwsImageExtensionUpdateReq]{
			ID:    dsMap[id].Image.ID,
			State: cloudMap[id].Image.State,
		}
		updateReq = append(updateReq, image)
	}

	if len(updateReq) > 0 {
		if _, err := da.dataCli.Aws.BatchUpdateImage(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateImage failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

// isAwsImageChange ...
func isAwsImageChange(cloud image.AwsImage, db *dataproto.ImageExtResult[dataproto.AwsImageExtensionResult]) bool {
	if cloud.State != db.State {
		return true
	}

	return false
}

func (da *imageAdaptor) getAwsImageDSSync(cloudIDs []string, req *protoimage.AwsImageSyncReq,
	cts *rest.Contexts) ([]string, map[string]*AwsDSImageSync, error) {

	if len(cloudIDs) > int(apicore.DefaultMaxPageLimit) {
		return nil, nil, fmt.Errorf("list aws image cloudIDs should <= %d", apicore.DefaultMaxPageLimit)
	}

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*AwsDSImageSync)

	dataReq := &dataproto.ImageListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.Aws,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: cloudIDs,
				},
				&filter.AtomRule{
					Field: "extension.region",
					Op:    filter.JSONEqual.Factory(),
					Value: req.Region,
				},
			},
		},
		Page: apicore.DefaultBasePage,
	}

	results, err := da.dataCli.Aws.ListImage(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
	if err != nil {
		logs.Errorf("from data-service list image failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return updateIDs, dsMap, err
	}

	if len(results.Details) > 0 {
		for _, detail := range results.Details {
			updateIDs = append(updateIDs, detail.CloudID)
			dsImageSync := new(AwsDSImageSync)
			dsImageSync.Image = detail
			dsMap[detail.CloudID] = dsImageSync
		}
	}

	return updateIDs, dsMap, nil
}

func (da *imageAdaptor) getAwsImageAllDS(req *protoimage.AwsImageSyncReq,
	cts *rest.Contexts) ([]string, error) {

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
						Value: enumor.Aws,
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

		results, err := da.dataCli.Aws.ListImage(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
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
