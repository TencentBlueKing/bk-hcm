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
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
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

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SyncAwsSecurityGroup sync aws security group and rules to hcm.
func SyncAwsSecurityGroup(kt *kit.Kit, req *SyncAwsSecurityGroupOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromAwsForSecurityGroupSync(kt, adaptor, req)
	if err != nil {
		return nil, err
	}

	commonReq := &hcservice.SecurityGroupSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  req.CloudIDs,
	}
	dsMap, err := GetDatasFromDSForSecurityGroupSync(kt, commonReq, dataCli)
	if err != nil {
		return nil, err
	}

	err = diffAwsSecurityGroupSync(kt, cloudMap, dsMap, req, dataCli, adaptor)
	if err != nil {
		return nil, err
	}

	return nil, nil
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
	for _, one := range result.SecurityGroups {
		sg := new(SecurityGroupSyncAwsDiff)
		sg.SecurityGroup = one
		cloudMap[*one.GroupId] = sg
	}

	return cloudMap, nil
}

func diffAwsSecurityGroupSync(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAwsDiff, dsMap map[string]*SecurityGroupSyncDS,
	req *SyncAwsSecurityGroupOption, dataCli *dataservice.Client, adaptor *cloudclient.CloudAdaptorClient) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync aws SecurityGroup delete operate rid: %s", kt.Rid)
		err := diffSecurityGroupSyncDelete(kt, deleteCloudIDs, dataCli)
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
		logs.Infof("do sync aws SecurityGroup update operate rid: %s", kt.Rid)
		err := diffAwsSecurityGroupSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
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
		logs.Infof("do sync aws SecurityGroup add operate rid: %s", kt.Rid)
		ids, err := diffAwsSecurityGroupSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
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
	req *SyncAwsSecurityGroupOption, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AwsSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AwsSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.AwsSecurityGroupExtension]{
			CloudID:   *cloudMap[id].SecurityGroup.GroupId,
			BkBizID:   constant.UnassignedBiz,
			Region:    req.Region,
			Name:      *cloudMap[id].SecurityGroup.GroupName,
			Memo:      cloudMap[id].SecurityGroup.Description,
			AccountID: req.AccountID,
			Extension: &corecloud.AwsSecurityGroupExtension{
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

func diffAwsSecurityGroupSyncUpdate(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAwsDiff,
	dsMap map[string]*SecurityGroupSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.AwsSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.AwsSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if *cloudMap[id].SecurityGroup.GroupName == dsMap[id].HcSecurityGroup.Name &&
			cloudMap[id].SecurityGroup.Description == dsMap[id].HcSecurityGroup.Memo {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.AwsSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: *cloudMap[id].SecurityGroup.GroupName,
			Memo: cloudMap[id].SecurityGroup.Description,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) > 0 {
		if err := dataCli.Aws.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
			updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
