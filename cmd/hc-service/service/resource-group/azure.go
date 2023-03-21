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

package resourcegroup

import (
	typesresourcegroup "hcm/pkg/adaptor/types/resource-group"
	"hcm/pkg/api/core"
	apicloudregion "hcm/pkg/api/core/cloud/region"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	apiregion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// AzureRGSync ...
type AzureRGSync struct {
	IsUpdate      bool
	ResourceGroup *typesresourcegroup.AzureResourceGroup
}

// AzureDSRGSync ...
type AzureDSRGSync struct {
	ResourceGroup apicloudregion.AzureRG
}

// SyncAzureRG sync all resource group
func (r *resourcegroup) SyncAzureRG(cts *rest.Contexts) (interface{}, error) {

	req := new(apiregion.AzureRGSyncReq)
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

	regions, err := client.ListResourceGroup(cts.Kit)
	if err != nil {
		logs.Errorf("request adaptor to list azure resource group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cloudAllIDs := make(map[string]bool)

	cloudMap := make(map[string]*AzureRGSync)
	cloudIDs := make([]string, 0, len(regions))
	for _, data := range regions {
		regionSync := new(AzureRGSync)
		regionSync.IsUpdate = false
		regionSync.ResourceGroup = data
		cloudMap[*data.Name] = regionSync
		cloudIDs = append(cloudIDs, *data.Name)
		cloudAllIDs[*data.Name] = true
	}

	updateIDs, dsMap, err := r.getAzureRGDSSync(cloudIDs, req, cts)
	if err != nil {
		logs.Errorf("request getAzureRGDSSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(updateIDs) > 0 {
		err := r.syncAzureRGUpdate(updateIDs, cloudMap, dsMap, cts)
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
		err := r.syncAzureRGAdd(addIDs, cts, req, cloudMap)
		if err != nil {
			logs.Errorf("request syncAzureRGAdd failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	dsIDs, err := r.getAzureRGAllDS(req, cts)
	if err != nil {
		logs.Errorf("request getAzureRGAllDS failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for _, id := range dsIDs {
		if _, ok := cloudAllIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(deleteIDs) > 0 {
		err := r.syncAzureRGDelete(cts, deleteIDs)
		if err != nil {
			logs.Errorf("request syncAzureRGDelete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func (r *resourcegroup) syncAzureRGUpdate(updateIDs []string, cloudMap map[string]*AzureRGSync,
	dsMap map[string]*AzureDSRGSync, cts *rest.Contexts) error {

	list := make([]protoregion.AzureRGBatchUpdate, 0, len(updateIDs))
	for _, id := range updateIDs {
		if *cloudMap[id].ResourceGroup.Location == dsMap[id].ResourceGroup.Location {
			continue
		}
		one := protoregion.AzureRGBatchUpdate{
			ID:       dsMap[id].ResourceGroup.ID,
			Location: *cloudMap[id].ResourceGroup.Location,
		}
		list = append(list, one)
	}

	updateReq := &protoregion.AzureRGBatchUpdateReq{
		ResourceGroups: list,
	}

	if len(updateReq.ResourceGroups) > 0 {
		if err := r.dataCli.Azure.ResourceGroup.BatchUpdateRG(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateRG failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}
	return nil
}

func (r *resourcegroup) syncAzureRGAdd(addIDs []string, cts *rest.Contexts, req *apiregion.AzureRGSyncReq,
	cloudMap map[string]*AzureRGSync) error {

	list := make([]protoregion.AzureRGBatchCreate, 0, len(addIDs))
	for _, id := range addIDs {
		one := protoregion.AzureRGBatchCreate{
			Name:      *cloudMap[id].ResourceGroup.Name,
			Type:      string(*cloudMap[id].ResourceGroup.Type),
			Location:  *cloudMap[id].ResourceGroup.Location,
			AccountID: req.AccountID,
		}
		list = append(list, one)
	}

	createReq := &protoregion.AzureRGBatchCreateReq{
		ResourceGroups: list,
	}
	_, err := r.dataCli.Azure.ResourceGroup.BatchCreateResourceGroup(cts.Kit.Ctx, cts.Kit.Header(),
		createReq)
	if err != nil {
		return err
	}

	return nil
}

func (r *resourcegroup) syncAzureRGDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	deleteReq := &protoregion.AzureRGBatchDeleteReq{
		Filter: tools.ContainersExpression("name", deleteCloudIDs),
	}

	err := r.dataCli.Azure.ResourceGroup.BatchDeleteResourceGroup(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return err
	}

	return nil
}

func (r *resourcegroup) getAzureRGDSSync(cloudIDs []string, req *apiregion.AzureRGSyncReq,
	cts *rest.Contexts) ([]string, map[string]*AzureDSRGSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*AzureDSRGSync)

	start := 0
	for {
		dataReq := &protoregion.AzureRGListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: constant.SyncTimingListAzureRG,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
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

		results, err := r.dataCli.Azure.ResourceGroup.ListResourceGroup(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list resource group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.Name)
				dsRegionSync := new(AzureDSRGSync)
				dsRegionSync.ResourceGroup = detail
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

func (r *resourcegroup) getAzureRGAllDS(req *apiregion.AzureRGSyncReq, cts *rest.Contexts) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {
		dataReq := &protoregion.AzureRGListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "type",
						Op:    filter.Equal.Factory(),
						Value: constant.SyncTimingListAzureRG,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
				},
			},
			Page: &core.BasePage{Start: uint32(start), Limit: core.DefaultMaxPageLimit},
		}

		results, err := r.dataCli.Azure.ResourceGroup.ListResourceGroup(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list resource group failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
