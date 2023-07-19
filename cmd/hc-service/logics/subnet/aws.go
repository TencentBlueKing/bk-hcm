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

// Package subnet defines subnet logics.
package subnet

import (
	"errors"
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service/subnet"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// AwsSubnetCreate create aws subnet.
func (s *Subnet) AwsSubnetCreate(kt *kit.Kit, opt *SubnetCreateOptions[hcservice.AwsSubnetCreateExt]) (
	*core.BatchCreateResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.adaptor.Aws(kt, opt.AccountID)
	if err != nil {
		return nil, err
	}

	// create aws subnets
	createReqs := make([]cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt], 0, len(opt.CreateReqs))
	for _, req := range opt.CreateReqs {
		awsCreateOpt := &adtysubnet.AwsSubnetCreateOption{
			Name:       req.Name,
			Memo:       req.Memo,
			CloudVpcID: opt.CloudVpcID,
			Extension: &adtysubnet.AwsSubnetCreateExt{
				Region:   opt.Region,
				Zone:     req.Extension.Zone,
				IPv4Cidr: req.Extension.IPv4Cidr,
				IPv6Cidr: req.Extension.IPv6Cidr,
			},
		}
		awsCreateRes, err := cli.CreateSubnet(kt, awsCreateOpt)
		if err != nil {
			return nil, err
		}

		createReqs = append(createReqs, convertAwsSubnetCreateReq(awsCreateRes, opt.AccountID, opt.BkBizID))
	}

	// create hcm subnets
	syncOpt := &SyncAwsOption{
		AccountID: opt.AccountID,
		Region:    opt.Region,
	}
	res, err := BatchCreateAwsSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync aws subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, kt.Rid)
		return nil, err
	}

	return res, nil
}

func convertAwsSubnetCreateReq(data *adtysubnet.AwsSubnet, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt]{
		AccountID:  accountID,
		CloudVpcID: data.CloudVpcID,
		CloudID:    data.CloudID,
		Name:       &data.Name,
		Region:     data.Extension.Region,
		Zone:       data.Extension.Zone,
		Ipv4Cidr:   data.Ipv4Cidr,
		Ipv6Cidr:   data.Ipv6Cidr,
		Memo:       data.Memo,
		BkBizID:    bizID,
		Extension: &cloud.AwsSubnetCreateExt{
			State:                       data.Extension.State,
			IsDefault:                   data.Extension.IsDefault,
			MapPublicIpOnLaunch:         data.Extension.MapPublicIpOnLaunch,
			AssignIpv6AddressOnCreation: data.Extension.AssignIpv6AddressOnCreation,
			HostnameType:                data.Extension.HostnameType,
		},
	}

	return subnetReq
}

// SyncAwsOption define aws sync option.
type SyncAwsOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate SyncAwsOption.
func (opt SyncAwsOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) == 0 {
		return errors.New("cloudIDs is required")
	}

	if len(opt.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloudIDs should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// BatchCreateAwsSubnet ...
// TODO right now this method is used by create subnet api to get created result, because sync method do not return it.
// TODO modify sync logics to return crud infos, then change this method to 'batchCreateAwsSubnet'.
func BatchCreateAwsSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *SyncAwsOption) (
	*core.BatchCreateResult, error) {

	cloudVpcIDs := make([]string, 0, len(createResources))
	for _, one := range createResources {
		cloudVpcIDs = append(cloudVpcIDs, one.CloudVpcID)
	}

	opt := &QueryVpcIDsAndSyncOption{
		Vendor:      enumor.Aws,
		AccountID:   req.AccountID,
		CloudVpcIDs: cloudVpcIDs,
		Region:      req.Region,
	}
	vpcMap, err := QueryVpcIDsAndSync(kt, adaptor, dataCli, opt)
	if err != nil {
		logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for index, resource := range createResources {
		one, exist := vpcMap[resource.CloudVpcID]
		if !exist {
			return nil, fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
		}

		createResources[index].VpcID = one
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.AwsSubnetCreateExt]{
		Subnets: createResources,
	}

	return dataCli.Aws.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
}
