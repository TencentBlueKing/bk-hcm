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
	cloudproto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	proto "hcm/pkg/api/web-server/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListSubnetWithIPCountInBiz 查询子网信息且带有可用IP数量。
func (svc *service) ListSubnetWithIPCountInBiz(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(core.ListWithoutFieldReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	pageOpt := &core.PageOption{
		EnableUnlimitedLimit: false,
		MaxLimit:             core.AggregationQueryMaxPageLimit,
		DisabledSort:         false,
	}
	if err := req.Page.Validate(pageOpt); err != nil {
		return nil, err
	}

	listReq := &core.ListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	listResult, err := svc.client.CloudServer().Subnet.ListInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID, listReq)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &core.ListResult{Count: listResult.Count}, nil
	}

	if len(listResult.Details) == 0 {
		return &proto.ListSubnetWithAvailIPCountResp{Details: make([]proto.ListSubnetWithAvailIPCountResult, 0)}, nil
	}

	ids := make([]string, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		ids = append(ids, one.ID)
	}

	countReq := &cloudproto.ListSubnetCountIPReq{
		IDs: ids,
	}
	idIPCountMap, err := svc.client.CloudServer().Subnet.ListCountIPInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID, countReq)
	if err != nil {
		logs.Errorf("list subnet count avail ip failed, err: %v, ids: %v, rid: %s", err, ids, cts.Kit.Rid)
		return nil, err
	}

	details := make([]proto.ListSubnetWithAvailIPCountResult, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		subnet := proto.ListSubnetWithAvailIPCountResult{
			BaseSubnet: one,
		}

		tmp, exist := idIPCountMap[one.ID]
		if exist {
			subnet.AvailableIPCount = tmp.AvailableIPCount
			subnet.TotalIPCount = tmp.TotalIPCount
			subnet.UsedIPCount = tmp.UsedIPCount
		}

		details = append(details, subnet)
	}

	return &proto.ListSubnetWithAvailIPCountResp{Details: details}, nil
}

// ListSubnetWithIPCountInRes 查询子网信息且带有可用IP数量。
func (svc *service) ListSubnetWithIPCountInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListWithoutFieldReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	pageOpt := &core.PageOption{
		EnableUnlimitedLimit: false,
		MaxLimit:             core.AggregationQueryMaxPageLimit,
		DisabledSort:         false,
	}
	if err := req.Page.Validate(pageOpt); err != nil {
		return nil, err
	}

	listReq := &core.ListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	listResult, err := svc.client.CloudServer().Subnet.ListInRes(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &core.ListResult{Count: listResult.Count}, nil
	}

	if len(listResult.Details) == 0 {
		return &proto.ListSubnetWithAvailIPCountResp{Details: make([]proto.ListSubnetWithAvailIPCountResult, 0)}, nil
	}

	ids := make([]string, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		ids = append(ids, one.ID)
	}

	countReq := &cloudproto.ListSubnetCountIPReq{
		IDs: ids,
	}
	idIPCountMap, err := svc.client.CloudServer().Subnet.ListCountIPInRes(cts.Kit.Ctx, cts.Kit.Header(), countReq)
	if err != nil {
		logs.Errorf("list subnet count avail ip failed, err: %v, ids: %v, rid: %s", err, ids, cts.Kit.Rid)
		return nil, err
	}

	details := make([]proto.ListSubnetWithAvailIPCountResult, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		subnet := proto.ListSubnetWithAvailIPCountResult{
			BaseSubnet: one,
		}

		tmp, exist := idIPCountMap[one.ID]
		if exist {
			subnet.AvailableIPCount = tmp.AvailableIPCount
			subnet.TotalIPCount = tmp.TotalIPCount
			subnet.UsedIPCount = tmp.UsedIPCount
		}

		details = append(details, subnet)
	}

	return &proto.ListSubnetWithAvailIPCountResp{Details: details}, nil
}
