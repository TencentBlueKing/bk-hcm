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
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncAzureSecurityGroupOption define sync azure sg and sg rule option.
type SyncAzureSecurityGroupOption struct {
	AccountID         string   `json:"account_id" validate:"required"`
	Region            string   `json:"region" validate:"required"`
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAzureSecurityGroupOption
func (opt SyncAzureSecurityGroupOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SyncAzureSecurityGroup sync azure security group and rules to hcm.
func SyncAzureSecurityGroup(kt *kit.Kit, req *SyncAzureSecurityGroupOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromAzureForSecurityGroupSync(kt, req, adaptor)
	if err != nil {
		return nil, err
	}

	dsMap, err := getDatasFromAzureDSForSecurityGroupSync(kt, req, dataCli)
	if err != nil {
		return nil, err
	}

	err = diffAzureSecurityGroupSync(kt, cloudMap, dsMap, req, dataCli, adaptor)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func getDatasFromAzureForSecurityGroupSync(kt *kit.Kit, req *SyncAzureSecurityGroupOption,
	ad *cloudclient.CloudAdaptorClient) (map[string]*SecurityGroupSyncAzureDiff, error) {

	client, err := ad.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &securitygroup.AzureListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	if len(req.CloudIDs) > 0 {
		listOpt.CloudIDs = req.CloudIDs
	}

	result, err := client.ListSecurityGroup(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list azure security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*SecurityGroupSyncAzureDiff)
	for _, one := range result {
		sg := new(SecurityGroupSyncAzureDiff)
		sg.SecurityGroup = one
		cloudMap[*one.ID] = sg
	}

	return cloudMap, nil
}

func getDatasFromAzureDSForSecurityGroupSync(kt *kit.Kit, req *SyncAzureSecurityGroupOption,
	dataCli *dataservice.Client) (map[string]*SecurityGroupSyncDS, error) {

	start := 0
	resultsHcm := make([]corecloud.BaseSecurityGroup, 0)
	for {
		dataReq := &dataproto.SecurityGroupListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "extension.resource_group_name",
						Op:    filter.JSONEqual.Factory(),
						Value: req.ResourceGroupName,
					},
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

		results, err := dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(),
			dataReq)

		if err != nil {
			logs.Errorf("from data-service list security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(results.Details) == 0 {
			break
		}

		resultsHcm = append(resultsHcm, results.Details...)
		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	dsMap := make(map[string]*SecurityGroupSyncDS)
	for _, result := range resultsHcm {
		sg := new(SecurityGroupSyncDS)
		sg.IsUpdated = false
		sg.HcSecurityGroup = result
		dsMap[result.CloudID] = sg
	}

	return dsMap, nil
}

func diffAzureSecurityGroupSync(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAzureDiff, dsMap map[string]*SecurityGroupSyncDS,
	req *SyncAzureSecurityGroupOption, dataCli *dataservice.Client, adaptor *cloudclient.CloudAdaptorClient) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync azure SecurityGroup delete operate rid: %s", kt.Rid)
		err := diffSecurityGroupSyncDelete(kt, deleteCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync delete azure security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range deleteCloudIDs {
			_, err := SyncAzureSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(updateCloudIDs) > 0 {
		logs.Infof("do sync azure SecurityGroup update operate rid: %s", kt.Rid)
		err := diffAzureSecurityGroupSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync update azure security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range updateCloudIDs {
			_, err := SyncAzureSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync azure security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(addCloudIDs) > 0 {
		logs.Infof("do sync azure SecurityGroup add operate rid: %s", kt.Rid)
		ids, err := diffAzureSecurityGroupSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync add azure security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range ids {
			_, err := SyncAzureSGRule(kt, req, adaptor, dataCli, id)
			if err != nil {
				logs.Errorf("sync azure security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	return nil
}

func diffAzureSecurityGroupSyncAdd(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAzureDiff,
	req *SyncAzureSecurityGroupOption, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.AzureSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.AzureSecurityGroupExtension]{
			CloudID:   *cloudMap[id].SecurityGroup.ID,
			BkBizID:   constant.UnassignedBiz,
			Region:    *cloudMap[id].SecurityGroup.Location,
			Name:      *cloudMap[id].SecurityGroup.Name,
			Memo:      nil,
			AccountID: req.AccountID,
			Extension: &corecloud.AzureSecurityGroupExtension{
				ResourceGroupName: req.ResourceGroupName,
				Etag:              cloudMap[id].SecurityGroup.Etag,
				FlushConnection:   cloudMap[id].SecurityGroup.Properties.FlushConnection,
				ResourceGUID:      cloudMap[id].SecurityGroup.Properties.ResourceGUID,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.Azure.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}

func diffAzureSecurityGroupSyncUpdate(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncAzureDiff,
	dsMap map[string]*SecurityGroupSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.AzureSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.AzureSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if *cloudMap[id].SecurityGroup.Name == dsMap[id].HcSecurityGroup.Name {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.AzureSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: *cloudMap[id].SecurityGroup.Name,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) > 0 {
		if err := dataCli.Azure.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
			updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
