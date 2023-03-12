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

package subnet

import (
	"errors"
	"fmt"

	subnetlogic "hcm/cmd/hc-service/logics/sync/subnet"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// QuerySubnetIDsAndSyncOption ...
type QuerySubnetIDsAndSyncOption struct {
	Vendor            enumor.Vendor `json:"vendor" validate:"required"`
	AccountID         string        `json:"account_id" validate:"required"`
	CloudSubnetIDs    []string      `json:"cloud_subnet_ids" validate:"required"`
	ResourceGroupName string        `json:"resource_group_name" validate:"omitempty"`
	Region            string        `json:"region" validate:"omitempty"`
	CloudVpcID        string        `json:"cloud_vpc_id" validate:"omitempty"`
}

// Validate QuerySubnetIDsAndSyncOption
func (opt *QuerySubnetIDsAndSyncOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudSubnetIDs) == 0 {
		return errors.New("cloud_subnet_ids is required")
	}

	if len(opt.CloudSubnetIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloud_subnet_ids should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// QuerySubnetIDsAndSync 查询subnet，如果不存在则同步完再进行查询.
func QuerySubnetIDsAndSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client, opt *QuerySubnetIDsAndSyncOption) (map[string]cloud.BaseSubnet, error) {

	cloudIDs := slice.Unique(opt.CloudSubnetIDs)
	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("cloud_id", cloudIDs),
		Page:   core.DefaultBasePage,
		Fields: []string{"id", "cloud_id", "vpc_id", "cloud_vpc_id"},
	}
	result, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("logics list subnet from db failed, err: %v, cloudIDs: %v, rid: %s", err, cloudIDs, kt.Rid)
		return nil, err
	}

	existMap := convSubnetCloudIDMap(result)
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

	// 如果有部分subnet不存在，则触发subnet同步
	err = batchSubnetSync(kt, adaptor, dataCli, opt, notExistCloudIDs)
	if err != nil {
		return nil, err
	}

	// 同步完，二次查询
	listReq = &core.ListReq{
		Filter: tools.ContainersExpression("cloud_id", notExistCloudIDs),
		Page:   core.DefaultBasePage,
		Fields: []string{"id", "cloud_id", "vpc_id", "cloud_vpc_id"},
	}
	notExistResult, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("logics list subnet from db failed, err: %v, cloudIDs: %v, rid: %s",
			err, notExistCloudIDs, kt.Rid)
		return nil, err
	}

	if len(notExistResult.Details) != len(cloudIDs) {
		return nil, fmt.Errorf("logics some subnet can not sync, cloudIDs: %v", notExistCloudIDs)
	}

	for cloudID, id := range convSubnetCloudIDMap(notExistResult) {
		existMap[cloudID] = id
	}

	return existMap, nil
}

func batchSubnetSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QuerySubnetIDsAndSyncOption, cloudIDs []string) error {

	switch opt.Vendor {
	case enumor.Aws:
		syncOpt := &subnetlogic.SyncAwsOption{
			AccountID: opt.AccountID,
			Region:    opt.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := subnetlogic.AwsSubnetSync(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	case enumor.TCloud:
		syncOpt := &subnetlogic.SyncTCloudOption{
			AccountID: opt.AccountID,
			Region:    opt.Region,
			CloudIDs:  cloudIDs,
		}
		if _, err := subnetlogic.TCloudSubnetSync(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	case enumor.HuaWei:
		syncOpt := &subnetlogic.SyncHuaWeiOption{
			AccountID:  opt.AccountID,
			Region:     opt.Region,
			CloudVpcID: opt.CloudVpcID,
			CloudIDs:   cloudIDs,
		}
		if _, err := subnetlogic.SyncHuaWeiSubnet(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	case enumor.Azure:
		syncOpt := &hcservice.AzureResourceSyncReq{
			AccountID:         opt.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudVpcID:        opt.CloudVpcID,
			CloudIDs:          cloudIDs,
		}
		if _, err := subnetlogic.AzureSubnetSync(kt, syncOpt, adaptor, dataCli); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown %s vendor", opt.Vendor)
	}
	return nil
}

func convSubnetCloudIDMap(result *protocloud.SubnetListResult) map[string]cloud.BaseSubnet {
	m := make(map[string]cloud.BaseSubnet, len(result.Details))
	for _, one := range result.Details {
		m[one.CloudID] = one
	}
	return m
}
