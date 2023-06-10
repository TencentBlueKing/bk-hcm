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
	"fmt"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	corerg "hcm/pkg/api/core/cloud/resource-group"
	protorg "hcm/pkg/api/data-service/cloud/resource-group"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablerg "hcm/pkg/dal/table/cloud/resource-group"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// InitAzureResourceGroupService initial the azure ResourceGroupName service
func InitAzureResourceGroupService(cap *capability.Capability) {
	svc := &azureRGSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("ListAzureResourceGroup", "POST", "/vendors/azure/resource_groups/list", svc.ListAzureResourceGroup)

	h.Add("DeleteAzureResourceGroup", "DELETE", "/vendors/azure/resource_groups/batch", svc.DeleteAzureResourceGroup)

	h.Add("CreateAzureResourceGroup", "POST", "/vendors/azure/resource_groups/batch/create", svc.CreateAzureResourceGroup)

	h.Add("UpdateAzureResourceGroup", "PUT", "/vendors/azure/resource_groups/batch/update", svc.UpdateAzureResourceGroup)

	h.Load(cap.WebService)
}

type azureRGSvc struct {
	dao dao.Set
}

// UpdateAzureResourceGroup update azure ResourceGroupName.
func (svc *azureRGSvc) UpdateAzureResourceGroup(cts *rest.Contexts) (interface{}, error) {

	req := new(protorg.AzureRGBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, one := range req.ResourceGroups {
			rule := &tablerg.AzureRGTable{
				Location: one.Location,
				Reviser:  cts.Kit.User,
			}

			flt := tools.EqualExpression("id", one.ID)
			if err := svc.dao.AzureRG().UpdateWithTx(cts.Kit, txn, flt, rule); err != nil {
				logs.Errorf("update azure resource group failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update azure resource group failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// CreateAzureResourceGroup create ResourceGroupName.
func (svc *azureRGSvc) CreateAzureResourceGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(protorg.AzureRGBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resourceGroups := make([]*tablerg.AzureRGTable, 0, len(req.ResourceGroups))
	for _, resourceGroup := range req.ResourceGroups {
		resourceGroups = append(resourceGroups, &tablerg.AzureRGTable{
			Name:      resourceGroup.Name,
			Type:      resourceGroup.Type,
			Location:  resourceGroup.Location,
			AccountID: resourceGroup.AccountID,
			Creator:   cts.Kit.User,
			Reviser:   cts.Kit.User,
		})
	}

	resourceGroupIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		resourceGroupIDs, err := svc.dao.AzureRG().CreateWithTx(cts.Kit, txn, resourceGroups)
		if err != nil {
			return nil, fmt.Errorf("batch create azure resource group failed, err: %v", err)
		}
		return resourceGroupIDs, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := resourceGroupIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create azure resource group but return id type is not string, id type: %v",
			reflect.TypeOf(resourceGroupIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// DeleteAzureResourceGroup delete azure resource group.
func (svc *azureRGSvc) DeleteAzureResourceGroup(cts *rest.Contexts) (interface{}, error) {

	req := new(protorg.AzureRGBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.dao.AzureRG().DeleteWithTx(cts.Kit, req.Filter); err != nil {
		logs.Errorf("delete azure resource group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListAzureResourceGroup list azure resource group with filter
func (svc *azureRGSvc) ListAzureResourceGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(protorg.AzureRGListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.AzureRG().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list azure resource group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list azure resource group failed, err: %v", err)
	}

	if req.Page.Count {
		return &protorg.AzureRGListResult{Count: result.Count}, nil
	}

	details := make([]corerg.AzureRG, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corerg.AzureRG{
			ID:        one.ID,
			Name:      one.Name,
			Type:      one.Type,
			Location:  one.Location,
			AccountID: one.AccountID,
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		})
	}

	return &protorg.AzureRGListResult{Details: details}, nil
}
