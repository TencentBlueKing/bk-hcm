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

package eip

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/common"
	"hcm/cmd/cloud-server/service/eip/aws"
	"hcm/cmd/cloud-server/service/eip/azure"
	"hcm/cmd/cloud-server/service/eip/gcp"
	"hcm/cmd/cloud-server/service/eip/huawei"
	"hcm/cmd/cloud-server/service/eip/tcloud"
	cloudproto "hcm/pkg/api/cloud-server/eip"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	datarelproto "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	hcproto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

type eipSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
	tcloud     *tcloud.TCloud
	aws        *aws.Aws
	azure      *azure.Azure
	gcp        *gcp.Gcp
	huawei     *huawei.HuaWei
}

// ListEip list eip.
func (svc *eipSvc) ListEip(cts *rest.Contexts) (interface{}, error) {
	return svc.listEip(cts, handler.ListResourceAuthRes)
}

// ListBizEip list biz eip.
func (svc *eipSvc) ListBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.listEip(cts, handler.ListBizAuthRes)
}

func (svc *eipSvc) listEip(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(cloudproto.EipListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.Eip, Action: meta.Find, Filter: req.Filter,
	})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &dataproto.EipListResult{Details: make([]*dataproto.EipResult, 0)}, nil
	}

	filterExp := expr
	if filterExp == nil {
		filterExp = tools.AllExpression()
	}

	resp, err := svc.client.DataService().Global.ListEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.EipListReq{
			Filter: filterExp,
			Page:   req.Page,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Details) == 0 {
		return &cloudproto.EipListResult{Details: make([]*cloudproto.EipResult, 0), Count: resp.Count}, nil
	}

	eipIDs := make([]string, len(resp.Details))
	for idx, eipData := range resp.Details {
		eipIDs[idx] = eipData.ID
	}

	rels, err := svc.client.DataService().Global.ListEipCvmRel(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&datarelproto.EipCvmRelListReq{
			Filter: tools.ContainersExpression("eip_id", eipIDs),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		return nil, err
	}

	eipIDToCvmID := make(map[string]string)
	for _, relData := range rels.Details {
		eipIDToCvmID[relData.EipID] = relData.CvmID
	}

	details := make([]*cloudproto.EipResult, len(resp.Details))
	for idx, eipData := range resp.Details {
		eipData.InstanceID = converter.ValToPtr(eipIDToCvmID[eipData.ID])
		details[idx] = &cloudproto.EipResult{
			CvmID:     eipIDToCvmID[eipData.ID],
			EipResult: eipData,
		}
		if eipIDToCvmID[eipData.ID] != "" {
			eipData.InstanceType = string(enumor.EipBindCvm)
		}
	}

	return &cloudproto.EipListResult{Details: details, Count: resp.Count}, nil
}

// RetrieveEip get eip.
func (svc *eipSvc) RetrieveEip(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveEip(cts, handler.ResValidWithAuth)
}

// RetrieveBizEip get biz eip.
func (svc *eipSvc) RetrieveBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveEip(cts, handler.BizValidWithAuth)
}

func (svc *eipSvc) retrieveEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	eipID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.EipCloudResType,
		eipID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer, ResType: meta.Eip,
		Action: meta.Find, BasicInfo: basicInfo,
	})
	if err != nil {
		return nil, err
	}

	rels, err := svc.client.DataService().Global.ListEipCvmRel(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&datarelproto.EipCvmRelListReq{
			Filter: tools.ContainersExpression("eip_id", []string{eipID}),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		return nil, err
	}

	var cvmID string
	if len(rels.Details) > 0 {
		cvmID = rels.Details[0].CvmID
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.tcloud.RetrieveEip(cts, eipID, cvmID)
	case enumor.Aws:
		return svc.aws.RetrieveEip(cts, eipID, cvmID)
	case enumor.HuaWei:
		return svc.huawei.RetrieveEip(cts, eipID, cvmID)
	case enumor.Gcp:
		return svc.gcp.RetrieveEip(cts, eipID, cvmID)
	case enumor.Azure:
		return svc.azure.RetrieveEip(cts, eipID, cvmID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}

// AssignEip ...
func (svc *eipSvc) AssignEip(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudproto.EipAssignReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authorizeEipAssignOp(cts.Kit, req.IDs); err != nil {
		return nil, err
	}

	// check if all eips are not assigned to biz, right now assigning resource twice is not allowed
	eipFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs}
	err := svc.checkEipsInBiz(cts.Kit, eipFilter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create assign audit
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.EipAuditResType, req.IDs, int64(req.BkBizID))
	if err != nil {
		logs.Errorf("create assign eip audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.client.DataService().Global.BatchUpdateEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.EipBatchUpdateReq{IDs: req.IDs, BkBizID: req.BkBizID},
	)
}

// BatchDeleteEip batch delete eip.
func (svc *eipSvc) BatchDeleteEip(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteEip(cts, handler.ResValidWithAuth)
}

// BatchDeleteBizEip batch delete biz eip.
func (svc *eipSvc) BatchDeleteBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteEip(cts, handler.BizValidWithAuth)
}

func (svc *eipSvc) batchDeleteEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.EipCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Eip,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// create delete audit.
	err = svc.audit.ResDeleteAudit(cts.Kit, enumor.EipAuditResType, req.IDs)
	if err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// TODO 判断 Eip 是否可删除

	succeeded := make([]string, 0)
	for _, eipID := range req.IDs {
		basicInfo, exists := basicInfoMap[eipID]
		if !exists {
			return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("id %s has no corresponding vendor", eipID))
		}

		deleteReq := &hcproto.EipDeleteReq{EipID: eipID, AccountID: basicInfo.AccountID}

		switch basicInfo.Vendor {
		case enumor.TCloud:
			err = svc.client.HCService().TCloud.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
		case enumor.Aws:
			err = svc.client.HCService().Aws.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
		case enumor.HuaWei:
			err = svc.client.HCService().HuaWei.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
		case enumor.Gcp:
			err = svc.client.HCService().Gcp.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
		case enumor.Azure:
			err = svc.client.HCService().Azure.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
		default:
			err = errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
		}

		if err != nil {
			return core.BatchOperateResult{
				Succeeded: succeeded,
				Failed: &core.FailedInfo{
					ID:    eipID,
					Error: err,
				},
			}, errf.NewFromErr(errf.PartialFailed, err)
		}

		succeeded = append(succeeded, eipID)
	}

	return core.BatchOperateResult{Succeeded: succeeded}, nil
}

