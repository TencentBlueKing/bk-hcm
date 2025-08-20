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
	"strconv"

	cloudclient "hcm/cmd/hc-service/logics/cloud-adaptor"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core"
	apicloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service/subnet"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GcpSubnetCreate create gcp subnet.
func (s *Subnet) GcpSubnetCreate(kt *kit.Kit, opt *SubnetCreateOptions[hcservice.GcpSubnetCreateExt]) (
	*core.BatchCreateResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.adaptor.Gcp(kt, opt.AccountID)
	if err != nil {
		return nil, err
	}

	vpcRes, err := s.getGcpVpcByCloudID(kt, opt.CloudVpcID)
	if err != nil {
		logs.Errorf("get gcp vpc by cloud id %s failed, err: %v, rid: %s", opt.CloudVpcID, err, kt.Rid)
		return nil, err
	}

	// create gcp subnets
	createdIDs := make([]string, 0, len(opt.CreateReqs))
	for _, req := range opt.CreateReqs {
		gcpCreateOpt := &adtysubnet.GcpSubnetCreateOption{
			Name:       req.Name,
			Memo:       req.Memo,
			CloudVpcID: vpcRes.Extension.SelfLink,
			Extension: &adtysubnet.GcpSubnetCreateExt{
				Region:                req.Extension.Region,
				IPv4Cidr:              req.Extension.IPv4Cidr,
				PrivateIpGoogleAccess: req.Extension.PrivateIpGoogleAccess,
				EnableFlowLogs:        req.Extension.EnableFlowLogs,
			},
		}
		createdID, err := cli.CreateSubnet(kt, gcpCreateOpt)
		if err != nil {
			logs.Errorf("create subnet failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		cloudID := strconv.FormatUint(createdID, 10)
		createdIDs = append(createdIDs, cloudID)
	}

	// get created subnets
	subnetRes, err := cli.ListSubnet(kt, &adtysubnet.GcpSubnetListOption{
		GcpListOption: adcore.GcpListOption{CloudIDs: createdIDs,
			Page: &adcore.GcpPage{PageSize: adcore.GcpQueryLimit}},
		Region: opt.Region,
	})
	if err != nil {
		logs.Errorf("get subnet failed, err: %v,s, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(subnetRes.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "created subnets are not found")
	}

	createReqs := make([]cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt], 0, len(subnetRes.Details))
	for _, subnet := range subnetRes.Details {
		createReqs = append(createReqs, convertGcpSubnetCreateReq(&subnet, opt.AccountID, opt.CloudVpcID,
			opt.BkBizID))
	}

	// create hcm subnets
	syncOpt := &SyncGcpOption{
		AccountID: opt.AccountID,
		Region:    opt.Region,
	}
	res, err := BatchCreateGcpSubnet(kt, createReqs, s.client.DataService(), s.adaptor, syncOpt)
	if err != nil {
		logs.Errorf("sync gcp subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, kt.Rid)
		return nil, err
	}

	return res, nil
}

func (s *Subnet) getGcpVpcByCloudID(kt *kit.Kit, cloudVpcID string) (*apicloud.Vpc[apicloud.GcpVpcExtension], error) {
	// get gcp vpc self link by cloud id
	vpcReq := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", cloudVpcID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"extension"},
	}
	vpcRes, err := s.client.DataService().Gcp.Vpc.ListVpcExt(kt, vpcReq)
	if err != nil {
		logs.Errorf("get vpc by cloud id %s failed, err: %v, rid: %s", cloudVpcID, err, kt.Rid)
		return nil, err
	}

	if len(vpcRes.Details) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "gcp vpc(cloud id: %s) not exists", cloudVpcID)
	}
	return &vpcRes.Details[0], nil
}

func convertGcpSubnetCreateReq(data *adtysubnet.GcpSubnet, accountID, cloudVpcID string,
	bizID int64) cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]{
		AccountID:  accountID,
		CloudVpcID: cloudVpcID,
		CloudID:    data.CloudID,
		Name:       &data.Name,
		Region:     data.Region,
		Ipv4Cidr:   data.Ipv4Cidr,
		Ipv6Cidr:   data.Ipv6Cidr,
		Memo:       data.Memo,
		BkBizID:    bizID,
		Extension: &cloud.GcpSubnetCreateExt{
			VpcSelfLink:           data.CloudVpcID,
			SelfLink:              data.Extension.SelfLink,
			StackType:             data.Extension.StackType,
			Ipv6AccessType:        data.Extension.Ipv6AccessType,
			GatewayAddress:        data.Extension.GatewayAddress,
			PrivateIpGoogleAccess: data.Extension.PrivateIpGoogleAccess,
			EnableFlowLogs:        data.Extension.EnableFlowLogs,
		},
	}

	return subnetReq
}

// SyncGcpOption define gcp sync option.
type SyncGcpOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// Validate SyncGcpOption.
func (opt SyncGcpOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.SelfLinks) == 0 && len(opt.CloudIDs) == 0 {
		return errors.New("self_links or cloud_ids is required")
	}

	if len(opt.SelfLinks) != 0 && len(opt.CloudIDs) != 0 {
		return errors.New("self_links or cloud_ids only one can be set")
	}

	if len(opt.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloudIDs should <= %d", core.DefaultMaxPageLimit)
	}

	if len(opt.SelfLinks) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("selfLinks should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// BatchCreateGcpSubnet ...
// TODO right now this method is used by create subnet api to get created result, because sync method do not return it.
// TODO modify sync logics to return crud infos, then change this method to 'batchCreateGcpSubnet'.
func BatchCreateGcpSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *SyncGcpOption) (
	*core.BatchCreateResult, error) {

	selfLinks := make([]string, 0, len(createResources))
	for _, one := range createResources {
		selfLinks = append(selfLinks, one.Extension.VpcSelfLink)
	}

	vpcMap, err := QueryVpcIDsAndSyncForGcp(kt, adaptor, dataCli, req.AccountID, selfLinks)
	if err != nil {
		logs.Errorf("query vpcIDs and sync for gcp failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for index, resource := range createResources {
		one, exist := vpcMap[resource.Extension.VpcSelfLink]
		if !exist {
			return nil, fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
		}

		createResources[index].VpcID = one.ID
		createResources[index].CloudVpcID = one.CloudID
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.GcpSubnetCreateExt]{
		Subnets: createResources,
	}
	return dataCli.Gcp.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
}
