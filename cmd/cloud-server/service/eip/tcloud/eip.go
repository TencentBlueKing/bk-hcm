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

package tcloud

import (
	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/adaptor/types/eip"
	cloudproto "hcm/pkg/api/cloud-server/eip"
	hcproto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// TCloud eip service.
type TCloud struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// NewTCloud init tcloud eip service.
func NewTCloud(client *client.ClientSet, authorizer auth.Authorizer, audit audit.Interface) *TCloud {
	return &TCloud{
		client:     client,
		authorizer: authorizer,
		audit:      audit,
	}
}

// AssociateEip associate eip.
func (t *TCloud) AssociateEip(cts *rest.Contexts, accountID string, req *cloudproto.AssociateReq) (
	interface{}, error) {

	return nil, t.client.HCService().TCloud.Eip.AssociateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.TCloudEipAssociateReq{
			AccountID:          accountID,
			CvmID:              req.CvmID,
			EipID:              req.EipID,
			NetworkInterfaceID: req.NetworkInterfaceID,
		},
	)
}

// DisassociateEip disassociate eip.
func (t *TCloud) DisassociateEip(cts *rest.Contexts, accountID, eipID, cvmID string) (interface{}, error) {

	err := t.client.HCService().TCloud.Eip.DisassociateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.TCloudEipDisassociateReq{
			AccountID: accountID,
			CvmID:     cvmID,
			EipID:     eipID,
		},
	)

	return nil, err
}

// CreateEip ...
func (t *TCloud) CreateEip(cts *rest.Contexts, bkBizID int64) (interface{}, error) {
	req := new(cloudproto.TCloudEipCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resp, err := t.client.HCService().TCloud.Eip.CreateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.TCloudEipCreateReq{
			AccountID: req.AccountID,
			BkBizID:   bkBizID,
			TCloudEipCreateOption: &eip.TCloudEipCreateOption{
				Region:          req.Region,
				EipName:         req.EipName,
				EipCount:        req.EipCount,
				ServiceProvider: req.ServiceProvider,
				AddressType:     req.AddressType,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RetrieveEip ...
func (t *TCloud) RetrieveEip(cts *rest.Contexts, eipID string, cvmID string) (*cloudproto.TCloudEipExtResult, error) {
	eipResp, err := t.client.DataService().TCloud.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
	if err != nil {
		return nil, err
	}

	eipResult := &cloudproto.TCloudEipExtResult{EipExtResult: eipResp, CvmID: cvmID}
	// 表示没有关联
	if cvmID == "" {
		return eipResult, nil
	}

	eipResult.InstanceType = string(enumor.EipBindCvm)
	eipResult.InstanceID = converter.ValToPtr(cvmID)

	return eipResult, nil
}
