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
	"fmt"

	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// TCloudSubnetBatchCreate batch create tcloud subnet.
func (s subnet) TCloudSubnetBatchCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudSubnetBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// TODO: 子网创建待梳理、优化；未直接复用同步逻辑
	res, err := s.subnet.TCloudSubnetCreate(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TCloudSubnetUpdate update tencent cloud subnet.
func (s subnet) TCloudSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(proto.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().TCloud.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.TCloud(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(adtysubnet.TCloudSubnetUpdateOption)
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.TCloudSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().TCloud.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudSubnetDelete delete tencent cloud subnet.
func (s subnet) TCloudSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().TCloud.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.TCloud(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
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

// GcpSubnetCountIP count gcp subnet subnets' available ips.
func (s subnet) GcpSubnetCountIP(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ListCountIPReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Page: core.NewDefaultBasePage(),
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs},
			},
		},
	}
	listResult, err := s.cs.DataService().Global.Subnet.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}

	if len(listResult.Details) != len(req.IDs) {
		return nil, fmt.Errorf("list subnet return count not right, query id count: %d, but return %d",
			len(req.IDs), len(listResult.Details))
	}

	cloudIDs := make([]string, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	cli, err := s.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &adtysubnet.GcpSubnetListOption{
		Region: req.Region,
		GcpListOption: adcore.GcpListOption{
			CloudIDs: cloudIDs,
			Page: &adcore.GcpPage{
				PageSize: adcore.GcpQueryLimit,
			},
		},
	}
	subnetRes, err := cli.ListSubnetWithIPNumber(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	if len(subnetRes.Details) != len(cloudIDs) {
		return nil, fmt.Errorf("list gcp subnet return count not right, query id count: %d, but return %d",
			len(cloudIDs), len(subnetRes.Details))
	}

	cloudIDMap := make(map[string]string)
	for _, one := range listResult.Details {
		cloudIDMap[one.CloudID] = one.ID
	}

	result := make(map[string]proto.AvailIPResult)
	for _, one := range subnetRes.Details {
		id, exist := cloudIDMap[one.CloudID]
		if !exist {
			return nil, fmt.Errorf("subnet: %s not found", one.CloudID)
		}

		result[id] = proto.AvailIPResult{
			AvailableIPCount: one.Extension.AvailableIPAddressCount,
			TotalIPCount:     one.Extension.TotalIpAddressCount,
			UsedIPCount:      one.Extension.UsedIpAddressCount,
		}
	}

	return result, nil
}

// TCloudListSubnetCountIP count tcloud subnets' available ips.
func (s subnet) TCloudListSubnetCountIP(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ListCountIPReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Page: core.NewDefaultBasePage(),
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs},
			},
		},
	}
	listResult, err := s.cs.DataService().Global.Subnet.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}

	if len(listResult.Details) != len(req.IDs) {
		return nil, fmt.Errorf("list subnet return count not right, query id count: %d, but return %d",
			len(req.IDs), len(listResult.Details))
	}

	cloudIDs := make([]string, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	cli, err := s.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &adcore.TCloudListOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	subnetRes, err := cli.ListSubnet(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	if len(subnetRes.Details) != len(cloudIDs) {
		return nil, fmt.Errorf("list tcloud subnet return count not right, query id count: %d, but return %d",
			len(cloudIDs), len(subnetRes.Details))
	}

	cloudIDMap := make(map[string]string)
	for _, one := range listResult.Details {
		cloudIDMap[one.CloudID] = one.ID
	}

	result := make(map[string]proto.AvailIPResult)
	for _, one := range subnetRes.Details {
		id, exist := cloudIDMap[one.CloudID]
		if !exist {
			return nil, fmt.Errorf("subnet: %s not found", one.CloudID)
		}

		result[id] = proto.AvailIPResult{
			AvailableIPCount: one.Extension.AvailableIPAddressCount,
			TotalIPCount:     one.Extension.TotalIpAddressCount,
			UsedIPCount:      one.Extension.UsedIpAddressCount,
		}
	}

	return result, nil
}
