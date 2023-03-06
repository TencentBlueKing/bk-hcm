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
	"hcm/pkg/adaptor/tcloud"
	typcore "hcm/pkg/adaptor/types/core"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// SyncTCloudSecurityGroupOption define sync tcloud sg and sg rule option.
type SyncTCloudSecurityGroupOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncTCloudSecurityGroupOption
func (opt SyncTCloudSecurityGroupOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SyncTCloudSecurityGroup sync tcloud security group and rules to hcm.
func SyncTCloudSecurityGroup(kt *kit.Kit, req *SyncTCloudSecurityGroupOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromTCloudForSecurityGroupSync(kt, req, adaptor)
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

	err = diffTCloudSecurityGroupSync(kt, cloudMap, dsMap, req, adaptor, dataCli)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func getDatasFromTCloudForSecurityGroupSync(kt *kit.Kit, req *SyncTCloudSecurityGroupOption,
	ad *cloudclient.CloudAdaptorClient) (map[string]*SecurityGroupSyncTCloudDiff, error) {

	client, err := ad.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudMap := make(map[string]*SecurityGroupSyncTCloudDiff)
	if len(req.CloudIDs) > 0 {
		cloudMap, err = getTCloudSGByCloudIDsSync(kt, client, req)
		if err != nil {
			logs.Errorf("request to list tcloud security group by cloud_ids failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	} else {
		cloudMap, err = getTCloudSGAllSync(kt, client, req)
		if err != nil {
			logs.Errorf("request to list all tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return cloudMap, nil
}

func getTCloudSGByCloudIDsSync(kt *kit.Kit, client *tcloud.TCloud,
	req *SyncTCloudSecurityGroupOption) (map[string]*SecurityGroupSyncTCloudDiff, error) {

	opt := &securitygroup.TCloudListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
	}

	datas, err := client.ListSecurityGroup(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list tcloud security group failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*SecurityGroupSyncTCloudDiff)
	for _, data := range datas {
		sg := new(SecurityGroupSyncTCloudDiff)
		sg.SecurityGroup = data
		cloudMap[*data.SecurityGroupId] = sg
	}

	return cloudMap, nil
}

func getTCloudSGAllSync(kt *kit.Kit, client *tcloud.TCloud,
	req *SyncTCloudSecurityGroupOption) (map[string]*SecurityGroupSyncTCloudDiff, error) {

	offset := 0
	datasCloud := []*vpc.SecurityGroup{}

	for {
		opt := &securitygroup.TCloudListOption{
			Region: req.Region,
			Page:   &typcore.TCloudPage{Offset: uint64(offset), Limit: uint64(typcore.TCloudQueryLimit)},
		}
		datas, err := client.ListSecurityGroup(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud security group failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}
		offset += len(datas)
		datasCloud = append(datasCloud, datas...)
		if len(datas) == 0 || uint(len(datas)) < typcore.TCloudQueryLimit {
			break
		}
	}

	cloudMap := make(map[string]*SecurityGroupSyncTCloudDiff)
	for _, data := range datasCloud {
		sg := new(SecurityGroupSyncTCloudDiff)
		sg.SecurityGroup = data
		cloudMap[*data.SecurityGroupId] = sg
	}

	return cloudMap, nil
}

func diffTCloudSecurityGroupSync(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncTCloudDiff,
	dsMap map[string]*SecurityGroupSyncDS, req *SyncTCloudSecurityGroupOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync tcloud SecurityGroup delete operate, rid: %s", kt.Rid)
		err := diffSecurityGroupSyncDelete(kt, deleteCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync delete tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range deleteCloudIDs {
			_, err := SyncTCloudSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(updateCloudIDs) > 0 {
		logs.Infof("do sync tcloud SecurityGroup update operate, rid: %s", kt.Rid)
		err := diffTCloudSecurityGroupSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync update tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range updateCloudIDs {
			_, err := SyncTCloudSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(addCloudIDs) > 0 {
		logs.Infof("do sync tcloud SecurityGroup add operate, rid: %s", kt.Rid)
		ids, err := diffTCloudSecurityGroupSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync add tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range ids {
			_, err := SyncTCloudSGRule(kt, req, adaptor, dataCli, id)
			if err != nil {
				logs.Errorf("sync tcloud security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	return nil
}

func diffTCloudSecurityGroupSyncUpdate(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncTCloudDiff,
	dsMap map[string]*SecurityGroupSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.TCloudSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if *cloudMap[id].SecurityGroup.SecurityGroupName == dsMap[id].HcSecurityGroup.Name &&
			cloudMap[id].SecurityGroup.SecurityGroupDesc == dsMap[id].HcSecurityGroup.Memo {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.TCloudSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: *cloudMap[id].SecurityGroup.SecurityGroupName,
			Memo: cloudMap[id].SecurityGroup.SecurityGroupDesc,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) > 0 {
		if err := dataCli.TCloud.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
			updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func diffTCloudSecurityGroupSyncAdd(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncTCloudDiff,
	req *SyncTCloudSecurityGroupOption, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.TCloudSecurityGroupExtension]{
			CloudID:   *cloudMap[id].SecurityGroup.SecurityGroupId,
			BkBizID:   constant.UnassignedBiz,
			Region:    req.Region,
			Name:      *cloudMap[id].SecurityGroup.SecurityGroupName,
			Memo:      cloudMap[id].SecurityGroup.SecurityGroupDesc,
			AccountID: req.AccountID,
			Extension: &corecloud.TCloudSecurityGroupExtension{
				CloudProjectID: cloudMap[id].SecurityGroup.ProjectId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return make([]string, 0), nil
	}

	results, err := dataCli.TCloud.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return results.IDs, nil
}