// AssociateEip associate eip.
func (svc *eipSvc) AssociateEip(cts *rest.Contexts) (interface{}, error) {
	return svc.associateEip(cts, handler.ResValidWithAuth)
}

// AssociateBizEip associate biz eip.
func (svc *eipSvc) AssociateBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.associateEip(cts, handler.BizValidWithAuth)
}

func (svc *eipSvc) associateEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	eipID, err := extractEipID(cts)
	if err != nil {
		return nil, err
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.EipCloudResType,
		eipID,
	)
	if err != nil {
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.tcloud.AssociateEip(cts, basicInfo, validHandler)
	case enumor.Aws:
		return svc.aws.AssociateEip(cts, basicInfo, validHandler)
	case enumor.HuaWei:
		return svc.huawei.AssociateEip(cts, basicInfo, validHandler)
	case enumor.Gcp:
		return svc.gcp.AssociateEip(cts, basicInfo, validHandler)
	case enumor.Azure:
		return svc.azure.AssociateEip(cts, basicInfo, validHandler)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}

// DisassociateEip disassociate eip.
func (svc *eipSvc) DisassociateEip(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateEip(cts, handler.ResValidWithAuth)
}

// DisassociateBizEip disassociate biz eip.
func (svc *eipSvc) DisassociateBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateEip(cts, handler.BizValidWithAuth)
}

func (svc *eipSvc) disassociateEip(
	cts *rest.Contexts,
	validHandler handler.ValidWithAuthHandler,
) (interface{}, error) {
	eipID, err := extractEipID(cts)
	if err != nil {
		return nil, err
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.EipCloudResType,
		eipID,
	)
	if err != nil {
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.tcloud.DisassociateEip(cts, basicInfo, validHandler)
	case enumor.Aws:
		return svc.aws.DisassociateEip(cts, basicInfo, validHandler)
	case enumor.HuaWei:
		return svc.huawei.DisassociateEip(cts, basicInfo, validHandler)
	case enumor.Gcp:
		return svc.gcp.DisassociateEip(cts, basicInfo, validHandler)
	case enumor.Azure:
		return svc.azure.DisassociateEip(cts, basicInfo, validHandler)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
}

// CreateEip create eip.
func (svc *eipSvc) CreateEip(cts *rest.Contexts) (interface{}, error) {
	bizID := int64(constant.UnassignedBiz)
	return svc.createEip(cts, bizID, handler.ResValidWithAuth)
}

// CreateBizEip create biz eip.
func (svc *eipSvc) CreateBizEip(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	return svc.createEip(cts, bkBizID, handler.BizValidWithAuth)
}

// CreateBizEip ...
func (svc *eipSvc) createEip(cts *rest.Contexts, bizID int64,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	accountID, err := common.ExtractAccountID(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// validate authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Eip,
		Action: meta.Disassociate, BasicInfo: common.GetCloudResourceBasicInfo(accountID, bizID)})
	if err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.AccountCloudResType,
		accountID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return svc.tcloud.CreateEip(cts, bizID)
	case enumor.Aws:
		return svc.aws.CreateEip(cts, bizID)
	case enumor.HuaWei:
		return svc.huawei.CreateEip(cts, bizID)
	case enumor.Gcp:
		return svc.gcp.CreateEip(cts, bizID)
	case enumor.Azure:
		return svc.azure.CreateEip(cts, bizID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}

func (svc *eipSvc) authorizeEipAssignOp(kt *kit.Kit, ids []string) error {
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.EipCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(kt.Ctx, kt.Header(), basicInfoReq)
	if err != nil {
		return err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{
			Type: meta.Eip, Action: meta.Assign,
			ResourceID: info.AccountID,
		}, BizID: info.BkBizID})
	}
	err = svc.authorizer.AuthorizeWithPerm(kt, authRes...)
	if err != nil {
		return err
	}

	return nil
}

// checkEipsInBiz check if eips are in the specified biz.
func (svc *eipSvc) checkEipsInBiz(kt *kit.Kit, rule filter.RuleFactory, bizID int64) error {
	req := &dataproto.EipListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: bizID}, rule,
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := svc.client.DataService().Global.ListEip(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count eips that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != nil && *result.Count != 0 {
		return fmt.Errorf("%d eips are already assigned", result.Count)
	}

	return nil
}

func extractEipID(cts *rest.Contexts) (string, error) {
	req := new(cloudproto.EipReq)
	reqData, err := ioutil.ReadAll(cts.Request.Request.Body)
	if err != nil {
		logs.Errorf("read request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", err
	}

	cts.Request.Request.Body = ioutil.NopCloser(bytes.NewReader(reqData))
	if err := cts.DecodeInto(req); err != nil {
		return "", err
	}

	if err := req.Validate(); err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	cts.Request.Request.Body = ioutil.NopCloser(bytes.NewReader(reqData))

	return req.EipID, nil
}
