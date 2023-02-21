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
	apicore "hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	protoimage "hcm/pkg/api/hc-service/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// GcpSyncImage sync gcp to hcm
func GcpSyncImage(da *imageAdaptor, cts *rest.Contexts) (interface{}, error) {
	req := new(protoimage.GcpImageSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := da.idaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, projectID := range []string{"centos-cloud", "windows-cloud"} {
		cloudAllIDs := make(map[string]bool)

		nextToken := ""
		for {
			opt := &image.GcpImageListOption{
				NextPageToken: nextToken,
				ProjectID:     projectID,
			}

			if nextToken != "" {
				opt.NextPageToken = nextToken
			}

			datas, token, err := client.ListImage(cts.Kit, opt)
			if err != nil {
				logs.Errorf("request adaptor to list gcp Image failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}

			if len(datas.Details) <= 0 {
				logs.Errorf("request adaptor to list gcp Image num <= 0, rid: %s", cts.Kit.Rid)
				return nil, err
			}

			cloudMap := make(map[string]*GcpImageSync)
			cloudIDs := make([]string, 0, len(datas.Details))
			for _, data := range datas.Details {
				imageSync := new(GcpImageSync)
				imageSync.IsUpdate = false
				imageSync.Image = data
				cloudMap[data.CloudID] = imageSync
				cloudIDs = append(cloudIDs, data.CloudID)
				cloudAllIDs[data.CloudID] = true
			}

			updateIDs, dsMap, err := da.getGcpImageDSSync(cloudIDs, req, cts, projectID)
			if err != nil {
				logs.Errorf("request getGcpImageDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}

			if len(updateIDs) > 0 {
				err := da.syncGcpImageUpdate(updateIDs, cloudMap, dsMap, cts)
				if err != nil {
					logs.Errorf("request syncGcpImageUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
				err := da.syncGcpImageAdd(addIDs, cts, req, cloudMap, projectID)
				if err != nil {
					logs.Errorf("request syncImageAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
					return nil, err
				}
			}

			if len(token) == 0 {
				break
			}

			nextToken = token
		}

		dsIDs, err := da.getGcpImageAllDS(req, cts, projectID)
		if err != nil {
			logs.Errorf("request getGcpImageAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
				opt := &image.GcpImageListOption{
					NextPageToken: nextToken,
					ProjectID:     projectID,
				}

				if nextToken != "" {
					opt.NextPageToken = nextToken
				}

				datas, token, err := client.ListImage(cts.Kit, opt)
				if err != nil {
					logs.Errorf("request adaptor to list gcp Image failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

				if len(token) == 0 {
					break
				}
				nextToken = token
			}

			err := da.syncImageDelete(cts, realDeleteIDs)
			if err != nil {
				logs.Errorf("request syncImageDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func (da *imageAdaptor) syncGcpImageAdd(addIDs []string, cts *rest.Contexts, req *protoimage.GcpImageSyncReq,
	cloudMap map[string]*GcpImageSync, projectID string) error {
	var createReq dataproto.ImageExtBatchCreateReq[dataproto.GcpImageExtensionCreateReq]

	for _, id := range addIDs {
		image := &dataproto.ImageExtCreateReq[dataproto.GcpImageExtensionCreateReq]{
			CloudID:      id,
			Name:         cloudMap[id].Image.Name,
			Architecture: cloudMap[id].Image.Architecture,
			Platform:     cloudMap[id].Image.Platform,
			State:        cloudMap[id].Image.State,
			Type:         cloudMap[id].Image.Type,
			Extension: &dataproto.GcpImageExtensionCreateReq{
				Region:    req.Region,
				ProjectID: projectID,
			},
		}
		createReq = append(createReq, image)
	}

	if len(createReq) <= 0 {
		return nil
	}

	_, err := da.dataCli.Gcp.BatchCreateImage(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create gcp image failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

func (da *imageAdaptor) syncGcpImageUpdate(updateIDs []string, cloudMap map[string]*GcpImageSync,
	dsMap map[string]*GcpDSImageSync, cts *rest.Contexts) error {

	var updateReq dataproto.ImageExtBatchUpdateReq[dataproto.GcpImageExtensionUpdateReq]

	for _, id := range updateIDs {
		if cloudMap[id].Image.State == dsMap[id].Image.State {
			continue
		}
		image := &dataproto.ImageExtUpdateReq[dataproto.GcpImageExtensionUpdateReq]{
			ID:    dsMap[id].Image.ID,
			State: cloudMap[id].Image.State,
		}
		updateReq = append(updateReq, image)
	}

	if len(updateReq) > 0 {
		if _, err := da.dataCli.Gcp.BatchUpdateImage(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateImage failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (da *imageAdaptor) getGcpImageDSSync(cloudIDs []string, req *protoimage.GcpImageSyncReq,
	cts *rest.Contexts, projectID string) ([]string, map[string]*GcpDSImageSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*GcpDSImageSync)

	start := 0
	for {

		dataReq := &dataproto.ImageListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.Gcp,
					},
					&filter.AtomRule{
						Field: "extension.region",
						Op:    filter.JSONEqual.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "extension.project_id",
						Op:    filter.JSONEqual.Factory(),
						Value: projectID,
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

		results, err := da.dataCli.Gcp.ListImage(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list image failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(GcpDSImageSync)
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

func (da *imageAdaptor) getGcpImageAllDS(req *protoimage.GcpImageSyncReq,
	cts *rest.Contexts, projectID string) ([]string, error) {

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
						Value: enumor.Gcp,
					},
					&filter.AtomRule{
						Field: "extension.region",
						Op:    filter.JSONEqual.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "extension.project_id",
						Op:    filter.JSONEqual.Factory(),
						Value: projectID,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		results, err := da.dataCli.Gcp.ListImage(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
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
