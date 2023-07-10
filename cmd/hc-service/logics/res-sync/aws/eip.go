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

package aws

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
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

	addEip, updateMap, delCloudIDs := common.Diff[*typeseip.AwsEip,
		*dataeip.EipExtResult[dataeip.AwsEipExtensionResult]](eipFromCloud, eipFromDB, isEipChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteEip(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addEip) > 0 {
		if err = cli.createEip(kt, params.AccountID, addEip, opt.BkBizID); err != nil {
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
func (cli *client) RemoveEipDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {

	req := &dataeip.EipListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ListEip(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list eip failed, err: %v, req: %v, rid: %s", enumor.Aws,
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

		var resultFromCloud []*typeseip.AwsEip
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
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
			if err = cli.deleteEip(kt, accountID, region, delCloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func (cli *client) deleteEip(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete eip, cloudIDs is required")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delEipFromCloud, err := cli.listEipFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delEipFromCloud) > 0 {
		logs.Errorf("[%s] validate eip not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Aws, checkParams, len(delEipFromCloud), kt.Rid)
		return fmt.Errorf("validate eip not exist failed, before delete")
	}

	deleteReq := &dataeip.EipDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteEip(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete eip failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to delete eip success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateEip(kt *kit.Kit, accountID string, updateMap map[string]*typeseip.AwsEip) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update eip, eips is required")
	}

	updateReq := make(dataeip.EipExtBatchUpdateReq[dataeip.AwsEipExtensionUpdateReq], 0, len(updateMap))
	for id, one := range updateMap {
		eip := &dataeip.EipExtUpdateReq[dataeip.AwsEipExtensionUpdateReq]{
			ID:     id,
			Status: converter.PtrToVal(one.Status),
			Extension: &dataeip.AwsEipExtensionUpdateReq{
				PublicIpv4Pool:          one.PublicIpv4Pool,
				Domain:                  one.Domain,
				PrivateIpAddress:        one.PrivateIpAddress,
				NetworkBorderGroup:      one.NetworkBorderGroup,
				NetworkInterfaceId:      one.NetworkInterfaceId,
				NetworkInterfaceOwnerId: one.NetworkInterfaceOwnerId,
			},
		}

		updateReq = append(updateReq, eip)
	}

	if _, err := cli.dbCli.Aws.BatchUpdateEip(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db eip failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to update eip success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createEip(kt *kit.Kit, accountID string, addEip []*typeseip.AwsEip, bizID int64) error {
	if len(addEip) == 0 {
		return fmt.Errorf("create eip, eips is required")
	}

	createReq := make(dataeip.EipExtBatchCreateReq[dataeip.AwsEipExtensionCreateReq], 0, len(addEip))
	for _, one := range addEip {
		tmpRes := &dataeip.EipExtCreateReq[dataeip.AwsEipExtensionCreateReq]{
			CloudID:   one.CloudID,
			Region:    one.Region,
			AccountID: accountID,
			Name:      one.Name,
			Status:    converter.PtrToVal(one.Status),
			PublicIp:  converter.PtrToVal(one.PublicIp),
			PrivateIp: converter.PtrToVal(one.PrivateIp),
			BkBizID:   bizID,
			Extension: &dataeip.AwsEipExtensionCreateReq{
				PublicIpv4Pool:          one.PublicIpv4Pool,
				Domain:                  one.Domain,
				PrivateIpAddress:        one.PrivateIpAddress,
				NetworkBorderGroup:      one.NetworkBorderGroup,
				NetworkInterfaceId:      one.NetworkInterfaceId,
				NetworkInterfaceOwnerId: one.NetworkInterfaceOwnerId,
			},
		}

		createReq = append(createReq, tmpRes)
	}

	if _, err := cli.dbCli.Aws.BatchCreateEip(kt.Ctx, kt.Header(), &createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create eip failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to create eip success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(addEip), kt.Rid)

	return nil
}

func (cli *client) listEipFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]*typeseip.AwsEip, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typeseip.AwsEipListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
	}
	result, err := cli.cloudCli.ListEip(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list eip from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listEipFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*dataeip.EipExtResult[dataeip.AwsEipExtensionResult], error) {

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
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: params.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Aws.ListEip(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list eip from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isEipChange(cloud *typeseip.AwsEip, db *dataeip.EipExtResult[dataeip.AwsEipExtensionResult]) bool {
	if converter.PtrToVal(cloud.Status) != db.Status {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.PublicIpv4Pool, db.Extension.PublicIpv4Pool) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Domain, db.Extension.Domain) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.PrivateIpAddress, db.Extension.PrivateIpAddress) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.NetworkBorderGroup, db.Extension.NetworkBorderGroup) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.NetworkInterfaceId, db.Extension.NetworkInterfaceId) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.NetworkInterfaceOwnerId, db.Extension.NetworkInterfaceOwnerId) {
		return true
	}

	return false
}
