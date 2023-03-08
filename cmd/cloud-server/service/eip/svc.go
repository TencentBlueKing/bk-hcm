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
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/eip/aws"
	"hcm/cmd/cloud-server/service/eip/azure"
	"hcm/cmd/cloud-server/service/eip/gcp"
	"hcm/cmd/cloud-server/service/eip/huawei"
	"hcm/cmd/cloud-server/service/eip/tcloud"
	cloudproto "hcm/pkg/api/cloud-server/eip"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
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
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Eip, Action: meta.Find, Filter: req.Filter})
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
	return svc.client.DataService().Global.ListEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.EipListReq{
			Filter: filterExp,
			Page:   req.Page,
		},
	)
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
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Eip,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
	case enumor.Aws:
		return svc.client.DataService().Aws.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
	case enumor.Azure:
		return svc.client.DataService().Azure.RetrieveEip(cts.Kit.Ctx, cts.Kit.Header(), eipID)
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

// DeleteEip delete eip.
func (svc *eipSvc) DeleteEip(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteEip(cts, handler.ResValidWithAuth)
}

// DeleteBizEip delete biz eip.
func (svc *eipSvc) DeleteBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteEip(cts, handler.BizValidWithAuth)
}

func (svc *eipSvc) deleteEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
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
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Eip,
		Action: meta.Delete, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// TODO 增加审计
	// TODO 判断 Eip 是否可删除

	deleteReq := &hcproto.EipDeleteReq{EipID: eipID, AccountID: basicInfo.AccountID}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.Aws:
		return nil, svc.client.HCService().Aws.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.HuaWei:
		return nil, svc.client.HCService().HuaWei.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.Gcp:
		return nil, svc.client.HCService().Gcp.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	case enumor.Azure:
		return nil, svc.client.HCService().Azure.Eip.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", basicInfo.Vendor))
	}
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
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.tcloud.AssociateEip(cts, validHandler)
	case enumor.Aws:
		return svc.aws.AssociateEip(cts, validHandler)
	case enumor.HuaWei:
		return svc.huawei.AssociateEip(cts, validHandler)
	case enumor.Gcp:
		return svc.gcp.AssociateEip(cts, validHandler)
	case enumor.Azure:
		return svc.azure.AssociateEip(cts, validHandler)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
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

func (svc *eipSvc) disassociateEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.tcloud.DisassociateEip(cts, validHandler)
	case enumor.Aws:
		return svc.aws.DisassociateEip(cts, validHandler)
	case enumor.HuaWei:
		return svc.huawei.DisassociateEip(cts, validHandler)
	case enumor.Gcp:
		return svc.gcp.DisassociateEip(cts, validHandler)
	case enumor.Azure:
		return svc.azure.DisassociateEip(cts, validHandler)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
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
