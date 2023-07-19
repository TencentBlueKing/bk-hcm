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

	subnetlogics "hcm/cmd/hc-service/logics/subnet"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	adtysubnet "hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// AzureSubnetCreate create azure subnet.
func (s subnet) AzureSubnetCreate(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SubnetCreateReq[proto.AzureSubnetCreateExt])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := s.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	azureCreateOpt := subnetlogics.ConvAzureCreateReq(req)
	azureCreateRes, err := cli.CreateSubnet(cts.Kit, azureCreateOpt)
	if err != nil {
		return nil, err
	}

	// sync hcm subnets and related route tables
	subnetSyncOpt := &subnetlogics.AzureSubnetSyncOptions{
		BkBizID:       req.BkBizID,
		AccountID:     req.AccountID,
		CloudVpcID:    azureCreateRes.CloudVpcID,
		ResourceGroup: azureCreateRes.Extension.ResourceGroupName,
		Subnets:       []adtysubnet.AzureSubnet{*azureCreateRes},
	}
	res, err := s.subnet.AzureSubnetSync(cts.Kit, subnetSyncOpt)
	if err != nil {
		return nil, err
	}

	return core.CreateResult{ID: res.IDs[0]}, nil
}

// AzureSubnetUpdate update azure subnet.
func (s subnet) AzureSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(proto.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(adtysubnet.AzureSubnetUpdateOption)
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.AzureSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().Azure.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AzureSubnetDelete delete azure subnet.
func (s subnet) AzureSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adtysubnet.AzureSubnetDeleteOption{
		AzureDeleteOption: adcore.AzureDeleteOption{
			BaseDeleteOption:  adcore.BaseDeleteOption{ResourceID: getRes.Name},
			ResourceGroupName: getRes.Extension.ResourceGroupName,
		},
		VpcID: getRes.CloudVpcID,
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

// AzureListSubnetCountIP count azure subnets' available ips.
func (s subnet) AzureListSubnetCountIP(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ListAzureCountIPReq)
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
				&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(),
					Value: req.ResourceGroupName},
				&filter.AtomRule{Field: "vpc_id", Op: filter.Equal.Factory(), Value: req.VpcID},
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

	cli, err := s.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	usageOpt := &types.AzureVpcListUsageOption{
		ResourceGroupName: req.ResourceGroupName,
		VpcID:             listResult.Details[0].CloudVpcID,
	}
	usages, err := cli.ListVpcUsage(cts.Kit, usageOpt)
	if err != nil {
		return nil, err
	}

	cloudIDIPCountMap := make(map[string]proto.AvailIPResult)
	for _, usage := range usages {
		cloudIDIPCountMap[converter.PtrToVal(usage.ID)] = proto.AvailIPResult{
			AvailableIPCount: uint64(converter.PtrToVal(usage.Limit) - converter.PtrToVal(usage.CurrentValue)),
			TotalIPCount:     uint64(converter.PtrToVal(usage.Limit)),
			UsedIPCount:      uint64(converter.PtrToVal(usage.CurrentValue)),
		}
	}

	result := make(map[string]proto.AvailIPResult)
	for _, one := range listResult.Details {
		count, exist := cloudIDIPCountMap[one.CloudID]
		if !exist {
			return nil, fmt.Errorf("subnet: %s not found", one.CloudID)
		}

		result[one.ID] = count
	}

	return result, nil
}
