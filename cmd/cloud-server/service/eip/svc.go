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

	"hcm/cmd/cloud-server/logics/async"
	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/eip"
	"hcm/cmd/cloud-server/service/common"
	"hcm/cmd/cloud-server/service/eip/aws"
	"hcm/cmd/cloud-server/service/eip/azure"
	"hcm/cmd/cloud-server/service/eip/gcp"
	"hcm/cmd/cloud-server/service/eip/huawei"
	"hcm/cmd/cloud-server/service/eip/tcloud"
	actioneip "hcm/cmd/task-server/logics/action/eip"
	cloudproto "hcm/pkg/api/cloud-server/eip"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
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
	eip        eip.Interface
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
		cts.Kit,
		&core.ListReq{
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
		cts.Kit,
		&core.ListReq{
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
	return svc.retrieveEip(cts, handler.ResOperateAuth)
}

// RetrieveBizEip get biz eip.
func (svc *eipSvc) RetrieveBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.retrieveEip(cts, handler.BizOperateAuth)
}

func (svc *eipSvc) retrieveEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	eipID := cts.PathParameter("id").String()

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.EipCloudResType, eipID)
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
		cts.Kit,
		&core.ListReq{
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

	return nil, eip.Assign(cts.Kit, svc.client.DataService(), req.IDs, req.BkBizID, false)
}

// BatchDeleteEip batch delete eip.
func (svc *eipSvc) BatchDeleteEip(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteEip(cts, handler.ResOperateAuth)
}

// BatchDeleteBizEip batch delete biz eip.
func (svc *eipSvc) BatchDeleteBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteEip(cts, handler.BizOperateAuth)
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
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
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

	tasks := make([]ts.CustomFlowTask, 0, len(req.IDs))
	var nextID = counter.NewNumStringCounter(1, 10)
	for _, eipID := range req.IDs {
		info, exists := basicInfoMap[eipID]
		if !exists {
			return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("id %s has no corresponding vendor", eipID))
		}
		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(nextID()),
			ActionName: enumor.ActionDeleteEIP,
			Params: actioneip.DeleteEIPOption{
				Vendor: info.Vendor,
				ID:     info.ID,
			},
			DependOn: nil,
		})
	}
	flowReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowDeleteEIP,
		Tasks: tasks,
	}

	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, flowReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateEip create eip.
func (svc *eipSvc) CreateEip(cts *rest.Contexts) (interface{}, error) {
	bizID := int64(constant.UnassignedBiz)
	return svc.createEip(cts, bizID, handler.ResOperateAuth)
}

// CreateBizEip create biz eip.
func (svc *eipSvc) CreateBizEip(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	return svc.createEip(cts, bkBizID, handler.BizOperateAuth)
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
	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, accountID)
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
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(kt, basicInfoReq)
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
