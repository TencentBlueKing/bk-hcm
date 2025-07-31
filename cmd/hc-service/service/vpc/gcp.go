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

// Package vpc defines vpc service.
package vpc

import (
	"fmt"
	"strconv"
	"strings"

	"hcm/cmd/hc-service/logics/subnet"
	"hcm/pkg/adaptor/gcp"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	adrt "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	dsrt "hcm/pkg/api/data-service/cloud/route-table"
	subnetproto "hcm/pkg/api/hc-service/subnet"
	hcservice "hcm/pkg/api/hc-service/vpc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
)

// GcpVpcCreate create gcp vpc.
func (v vpc) GcpVpcCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.VpcCreateReq[hcservice.GcpVpcCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	adaptor, err := v.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create gcp vpc
	opt := &types.GcpVpcCreateOption{
		AccountID: req.AccountID,
		Name:      req.Name,
		Memo:      req.Memo,
		Extension: &types.GcpVpcCreateExt{
			AutoCreateSubnetworks: req.Extension.AutoCreateSubnetworks,
			EnableUlaInternalIpv6: req.Extension.EnableUlaInternalIpv6,
			InternalIpv6Range:     req.Extension.InternalIpv6Range,
			RoutingMode:           req.Extension.RoutingMode,
		},
	}
	vpcID, err := adaptor.CreateVpc(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// get created vpc info
	listOpt := &types.GcpListOption{CloudIDs: []string{strconv.FormatUint(vpcID, 10)},
		Page: &adcore.GcpPage{PageSize: adcore.GcpQueryLimit}}
	listRes, err := adaptor.ListVpc(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("fail to ListVpc, err: %v, vpcID, rid: %s", err, vpcID, cts.Kit.Rid)
		return nil, err
	}
	if len(listRes.Details) != 1 {
		return nil, errf.Newf(errf.Aborted, "get created vpc detail, but vpcResult count is invalid")
	}
	vpcCreated := listRes.Details[0]

	vpcCreatedID, err := v.createGcpVpcForDB(cts.Kit, &vpcCreated, req)
	if err != nil {
		logs.Errorf("create gcp vpc for db failed, err: %v, vpcID: %s, rid: %s", err, vpcCreated.CloudID, cts.Kit.Rid)
		return nil, err
	}

	// create gcp subnets
	if len(req.Extension.Subnets) == 0 {
		return core.CreateResult{ID: vpcCreatedID}, nil
	}

	regionSubnetMap := make(map[string][]subnetproto.SubnetCreateReq[subnetproto.GcpSubnetCreateExt])
	for _, s := range req.Extension.Subnets {
		regionSubnetMap[s.Extension.Region] = append(regionSubnetMap[s.Extension.Region], s)
	}

	for region, subnets := range regionSubnetMap {
		err = v.createGcpSubnetWithRetry(cts.Kit, constant.UnassignedBiz, req.AccountID, vpcCreated.CloudID, region,
			subnets)
		if err != nil {
			return nil, err
		}
	}

	err = v.createGeneratedRoute(cts.Kit, adaptor, vpcCreated.Extension.SelfLink)
	if err != nil {
		logs.Errorf("create gcp vpc and subnet success, but create route fail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return core.CreateResult{ID: vpcCreatedID}, nil
}

func (v vpc) createGcpVpcForDB(kt *kit.Kit, vpcCreated *types.GcpVpc,
	req *hcservice.VpcCreateReq[hcservice.GcpVpcCreateExt]) (string, error) {

	// create hcm vpc
	createReq := &cloud.VpcBatchCreateReq[cloud.GcpVpcCreateExt]{
		Vpcs: []cloud.VpcCreateReq[cloud.GcpVpcCreateExt]{convertGcpVpcCreateReq(req, vpcCreated)},
	}
	vpcResult, err := v.cs.DataService().Gcp.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("vpc created on cloud, but fail to BatchCreate vpc on db, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(vpcResult.IDs) != 1 {
		return "", errf.New(errf.Aborted, "create vpcResult is invalid")
	}
	return vpcResult.IDs[0], nil
}

const maxRetryCount = 10

func (v vpc) createGeneratedRoute(kt *kit.Kit, adaptor *gcp.Gcp, network string) error {
	routeOpt := &adrt.GcpListOption{
		Page:    &adcore.GcpPage{PageSize: adcore.GcpQueryLimit},
		Network: []string{network},
	}
	routeResp, err := adaptor.ListRoute(kt, routeOpt)
	if err != nil {
		logs.Errorf("[%s] list route from cloud failed, err: %v, network: %s, opt: %v, rid: %s", enumor.Gcp,
			err, network, routeOpt, kt.Rid)
		return err
	}
	if len(routeResp.Details) == 0 {
		return nil
	}
	routes := slice.Map(routeResp.Details, func(r adrt.GcpRoute) dsrt.GcpRouteCreateReq {
		return dsrt.GcpRouteCreateReq{
			CloudID:          r.CloudID,
			SelfLink:         r.SelfLink,
			Network:          r.Network,
			Name:             r.Name,
			DestRange:        r.DestRange,
			NextHopGateway:   r.NextHopIp,
			NextHopIlb:       r.NextHopIlb,
			NextHopInstance:  r.NextHopInstance,
			NextHopIp:        r.NextHopGateway,
			NextHopNetwork:   r.NextHopNetwork,
			NextHopPeering:   r.NextHopPeering,
			NextHopVpnTunnel: r.NextHopVpnTunnel,
			Priority:         r.Priority,
			RouteStatus:      r.RouteStatus,
			RouteType:        r.RouteType,
			Tags:             r.Tags,
			Memo:             r.Memo,
		}
	})

	_, err = v.cs.DataService().Gcp.RouteTable.BatchCreateRoute(kt, &dsrt.GcpRouteBatchCreateReq{GcpRoutes: routes})
	if err != nil {
		logs.Errorf("fail to create gcp route, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// createGcpSubnetWithRetry create gcp subnet, retry when vpc is not ready.
func (v vpc) createGcpSubnetWithRetry(kt *kit.Kit, bizID int64, accountID, cloudVpcID, region string,
	subnets []subnetproto.SubnetCreateReq[subnetproto.GcpSubnetCreateExt]) error {

	rty := retry.NewRetryPolicy(maxRetryCount, [2]uint{10000, 15000})

	for {
		if rty.RetryCount() == maxRetryCount {
			return fmt.Errorf("create subnet failed count exceeds %d", maxRetryCount)
		}

		subnetCreateOpt := &subnet.SubnetCreateOptions[subnetproto.GcpSubnetCreateExt]{
			BkBizID:    bizID,
			AccountID:  accountID,
			Region:     region,
			CloudVpcID: cloudVpcID,
			CreateReqs: subnets,
		}
		_, err := v.subnet.GcpSubnetCreate(kt, subnetCreateOpt)
		if err != nil {
			if strings.Contains(err.Error(), "resourceNotReady") {
				rty.Sleep()
				continue
			}

			logs.Errorf("create subnet failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		return nil
	}
}

func convertGcpVpcCreateReq(req *hcservice.VpcCreateReq[hcservice.GcpVpcCreateExt],
	data *types.GcpVpc) cloud.VpcCreateReq[cloud.GcpVpcCreateExt] {

	vpcReq := cloud.VpcCreateReq[cloud.GcpVpcCreateExt]{
		AccountID: req.AccountID,
		CloudID:   data.CloudID,
		BkBizID:   constant.UnassignedBiz,
		Name:      &data.Name,
		Region:    data.Region,
		Category:  req.Category,
		Memo:      req.Memo,
		Extension: &cloud.GcpVpcCreateExt{
			SelfLink:              data.Extension.SelfLink,
			AutoCreateSubnetworks: data.Extension.AutoCreateSubnetworks,
			EnableUlaInternalIpv6: data.Extension.EnableUlaInternalIpv6,
			InternalIpv6Range:     data.Extension.InternalIpv6Range,
			Mtu:                   data.Extension.Mtu,
			RoutingMode:           data.Extension.RoutingMode,
		},
	}

	return vpcReq
}

// GcpVpcUpdate update gcp vpc.
func (v vpc) GcpVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Gcp.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.GcpVpcUpdateOption{
		ResourceID: getRes.CloudID,
		Data:       &types.BaseVpcUpdateData{Memo: req.Memo},
	}
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Gcp.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GcpVpcDelete delete gcp vpc.
func (v vpc) GcpVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Gcp.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseDeleteOption{
		ResourceID: getRes.CloudID,
	}
	err = cli.DeleteVpc(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = v.cs.DataService().Global.Vpc.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
