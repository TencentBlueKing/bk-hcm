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

package subnetcvmrel

import (
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// ListWithSubnet list subnet cvm relations with subnet details.
func (svc *subnetCvmRelSvc) ListWithSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SubnetCvmRelWithSubnetListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	details, err := svc.dao.SubnetCvmRel().ListJoinSubnet(cts.Kit, req.CvmIDs)
	if err != nil {
		logs.Errorf("list subnet cvm rels join subnet failed, err: %v, cvmIDs: %v, rid: %s", err,
			req.CvmIDs, cts.Kit.Rid)
		return nil, err
	}

	subnets := make([]corecloud.SubnetCvmRelWithBaseSubnet, 0, len(details.Details))
	for _, one := range details.Details {
		subnets = append(subnets, corecloud.SubnetCvmRelWithBaseSubnet{
			BaseSubnet: corecloud.BaseSubnet{
				ID:                one.ID,
				Vendor:            one.Vendor,
				AccountID:         one.AccountID,
				CloudVpcID:        one.CloudVpcID,
				CloudRouteTableID: converter.PtrToVal(one.CloudRouteTableID),
				CloudID:           one.CloudID,
				Name:              converter.PtrToVal(one.Name),
				Region:            one.Region,
				Zone:              one.Zone,
				Ipv4Cidr:          one.Ipv4Cidr,
				Ipv6Cidr:          one.Ipv6Cidr,
				Memo:              one.Memo,
				VpcID:             one.VpcID,
				RouteTableID:      converter.PtrToVal(one.RouteTableID),
				BkBizID:           one.BkBizID,
				Revision: &core.Revision{
					Creator:   one.Creator,
					Reviser:   one.Reviser,
					CreatedAt: one.CreatedAt,
					UpdatedAt: one.UpdatedAt,
				},
			},
			CvmID:        one.CvmID,
			RelCreator:   one.RelCreator,
			RelCreatedAt: one.RelCreatedAt,
		})
	}

	return subnets, nil
}
