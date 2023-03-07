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
	"hcm/pkg/adaptor/huawei"
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

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// SyncHuaWeiSecurityGroupOption define sync huawei sg and sg rule option.
type SyncHuaWeiSecurityGroupOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncHuaWeiSecurityGroupOption
func (opt SyncHuaWeiSecurityGroupOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SyncHuaWeiSecurityGroup sync huawei security group and rules to hcm.
func SyncHuaWeiSecurityGroup(kt *kit.Kit, req *SyncHuaWeiSecurityGroupOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	cloudMap, err := getDatasFromHuaWeiForSecurityGroupSync(kt, adaptor, req)
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

	err = diffHWSecurityGroupSync(kt, cloudMap, dsMap, req, adaptor, dataCli)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func getDatasFromHuaWeiForSecurityGroupSync(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient,
	req *SyncHuaWeiSecurityGroupOption) (map[string]*SecurityGroupSyncHuaWeiDiff, error) {

	client, err := ad.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudMap := make(map[string]*SecurityGroupSyncHuaWeiDiff)
	if len(req.CloudIDs) > 0 {
		cloudMap, err = getHuaWeiSGByCloudIDsSync(kt, client, req)
		if err != nil {
			logs.Errorf("request to list huawei security group by cloud_ids failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	} else {
		cloudMap, err = getHuaWeiSGAllSync(kt, client, req)
		if err != nil {
			logs.Errorf("request to list all huawei security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return cloudMap, nil
}

func getHuaWeiSGByCloudIDsSync(kt *kit.Kit, client *huawei.HuaWei,
	req *SyncHuaWeiSecurityGroupOption) (map[string]*SecurityGroupSyncHuaWeiDiff, error) {

	opt := &securitygroup.HuaWeiListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
	}

	datas, _, err := client.ListSecurityGroup(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list huawei security group failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*SecurityGroupSyncHuaWeiDiff)
	for _, data := range *datas {
		sg := new(SecurityGroupSyncHuaWeiDiff)
		sg.SecurityGroup = data
		cloudMap[data.Id] = sg
	}

	return cloudMap, nil
}

func getHuaWeiSGAllSync(kt *kit.Kit, client *huawei.HuaWei,
	req *SyncHuaWeiSecurityGroupOption) (map[string]*SecurityGroupSyncHuaWeiDiff, error) {

	datasCloud := []model.SecurityGroup{}

	limit := int32(typcore.HuaWeiQueryLimit)
	var marker *string = nil
	for {
		opt := &securitygroup.HuaWeiListOption{
			Region: req.Region,
			Page:   &typcore.HuaWeiPage{Limit: &limit, Marker: marker},
		}

		datas, pageInfo, err := client.ListSecurityGroup(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		datasCloud = append(datasCloud, *datas...)

		marker = pageInfo.NextMarker
		if len(*datas) == 0 || pageInfo.NextMarker == nil {
			break
		}
	}

	cloudMap := make(map[string]*SecurityGroupSyncHuaWeiDiff)
	for _, data := range datasCloud {
		sg := new(SecurityGroupSyncHuaWeiDiff)
		sg.SecurityGroup = data
		cloudMap[data.Id] = sg
	}

	return cloudMap, nil
}

func diffHWSecurityGroupSync(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncHuaWeiDiff, dsMap map[string]*SecurityGroupSyncDS,
	req *SyncHuaWeiSecurityGroupOption, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) error {

	addCloudIDs := getAddCloudIDs(cloudMap, dsMap)
	deleteCloudIDs, updateCloudIDs := getDeleteAndUpdateCloudIDs(dsMap)

	if len(deleteCloudIDs) > 0 {
		logs.Infof("do sync huawei SecurityGroup delete operate, rid: %s", kt.Rid)
		err := diffSecurityGroupSyncDelete(kt, deleteCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync delete huawei security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range deleteCloudIDs {
			_, err := SyncHuaWeiSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync huawei security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(updateCloudIDs) > 0 {
		logs.Infof("do sync huawei SecurityGroup update operate, rid: %s", kt.Rid)
		err := diffHWSecurityGroupSyncUpdate(kt, cloudMap, dsMap, updateCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync update huawei security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range updateCloudIDs {
			_, err := SyncHuaWeiSGRule(kt, req, adaptor, dataCli, dsMap[id].HcSecurityGroup.ID)
			if err != nil {
				logs.Errorf("sync huawei security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	if len(addCloudIDs) > 0 {
		logs.Infof("do sync huawei SecurityGroup add operate, rid: %s", kt.Rid)
		ids, err := diffHWSecurityGroupSyncAdd(kt, cloudMap, req, addCloudIDs, dataCli)
		if err != nil {
			logs.Errorf("sync add huawei security group failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for _, id := range ids {
			_, err := SyncHuaWeiSGRule(kt, req, adaptor, dataCli, id)
			if err != nil {
				logs.Errorf("sync huawei security group rule failed, err: %v, rid: %s", err, kt.Rid)
			}
		}
	}

	return nil
}

func diffHWSecurityGroupSyncAdd(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncHuaWeiDiff,
	req *SyncHuaWeiSecurityGroupOption, addCloudIDs []string, dataCli *dataservice.Client) ([]string, error) {

	createReq := &protocloud.SecurityGroupBatchCreateReq[corecloud.HuaWeiSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchCreate[corecloud.HuaWeiSecurityGroupExtension]{},
	}

	for _, id := range addCloudIDs {
		securityGroup := protocloud.SecurityGroupBatchCreate[corecloud.HuaWeiSecurityGroupExtension]{
			CloudID:   cloudMap[id].SecurityGroup.Id,
			BkBizID:   constant.UnassignedBiz,
			Region:    req.Region,
			Name:      cloudMap[id].SecurityGroup.Name,
			Memo:      &cloudMap[id].SecurityGroup.Description,
			AccountID: req.AccountID,
			Extension: &corecloud.HuaWeiSecurityGroupExtension{
				CloudProjectID:           cloudMap[id].SecurityGroup.ProjectId,
				CloudEnterpriseProjectID: cloudMap[id].SecurityGroup.EnterpriseProjectId,
			},
		}
		createReq.SecurityGroups = append(createReq.SecurityGroups, securityGroup)
	}

	if len(createReq.SecurityGroups) <= 0 {
		return make([]string, 0), nil
	}

	ids, err := dataCli.HuaWei.SecurityGroup.BatchCreateSecurityGroup(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("request dataservice to BatchCreateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids.IDs, nil
}

func diffHWSecurityGroupSyncUpdate(kt *kit.Kit, cloudMap map[string]*SecurityGroupSyncHuaWeiDiff,
	dsMap map[string]*SecurityGroupSyncDS, updateCloudIDs []string, dataCli *dataservice.Client) error {

	updateReq := &protocloud.SecurityGroupBatchUpdateReq[corecloud.HuaWeiSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[corecloud.HuaWeiSecurityGroupExtension]{},
	}

	for _, id := range updateCloudIDs {
		if cloudMap[id].SecurityGroup.Name == dsMap[id].HcSecurityGroup.Name &&
			cloudMap[id].SecurityGroup.Description == *dsMap[id].HcSecurityGroup.Memo {
			continue
		}
		securityGroup := protocloud.SecurityGroupBatchUpdate[corecloud.HuaWeiSecurityGroupExtension]{
			ID:   dsMap[id].HcSecurityGroup.ID,
			Name: cloudMap[id].SecurityGroup.Name,
			Memo: &cloudMap[id].SecurityGroup.Description,
		}
		updateReq.SecurityGroups = append(updateReq.SecurityGroups, securityGroup)
	}

	if len(updateReq.SecurityGroups) > 0 {
		if err := dataCli.HuaWei.SecurityGroup.BatchUpdateSecurityGroup(kt.Ctx, kt.Header(),
			updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}
