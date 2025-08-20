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
	logicsrt "hcm/cmd/hc-service/logics/route-table"
	"hcm/cmd/hc-service/logics/subnet"
	"hcm/cmd/hc-service/service/sync/handler"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	apicloud "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	subnetproto "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/api/hc-service/sync"
	hcservice "hcm/pkg/api/hc-service/vpc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// HuaWeiVpcCreate create huawei vpc.
func (v vpc) HuaWeiVpcCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.VpcCreateReq[hcservice.HuaWeiVpcCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := v.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// create huawei vpc
	opt := &types.HuaWeiVpcCreateOption{
		AccountID: req.AccountID,
		Name:      req.Name,
		Memo:      req.Memo,
		Extension: &types.HuaWeiVpcCreateExt{
			Region:              req.Extension.Region,
			IPv4Cidr:            req.Extension.IPv4Cidr,
			EnterpriseProjectID: req.Extension.EnterpriseProjectID,
		},
	}
	err = cli.CreateVpc(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// get created vpc info
	listOpt := &types.HuaWeiVpcListOption{
		HuaWeiListOption: adcore.HuaWeiListOption{
			Region: req.Extension.Region,
		},
		Names: []string{req.Name},
	}
	listRes, err := cli.ListVpc(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	if len(listRes.Details) != 1 {
		return nil, errf.Newf(errf.Aborted, "get created vpc detail, but result count is invalid")
	}

	vpcCreatedID, err := v.createHuaWeiVpcForDB(cts.Kit, req, listRes.Details[0])
	if err != nil {
		logs.Errorf("create huawei vpc for db failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// create huawei subnets
	if len(req.Extension.Subnets) == 0 {
		return core.CreateResult{ID: vpcCreatedID}, nil
	}

	subnetCreateOpt := &subnet.SubnetCreateOptions[subnetproto.HuaWeiSubnetCreateExt]{
		BkBizID:    constant.UnassignedBiz,
		AccountID:  req.AccountID,
		Region:     listRes.Details[0].Region,
		CloudVpcID: listRes.Details[0].CloudID,
		CreateReqs: req.Extension.Subnets,
	}
	_, err = v.subnet.HuaWeiSubnetCreate(cts.Kit, subnetCreateOpt)
	if err != nil {
		return nil, err
	}

	err = v.syncHuaweiRouteTable(cts, req.AccountID, req.Extension.Region)
	if err != nil {
		return nil, err
	}

	return core.CreateResult{ID: vpcCreatedID}, nil
}

func (v vpc) syncHuaweiRouteTable(cts *rest.Contexts, accountID, region string) error {

	syncCli, err := v.syncCli.HuaWei(cts.Kit, accountID)
	if err != nil {
		return err
	}

	err = handler.ResourceSync(cts, &logicsrt.HuaWeiRouteTableHandler{
		DisablePrepare: true,
		Cli:            v.syncCli,
		Request: &sync.HuaWeiSyncReq{
			AccountID: accountID,
			Region:    region,
		},
		SyncCli: syncCli,
	})
	if err != nil {
		logs.Errorf("route table sync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}
	return nil
}

func (v vpc) createHuaWeiVpcForDB(kt *kit.Kit, req *hcservice.VpcCreateReq[hcservice.HuaWeiVpcCreateExt],
	huaweiVpc types.HuaWeiVpc) (string, error) {

	// create hcm vpc
	createReq := &cloud.VpcBatchCreateReq[cloud.HuaWeiVpcCreateExt]{
		Vpcs: []cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt]{convertHuaWeiVpcCreateReq(req, &huaweiVpc)},
	}
	result, err := v.cs.DataService().HuaWei.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		return "", err
	}

	if len(result.IDs) != 1 {
		return "", errf.New(errf.Aborted, "create result is invalid")
	}
	return result.IDs[0], nil
}

func convertHuaWeiVpcCreateReq(req *hcservice.VpcCreateReq[hcservice.HuaWeiVpcCreateExt],
	data *types.HuaWeiVpc) cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt] {

	vpcReq := cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt]{
		AccountID: req.AccountID,
		CloudID:   data.CloudID,
		BkBizID:   constant.UnassignedBiz,
		Name:      &data.Name,
		Region:    data.Region,
		Category:  req.Category,
		Memo:      req.Memo,
		Extension: &cloud.HuaWeiVpcCreateExt{
			Cidr:                make([]cloud.HuaWeiCidr, 0, len(data.Extension.Cidr)),
			Status:              data.Extension.Status,
			EnterpriseProjectID: data.Extension.EnterpriseProjectId,
		},
	}

	for _, cidr := range data.Extension.Cidr {
		vpcReq.Extension.Cidr = append(vpcReq.Extension.Cidr, cloud.HuaWeiCidr{
			Type: cidr.Type,
			Cidr: cidr.Cidr,
		})
	}

	return vpcReq
}

// HuaWeiVpcUpdate update huawei vpc.
func (v vpc) HuaWeiVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().HuaWei.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.HuaWeiVpcUpdateOption{
		VpcUpdateOption: types.VpcUpdateOption{
			ResourceID: getRes.CloudID,
			Data:       &types.BaseVpcUpdateData{Memo: req.Memo},
		},
		Region: getRes.Region,
	}
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.HuaWeiVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().HuaWei.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// HuaWeiVpcDelete delete huawei vpc.
func (v vpc) HuaWeiVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().HuaWei.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseRegionalDeleteOption{
		BaseDeleteOption: adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
		Region:           getRes.Region,
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

// HuaWeiVpcCount count huawei vpc.
func (v vpc) HuaWeiVpcCount(cts *rest.Contexts) (interface{}, error) {
	req := new(apicloud.HuaWeiSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := v.ad.Adaptor().HuaWei(&types.BaseSecret{
		CloudSecretID:  req.CloudSecretID,
		CloudSecretKey: req.CloudSecretKey,
	})
	if err != nil {
		return nil, err
	}

	return client.CountAllResources(cts.Kit, enumor.HuaWeiVpcProviderType)
}
