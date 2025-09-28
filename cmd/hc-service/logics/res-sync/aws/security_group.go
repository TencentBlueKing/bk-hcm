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
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/aws"
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
	"hcm/pkg/tools/slice"
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

	addSlice, updateMap, delCloudIDs := common.Diff[
		securitygroup.AwsSG, cloudcore.SecurityGroup[cloudcore.AwsSecurityGroupExtension]](
		sgFromCloud, sgFromDB, isSGChange)

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
		if err = cli.updateSG(kt, params.AccountID, params.Region, updateMap); err != nil {
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
		AccountID: params.AccountID,
		Region:    params.Region,
		CloudIDs:  cloudSGIDs,
	}
	_, err = cli.SecurityGroupRule(kt, sgRuleParams, &SyncSGRuleOption{})
	if err != nil {
		logs.Errorf("[%s] sg sync sgRule failed. err: %v, accountID: %s, region: %s, rid: %s",
			err, enumor.Aws, params.AccountID, params.Region, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) updateSG(kt *kit.Kit, accountID string, region string,
	updateMap map[string]securitygroup.AwsSG) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("sg updateMap is <= 0, not update")
	}

	securityGroups := make([]protocloud.SecurityGroupBatchUpdate[cloudcore.AwsSecurityGroupExtension], 0)

	cloudVpcIDs := make([]string, 0)
	for _, one := range updateMap {
		cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(one.VpcId))
	}

	opt := &QueryVpcIDsAndSyncOption{
		AccountID:   accountID,
		Region:      region,
		CloudVpcIDs: cloudVpcIDs,
	}
	vpcMap, err := cli.queryVpcIDsAndSync(kt, opt)
	if err != nil {
		logs.Errorf("[%s] request QueryVpcIDsAndSync failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	for id, one := range updateMap {
		securityGroup := protocloud.SecurityGroupBatchUpdate[cloudcore.AwsSecurityGroupExtension]{
			ID:   id,
			Name: converter.PtrToVal(one.GroupName),
			Memo: one.Description,
			Extension: &cloudcore.AwsSecurityGroupExtension{
				VpcID:        vpcMap[converter.PtrToVal(one.VpcId)],
				CloudVpcID:   one.VpcId,
				CloudOwnerID: one.OwnerId,
			},
		}

		securityGroups = append(securityGroups, securityGroup)
	}

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[cloudcore.AwsSecurityGroupExtension]{
		SecurityGroups: securityGroups,
	}
	if err := cli.dbCli.Aws.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
		updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to update sg success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSG(kt *kit.Kit, accountID string, region string,
	addSlice []securitygroup.AwsSG) ([]string, error) {

	if len(addSlice) <= 0 {
		return nil, fmt.Errorf("sg addSlice is <= 0, not create")
	}

	createReq := &protocloud.SecurityGroupBatchCreateReq[cloudcore.AwsSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[cloudcore.AwsSecurityGroupExtension]{},
	}

	cloudVpcIDs := make([]string, 0)
	for _, one := range addSlice {
		cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(one.VpcId))
	}

	opt := &QueryVpcIDsAndSyncOption{
		AccountID:   accountID,
		Region:      region,
		CloudVpcIDs: cloudVpcIDs,
	}
	vpcMap, err := cli.queryVpcIDsAndSync(kt, opt)
	if err != nil {
		logs.Errorf("[%s] request QueryVpcIDsAndSync failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return nil, err
	}

	for _, one := range addSlice {
		securityGroup := protocloud.SecurityGroupBatchCreate[cloudcore.AwsSecurityGroupExtension]{
			CloudID:   converter.PtrToVal(one.GroupId),
			BkBizID:   constant.UnassignedBiz,
			Region:    region,
			Name:      converter.PtrToVal(one.GroupName),
			Memo:      one.Description,
			AccountID: accountID,
			MgmtBizID: constant.UnassignedBiz,
			Extension: &cloudcore.AwsSecurityGroupExtension{
				VpcID:        vpcMap[converter.PtrToVal(one.VpcId)],
				CloudVpcID:   one.VpcId,
				CloudOwnerID: one.OwnerId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	results, err := cli.dbCli.Aws.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return nil, err
	}

	logs.Infof("[%s] sync sg to create sg success, accountID: %s, count: %d, rid: %s", enumor.Aws,
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
			enumor.Aws, checkParams, len(delSGFromCloud), kt.Rid)
		return fmt.Errorf("validate sg not exist failed, before delete")
	}

	deleteReq := &protocloud.SecurityGroupBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.SecurityGroup.BatchDeleteSecurityGroup(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete sg failed, err: %v, rid: %s", enumor.Aws,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sg to delete sg success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listSGFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]securitygroup.AwsSG, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &securitygroup.AwsListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
	}
	result, _, err := cli.cloudCli.ListSecurityGroup(kt, opt)
	if err != nil {
		if strings.Contains(err.Error(), aws.ErrSGNotFound) {
			return nil, nil
		}

		logs.Errorf("[%s] list sg from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws,
			err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil
}

func (cli *client) listSGFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.SecurityGroup[cloudcore.AwsSecurityGroupExtension], error) {

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
	result, err := cli.dbCli.Aws.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list sg from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws,
			err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

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
		resultFromDB, err := cli.dbCli.Aws.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list sg failed, err: %v, req: %v, rid: %s", enumor.Aws,
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
		opt := &securitygroup.AwsListOption{
			Region:   params.Region,
			CloudIDs: []string{one},
		}
		_, _, err := cli.cloudCli.ListSecurityGroup(kt, opt)
		if err != nil {
			if strings.Contains(err.Error(), aws.ErrSGNotFound) {
				delCloudIDs = append(delCloudIDs, one)
			}
		}
	}

	return delCloudIDs, nil
}

func isSGChange(cloud securitygroup.AwsSG, db cloudcore.SecurityGroup[cloudcore.AwsSecurityGroupExtension]) bool {

	if converter.PtrToVal(cloud.GroupName) != db.BaseSecurityGroup.Name {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Description, db.BaseSecurityGroup.Memo) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.VpcId, db.Extension.CloudVpcID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.OwnerId, db.Extension.CloudOwnerID) {
		return true
	}

	return false
}

func (cli *client) queryVpcIDsAndSync(kt *kit.Kit, opt *QueryVpcIDsAndSyncOption) (map[string]string, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	cloudVpcIDs := slice.Unique(opt.CloudVpcIDs)
	listParams := &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  cloudVpcIDs,
	}
	result, err := cli.listVpcFromDB(kt, listParams)
	if err != nil {
		logs.Errorf("list vpc from db failed, err: %v, cloudIDs: %v, rid: %s", err, cloudVpcIDs, kt.Rid)
		return nil, err
	}

	existVpcMap := convVpcCloudIDMap(result)

	// 如果相等，则Vpc全部同步到了db
	if len(result) == len(cloudVpcIDs) {
		return existVpcMap, nil
	}

	notExistCloudID := make([]string, 0)
	for _, cloudID := range cloudVpcIDs {
		if _, exist := existVpcMap[cloudID]; !exist {
			notExistCloudID = append(notExistCloudID, cloudID)
		}
	}

	params := &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  notExistCloudID,
	}
	if _, err = cli.Vpc(kt, params, new(SyncVpcOption)); err != nil {
		return nil, err
	}

	// 同步完，二次查询
	listParams = &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  notExistCloudID,
	}
	notExistResult, err := cli.listVpcFromDB(kt, listParams)
	if err != nil {
		logs.Errorf("list vpc from db failed, err: %v, cloudIDs: %v, rid: %s", err, cloudVpcIDs, kt.Rid)
		return nil, err
	}

	if len(notExistResult) != len(cloudVpcIDs) {
		return nil, fmt.Errorf("some vpc can not sync, cloudIDs: %v", notExistCloudID)
	}

	for cloudID, id := range convVpcCloudIDMap(notExistResult) {
		existVpcMap[cloudID] = id
	}

	return existVpcMap, nil
}

func convVpcCloudIDMap(result []cloudcore.Vpc[cloudcore.AwsVpcExtension]) map[string]string {
	m := make(map[string]string, len(result))
	for _, one := range result {
		m[one.CloudID] = one.ID
	}
	return m
}
