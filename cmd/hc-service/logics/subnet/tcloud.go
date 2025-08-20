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

// Package subnet defines subnet service.
package subnet

import (
	"errors"
	"fmt"

	cloudclient "hcm/cmd/hc-service/logics/cloud-adaptor"
	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
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
	"hcm/pkg/tools/converter"
)

// SyncTCloudOption define tcloud sync option.
type SyncTCloudOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate SyncTCloudOption.
func (opt SyncTCloudOption) Validate() error {
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

// TCloudSubnetCreate create tcloud subnet.
func (s *Subnet) TCloudSubnetCreate(kt *kit.Kit, req *hcservice.TCloudSubnetBatchCreateReq) (
	*core.BatchCreateResult, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create tencent cloud subnets
	subnetReqs := make([]adtysubnet.TCloudOneSubnetCreateOpt, 0, len(req.Subnets))
	for _, subnetReq := range req.Subnets {
		subnetReqs = append(subnetReqs, adtysubnet.TCloudOneSubnetCreateOpt{
			IPv4Cidr:          subnetReq.IPv4Cidr,
			Name:              subnetReq.Name,
			Zone:              subnetReq.Zone,
			CloudRouteTableID: subnetReq.CloudRouteTableID,
			Memo:              subnetReq.Memo,
		})
	}

	tcloudCreateOpt := &adtysubnet.TCloudSubnetsCreateOption{
		AccountID:  req.AccountID,
		Region:     req.Region,
		CloudVpcID: req.CloudVpcID,
		Subnets:    subnetReqs,
	}
	createdSubnets, err := cli.CreateSubnets(kt, tcloudCreateOpt)
	if err != nil {
		return nil, err
	}

	// sync hcm subnets and related route tables
	createReqs := make([]cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt], 0, len(createdSubnets))
	cloudRTIDs := make([]string, 0, len(req.Subnets))
	for _, subnet := range createdSubnets {
		createReq := convertTCloudSubnetCreateReq(&subnet, req.AccountID, req.BkBizID)
		createReqs = append(createReqs, createReq)
		cloudRTIDs = append(cloudRTIDs, createReq.CloudRouteTableID)
	}

	syncOpt := &SyncTCloudOption{
		AccountID: req.AccountID,
		Region:    req.Region,
	}
	res, err := BatchCreateTCloudSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync tcloud subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, kt.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(s.client.DataService(), cli)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudRTIDs,
	}

	_, err = syncClient.RouteTable(kt, params, &synctcloud.SyncRouteTableOption{})
	if err != nil {
		logs.Errorf("sync tcloud route-table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return res, nil
}

func convertTCloudSubnetCreateReq(data *adtysubnet.TCloudSubnet, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt]{
		AccountID:         accountID,
		CloudVpcID:        data.CloudVpcID,
		BkBizID:           bizID,
		CloudRouteTableID: converter.PtrToVal(data.Extension.CloudRouteTableID),
		CloudID:           data.CloudID,
		Name:              &data.Name,
		Region:            data.Region,
		Zone:              data.Extension.Zone,
		Ipv4Cidr:          data.Ipv4Cidr,
		Ipv6Cidr:          data.Ipv6Cidr,
		Memo:              data.Memo,
		Extension: &cloud.TCloudSubnetCreateExt{
			IsDefault:         data.Extension.IsDefault,
			CloudNetworkAclID: data.Extension.CloudNetworkAclID,
		},
	}

	return subnetReq
}

// BatchCreateTCloudSubnet ...
func BatchCreateTCloudSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *SyncTCloudOption) (
	*core.BatchCreateResult, error) {

	cloudVpcIDs := make([]string, 0, len(createResources))
	for _, one := range createResources {
		cloudVpcIDs = append(cloudVpcIDs, one.CloudVpcID)
	}

	opt := &QueryVpcIDsAndSyncOption{
		Vendor:      enumor.TCloud,
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

	createReq := &cloud.SubnetBatchCreateReq[cloud.TCloudSubnetCreateExt]{
		Subnets: createResources,
	}

	res, err := dataCli.TCloud.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		return nil, err
	}

	return res, nil
}
