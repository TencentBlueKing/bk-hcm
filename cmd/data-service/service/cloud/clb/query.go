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

package clb

import (
	"fmt"

	"hcm/pkg/api/core"
	coreclb "hcm/pkg/api/core/cloud/clb"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableclb "hcm/pkg/dal/table/cloud/clb"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// ListLoadBalancer list clb.
func (svc *clbSvc) ListLoadBalancer(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list clb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list clb failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ClbListResult{Count: result.Count}, nil
	}

	details := make([]coreclb.BaseClb, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseClb(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.ClbListResult{Details: details}, nil
}

func convTableToBaseClb(one *tableclb.LoadBalancerTable) *coreclb.BaseClb {
	return &coreclb.BaseClb{
		ID:                   one.ID,
		CloudID:              one.CloudID,
		Name:                 one.Name,
		Vendor:               one.Vendor,
		AccountID:            one.AccountID,
		BkBizID:              one.BkBizID,
		Region:               one.Region,
		Zones:                one.Zones,
		BackupZones:          one.BackupZones,
		VpcID:                one.VpcID,
		CloudVpcID:           one.CloudVpcID,
		SubnetID:             one.SubnetID,
		CloudSubnetID:        one.CloudSubnetID,
		PrivateIPv4Addresses: one.PrivateIPv4Addresses,
		PrivateIPv6Addresses: one.PrivateIPv6Addresses,
		PublicIPv4Addresses:  one.PublicIPv4Addresses,
		PublicIPv6Addresses:  one.PublicIPv6Addresses,
		Domain:               one.Domain,
		Status:               one.Status,
		CloudCreatedTime:     one.CloudCreatedTime,
		CloudStatusTime:      one.CloudStatusTime,
		CloudExpiredTime:     one.CloudExpiredTime,
		Memo:                 converter.ValToPtr(one.Memo),
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// ListLoadBalancerExt list clb ext.
func (svc *clbSvc) ListLoadBalancerExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.EipListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return convClbListResult[coreclb.TCloudClbExtension](data.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// GetLoadBalancer ...
func (svc *clbSvc) GetLoadBalancer(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "clb id is required")
	}

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.LoadBalancer().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list clb(%s) failed, err: %v, rid: %s", err, id, cts.Kit.Rid)
		return nil, fmt.Errorf("get clb failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "load balancer not found")
	}

	clbTable := result.Details[0]
	switch clbTable.Vendor {
	case enumor.TCloud:
		return convTableToBaseClb(&clbTable), nil
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", vendor)
	}
}

func convClbListResult[T coreclb.Extension](tables []tableclb.LoadBalancerTable) (
	*protocloud.ClbExtListResult[T], error) {

	details := make([]coreclb.Clb[T], 0, len(tables))
	for _, one := range tables {
		tmpClb := convTableToBaseClb(&one)
		extension := new(T)
		details = append(details, coreclb.Clb[T]{
			BaseClb:   *tmpClb,
			Extension: extension,
		})
	}

	return &protocloud.ClbExtListResult[T]{
		Details: details,
	}, nil
}
