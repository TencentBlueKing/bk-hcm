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

package tcloud

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/tcloud"
	typecore "hcm/pkg/adaptor/types/core"
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

	addSlice, updateMap, delCloudIDs := common.Diff[securitygroup.TCloudSG, cloudcore.SecurityGroup[cloudcore.TCloudSecurityGroupExtension]](
		sgFromCloud, sgFromDB, isSGChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSG(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		_, err = cli.createSG(kt, params.AccountID, params.Region, addSlice)
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

	sgIDs := make([]string, 0, len(sgFromDB))
	for _, one := range sgFromDB {
		sgIDs = append(sgIDs, one.ID)
	}

	sgRuleParams := &SyncBaseParams{
		AccountID: params.AccountID,
		Region:    params.Region,
		CloudIDs:  sgIDs,
	}
	_, err = cli.SecurityGroupRule(kt, sgRuleParams, &SyncSGRuleOption{})
	if err != nil {
		logs.Errorf("[%s] sg sync sgRule failed. err: %v, accountID: %s, region: %s, rid: %s",
			enumor.TCloud, err, params.AccountID, params.Region, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) updateSG(kt *kit.Kit, accountID string,
	updateMap map[string]securitygroup.TCloudSG) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("sg updateMap is <= 0, not update")
	}

	securityGroups := make([]protocloud.SecurityGroupBatchUpdate[cloudcore.TCloudSecurityGroupExtension], 0)

	for id, one := range updateMap {
		tagMap := core.TagMap{}
		for _, tag := range one.TagSet {
			tagMap.Set(converter.PtrToVal(tag.Key), converter.PtrToVal(tag.Value))
		}

		securityGroup := protocloud.SecurityGroupBatchUpdate[cloudcore.TCloudSecurityGroupExtension]{
			ID:   id,
			Name: converter.PtrToVal(one.SecurityGroupName),
			Memo: one.SecurityGroupDesc,
			Extension: &cloudcore.TCloudSecurityGroupExtension{
				CloudProjectID: one.ProjectId,
			},
			CloudCreatedTime: converter.PtrToVal(one.CreatedTime),
			CloudUpdateTime:  converter.PtrToVal(one.UpdateTime),
			Tags:             tagMap,
		}

		securityGroups = append(securityGroups, securityGroup)
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[cloudcore.TCloudSecurityGroupExtension]{
		SecurityGroups: securityGroups,
	}
	if err := cli.dbCli.TCloud.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
		updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to update sg success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSG(kt *kit.Kit, accountID string, region string,
	addSlice []securitygroup.TCloudSG) ([]string, error) {

	if len(addSlice) <= 0 {
		return nil, fmt.Errorf("sg addSlice is <= 0, not create")
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[cloudcore.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[cloudcore.TCloudSecurityGroupExtension]{},
	}

	for _, one := range addSlice {
		tagMap := core.TagMap{}
		for _, tag := range one.TagSet {
			tagMap.Set(converter.PtrToVal(tag.Key), converter.PtrToVal(tag.Value))
		}
		securityGroup := protocloud.SecurityGroupBatchCreate[cloudcore.TCloudSecurityGroupExtension]{
			CloudID:   converter.PtrToVal(one.SecurityGroupId),
			BkBizID:   constant.UnassignedBiz,
			Region:    region,
			Name:      converter.PtrToVal(one.SecurityGroupName),
			Memo:      one.SecurityGroupDesc,
			AccountID: accountID,
			Extension: &cloudcore.TCloudSecurityGroupExtension{
				CloudProjectID: one.ProjectId,
			},
			CloudCreatedTime: converter.PtrToVal(one.CreatedTime),
			CloudUpdateTime:  converter.PtrToVal(one.UpdateTime),
			Tags:             tagMap,
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	results, err := cli.dbCli.TCloud.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud security group failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return nil, err
	}

	logs.Infof("[%s] sync sg to create sg success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
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
			enumor.TCloud, checkParams, len(delSGFromCloud), kt.Rid)
		return fmt.Errorf("validate sg not exist failed, before delete")
	}

	deleteReq := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.SecurityGroup.BatchDeleteSecurityGroup(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete sg failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to delete sg success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listSGFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]securitygroup.TCloudSG, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &securitygroup.TCloudListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &typecore.TCloudPage{
			Offset: 0,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListSecurityGroupNew(kt, opt)
	if err != nil {
		if strings.Contains(err.Error(), tcloud.ErrNotFound) {
			return nil, nil
		}

		logs.Errorf("[%s] list sg from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listSGFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.SecurityGroup[cloudcore.TCloudSecurityGroupExtension], error) {

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
	result, err := cli.dbCli.TCloud.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list sg from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.TCloud,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// RemoveSecurityGroupDeleteFromCloud remove security group delete from cloud
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
		resultFromDB, err := cli.dbCli.TCloud.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list sg failed, err: %v, req: %v, rid: %s", enumor.TCloud,
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

		var delCloudIDs []string
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			delCloudIDs, err = cli.listRemoveSGID(kt, params)
			if err != nil {
				return err
			}
		}

		if len(delCloudIDs) != 0 {
			if err = cli.deleteSG(kt, accountID, region, delCloudIDs); err != nil {
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

func (cli *client) listRemoveSGID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delCloudIDs := make([]string, 0)
	for _, one := range params.CloudIDs {
		opt := &securitygroup.TCloudListOption{
			Region:   params.Region,
			CloudIDs: []string{one},
		}
		_, err := cli.cloudCli.ListSecurityGroupNew(kt, opt)
		if err != nil {
			if strings.Contains(err.Error(), tcloud.ErrNotFound) {
				delCloudIDs = append(delCloudIDs, one)
			}
		}
	}

	return delCloudIDs, nil
}

func isSGChange(cloud securitygroup.TCloudSG, db cloudcore.SecurityGroup[cloudcore.TCloudSecurityGroupExtension]) bool {

	if converter.PtrToVal(cloud.SecurityGroupName) != db.BaseSecurityGroup.Name {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SecurityGroupDesc, db.BaseSecurityGroup.Memo) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.ProjectId, db.Extension.CloudProjectID) {
		return true
	}

	if converter.PtrToVal(cloud.CreatedTime) != db.BaseSecurityGroup.CloudCreatedTime {
		return true
	}

	if converter.PtrToVal(cloud.UpdateTime) != db.BaseSecurityGroup.CloudUpdateTime {
		return true
	}

	if len(cloud.TagSet) != len(db.BaseSecurityGroup.Tags) {
		return true
	}

	for _, tag := range cloud.TagSet {
		value, ok := db.BaseSecurityGroup.Tags.Get(converter.PtrToVal(tag.Key))
		if !ok {
			return true
		}
		if value != converter.PtrToVal(tag.Value) {
			return true
		}
	}

	return false
}
