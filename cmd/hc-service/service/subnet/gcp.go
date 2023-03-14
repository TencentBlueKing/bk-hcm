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
	"strconv"

	syncsubnet "hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/cmd/hc-service/service/sync"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GcpSubnetCreate create gcp subnet.
func (s subnet) GcpSubnetCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.SubnetCreateReq[hcservice.GcpSubnetCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// get gcp vpc self link by cloud id
	vpcReq := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", req.CloudVpcID),
		Page:   core.DefaultBasePage,
		Fields: []string{"extension"},
	}
	vpcRes, err := s.cs.DataService().Gcp.Vpc.ListVpcExt(cts.Kit.Ctx, cts.Kit.Header(), vpcReq)
	if err != nil {
		logs.Errorf("get vpc by cloud id %s failed, err: %v, rid: %s", req.CloudVpcID, err, cts.Kit.Rid)
		return nil, err
	}

	if len(vpcRes.Details) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "gcp vpc(cloud id: %s) not exists", req.CloudVpcID)
	}

	// create gcp subnet
	gcpCreateOpt := &types.GcpSubnetCreateOption{
		Name:       req.Name,
		Memo:       req.Memo,
		CloudVpcID: vpcRes.Details[0].Extension.SelfLink,
		Extension: &types.GcpSubnetCreateExt{
			Region:                req.Extension.Region,
			IPv4Cidr:              req.Extension.IPv4Cidr,
			PrivateIpGoogleAccess: req.Extension.PrivateIpGoogleAccess,
			EnableFlowLogs:        req.Extension.EnableFlowLogs,
		},
	}
	createdID, err := cli.CreateSubnet(cts.Kit, gcpCreateOpt)
	if err != nil {
		return nil, err
	}

	// get created subnet
	subnetRes, err := cli.ListSubnet(cts.Kit, &types.GcpSubnetListOption{
		GcpListOption: adcore.GcpListOption{CloudIDs: []string{strconv.FormatUint(createdID, 10)}},
		Region:        req.Extension.Region,
	})
	if err != nil {
		logs.Errorf("get subnet failed, err: %v,s, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(subnetRes.Details) == 0 {
		return nil, errf.New(errf.RecordNotFound, "created subnet is not found")
	}

	// create hcm subnet
	sync.SleepBeforeSync()

	syncOpt := &syncsubnet.SyncGcpOption{
		AccountID: req.AccountID,
		Region:    req.Extension.Region,
	}
	createReqs := []cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]{convertGcpSubnetCreateReq(&subnetRes.Details[0],
		req.CloudVpcID, req.AccountID, req.BkBizID)}
	res, err := syncsubnet.BatchCreateGcpSubnet(cts.Kit, createReqs, s.cs.DataService(), s.ad, syncOpt)
	if err != nil {
		logs.Errorf("sync gcp subnet failed, err: %v, reqs: %+v, rid: %s", err, createReqs, cts.Kit.Rid)
		return nil, err
	}

	return core.CreateResult{ID: res.IDs[0]}, nil
}

func convertGcpSubnetCreateReq(data *types.GcpSubnet, cloudVpcID, accountID string,
	bizID int64) cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt] {

	subnetReq := cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]{
		AccountID:  accountID,
		CloudVpcID: cloudVpcID,
		CloudID:    data.CloudID,
		Name:       &data.Name,
		Region:     data.Extension.Region,
		Ipv4Cidr:   data.Ipv4Cidr,
		Ipv6Cidr:   data.Ipv6Cidr,
		Memo:       data.Memo,
		BkBizID:    bizID,
		Extension: &cloud.GcpSubnetCreateExt{
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

// GcpSubnetUpdate update gcp subnet.
func (s subnet) GcpSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().Gcp.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.GcpSubnetUpdateOption{
		SubnetUpdateOption: types.SubnetUpdateOption{
			ResourceID: getRes.CloudID,
			Data:       &types.BaseSubnetUpdateData{Memo: req.Memo},
		},
		Region: getRes.Region,
	}
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.GcpSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().Gcp.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GcpSubnetDelete delete gcp subnet.
func (s subnet) GcpSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().Gcp.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	if getRes.Extension == nil {
		return nil, errf.New(errf.InvalidParameter, "subnet extension is empty")
	}

	delOpt := &adcore.BaseRegionalDeleteOption{
		BaseDeleteOption: adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
		Region:           getRes.Region,
	}
	err = cli.DeleteSubnet(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = s.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
