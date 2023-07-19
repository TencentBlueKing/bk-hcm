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

	syncazure "hcm/cmd/hc-service/logics/res-sync/azure"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/hc-service/subnet"
	hcservice "hcm/pkg/api/hc-service/vpc"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// ConvAzureCreateReq convert hc-service azure subnet create request to adaption options.
func ConvAzureCreateReq(req *subnet.SubnetCreateReq[subnet.AzureSubnetCreateExt]) *adtysubnet.AzureSubnetCreateOption {
	return &adtysubnet.AzureSubnetCreateOption{
		Name:       req.Name,
		Memo:       req.Memo,
		CloudVpcID: req.CloudVpcID,
		Extension: &adtysubnet.AzureSubnetCreateExt{
			ResourceGroup:        req.Extension.ResourceGroup,
			IPv4Cidr:             req.Extension.IPv4Cidr,
			IPv6Cidr:             req.Extension.IPv6Cidr,
			CloudRouteTableID:    req.Extension.CloudRouteTableID,
			NatGateway:           req.Extension.NatGateway,
			NetworkSecurityGroup: req.Extension.NetworkSecurityGroup,
		},
	}
}

// AzureSubnetSync sync azure subnets.
func (s *Subnet) AzureSubnetSync(kt *kit.Kit, req *AzureSubnetSyncOptions) (*core.BatchCreateResult, error) {
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createReqs := make([]cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt], 0, len(req.Subnets))

	routeTableIDs := make([]string, 0, len(req.Subnets))
	for _, subnet := range req.Subnets {
		createReqs = append(createReqs, convertAzureSubnetCreateReq(&subnet, req.AccountID, constant.UnassignedBiz))
		if subnet.Extension.CloudRouteTableID != nil {
			routeTableIDs = append(routeTableIDs, *subnet.Extension.CloudRouteTableID)
		}
	}

	syncOpt := &hcservice.AzureResourceSyncReq{
		AccountID:         req.AccountID,
		ResourceGroupName: req.ResourceGroup,
		CloudVpcID:        req.CloudVpcID,
	}

	res, err := BatchCreateAzureSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync azure subnet failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if len(routeTableIDs) != 0 {
		cli, err := s.adaptor.Azure(kt, req.AccountID)
		if err != nil {
			return nil, err
		}

		syncClient := syncazure.NewClient(s.client.DataService(), cli)

		params := &syncazure.SyncBaseParams{
			AccountID:         req.AccountID,
			ResourceGroupName: req.ResourceGroup,
			CloudIDs:          routeTableIDs,
		}

		_, err = syncClient.RouteTable(kt, params, &syncazure.SyncRouteTableOption{})
		if err != nil {
			logs.Errorf("sync azure route-table failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

	}

	return res, nil
}

func convertAzureSubnetCreateReq(data *adtysubnet.AzureSubnet, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]{
		AccountID:         accountID,
		CloudVpcID:        data.CloudVpcID,
		CloudID:           data.CloudID,
		Name:              &data.Name,
		Ipv4Cidr:          data.Ipv4Cidr,
		Ipv6Cidr:          data.Ipv6Cidr,
		Memo:              data.Memo,
		BkBizID:           bizID,
		CloudRouteTableID: converter.PtrToVal(data.Extension.CloudRouteTableID),
		Extension: &cloud.AzureSubnetCreateExt{
			ResourceGroupName:    data.Extension.ResourceGroupName,
			NatGateway:           data.Extension.NatGateway,
			CloudSecurityGroupID: data.Extension.NetworkSecurityGroup,
		},
	}

	return subnetReq
}

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

// BatchCreateAzureSubnet ...
// TODO right now this method is used by create subnet api to get created result, because sync method do not return it.
// TODO modify sync logics to return crud infos, then change this method to 'batchCreateAzureSubnet'.
func BatchCreateAzureSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *hcservice.AzureResourceSyncReq) (
	*core.BatchCreateResult, error) {

	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}

	listVpcOpt := &QueryVpcIDsAndSyncOption{
		Vendor:            enumor.Azure,
		AccountID:         req.AccountID,
		CloudVpcIDs:       []string{req.CloudVpcID},
		ResourceGroupName: req.ResourceGroupName,
	}
	vpcMap, err := QueryVpcIDsAndSync(kt, adaptor, dataCli, listVpcOpt)
	if err != nil {
		logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	createRes := &core.BatchCreateResult{IDs: make([]string, 0)}
	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]
		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		cloudSGIDs := make([]string, 0)
		for _, one := range newResources {
			if len(one.Extension.CloudSecurityGroupID) != 0 {
				cloudSGIDs = append(cloudSGIDs, one.Extension.CloudSecurityGroupID)
			}
		}

		listSGOpt := &QuerySecurityGroupIDsAndSyncOption{
			Vendor:                enumor.Azure,
			AccountID:             req.AccountID,
			ResourceGroupName:     req.ResourceGroupName,
			CloudSecurityGroupIDs: cloudSGIDs,
		}
		securityGroupMap, err := QuerySecurityGroupIDsAndSync(kt, adaptor, dataCli, listSGOpt)
		if err != nil {
			return nil, err
		}

		for index, resource := range newResources {
			vpcID, exist := vpcMap[resource.CloudVpcID]
			if !exist {
				return nil, fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
			}
			newResources[index].VpcID = vpcID

			if len(resource.Extension.CloudSecurityGroupID) != 0 {
				sgID, exist := securityGroupMap[resource.Extension.CloudSecurityGroupID]
				if !exist {
					return nil, fmt.Errorf("security group: %s not sync from cloud", resource.Extension.CloudSecurityGroupID)
				}
				newResources[index].Extension.SecurityGroupID = sgID
			}
		}

		createReq := &cloud.SubnetBatchCreateReq[cloud.AzureSubnetCreateExt]{
			Subnets: newResources,
		}
		res, err := dataCli.Azure.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			return nil, err
		}
		createRes.IDs = append(createRes.IDs, res.IDs...)
	}

	return createRes, nil
}
