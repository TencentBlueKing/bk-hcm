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

package huawei

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
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
	"hcm/pkg/tools/converter"
)

// SyncSGOption ...
type SyncSGOption struct {
}

// Validate ...
func (opt SyncSGOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SecurityGroup ...
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

	addSlice, updateMap, delCloudIDs := common.Diff[securitygroup.HuaWeiSG,
		cloudcore.SecurityGroup[cloudcore.HuaWeiSecurityGroupExtension]](sgFromCloud, sgFromDB, isSGChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteSG(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		_, err := cli.createSG(kt, params.AccountID, params.Region, addSlice)
		if err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateSG(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	// 同步安全组规则
	sgFromDB, err = cli.listSGFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	cloudSGIDs := make([]string, 0, len(sgFromDB))
	sgMap := make(map[string]string)
	for _, one := range sgFromDB {
		cloudSGIDs = append(cloudSGIDs, one.CloudID)
		sgMap[one.CloudID] = one.ID
	}

	sgRuleParams := &SyncBaseParams{
		AccountID: params.AccountID,
		Region:    params.Region,
		CloudIDs:  cloudSGIDs,
	}
	_, err = cli.SecurityGroupRule(kt, sgRuleParams, &SyncSGRuleOption{})
	if err != nil {
		logs.Errorf("[%s] sg sync sgRule failed. err: %v, accountID: %s, region: %s, rid: %s",
			err, enumor.HuaWei, params.AccountID, params.Region, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) updateSG(kt *kit.Kit, accountID string,
	updateMap map[string]securitygroup.HuaWeiSG) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("sg updateMap is <= 0, not update")
	}

	securityGroups := make([]protocloud.SecurityGroupBatchUpdate[cloudcore.HuaWeiSecurityGroupExtension], 0)

	for id, one := range updateMap {
		securityGroup := protocloud.SecurityGroupBatchUpdate[cloudcore.HuaWeiSecurityGroupExtension]{
			ID:   id,
			Name: one.Name,
			Memo: converter.ValToPtr(one.Description),
			Extension: &cloudcore.HuaWeiSecurityGroupExtension{
				CloudProjectID:           one.ProjectId,
				CloudEnterpriseProjectID: one.EnterpriseProjectId,
			},
		}

		securityGroups = append(securityGroups, securityGroup)
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[cloudcore.HuaWeiSecurityGroupExtension]{
		SecurityGroups: securityGroups,
	}
	if err := cli.dbCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
		updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to update sg success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSG(kt *kit.Kit, accountID string, region string,
	addSlice []securitygroup.HuaWeiSG) ([]string, error) {

	if len(addSlice) <= 0 {
		return nil, fmt.Errorf("sg addSlice is <= 0, not create")
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[cloudcore.HuaWeiSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[cloudcore.HuaWeiSecurityGroupExtension]{},
	}

	for _, one := range addSlice {
		securityGroup := protocloud.SecurityGroupBatchCreate[cloudcore.HuaWeiSecurityGroupExtension]{
			CloudID:   one.Id,
			BkBizID:   constant.UnassignedBiz,
			Region:    region,
			Name:      one.Name,
			Memo:      converter.ValToPtr(one.Description),
			AccountID: accountID,
			MgmtBizID: constant.UnassignedBiz,
			Extension: &cloudcore.HuaWeiSecurityGroupExtension{
				CloudProjectID:           one.SecurityGroup.ProjectId,
				CloudEnterpriseProjectID: one.SecurityGroup.EnterpriseProjectId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	results, err := cli.dbCli.HuaWei.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return nil, err
	}

	logs.Infof("[%s] sync sg to create sg success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(addSlice), kt.Rid)

	return results.IDs, nil
}

func (cli *client) deleteSG(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("sg delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delSGFromCloud, err := cli.listSGFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delSGFromCloud) > 0 {
		logs.Errorf("[%s] validate sg not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.HuaWei, checkParams, len(delSGFromCloud), kt.Rid)
		return fmt.Errorf("validate sg not exist failed, before delete")
	}

	deleteReq := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.SecurityGroup.BatchDeleteSecurityGroup(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete sg failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to delete sg success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listSGFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]securitygroup.HuaWeiSG, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &securitygroup.HuaWeiListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &adcore.HuaWeiPage{
			Limit: converter.ValToPtr(int32(adcore.HuaWeiQueryLimit)),
		},
	}
	result, _, err := cli.cloudCli.ListSecurityGroup(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list sg from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listSGFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.SecurityGroup[cloudcore.HuaWeiSecurityGroupExtension], error) {

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
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: params.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.HuaWei.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list sg from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// RemoveSecurityGroupDeleteFromCloud ...
func (cli *client) RemoveSecurityGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
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
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: region,
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.HuaWei.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list sg failed, err: %v, req: %v, rid: %s", enumor.HuaWei,
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
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listSGFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.Id)
			}

			cloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if len(cloudIDs) > 0 {
				if err := cli.deleteSG(kt, accountID, region, cloudIDs); err != nil {
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

func isSGChange(cloud securitygroup.HuaWeiSG, db cloudcore.SecurityGroup[cloudcore.HuaWeiSecurityGroupExtension]) bool {

	if cloud.Name != db.BaseSecurityGroup.Name {
		return true
	}

	if cloud.Description != converter.PtrToVal(db.BaseSecurityGroup.Memo) {
		return true
	}

	if cloud.ProjectId != db.Extension.CloudProjectID {
		return true
	}

	if cloud.EnterpriseProjectId != db.Extension.CloudEnterpriseProjectID {
		return true
	}

	return false
}
