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

package sgcomrel

import (
	"strconv"

	corecloud "hcm/pkg/api/core/cloud"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListWithCVMSummary list cvm that related to security group.
// only summary information will be displayed to avoid the risk of exceeding authority.
func (svc *sgComRelSvc) ListWithCVMSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SGCommonRelListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}

	result, err := svc.dao.SGCommonRel().ListJoinCVM(cts.Kit, req.SGIDs, opt)
	if err != nil {
		logs.Errorf("list sg common rels join cvm failed, err: %v, sgIDs: %v, rid: %s", err, req.SGIDs,
			cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.SGCommonRelWithCVMListResp{Count: result.Count}, nil
	}

	details := make([]corecloud.SGCommonRelWithCVMSummary, len(result.Details))
	for i, one := range result.Details {
		details[i] = corecloud.SGCommonRelWithCVMSummary{
			SummaryCVM: corecvm.SummaryCVM{
				ID:                   one.ID,
				CloudID:              one.CloudID,
				Name:                 one.Name,
				Vendor:               one.Vendor,
				BkBizID:              one.BkBizID,
				Region:               one.Region,
				Zone:                 one.Zone,
				CloudVpcIDs:          one.CloudVpcIDs,
				CloudSubnetIDs:       one.CloudSubnetIDs,
				Status:               one.Status,
				PrivateIPv4Addresses: one.PrivateIPv4Addresses,
				PrivateIPv6Addresses: one.PrivateIPv6Addresses,
				PublicIPv4Addresses:  one.PublicIPv4Addresses,
				PublicIPv6Addresses:  one.PublicIPv6Addresses,
			},
			SecurityGroupId: one.SecurityGroupID,
			RelCreator:      one.RelCreator,
			RelCreatedAt:    one.RelCreatedAt.String(),
		}
	}

	return &protocloud.SGCommonRelWithCVMListResp{Details: details}, nil
}

// ListWithLBSummary list load balancer that related to security group.
// only summary information will be displayed to avoid the risk of exceeding authority.
func (svc *sgComRelSvc) ListWithLBSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SGCommonRelListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}

	result, err := svc.dao.SGCommonRel().ListJoinLoadBalancer(cts.Kit, req.SGIDs, opt)
	if err != nil {
		logs.Errorf("list sg common rels join load balancer failed, err: %v, sgIDs: %v, rid: %s", err,
			req.SGIDs, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &protocloud.SGCommonRelWithLBListResp{Count: result.Count}, nil
	}

	details := make([]corecloud.SGCommonRelWithLBSummary, len(result.Details))
	for i, one := range result.Details {
		details[i] = corecloud.SGCommonRelWithLBSummary{
			SummaryBalancer: corelb.SummaryBalancer{
				ID:                   one.ID,
				CloudID:              one.CloudID,
				Name:                 one.Name,
				BkBizID:              one.BkBizID,
				IPVersion:            enumor.IPAddressType(one.IPVersion),
				LoadBalancerType:     one.LBType,
				Region:               one.Region,
				Zones:                one.Zones,
				BackupZones:          one.BackupZones,
				VpcID:                one.VpcID,
				CloudVpcID:           one.CloudVpcID,
				Domain:               one.Domain,
				Status:               one.Status,
				Memo:                 one.Memo,
				PrivateIPv4Addresses: one.PrivateIPv4Addresses,
				PrivateIPv6Addresses: one.PrivateIPv6Addresses,
				PublicIPv4Addresses:  one.PublicIPv4Addresses,
				PublicIPv6Addresses:  one.PublicIPv6Addresses,
			},
			SecurityGroupId: one.SecurityGroupID,
			RelCreator:      one.RelCreator,
			RelCreatedAt:    one.RelCreatedAt.String(),
		}
	}

	return &protocloud.SGCommonRelWithLBListResp{Details: details}, nil
}

// CountSGRelatedResBizInfo 统计安全组关联的资源 所属业务情况, 例如 绑定的cvm有多少个在某个业务下
func (svc *sgComRelSvc) CountSGRelatedResBizInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SGCommonRelCountBizInfoReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch req.ResType {
	case enumor.CvmCloudResType:
		resp, err := svc.dao.SGCommonRel().CountCVMRelatedResGroupByBiz(cts.Kit, req.SgID)
		if err != nil {
			logs.Errorf("count cvm related res group by biz failed, err: %v, sgID: %s, rid: %s",
				err, req.SgID, cts.Kit.Rid)
			return nil, err
		}
		return convertCountResult(cts.Kit, resp)
	case enumor.LoadBalancerCloudResType:
		resp, err := svc.dao.SGCommonRel().CountLoadBalancerRelatedResGroupByBiz(cts.Kit, req.SgID)
		if err != nil {
			logs.Errorf("count load balancer related res group by biz failed, err: %v, sgID: %s, rid: %s",
				err, req.SgID, cts.Kit.Rid)
			return nil, err
		}
		return convertCountResult(cts.Kit, resp)
	default:
		return nil, errf.Newf(errf.InvalidParameter,
			"unsupported resource type: %s for count sg related res biz info", req.ResType)
	}

}

func convertCountResult(kt *kit.Kit, items []types.CountResult) (*protocloud.SGCommonRelCountBizInfoResp, error) {
	result := &protocloud.SGCommonRelCountBizInfoResp{}
	for _, item := range items {
		bkBizID, err := strconv.ParseInt(item.GroupField, 0, 64)
		if err != nil {
			logs.Errorf("parse int failed, err: %v, BkBizID: %s, rid: %s",
				err, item.GroupField, kt.Rid)
			return nil, err
		}
		result.Items = append(result.Items, protocloud.ListSGRelBusinessItem{
			BkBizID:  bkBizID,
			ResCount: int64(item.Count),
		})
	}
	return result, nil
}
