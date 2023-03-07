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

package region

import (
	"hcm/pkg/api/core"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	apiregion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// SyncAzureRegion sync all region
func (r *region) SyncAzureRegion(cts *rest.Contexts) (interface{}, error) {

	req := new(apiregion.AzureRegionSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := r.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	regions, err := client.ListRegion(cts.Kit)
	if err != nil {
		logs.Errorf("request adaptor to list azure region failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudAllIDs := make(map[string]bool)

	cloudMap := make(map[string]*AzureRegionSync)
	cloudIDs := make([]string, 0, len(regions))
	for _, data := range regions {
		regionSync := new(AzureRegionSync)
		regionSync.IsUpdate = false
		regionSync.Region = data
		cloudMap[*data.Name] = regionSync
		cloudIDs = append(cloudIDs, *data.Name)
		cloudAllIDs[*data.Name] = true
	}

	// TODO: 为后续如果有更新需求预留
	updateIDs, _, err := r.getAzureRegionDSSync(cts, cloudIDs, req)
	if err != nil {
		logs.Errorf("request getAzureRegionDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(updateIDs) > 0 {
		// TODO: 目前都是只读字段，先不做更新
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
		err := r.syncAzureRegionAdd(addIDs, cts, req, cloudMap)
		if err != nil {
			logs.Errorf("request syncAzureRegionAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := r.getAzureRegionAllDS(cts, req)
	if err != nil {
		logs.Errorf("request getAzureRegionAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for _, id := range dsIDs {
		if _, ok := cloudAllIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(deleteIDs) > 0 {
		err := r.syncAzureRegionDelete(cts, deleteIDs)
		if err != nil {
			logs.Errorf("request syncRegionDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (r *region) syncAzureRegionAdd(addIDs []string, cts *rest.Contexts, req *apiregion.AzureRegionSyncReq,
	cloudMap map[string]*AzureRegionSync) error {

	list := make([]protoregion.AzureRegionBatchCreate, 0, len(addIDs))
	for _, id := range addIDs {
		one := protoregion.AzureRegionBatchCreate{
			Cloud_ID:          *cloudMap[id].Region.ID,
			Name:              *cloudMap[id].Region.Name,
			Type:              string(*cloudMap[id].Region.Type),
			DisplayName:       *cloudMap[id].Region.DisplayName,
			RegionDisplayName: *cloudMap[id].Region.RegionalDisplayName,
			RegionType:        string(*cloudMap[id].Region.Metadata.RegionType),
		}
		list = append(list, one)
	}

	createReq := &protoregion.AzureRegionBatchCreateReq{
		Regions: list,
	}
	_, err := r.dataCli.Azure.Region.BatchCreateRegion(cts.Kit.Ctx, cts.Kit.Header(),
		createReq)
	if err != nil {
		return err
	}

	return nil
}

func (r *region) syncAzureRegionDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	deleteReq := &protoregion.AzureRegionBatchDeleteReq{
		Filter: tools.ContainersExpression("name", deleteCloudIDs),
	}

	err := r.dataCli.Azure.Region.BatchDeleteRegion(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return err
	}

	return nil
}

func (r *region) getAzureRegionDSSync(cts *rest.Contexts, cloudIDs []string,
	req *apiregion.AzureRegionSyncReq) ([]string, map[string]*AzureDSRegionSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*AzureDSRegionSync)

	start := 0
	for {
		dataReq := &protoregion.AzureRegionListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: constant.SyncTimingListAzureRegion,
					},
					&filter.AtomRule{
						Field: "name",
						Op:    filter.In.Factory(),
						Value: cloudIDs,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: core.DefaultMaxPageLimit},
		}

		results, err := r.dataCli.Azure.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list public region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.Name)
				dsRegionSync := new(AzureDSRegionSync)
				dsRegionSync.Region = detail
				dsMap[detail.Name] = dsRegionSync
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return updateIDs, dsMap, nil
}

func (r *region) getAzureRegionAllDS(cts *rest.Contexts, req *apiregion.AzureRegionSyncReq) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {
		dataReq := &protoregion.AzureRegionListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: constant.SyncTimingListAzureRegion,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: core.DefaultMaxPageLimit},
		}

		results, err := r.dataCli.Azure.Region.ListRegion(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list public region failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return dsIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsIDs = append(dsIDs, detail.Name)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsIDs, nil
}
