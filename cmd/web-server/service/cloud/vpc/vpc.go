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

package vpc

import (
	"hcm/pkg/api/core"
	proto "hcm/pkg/api/web-server/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListVpcWithSubnetCount 查询vpc列表和该vpc下的子网数量，以及当前可用区下的子网数量，用于申请主机。
func (svc *service) ListVpcWithSubnetCount(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.ListVpcWithSubnetCountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listVpcReq := &core.ListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	vpcResult, err := svc.client.CloudServer().Vpc.ListInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID, listVpcReq)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %svc", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &proto.ListVpcWithSubnetCountResult{Count: vpcResult.Count}, nil
	}

	details := make([]proto.VpcWithSubnetCount, 0, len(vpcResult.Details))
	for _, one := range vpcResult.Details {
		vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCount(cts.Kit, one.ID, req.Zone, bizID)
		if err != nil {
			return nil, err
		}

		details = append(details, proto.VpcWithSubnetCount{
			BaseVpc:                one,
			SubnetCount:            vpcSubnetCount,
			CurrentZoneSubnetCount: zoneSubnetCount,
		})
	}

	return &proto.ListVpcWithSubnetCountResult{Details: details}, nil
}

func (svc *service) getVpcSubnetCount(kt *kit.Kit, vpcID, zone string, bizID int64) (uint64, uint64, error) {
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vpc_id",
					Op:    filter.Equal.Factory(),
					Value: vpcID,
				},
			},
		},
		Page: core.CountPage,
	}
	vpcResult, err := svc.client.CloudServer().Subnet.ListInBiz(kt.Ctx, kt.Header(), bizID, req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %svc", err, kt.Rid)
		return 0, 0, err
	}

	req = &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vpc_id",
					Op:    filter.Equal.Factory(),
					Value: vpcID,
				},
				&filter.AtomRule{
					Field: "zone",
					Op:    filter.Equal.Factory(),
					Value: zone,
				},
			},
		},
		Page: core.CountPage,
	}
	zoneResult, err := svc.client.CloudServer().Subnet.ListInBiz(kt.Ctx, kt.Header(), bizID, req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %svc", err, kt.Rid)
		return 0, 0, err
	}
	return vpcResult.Count, zoneResult.Count, nil
}
