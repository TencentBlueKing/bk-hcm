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

package huawei

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/adaptor/types/eip"
	cloudproto "hcm/pkg/api/cloud-server/eip"
	"hcm/pkg/api/core"
	hcproto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// HuaWei eip service.
type HuaWei struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// NewHuaWei init huawei eip service.
func NewHuaWei(client *client.ClientSet, authorizer auth.Authorizer, audit audit.Interface) *HuaWei {
	return &HuaWei{
		client:     client,
		authorizer: authorizer,
		audit:      audit,
	}
}

// AssociateEip associate eip.
func (h *HuaWei) AssociateEip(cts *rest.Contexts, accountID string, req *cloudproto.AssociateReq) (interface{}, error) {

	rels, err := h.client.DataService().Global.NetworkInterfaceCvmRel.ListNetworkCvmRels(
		cts.Kit,
		&core.ListReq{
			Filter: tools.EqualExpression("network_interface_id", req.NetworkInterfaceID),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(rels.Details) == 0 {
		return nil, fmt.Errorf("network interface %s not found", req.NetworkInterfaceID)
	}

	return nil, h.client.HCService().HuaWei.Eip.AssociateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.HuaWeiEipAssociateReq{
			AccountID:          accountID,
			CvmID:              rels.Details[0].CvmID,
			EipID:              req.EipID,
			NetworkInterfaceID: req.NetworkInterfaceID,
		},
	)
}

// DisassociateEip disassociate eip.
func (h *HuaWei) DisassociateEip(cts *rest.Contexts, accountID, eipID, cvmID string) (interface{}, error) {

	err := h.client.HCService().HuaWei.Eip.DisassociateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.HuaWeiEipDisassociateReq{
			AccountID: accountID,
			CvmID:     cvmID,
			EipID:     eipID,
		},
	)

	return nil, err
}

// CreateEip ...
func (h *HuaWei) CreateEip(cts *rest.Contexts, bkBizID int64) (interface{}, error) {
	req := new(cloudproto.HuaWeiEipCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resp, err := h.client.HCService().HuaWei.Eip.CreateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.HuaWeiEipCreateReq{
			AccountID: req.AccountID,
			BkBizID:   bkBizID,
			HuaWeiEipCreateOption: &eip.HuaWeiEipCreateOption{
				Region:                req.Region,
				EipName:               req.EipName,
				EipType:               req.EipType,
				EipCount:              req.EipCount,
				InternetChargeType:    req.InternetChargeType,
				InternetChargePrepaid: req.InternetChargePrepaid,
				BandwidthOption:       req.BandwidthOption,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RetrieveEip ...
func (h *HuaWei) RetrieveEip(cts *rest.Contexts, eipID string, cvmID string) (*cloudproto.HuaWeiEipExtResult, error) {
	eipResp, err := h.client.DataService().HuaWei.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
	if err != nil {
		return nil, err
	}

	// 表示没有关联
	if cvmID == "" {
		return &cloudproto.HuaWeiEipExtResult{EipExtResult: eipResp, CvmID: cvmID}, nil
	}

	rels, err := h.client.DataService().Global.NetworkInterfaceCvmRel.ListNetworkCvmRels(
		cts.Kit,
		&core.ListReq{Filter: tools.ContainersExpression("cvm_id", []string{cvmID}), Page: core.NewDefaultBasePage()},
	)
	if err != nil {
		return nil, err
	}

	if rels == nil || rels.Details == nil || len(rels.Details) == 0 {
		return nil, fmt.Errorf("cvm(%s) has no networkinterface", cvmID)
	}

	niIDs := make([]string, len(rels.Details))
	for idx, rel := range rels.Details {
		niIDs[idx] = rel.NetworkInterfaceID
	}

	nis, err := h.client.DataService().Global.NetworkInterface.List(
		cts.Kit,
		&core.ListReq{Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{
					Field: "id",
					Op:    filter.In.Factory(),
					Value: niIDs,
				}, &filter.Expression{Op: filter.Or, Rules: []filter.RuleFactory{
					filter.AtomRule{
						Field: "public_ipv4",
						Op:    filter.JSONContains.Factory(),
						Value: eipResp.PublicIp,
					},
					filter.AtomRule{
						Field: "public_ipv6",
						Op:    filter.JSONContains.Factory(),
						Value: eipResp.PublicIp,
					},
				}},
			},
		}, Page: core.NewDefaultBasePage()})
	if err != nil {
		return nil, err
	}

	if nis == nil || nis.Details == nil {
		return nil, fmt.Errorf("eip(%s) not associated with cvm(%s)", eipResp.PublicIp, cvmID)
	}

	eipResult := &cloudproto.HuaWeiEipExtResult{EipExtResult: eipResp}
	if len(nis.Details) > 0 {
		eipResult.CvmID = cvmID
		eipResult.InstanceType = string(enumor.EipBindCvm)
		eipResult.InstanceID = converter.ValToPtr(nis.Details[0].ID)
	}

	return eipResult, nil
}
