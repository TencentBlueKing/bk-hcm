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
	typescore "hcm/pkg/adaptor/types/core"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
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

// SyncSGOption ...
type SyncSGOption struct {
}

// Validate ...
func (opt SyncSGOption) Validate() error {
	return validator.Validate.Struct(opt)
}

func (cli *client) SecurityGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgFromCloud, err := cli.listSGFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	sgFromDB, err := cli.listSGFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(sgFromCloud) == 0 && len(sgFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[securitygroup.AzureSecurityGroup, cloudcore.SecurityGroup[cloudcore.AzureSecurityGroupExtension]](
		sgFromCloud, sgFromDB, isSGChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteSG(kt, params.AccountID, params.ResourceGroupName, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		_, err := cli.createSG(kt, params.AccountID, params.ResourceGroupName, addSlice)
		if err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateSG(kt, params.AccountID, params.ResourceGroupName, updateMap); err != nil {
			return nil, err
		}
	}

	// 同步安全组规则
	sgFromDB, err = cli.listSGFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	cloudSGIDs := make([]string, 0, len(sgFromDB))
	for _, one := range sgFromDB {
		cloudSGIDs = append(cloudSGIDs, one.CloudID)
	}

	sgRuleParams := &SyncBaseParams{
		AccountID:         params.AccountID,
		ResourceGroupName: params.ResourceGroupName,
		CloudIDs:          cloudSGIDs,
	}
	_, err = cli.SecurityGroupRule(kt, sgRuleParams, &SyncSGRuleOption{})
	if err != nil {
		logs.Errorf("[%s] sg sync sgRule failed. err: %v, accountID: %s, resGroupName: %s, rid: %s",
			err, enumor.Azure, params.AccountID, params.ResourceGroupName, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) createSG(kt *kit.Kit, accountID string, resGroupName string,
	addSlice []securitygroup.AzureSecurityGroup) ([]string, error) {

	if len(addSlice) <= 0 {
		return nil, fmt.Errorf("sg addSlice is <= 0, not create")
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[cloudcore.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[cloudcore.AzureSecurityGroupExtension]{},
	}

	for _, one := range addSlice {
		securityGroup := protocloud.SecurityGroupBatchCreate[cloudcore.AzureSecurityGroupExtension]{
			CloudID: converter.PtrToVal(one.ID),
			BkBizID: constant.UnassignedBiz,
			Region:  converter.PtrToVal(one.Location),
			Name:    converter.PtrToVal(one.Name),
			Memo:    nil,
			// 无该字段
			MgmtBizID: constant.UnassignedBiz,
			AccountID: accountID,
			Extension: &cloudcore.AzureSecurityGroupExtension{
				ResourceGroupName: resGroupName,
				Etag:              one.Etag,
				FlushConnection:   one.FlushConnection,
				ResourceGUID:      one.ResourceGUID,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	results, err := cli.dbCli.Azure.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s",
			enumor.Azure, err, kt.Rid)
		return nil, err
	}

	logs.Infof("[%s] sync sg to create sg success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(addSlice), kt.Rid)

	return results.IDs, nil
}

func (cli *client) updateSG(kt *kit.Kit, accountID string, resGroupName string,
	updateMap map[string]securitygroup.AzureSecurityGroup) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("sg updateMap is <= 0, not update")
	}

	securityGroups := make([]protocloud.SecurityGroupBatchUpdate[cloudcore.AzureSecurityGroupExtension], 0)

	for id, one := range updateMap {
		securityGroup := protocloud.SecurityGroupBatchUpdate[cloudcore.AzureSecurityGroupExtension]{
			ID:   id,
			Name: converter.PtrToVal(one.Name),
			Extension: &cloudcore.AzureSecurityGroupExtension{
				ResourceGroupName: resGroupName,
				Etag:              one.Etag,
				FlushConnection:   one.FlushConnection,
				ResourceGUID:      one.ResourceGUID,
			},
		}

		securityGroups = append(securityGroups, securityGroup)
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[cloudcore.AzureSecurityGroupExtension]{
		SecurityGroups: securityGroups,
	}
	if err := cli.dbCli.Azure.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
		updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to update sg success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteSG(kt *kit.Kit, accountID string, resGroupName string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("sg delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          delCloudIDs,
	}
	delSGFromCloud, err := cli.listSGFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delSGFromCloud) > 0 {
		logs.Errorf("[%s] validate sg not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delSGFromCloud), kt.Rid)
		return fmt.Errorf("validate sg not exist failed, before delete")
	}

	deleteReq := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.SecurityGroup.BatchDeleteSecurityGroup(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete sg failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to delete sg success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listSGFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]securitygroup.AzureSecurityGroup, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typescore.AzureListByIDOption{
		ResourceGroupName: params.ResourceGroupName,
		CloudIDs:          params.CloudIDs,
	}
	result, err := cli.cloudCli.ListSecurityGroupByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list sg from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	sgs := make([]securitygroup.AzureSecurityGroup, 0, len(result))
	for _, one := range result {
		sgs = append(sgs, converter.PtrToVal(one))
	}

	return sgs, nil
}

func (cli *client) listSGFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.SecurityGroup[cloudcore.AzureSecurityGroupExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
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
	result, err := cli.dbCli.Azure.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list sg from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) RemoveSecurityGroupDeleteFromCloud(kt *kit.Kit, accountID string, resGroupName string) error {
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: accountID,
				},
				&filter.AtomRule{
					Field: "extension.resource_group_name",
					Op:    filter.JSONEqual.Factory(),
					Value: resGroupName,
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Azure.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list sg failed, err: %v, req: %v, rid: %s", enumor.Azure,
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

		params := &SyncBaseParams{
			AccountID:         accountID,
			ResourceGroupName: resGroupName,
			CloudIDs:          cloudIDs,
		}
		resultFromCloud, err := cli.listSGFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.ID))
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if err := cli.deleteSG(kt, accountID, resGroupName, cloudIDs); err != nil {
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

func isSGChange(cloud securitygroup.AzureSecurityGroup,
	db cloudcore.SecurityGroup[cloudcore.AzureSecurityGroupExtension]) bool {

	if converter.PtrToVal(cloud.Name) != db.BaseSecurityGroup.Name {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Etag, db.Extension.Etag) {
		return true
	}

	if !assert.IsPtrBoolEqual(cloud.FlushConnection, db.Extension.FlushConnection) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.ResourceGUID, db.Extension.ResourceGUID) {
		return true
	}

	return false
}
