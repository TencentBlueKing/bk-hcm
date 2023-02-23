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
	"hcm/pkg/adaptor/types/eip"
	apicore "hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	protoeip "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// GcpSyncEip sync gcp to hcm
func GcpSyncEip(ea *eipAdaptor, cts *rest.Contexts) (interface{}, error) {

	req, err := ea.decodeEipSyncReq(cts)
	if err != nil {
		logs.Errorf("request decodeEipSyncReq failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// get cloud datas by page
	client, err := ea.adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	cloudAllIDs := make(map[string]bool)
	for {
		opt := &eip.GcpEipListOption{
			Region:    req.Region,
			PageToken: nextToken,
		}

		if nextToken != "" {
			opt.PageToken = nextToken
		}

		datas, err := client.ListEip(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list gcp eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*GcpEipSync)
		cloudIDs := make([]string, 0, len(datas.Details))
		for _, data := range datas.Details {
			eipSync := new(GcpEipSync)
			eipSync.IsUpdate = false
			eipSync.Eip = data
			cloudMap[data.CloudID] = eipSync
			cloudIDs = append(cloudIDs, data.CloudID)
			cloudAllIDs[data.CloudID] = true
		}

		updateIDs, dsMap, err := ea.getGcpEipDSSync(cloudIDs, req, cts)
		if err != nil {
			logs.Errorf("request getGcpEipDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := ea.syncGcpEipUpdate(updateIDs, cloudMap, dsMap, cts)
			if err != nil {
				logs.Errorf("request syncGcpEipUpdate failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
			err := ea.syncGcpEipAdd(addIDs, cts, req, cloudMap)
			if err != nil {
				logs.Errorf("request syncGcpEipAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}

		if len(datas.NextPageToken) == 0 {
			break
		}
		nextToken = datas.NextPageToken
	}

	dsIDs, err := ea.getGcpEipAllDS(req, cts)
	if err != nil {
		logs.Errorf("request getGcpEipAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
			opt := &eip.GcpEipListOption{
				Region:    req.Region,
				PageToken: nextToken,
			}

			if nextToken != "" {
				opt.PageToken = nextToken
			}

			datas, err := client.ListEip(cts.Kit, opt)
			if err != nil {
				logs.Errorf("request adaptor to list gcp eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

			if len(datas.NextPageToken) == 0 {
				break
			}
			nextToken = datas.NextPageToken
		}

		err := ea.syncEipDelete(cts, realDeleteIDs)
		if err != nil {
			logs.Errorf("request syncEipDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (ea *eipAdaptor) syncGcpEipAdd(addIDs []string, cts *rest.Contexts, req *protoeip.EipSyncReq,
	cloudMap map[string]*GcpEipSync) error {

	var createReq dataproto.EipExtBatchCreateReq[dataproto.GcpEipExtensionCreateReq]

	for _, id := range addIDs {
		publicImage := &dataproto.EipExtCreateReq[dataproto.GcpEipExtensionCreateReq]{
			CloudID:   id,
			Region:    req.Region,
			AccountID: req.AccountID,
			Name:      cloudMap[id].Eip.Name,
			Status:    getStrPtrVal(cloudMap[id].Eip.Status),
			PublicIp:  getStrPtrVal(cloudMap[id].Eip.PublicIp),
			PrivateIp: getStrPtrVal(cloudMap[id].Eip.PrivateIp),
			Extension: &dataproto.GcpEipExtensionCreateReq{
				AddressType:  cloudMap[id].Eip.AddressType,
				Description:  cloudMap[id].Eip.Description,
				IpVersion:    cloudMap[id].Eip.IpVersion,
				NetworkTier:  cloudMap[id].Eip.NetworkTier,
				PrefixLength: cloudMap[id].Eip.PrefixLength,
				Purpose:      cloudMap[id].Eip.Purpose,
				Network:      cloudMap[id].Eip.Network,
				Subnetwork:   cloudMap[id].Eip.Subnetwork,
				SelfLink:     cloudMap[id].Eip.SelfLink,
			},
		}
		createReq = append(createReq, publicImage)
	}

	if len(createReq) > 0 {
		_, err := ea.dataCli.Gcp.BatchCreateEip(cts.Kit.Ctx, cts.Kit.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create gcp eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (ea *eipAdaptor) syncGcpEipUpdate(updateIDs []string, cloudMap map[string]*GcpEipSync,
	dsMap map[string]*GcpDSEipSync, cts *rest.Contexts) error {

	var updateReq dataproto.EipExtBatchUpdateReq[dataproto.GcpEipExtensionUpdateReq]

	for _, id := range updateIDs {
		if cloudMap[id].Eip.Status != nil && *cloudMap[id].Eip.Status == dsMap[id].Eip.Status {
			continue
		}

		publicImage := &dataproto.EipExtUpdateReq[dataproto.GcpEipExtensionUpdateReq]{
			ID:     dsMap[id].Eip.ID,
			Status: *cloudMap[id].Eip.Status,
		}

		updateReq = append(updateReq, publicImage)
	}

	if len(updateReq) > 0 {
		if _, err := ea.dataCli.Gcp.BatchUpdateEip(cts.Kit.Ctx, cts.Kit.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateEip failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}

	return nil
}

func (ea *eipAdaptor) getGcpEipDSSync(cloudIDs []string, req *protoeip.EipSyncReq,
	cts *rest.Contexts) ([]string, map[string]*GcpDSEipSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*GcpDSEipSync)

	start := 0
	for {

		dataReq := &dataproto.EipListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.Gcp,
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

		results, err := ea.dataCli.Gcp.ListEip(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(GcpDSEipSync)
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

func (ea *eipAdaptor) getGcpEipAllDS(req *protoeip.EipSyncReq, cts *rest.Contexts) ([]string, error) {
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
						Value: enumor.Gcp,
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

		results, err := ea.dataCli.Gcp.ListEip(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
