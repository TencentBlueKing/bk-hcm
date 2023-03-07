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

package eip

import (
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/eip"
	apicore "hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	protoeip "hcm/pkg/api/hc-service/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// SyncTCloudEip sync eip self
func SyncTCloudEip(kt *kit.Kit, req *protoeip.EipSyncReq,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	client, err := ad.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	offset := 0
	cloudAllIDs := make(map[string]bool)
	for {
		opt := &eip.TCloudEipListOption{
			Region: req.Region,
			Page:   &core.TCloudPage{Offset: uint64(offset), Limit: uint64(filter.DefaultMaxInLimit)},
		}
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		}

		datas, err := client.ListEip(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*TCloudEipSync)
		cloudIDs := make([]string, 0, len(datas.Details))
		for _, data := range datas.Details {
			eipSync := new(TCloudEipSync)
			eipSync.IsUpdate = false
			eipSync.Eip = data
			cloudMap[data.CloudID] = eipSync
			cloudIDs = append(cloudIDs, data.CloudID)
			cloudAllIDs[data.CloudID] = true
		}

		updateIDs, dsMap, err := getTCloudEipDSSync(kt, cloudIDs, req, dataCli)
		if err != nil {
			logs.Errorf("request getTCloudEipDSSync failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := syncTCloudEipUpdate(kt, updateIDs, cloudMap, dsMap, dataCli)
			if err != nil {
				logs.Errorf("request syncTCloudEipUpdate failed, err: %v, rid: %s", err, kt.Rid)
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
			err := syncTCloudEipAdd(kt, addIDs, req, cloudMap, dataCli)
			if err != nil {
				logs.Errorf("request syncTCloudEipAdd failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		offset += len(datas.Details)
		if uint(len(datas.Details)) < filter.DefaultMaxInLimit {
			break
		}
	}

	dsIDs, err := getTCloudEipAllDS(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getTCloudEipAllDS failed, err: %v, rid: %s", err, kt.Rid)
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

		offset := 0
		for {
			opt := &eip.TCloudEipListOption{
				Region: req.Region,
				Page:   &core.TCloudPage{Offset: uint64(offset), Limit: uint64(filter.DefaultMaxInLimit)},
			}

			datas, err := client.ListEip(kt, opt)
			if err != nil {
				logs.Errorf("request adaptor to list tcloud eip failed, err: %v, rid: %s", err, kt.Rid)
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

			offset += len(datas.Details)
			if uint(len(datas.Details)) < filter.DefaultMaxInLimit {
				break
			}
		}

		if len(realDeleteIDs) > 0 {
			err := syncEipDelete(kt, realDeleteIDs, dataCli)
			if err != nil {
				logs.Errorf("request syncEipDelete failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func syncTCloudEipAdd(kt *kit.Kit, addIDs []string, req *protoeip.EipSyncReq,
	cloudMap map[string]*TCloudEipSync, dataCli *dataservice.Client) error {

	createReq := make(dataproto.EipExtBatchCreateReq[dataproto.TCloudEipExtensionCreateReq], 0, len(addIDs))
	for _, id := range addIDs {
		publicImage := &dataproto.EipExtCreateReq[dataproto.TCloudEipExtensionCreateReq]{
			CloudID:    id,
			Region:     req.Region,
			AccountID:  req.AccountID,
			Name:       cloudMap[id].Eip.Name,
			InstanceId: converter.PtrToVal(cloudMap[id].Eip.InstanceId),
			Status:     converter.PtrToVal(cloudMap[id].Eip.Status),
			PublicIp:   converter.PtrToVal(cloudMap[id].Eip.PublicIp),
			PrivateIp:  converter.PtrToVal(cloudMap[id].Eip.PrivateIp),
			Extension: &dataproto.TCloudEipExtensionCreateReq{
				Bandwidth: cloudMap[id].Eip.Bandwidth,
			},
		}
		createReq = append(createReq, publicImage)
	}

	if len(createReq) > 0 {
		_, err := dataCli.TCloud.BatchCreateEip(kt.Ctx, kt.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create tcloud eip failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func syncTCloudEipUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*TCloudEipSync,
	dsMap map[string]*TCloudDSEipSync, dataCli *dataservice.Client) error {

	var updateReq dataproto.EipExtBatchUpdateReq[dataproto.TCloudEipExtensionUpdateReq]

	for _, id := range updateIDs {
		if cloudMap[id].Eip.Status != nil && *cloudMap[id].Eip.Status == dsMap[id].Eip.Status {
			continue
		}

		publicImage := &dataproto.EipExtUpdateReq[dataproto.TCloudEipExtensionUpdateReq]{
			ID:     dsMap[id].Eip.ID,
			Status: *cloudMap[id].Eip.Status,
		}

		updateReq = append(updateReq, publicImage)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.TCloud.BatchUpdateEip(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateEip failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getTCloudEipDSSync(kt *kit.Kit, cloudIDs []string, req *protoeip.EipSyncReq,
	dataCli *dataservice.Client) ([]string, map[string]*TCloudDSEipSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*TCloudDSEipSync)

	start := 0
	for {
		dataReq := &dataproto.EipListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.TCloud,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
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

		results, err := dataCli.TCloud.ListEip(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list eip failed, err: %v, rid: %s", err, kt.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(TCloudDSEipSync)
				dsImageSync.Eip = detail
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

func getTCloudEipAllDS(kt *kit.Kit, req *protoeip.EipSyncReq,
	dataCli *dataservice.Client) ([]string, error) {
	start := 0
	dsIDs := make([]string, 0)
	for {
		dataReq := &dataproto.EipListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.TCloud,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.TCloud.ListEip(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list eip failed, err: %v, rid: %s", err, kt.Rid)
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
