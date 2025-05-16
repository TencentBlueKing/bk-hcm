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
	adcore "hcm/pkg/adaptor/types/core"
	typeseip "hcm/pkg/adaptor/types/eip"
	"hcm/pkg/api/core"
	dataeip "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncEipOption ...
type SyncEipOption struct {
	// BkBizID Eip创建时，通过同步写入DB，需要传入业务ID
	BkBizID int64 `json:"bk_biz_id" validate:"omitempty"`
}

// Validate ...
func (opt SyncEipOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Eip ...
func (cli *client) Eip(kt *kit.Kit, params *SyncBaseParams, opt *SyncEipOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	eipFromCloud, err := cli.listEipFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	eipFromDB, err := cli.listEipFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(eipFromCloud) == 0 && len(eipFromDB) == 0 {
		return new(SyncResult), nil
	}

	addEip, updateMap, delCloudIDs := common.Diff[*typeseip.AzureEip,
		*dataeip.EipExtResult[dataeip.AzureEipExtensionResult]](eipFromCloud, eipFromDB, isEipChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteEip(kt, params.AccountID, params.ResourceGroupName, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addEip) > 0 {
		if err = cli.createEip(kt, params.AccountID, opt.BkBizID, addEip); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateEip(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// RemoveEipDeleteFromCloud ...
func (cli *client) RemoveEipDeleteFromCloud(kt *kit.Kit, accountID string, resGroupName string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(),
					Value: resGroupName},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ListEip(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list eip failed, err: %v, req: %v, rid: %s", enumor.Azure,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		var resultFromCloud []*typeseip.AzureEip
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID:         accountID,
				ResourceGroupName: resGroupName,
				CloudIDs:          cloudIDs,
			}
			resultFromCloud, err = cli.listEipFromCloud(kt, params)
			if err != nil {
				return err
			}
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.CloudID)
			}

			delCloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(delCloudIDs) > 0 {
				if err = cli.deleteEip(kt, accountID, resGroupName, delCloudIDs); err != nil {
					return err
				}
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func (cli *client) deleteEip(kt *kit.Kit, accountID string, resGroupName string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("eip delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          delCloudIDs,
	}
	delEipFromCloud, err := cli.listEipFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delEipFromCloud) > 0 {
		logs.Errorf("[%s] validate eip not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delEipFromCloud), kt.Rid)
		return fmt.Errorf("validate eip not exist failed, before delete")
	}

	deleteReq := &dataeip.EipDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteEip(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete eip failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to delete eip success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateEip(kt *kit.Kit, accountID string, updateMap map[string]*typeseip.AzureEip) error {
	if len(updateMap) <= 0 {
		return fmt.Errorf("eip updateMap is <= 0, not delete")
	}

	updateReq := dataeip.EipExtBatchUpdateReq[dataeip.AzureEipExtensionUpdateReq]{}
	for id, one := range updateMap {
		eip := &dataeip.EipExtUpdateReq[dataeip.AzureEipExtensionUpdateReq]{
			ID:     id,
			Status: converter.PtrToVal(one.Status),
			Extension: &dataeip.AzureEipExtensionUpdateReq{
				ResourceGroupName: one.ResourceGroupName,
				IpConfigurationID: one.IpConfigurationID,
				SKU:               one.SKU,
				SKUTier:           one.SKUTier,
				Zones:             one.Zones,
				Location:          one.Location,
				Fqdn:              one.Fqdn,
			},
		}

		updateReq = append(updateReq, eip)
	}

	if _, err := cli.dbCli.Azure.BatchUpdateEip(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db eip failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to update eip success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createEip(kt *kit.Kit, accountID string, bizID int64, addEip []*typeseip.AzureEip) error {
	if len(addEip) <= 0 {
		return fmt.Errorf("eip addEip is <= 0, not delete")
	}

	request := dataeip.EipExtBatchCreateReq[dataeip.AzureEipExtensionCreateReq]{}
	for _, one := range addEip {
		tmpRes := &dataeip.EipExtCreateReq[dataeip.AzureEipExtensionCreateReq]{
			CloudID:    one.CloudID,
			Region:     one.Region,
			AccountID:  accountID,
			Name:       one.Name,
			InstanceId: one.InstanceId,
			Status:     converter.PtrToVal(one.Status),
			PublicIp:   converter.PtrToVal(one.PublicIp),
			PrivateIp:  converter.PtrToVal(one.PrivateIp),
			BkBizID:    bizID,
			Extension: &dataeip.AzureEipExtensionCreateReq{
				ResourceGroupName: one.ResourceGroupName,
				IpConfigurationID: one.IpConfigurationID,
				SKU:               one.SKU,
				SKUTier:           one.SKUTier,
				Zones:             one.Zones,
				Location:          one.Location,
				Fqdn:              one.Fqdn,
			},
		}

		request = append(request, tmpRes)
	}

	if _, err := cli.dbCli.Azure.BatchCreateEip(kt.Ctx, kt.Header(), &request); err != nil {
		logs.Errorf("[%s] request dataservice to batch create eip failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to create eip success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(addEip), kt.Rid)

	return nil
}

func (cli *client) listEipFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]*typeseip.AzureEip, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adcore.AzureListByIDOption{
		ResourceGroupName: params.ResourceGroupName,
		CloudIDs:          params.CloudIDs,
	}
	result, err := cli.cloudCli.ListEipByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list eip from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listEipFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*dataeip.EipExtResult[dataeip.AzureEipExtensionResult], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &dataeip.EipListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "extension.resource_group_name",
					Op:    filter.JSONEqual.Factory(),
					Value: params.ResourceGroupName,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.ListEip(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list eip from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isEipChange(item *typeseip.AzureEip, info *dataeip.EipExtResult[dataeip.AzureEipExtensionResult]) bool {
	if converter.PtrToVal(item.Status) != info.Status {
		return true
	}

	if !assert.IsPtrStringEqual(item.InstanceId, info.InstanceID) {
		return true
	}

	if !assert.IsPtrStringEqual(item.IpConfigurationID, info.Extension.IpConfigurationID) {
		return true
	}

	if !assert.IsPtrStringEqual(item.SKU, info.Extension.SKU) {
		return true
	}

	if !assert.IsPtrStringEqual(item.SKUTier, info.Extension.SKUTier) {
		return true
	}

	if !assert.IsPtrStringSliceEqual(item.Zones, info.Extension.Zones) {
		return true
	}

	if !assert.IsPtrStringEqual(item.Location, info.Extension.Location) {
		return true
	}

	if !assert.IsPtrStringEqual(item.Fqdn, info.Extension.Fqdn) {
		return true
	}

	return false
}
