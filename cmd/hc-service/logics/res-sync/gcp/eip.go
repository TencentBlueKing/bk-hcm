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

package gcp

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
	Region string `json:"region" validate:"required"`
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

	eipFromCloud, err := cli.listEipFromCloud(kt, params, opt.Region)
	if err != nil {
		return nil, err
	}

	eipFromDB, err := cli.listEipFromDB(kt, params, opt.Region)
	if err != nil {
		return nil, err
	}

	if len(eipFromCloud) == 0 && len(eipFromDB) == 0 {
		return new(SyncResult), nil
	}

	addEip, updateMap, delCloudIDs := common.Diff[*typeseip.GcpEip,
		*dataeip.EipExtResult[dataeip.GcpEipExtensionResult]](eipFromCloud, eipFromDB, isEipChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteEip(kt, params.AccountID, opt.Region, delCloudIDs); err != nil {
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
			logs.Errorf("[%s] request dataservice to list eip failed, err: %v, req: %v, rid: %s", enumor.Gcp,
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

		var resultFromCloud []*typeseip.GcpEip
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				CloudIDs:  cloudIDs,
			}
			resultFromCloud, err = cli.listEipFromCloud(kt, params, region)
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
		CloudIDs:  delCloudIDs,
	}
	delEipFromCloud, err := cli.listEipFromCloud(kt, checkParams, region)
	if err != nil {
		return err
	}

	if len(delEipFromCloud) > 0 {
		logs.Errorf("[%s] validate eip not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delEipFromCloud), kt.Rid)
		return fmt.Errorf("validate eip not exist failed, before delete")
	}

	deleteReq := &dataeip.EipDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if _, err = cli.dbCli.Global.DeleteEip(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete eip failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to delete eip success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateEip(kt *kit.Kit, accountID string, updateMap map[string]*typeseip.GcpEip) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update eip, eips is required")
	}

	updateReq := make(dataeip.EipExtBatchUpdateReq[dataeip.GcpEipExtensionUpdateReq], 0, len(updateMap))
	for id, one := range updateMap {
		eip := &dataeip.EipExtUpdateReq[dataeip.GcpEipExtensionUpdateReq]{
			ID:     id,
			Status: converter.PtrToVal(one.Status),
			Extension: &dataeip.GcpEipExtensionUpdateReq{
				AddressType:  one.AddressType,
				Description:  one.Description,
				IpVersion:    one.IpVersion,
				NetworkTier:  one.NetworkTier,
				PrefixLength: one.PrefixLength,
				Purpose:      one.Purpose,
				Network:      one.Network,
				Subnetwork:   one.Subnetwork,
				SelfLink:     one.SelfLink,
				Users:        one.Users,
			},
		}

		updateReq = append(updateReq, eip)
	}

	if _, err := cli.dbCli.Gcp.BatchUpdateEip(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db eip failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to update eip success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createEip(kt *kit.Kit, accountID string, addEip []*typeseip.GcpEip, bizID int64) error {
	if len(addEip) == 0 {
		return fmt.Errorf("create eip, eips is required")
	}

	createReq := make(dataeip.EipExtBatchCreateReq[dataeip.GcpEipExtensionCreateReq], 0, len(addEip))
	for _, one := range addEip {
		tmpRes := &dataeip.EipExtCreateReq[dataeip.GcpEipExtensionCreateReq]{
			CloudID:   one.CloudID,
			Region:    one.Region,
			AccountID: accountID,
			Name:      one.Name,
			Status:    converter.PtrToVal(one.Status),
			PublicIp:  converter.PtrToVal(one.PublicIp),
			PrivateIp: converter.PtrToVal(one.PrivateIp),
			BkBizID:   bizID,
			Extension: &dataeip.GcpEipExtensionCreateReq{
				AddressType:  one.AddressType,
				Description:  one.Description,
				IpVersion:    one.IpVersion,
				NetworkTier:  one.NetworkTier,
				PrefixLength: one.PrefixLength,
				Purpose:      one.Purpose,
				Network:      one.Network,
				Subnetwork:   one.Subnetwork,
				SelfLink:     one.SelfLink,
				Users:        one.Users,
			},
		}

		createReq = append(createReq, tmpRes)
	}

	if _, err := cli.dbCli.Gcp.BatchCreateEip(kt.Ctx, kt.Header(), &createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create eip failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync eip to create eip success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addEip), kt.Rid)

	return nil
}

func (cli *client) listEipFromCloud(kt *kit.Kit, params *SyncBaseParams, region string) ([]*typeseip.GcpEip, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typeseip.GcpEipListOption{
		Region:   region,
		CloudIDs: params.CloudIDs,
		Page: &adcore.GcpPage{
			PageSize: adcore.GcpQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListEip(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list eip from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listEipFromDB(kt *kit.Kit, params *SyncBaseParams, region string) (
	[]*dataeip.EipExtResult[dataeip.GcpEipExtensionResult], error) {

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
					Value: region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.ListEip(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list eip from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isEipChange(cloud *typeseip.GcpEip, db *dataeip.EipExtResult[dataeip.GcpEipExtensionResult]) bool {
	if converter.PtrToVal(cloud.Status) != db.Status {
		return true
	}

	if cloud.AddressType != db.Extension.AddressType {
		return true
	}

	if cloud.Description != db.Extension.Description {
		return true
	}

	if cloud.IpVersion != db.Extension.IpVersion {
		return true
	}

	if cloud.NetworkTier != db.Extension.NetworkTier {
		return true
	}

	if cloud.PrefixLength != db.Extension.PrefixLength {
		return true
	}

	if cloud.Purpose != db.Extension.Purpose {
		return true
	}

	if cloud.Network != db.Extension.Network {
		return true
	}

	if cloud.Subnetwork != db.Extension.Subnetwork {
		return true
	}

	if cloud.SelfLink != db.Extension.SelfLink {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.Users, cloud.Users) {
		return true
	}

	return false
}
