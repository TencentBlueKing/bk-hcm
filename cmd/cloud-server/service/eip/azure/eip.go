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

package azure

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/adaptor/types/eip"
	cloudproto "hcm/pkg/api/cloud-server/eip"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	datarelproto "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	hcproto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/hooks/handler"
)

// Azure eip service.
type Azure struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// NewAzure init azure eip service.
func NewAzure(client *client.ClientSet, authorizer auth.Authorizer, audit audit.Interface) *Azure {
	return &Azure{
		client:     client,
		authorizer: authorizer,
		audit:      audit,
	}
}

// AssociateEip associate eip.
func (a *Azure) AssociateEip(
	cts *rest.Contexts,
	basicInfo *types.CloudResourceBasicInfo,
	validHandler handler.ValidWithAuthHandler,
) (interface{}, error) {
	req := new(cloudproto.AzureEipAssociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// TODO 判断 Eip 是否可关联

	// validate biz and authorize
	err := validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: a.authorizer, ResType: meta.Eip,
		Action: meta.Associate, BasicInfo: basicInfo,
	})
	if err != nil {
		return nil, err
	}

	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.EipAuditResType,
		ResID:             req.EipID,
		Action:            protoaudit.Associate,
		AssociatedResType: enumor.NetworkInterfaceAuditResType,
		AssociatedResID:   req.NetworkInterfaceID,
	}
	err = a.audit.ResOperationAudit(cts.Kit, operationInfo)
	if err != nil {
		logs.Errorf("create associate eip audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	rels, err := a.client.DataService().Global.NetworkInterfaceCvmRel.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&core.ListReq{
			Filter: tools.ContainersExpression("network_interface_id", req.NetworkInterfaceID),
			Page:   core.DefaultBasePage,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(rels.Details) == 0 {
		return nil, fmt.Errorf("network interface %s not found", req.NetworkInterfaceID)
	}

	return nil, a.client.HCService().Azure.Eip.AssociateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AzureEipAssociateReq{
			AccountID:          basicInfo.AccountID,
			CvmID:              rels.Details[0].CvmID,
			EipID:              req.EipID,
			NetworkInterfaceID: req.NetworkInterfaceID,
		},
	)
}

// DisassociateEip disassociate eip.
func (a *Azure) DisassociateEip(
	cts *rest.Contexts,
	basicInfo *types.CloudResourceBasicInfo,
	validHandler handler.ValidWithAuthHandler,
) (interface{}, error) {
	req := new(cloudproto.AzureEipDisassociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// validate biz and authorize
	err := validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: a.authorizer, ResType: meta.Eip,
		Action: meta.Disassociate, BasicInfo: basicInfo,
	})
	if err != nil {
		return nil, err
	}

	rels, err := a.client.DataService().Global.ListEipCvmRel(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&datarelproto.EipCvmRelListReq{
			Filter: tools.ContainersExpression("eip_id", []string{req.EipID}),
			Page:   core.DefaultBasePage,
		},
	)
	if len(rels.Details) == 0 {
		return nil, fmt.Errorf("eip(%s) not associated", req.EipID)
	}

	cvmID := rels.Details[0].CvmID

	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.EipAuditResType,
		ResID:             req.EipID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: enumor.NetworkInterfaceAuditResType,
		AssociatedResID:   req.NetworkInterfaceID,
	}
	err = a.audit.ResOperationAudit(cts.Kit, operationInfo)
	if err != nil {
		logs.Errorf("create disassociate eip audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, a.client.HCService().Azure.Eip.DisassociateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AzureEipDisassociateReq{
			AccountID:          basicInfo.AccountID,
			CvmID:              cvmID,
			EipID:              req.EipID,
			NetworkInterfaceID: req.NetworkInterfaceID,
		},
	)
}

// CreateEip ...
func (a *Azure) CreateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudproto.AzureEipCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bkBizID, err := cts.PathParameter("bk_biz_id").Uint64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = a.checkAzureEipParams(req); err != nil {
		return nil, err
	}

	// validate biz and authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Eip, Action: meta.Create}, BizID: int64(bkBizID)}
	err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.HCService().Azure.Eip.CreateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AzureEipCreateReq{
			AccountID: req.AccountID,
			AzureEipCreateOption: &eip.AzureEipCreateOption{
				ResourceGroupName:    req.ResourceGroupName,
				EipName:              req.EipName,
				Region:               req.Region,
				Zone:                 req.Zone,
				SKUName:              req.SKUName,
				SKUTier:              req.SKUTier,
				AllocationMethod:     req.AllocationMethod,
				IPVersion:            req.IPVersion,
				IdleTimeoutInMinutes: req.IdleTimeoutInMinutes,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// 分配业务
	_, err = a.client.DataService().Global.BatchUpdateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.EipBatchUpdateReq{IDs: resp.IDs, BkBizID: bkBizID},
	)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RetrieveEip ...
func (a *Azure) RetrieveEip(cts *rest.Contexts, eipID string, cvmID string) (*cloudproto.AzureEipExtResult, error) {
	eipResp, err := a.client.DataService().Azure.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
	if err != nil {
		return nil, err
	}

	// 表示没有关联
	if cvmID == "" {
		return &cloudproto.AzureEipExtResult{EipExtResult: eipResp, CvmID: cvmID}, nil
	}

	rels, err := a.client.DataService().Global.NetworkInterfaceCvmRel.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&core.ListReq{Filter: tools.ContainersExpression("cvm_id", []string{cvmID})},
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

	nis, err := a.client.DataService().Global.NetworkInterface.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
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
		}, Page: core.DefaultBasePage})
	if err != nil {
		return nil, err
	}

	if nis == nil || nis.Details == nil || len(nis.Details) == 0 {
		return nil, fmt.Errorf("eip(%s) not associated with cvm(%s)", eipResp.PublicIp, cvmID)
	}

	eipResult := &cloudproto.AzureEipExtResult{EipExtResult: eipResp, CvmID: cvmID}
	eipResult.InstanceType = "NI"
	eipResult.InstanceId = nis.Details[0].ID

	return eipResult, nil
}

// checkAzureEipParams check azure eip params
func (a *Azure) checkAzureEipParams(req *cloudproto.AzureEipCreateReq) error {
	if !assert.IsSameCaseString(req.ResourceGroupName) {
		return errf.New(errf.InvalidParameter, "resource_group_name can only be lowercase")
	}

	if !assert.IsSameCaseString(req.EipName) {
		return errf.New(errf.InvalidParameter, "eip_name can only be lowercase")
	}

	if !assert.IsSameCaseNoSpaceString(req.Region) {
		return errf.New(errf.InvalidParameter, "region can only be lowercase")
	}

	if !assert.IsSameCaseNoSpaceString(req.Zone) {
		return errf.New(errf.InvalidParameter, "zone can only be lowercase")
	}

	return nil
}
