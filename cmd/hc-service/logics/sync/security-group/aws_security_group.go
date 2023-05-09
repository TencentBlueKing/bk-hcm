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

package securitygroup

import (
	"fmt"

	"hcm/cmd/hc-service/logics/sync/logics"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	typescore "hcm/pkg/adaptor/types/core"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncAwsSecurityGroupOption define sync aws sg and sg rule option.
type SyncAwsSecurityGroupOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAwsSecurityGroupOption
func (opt SyncAwsSecurityGroupOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.SGBatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.SGBatchOperationMaxLimit)
	}

	return nil
}

// SyncAwsSecurityGroup sync aws security group and rules to hcm.
func SyncAwsSecurityGroup(kt *kit.Kit, req *SyncAwsSecurityGroupOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cloudMap, err := getDatasFromAwsForSecurityGroupSync(kt, adaptor, req)
	if err != nil {
		return nil, err
	}

	dsMap, err := getDatasFromDSForAwsSGSync(kt, req, dataCli)
	if err != nil {
		return nil, err
	}

	err = diffAwsSecurityGroupSync(kt, cloudMap, dsMap, req, dataCli, adaptor)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func getDatasFromDSForAwsSGSync(kt *kit.Kit, req *SyncAwsSecurityGroupOption,
	dataCli *dataservice.Client) (map[string]*AwsSecurityGroupSyncDS, error) {

	start := 0
	resultsHcm := make([]corecloud.SecurityGroup[corecloud.AwsSecurityGroupExtension], 0)
	for {
		dataReq := &core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				},
			},
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Aws.SecurityGroup.ListSecurityGroupExt(kt.Ctx, kt.Header(),
			dataReq)
		if err != nil {
			logs.Errorf("from data-service list security group failed, err: %v, rid: %s", err, kt.Rid)
		}

		if results == nil || len(results.Details) == 0 {
			break
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	dsMap := make(map[string]*AwsSecurityGroupSyncDS)
	for _, result := range resultsHcm {
		sg := new(AwsSecurityGroupSyncDS)
		sg.IsUpdated = false
		sg.HcSecurityGroup = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

func getDatasFromAwsForSecurityGroupSync(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient,
	req *SyncAwsSecurityGroupOption) (map[string]*SecurityGroupSyncAwsDiff, error) {

	client, err := ad.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &securitygroup.AwsListOption{
		Region: req.Region,
	}
	if len(req.CloudIDs) > 0 {
		listOpt.CloudIDs = req.CloudIDs
	}

	result, err := client.ListSecurityGroup(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*SecurityGroupSyncAwsDiff)
	if result != nil {
		for _, one := range result.SecurityGroups {
			sg := new(SecurityGroupSyncAwsDiff)
			sg.SecurityGroup = one
			cloudMap[*one.GroupId] = sg
		}
	}

	return cloudMap, nil
}

func diffAwsSecurityGroupSync(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAwsDiff, dsMap map[string]*AwsSecurityGroupSyncDS,
	req *SyncAwsSecurityGroupOption, dataCli *dataservice.Client, adaptor *cloudclient.CloudAdaptorClient) error {

	addCloudIDs := make([]string, 0)
	for id := range cloudMap {
		if _, ok := dsMap[id]; !ok {
			addCloudIDs = append(addCloudIDs, id)
		} else {
			dsMap[id].IsUpdated = true
		}
	}

	deleteCloudIDs := make([]string, 0)
	updateCloudIDs := make([]string, 0)
	for id, one := range dsMap {
		if !one.IsUpdated {
			deleteCloudIDs = append(deleteCloudIDs, id)
		} else {
			updateCloudIDs = append(updateCloudIDs, id)
		}
	}

	if len(deleteCloudIDs) > 0 {
		err := DiffSecurityGroupSyncDelete(kt, deleteCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync delete aws security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range deleteCloudIDs {
			_, err := SyncAwsSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync aws security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(updateCloudIDs) > 0 {
		err := diffAwsSecurityGroupSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli, req, adaptor)
		if err != nil {
			logs.Errorf("sync update aws security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range updateCloudIDs {
			_, err := SyncAwsSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync aws security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(addCloudIDs) > 0 {
		ids, err := diffAwsSecurityGroupSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli, adaptor)
		if err != nil {
			logs.Errorf("sync add aws security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range ids {
			_, err := SyncAwsSGRule(kt, req, adaptor, dataCli, id)
			if err != nil {
				logs.Errorf("sync aws security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	return nil
}

func diffAwsSecurityGroupSyncAdd(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAwsDiff,
	req *SyncAwsSecurityGroupOption, addCloudIDs []string, dataCli *dataservice.Client, ad *cloudclient.CloudAdaptorClient) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AwsSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AwsSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {

		vpcID := ""
		if cloudMap[id].SecurityGroup.VpcId != nil {
			opt := &logics.QueryVpcIDsAndSyncOption{
				Vendor:      enumor.Aws,
				AccountID:   req.AccountID,
				Region:      req.Region,
				CloudVpcIDs: []string{*cloudMap[id].SecurityGroup.VpcId},
			}
			vpcMap, err := logics.QueryVpcIDsAndSync(kt, ad, dataCli, opt)
			if err != nil {
				logs.Errorf("request QueryVpcIDsAndSync failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			if id, ok := vpcMap[*cloudMap[id].SecurityGroup.VpcId]; ok {
				vpcID = id
			}
		}

		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.AwsSecurityGroupExtension]{
			CloudID:   converter.PtrToVal(cloudMap[id].SecurityGroup.GroupId),
			BkBizID:   constant.UnassignedBiz,
			Region:    req.Region,
			Name:      converter.PtrToVal(cloudMap[id].SecurityGroup.GroupName),
			Memo:      cloudMap[id].SecurityGroup.Description,
			AccountID: req.AccountID,
			Extension: &corecloud.AwsSecurityGroupExtension{
				VpcID:        vpcID,
				CloudVpcID:   cloudMap[id].SecurityGroup.VpcId,
				CloudOwnerID: cloudMap[id].SecurityGroup.OwnerId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.Aws.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}

func isAwsSGChange(db *AwsSecurityGroupSyncDS, cloud *SecurityGroupSyncAwsDiff, vpcID string) bool {

	if converter.PtrToVal(cloud.SecurityGroup.GroupName) != db.HcSecurityGroup.BaseSecurityGroup.Name {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SecurityGroup.Description, db.HcSecurityGroup.BaseSecurityGroup.Memo) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SecurityGroup.VpcId, db.HcSecurityGroup.Extension.CloudVpcID) {
		return true
	}

	if vpcID != db.HcSecurityGroup.Extension.VpcID {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SecurityGroup.OwnerId, db.HcSecurityGroup.Extension.CloudOwnerID) {
		return true
	}

	return false
}

func diffAwsSecurityGroupSyncUpdate(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAwsDiff, dsMap map[string]*AwsSecurityGroupSyncDS,
	updateCloudIDs []string, dataCli *dataservice.Client, req *SyncAwsSecurityGroupOption, ad *cloudclient.CloudAdaptorClient) error {

	securityGroups := make([]protocloud.SecurityGroupBatchUpdate[corecloud.AwsSecurityGroupExtension], 0)

	for _, id := range updateCloudIDs {

		vpcID := ""
		if cloudMap[id].SecurityGroup.VpcId != nil {
			opt := &logics.QueryVpcIDsAndSyncOption{
				Vendor:      enumor.Aws,
				AccountID:   req.AccountID,
				Region:      req.Region,
				CloudVpcIDs: []string{*cloudMap[id].SecurityGroup.VpcId},
			}
			vpcMap, err := logics.QueryVpcIDsAndSync(kt, ad, dataCli, opt)
			if err != nil {
				logs.Errorf("request QueryVpcIDsAndSync failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			if id, ok := vpcMap[*cloudMap[id].SecurityGroup.VpcId]; ok {
				vpcID = id
			}
		}

		if !isAwsSGChange(dsMap[id], cloudMap[id], vpcID) {
			continue
		}

		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.AwsSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: converter.PtrToVal(cloudMap[id].SecurityGroup.GroupName),
			Memo: cloudMap[id].SecurityGroup.Description,
			Extension: &corecloud.AwsSecurityGroupExtension{
				VpcID:        vpcID,
				CloudVpcID:   cloudMap[id].SecurityGroup.VpcId,
				CloudOwnerID: cloudMap[id].SecurityGroup.OwnerId,
			},
		}

		securityGroups = append(securityGroups, securityGroup)
	}

	if len(securityGroups) > 0 {
		elems := slice.Split(securityGroups, typescore.TCloudQueryLimit)
		for _, partSecurityGroups := range elems {
			updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.AwsSecurityGroupExtension]{
				SecurityGroups: partSecurityGroups,
			}
			if err := dataCli.Aws.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
				updateReq); err != nil {
				logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}

	return nil
}
