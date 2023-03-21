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
	"errors"
	"fmt"

	securitygrouplogics "hcm/cmd/hc-service/logics/sync/security-group"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// QuerySecurityGroupIDsAndSyncOption ...
type QuerySecurityGroupIDsAndSyncOption struct {
	Vendor                enumor.Vendor `json:"vendor" validate:"required"`
	AccountID             string        `json:"account_id" validate:"required"`
	CloudSecurityGroupIDs []string      `json:"cloud_security_group_ids" validate:"required"`
	ResourceGroupName     string        `json:"resource_group_name" validate:"omitempty"`
	Region                string        `json:"region" validate:"omitempty"`
}

// Validate QuerySecurityGroupIDsAndSyncOption
func (opt *QuerySecurityGroupIDsAndSyncOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudSecurityGroupIDs) == 0 {
		return errors.New("cloud_security_group_ids is required")
	}

	if len(opt.CloudSecurityGroupIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloud_security_group_ids should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// QuerySecurityGroupIDsAndSync 查询安全组，如果不存在则同步完再进行查询.
func QuerySecurityGroupIDsAndSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client, opt *QuerySecurityGroupIDsAndSyncOption) (map[string]string, error) {
	if len(opt.CloudSecurityGroupIDs) <= 0 {
		return make(map[string]string), nil
	}

	cloudIDs := slice.Unique(opt.CloudSecurityGroupIDs)

	listReq := &protocloud.SecurityGroupListReq{
		Filter: tools.ContainersExpression("cloud_id", cloudIDs),
		Page:   core.DefaultBasePage,
		Field:  []string{"id", "cloud_id"},
	}
	result, err := dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("logics list security group from db failed, err: %v, cloudIDs: %v, rid: %s",
			err, cloudIDs, kt.Rid)
		return nil, err
	}

	existMap := convSecurityGroupCloudIDMap(result)

	// 如果相等，则全部同步到了db
	if len(result.Details) == len(cloudIDs) {
		return existMap, nil
	}

	notExistCloudIDs := make([]string, 0)
	for _, cloudID := range cloudIDs {
		if _, exist := existMap[cloudID]; !exist {
			notExistCloudIDs = append(notExistCloudIDs, cloudID)
		}
	}

	// 如果有部分不存在，则触发同步
	err = batchSecurityGroupSync(kt, adaptor, dataCli, opt, notExistCloudIDs)
	if err != nil {
		return nil, err
	}

	// 同步完，二次查询
	listReq = &protocloud.SecurityGroupListReq{
		Filter: tools.ContainersExpression("cloud_id", notExistCloudIDs),
		Page:   core.DefaultBasePage,
		Field:  []string{"id", "cloud_id"},
	}
	notExistResult, err := dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("logics list security group from db failed, err: %v, notExistCloudIDs: %v, rid: %s",
			err, notExistCloudIDs, kt.Rid)
		return nil, err
	}

	for cloudID, id := range convSecurityGroupCloudIDMap(notExistResult) {
		existMap[cloudID] = id
	}

	return existMap, nil
}

func batchSecurityGroupSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QuerySecurityGroupIDsAndSyncOption, cloudIDs []string) error {

	switch opt.Vendor {
	case enumor.Aws:
		syncOpt := &securitygrouplogics.SyncAwsSecurityGroupOption{
			AccountID: opt.AccountID,
			Region:    opt.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := securitygrouplogics.SyncAwsSecurityGroup(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	case enumor.TCloud:
		syncOpt := &securitygrouplogics.SyncTCloudSecurityGroupOption{
			AccountID: opt.AccountID,
			Region:    opt.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := securitygrouplogics.SyncTCloudSecurityGroup(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	case enumor.HuaWei:
		syncOpt := &securitygrouplogics.SyncHuaWeiSecurityGroupOption{
			AccountID: opt.AccountID,
			Region:    opt.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := securitygrouplogics.SyncHuaWeiSecurityGroup(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	case enumor.Azure:
		syncOpt := &securitygrouplogics.SyncAzureSecurityGroupOption{
			AccountID:         opt.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          cloudIDs,
		}
		if _, err := securitygrouplogics.SyncAzureSecurityGroup(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown %s vendor", opt.Vendor)
	}
	return nil
}

func convSecurityGroupCloudIDMap(result *protocloud.SecurityGroupListResult) map[string]string {
	m := make(map[string]string, len(result.Details))
	for _, one := range result.Details {
		m[one.CloudID] = one.ID
	}
	return m
}
