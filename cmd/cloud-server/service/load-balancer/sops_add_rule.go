/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package loadbalancer

import (
	"encoding/json"
	"fmt"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cloudserver "hcm/pkg/api/cloud-server"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// BatchBizRuleOnline batch biz rule online
func (svc *lbSvc) BatchBizRuleOnline(cts *rest.Contexts) (any, error) {
	return svc.batchBizRuleOnline(cts, handler.BizOperateAuth)
}

// BatchRuleOnline batch rule online
func (svc *lbSvc) BatchRuleOnline(cts *rest.Contexts) (any, error) {
	return svc.batchBizRuleOnline(cts, handler.ResOperateAuth)
}

func (svc *lbSvc) batchBizRuleOnline(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch sops rule online request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// authorized instances
	basicInfo := &types.CloudResourceBasicInfo{
		AccountID: req.AccountID,
	}
	err := authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TargetGroup,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		logs.Errorf("batch sops rule online auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get sops account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("get sops account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.buildCreateTcloudRule(cts, req.Data, accountInfo.AccountID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

func (svc *lbSvc) buildCreateTcloudRule(cts *rest.Contexts, body json.RawMessage, accountID string) (any, error) {
	req := new(cslb.TCloudSopsRuleBatchCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// TODO 拼凑调整参数，请求创建规则的批量异步任务接口
	var bindRSReqParam []*cloud.BatchOperationReq[*lblogic.BindRSRecord]
	// TODO 参数检查preview
	result, err := svc.previewValidate(cts, req.BindRSRecords)
	if err != nil {
		return nil, fmt.Errorf("batch sops rule online, preview validate err, err: %s", err)
	}
	switch value := result.(type) {
	case []*cloud.BatchOperationValidateError:
		return nil, fmt.Errorf("batch sops rule online, preview validate err, err: %s", value)
	case []*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord]:
		bindRSReqParam = convertBatchOperationPreviewResultToReq(value)
	default:
		return nil, fmt.Errorf("batch sops rule online, preview validate result type is invalid")
	}

	// TODO 提交异步任务
	logs.Infof("batch sops rule online, request bind rs api, request param: %v", bindRSReqParam)
	//return svc.BindRS(bindRSReqParam)
	return nil, nil
}

func (svc *lbSvc) previewValidate(cts *rest.Contexts, bindRSRecords []lblogic.BindRSRecord) (any, error) {
	// 错误校验列表
	errList := make([]*cloud.BatchOperationValidateError, 0)

	recordMap := make(map[string]struct{})
	resultMap := make(map[string]*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord])
	for _, record := range bindRSRecords {
		key := record.GetKey()
		err := record.Validate()
		if err != nil {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s %v", key, err),
			})
		}
		if _, ok := recordMap[key]; ok {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s duplicate record", key),
			})
		}
		recordMap[key] = struct{}{}

		validateErrs := record.CheckWithDataService(cts, svc.client.DataService(), svc.cvmLgc)
		if len(validateErrs) > 0 {
			errList = append(errList, validateErrs...)
			continue
		}

		lb, _ := record.GetLoadBalancer(cts, svc.client.DataService())
		if lb == nil {
			continue
		}
		previewResp, ok := resultMap[lb.ID]
		if !ok {
			previewResp = &cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord]{
				ClbID:     lb.ID,
				ClbName:   lb.Name,
				Vip:       record.VIP,
				Listeners: make([]*lblogic.BindRSRecord, 0),
			}
			resultMap[lb.ID] = previewResp
		}
		previewResp.Listeners = append(previewResp.Listeners, &record)
		previewResp.NewRsCount += len(record.RSInfos)
	}

	for lbID, previewResp := range resultMap {
		_, err := svc.checkResFlowRel(cts.Kit, lbID, enumor.LoadBalancerCloudResType)
		if err != nil {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s 已有任务执行中，不支持变更", previewResp.Vip),
				Ext:    lbID,
			})
		}
	}

	if len(errList) > 0 {
		return errList, nil
	}

	result := make([]*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord], 0, len(resultMap))
	for _, pr := range resultMap {
		result = append(result, pr)
	}
	return result, nil
}

func convertBatchOperationPreviewResultToReq(previewResultList []*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord]) []*cloud.BatchOperationReq[*lblogic.BindRSRecord] {
	batchOperationReqList := make([]*cloud.BatchOperationReq[*lblogic.BindRSRecord], 0, len(previewResultList))
	for _, previewResult := range previewResultList {
		req := &cloud.BatchOperationReq[*lblogic.BindRSRecord]{
			ClbID:             previewResult.ClbID,
			ClbName:           previewResult.ClbName,
			Vip:               previewResult.Vip,
			NewRsCount:        previewResult.NewRsCount,
			UpdateWeightCount: previewResult.UpdateWeightCount,
			Listeners:         previewResult.Listeners,
		}
		batchOperationReqList = append(batchOperationReqList, req)
	}

	return batchOperationReqList
}
