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
	corecloud "hcm/pkg/api/core/cloud"
	proto "hcm/pkg/api/web-server/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListVpcWithSubnetCountInBiz 查询vpc列表和该vpc下的子网数量，以及当前可用区下的子网数量，用于申请主机。
func (svc *service) ListVpcWithSubnetCountInBiz(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

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
	if req.Page.Count {
		vpcResult, err := svc.client.CloudServer().Vpc.ListInBiz(cts.Kit, bizID, listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return &core.ListResult{Count: vpcResult.Count}, nil
	}

	switch vendor {
	case enumor.TCloud:
		vpcResult, err := svc.client.CloudServer().Vpc.TCloudListExtInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID,
			listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.TCloudVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCount(cts.Kit, one.ID, req.Zone, bizID)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.TCloudVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.TCloudVpcExtension]{Details: details}, nil

	case enumor.Aws:
		vpcResult, err := svc.client.CloudServer().Vpc.AwsListExtInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID, listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.AwsVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCount(cts.Kit, one.ID, req.Zone, bizID)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.AwsVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.AwsVpcExtension]{Details: details}, nil
	case enumor.Gcp:
		vpcResult, err := svc.client.CloudServer().Vpc.GcpListExtInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID, listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.GcpVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCount(cts.Kit, one.ID, req.Zone, bizID)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.GcpVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.GcpVpcExtension]{Details: details}, nil
	case enumor.Azure:
		vpcResult, err := svc.client.CloudServer().Vpc.AzureListExtInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID,
			listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.AzureVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCount(cts.Kit, one.ID, req.Zone, bizID)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.AzureVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.AzureVpcExtension]{Details: details}, nil
	case enumor.HuaWei:
		vpcResult, err := svc.client.CloudServer().Vpc.HuaWeiListExtInBiz(cts.Kit.Ctx, cts.Kit.Header(), bizID,
			listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.HuaWeiVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCount(cts.Kit, one.ID, req.Zone, bizID)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.HuaWeiVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.HuaWeiVpcExtension]{Details: details}, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

func (svc *service) getVpcSubnetCountInRes(kt *kit.Kit, vpcID, zone string) (uint64, uint64, error) {
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
		Page: core.NewCountPage(),
	}
	vpcResult, err := svc.client.CloudServer().Subnet.ListInRes(kt, req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, kt.Rid)
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
		Page: core.NewCountPage(),
	}
	zoneResult, err := svc.client.CloudServer().Subnet.ListInRes(kt, req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, kt.Rid)
		return 0, 0, err
	}
	return vpcResult.Count, zoneResult.Count, nil
}

// ListVpcWithSubnetCountInRes 查询vpc列表和该vpc下的子网数量，以及当前可用区下的子网数量，用于申请主机。
func (svc *service) ListVpcWithSubnetCountInRes(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

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
	if req.Page.Count {
		vpcResult, err := svc.client.CloudServer().Vpc.ListInRes(cts.Kit, listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return &core.ListResult{Count: vpcResult.Count}, nil
	}

	switch vendor {
	case enumor.TCloud:
		vpcResult, err := svc.client.CloudServer().Vpc.TCloudListExtInRes(cts.Kit.Ctx, cts.Kit.Header(), listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %svc", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.TCloudVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCountInRes(cts.Kit, one.ID, req.Zone)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.TCloudVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.TCloudVpcExtension]{Details: details}, nil

	case enumor.Aws:
		vpcResult, err := svc.client.CloudServer().Vpc.AwsListExtInRes(cts.Kit.Ctx, cts.Kit.Header(), listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %svc", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.AwsVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCountInRes(cts.Kit, one.ID, req.Zone)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.AwsVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.AwsVpcExtension]{Details: details}, nil
	case enumor.Gcp:
		vpcResult, err := svc.client.CloudServer().Vpc.GcpListExtInRes(cts.Kit.Ctx, cts.Kit.Header(), listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %svc", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.GcpVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCountInRes(cts.Kit, one.ID, req.Zone)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.GcpVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.GcpVpcExtension]{Details: details}, nil
	case enumor.Azure:
		vpcResult, err := svc.client.CloudServer().Vpc.AzureListExtInRes(cts.Kit.Ctx, cts.Kit.Header(), listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %svc", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.AzureVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCountInRes(cts.Kit, one.ID, req.Zone)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.AzureVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.AzureVpcExtension]{Details: details}, nil
	case enumor.HuaWei:
		vpcResult, err := svc.client.CloudServer().Vpc.HuaWeiListExtInRes(cts.Kit.Ctx, cts.Kit.Header(), listVpcReq)
		if err != nil {
			logs.Errorf("list vpc failed, err: %v, rid: %svc", err, cts.Kit.Rid)
			return nil, err
		}
		details := make([]proto.VpcWithSubnetCount[corecloud.HuaWeiVpcExtension], 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			vpcSubnetCount, zoneSubnetCount, err := svc.getVpcSubnetCountInRes(cts.Kit, one.ID, req.Zone)
			if err != nil {
				return nil, err
			}

			details = append(details, proto.VpcWithSubnetCount[corecloud.HuaWeiVpcExtension]{
				Vpc:                    one,
				SubnetCount:            vpcSubnetCount,
				CurrentZoneSubnetCount: zoneSubnetCount,
			})
		}

		return &proto.ListVpcWithSubnetCountResult[corecloud.HuaWeiVpcExtension]{Details: details}, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
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
		Page: core.NewCountPage(),
	}
	vpcResult, err := svc.client.CloudServer().Subnet.ListInBiz(kt, bizID, req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %svc", err, kt.Rid)
		return 0, 0, err
	}
	// 没有指定zone的时候，当前zone下的子网为0
	if len(zone) == 0 {
		return vpcResult.Count, 0, nil
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
		Page: core.NewCountPage(),
	}
	zoneResult, err := svc.client.CloudServer().Subnet.ListInBiz(kt, bizID, req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %svc", err, kt.Rid)
		return 0, 0, err
	}
	return vpcResult.Count, zoneResult.Count, nil
}
