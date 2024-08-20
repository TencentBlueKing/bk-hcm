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
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
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
	return svc.batchRuleOnline(cts, handler.BizOperateAuth)
}

// BatchResRuleOnline batch rule online
func (svc *lbSvc) BatchResRuleOnline(cts *rest.Contexts) (any, error) {
	return svc.batchRuleOnline(cts, handler.ResOperateAuth)
}

func (svc *lbSvc) batchRuleOnline(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (any, error) {
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
	// TODO 这里的内容依赖clb excel导入的代码，暂时注释，等待clb excel导入合并后再开放
	//case enumor.TCloud:
	//	return svc.buildCreateTcloudRule(cts, req.Data, accountInfo.AccountID, accountInfo.BkBizID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

// TODO 这里的内容依赖clb excel导入的代码，暂时注释，等待clb excel导入合并后再开放
//
//func (svc *lbSvc) buildCreateTCloudRule(cts *rest.Contexts, body json.RawMessage, accountID string, BkBizID int64) (any, error) {
//	req := new(cslb.TCloudSopsRuleBatchCreateReq)
//	if err := json.Unmarshal(body, req); err != nil {
//		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
//	}
//
//	bindRsRecord := svc.convBindRsRecord(req.BindRSRecords)
//	// 参数检查preview
//	result, err := svc.bindRSPreview(cts, bindRsRecord, make([]*cloud.BatchOperationValidateError, 0), BkBizID)
//	if err != nil {
//		return nil, fmt.Errorf("batch sops rule online, preview validate err, err: %s", err)
//	}
//
//	var bindRSReqParam *cloud.BatchOperationReq[*lblogic.BindRSRecord]
//	switch value := result.(type) {
//	case []*cloud.BatchOperationValidateError:
//		return nil, fmt.Errorf("batch sops rule online, preview validate err, err: %s", value)
//	case []*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord]:
//		// 转换preview返回结果为请求参数
//		bindRSReqParam = convBatchOperationPreviewResultToReq(value, accountID)
//	default:
//		return nil, fmt.Errorf("batch sops rule online, preview validate result type is invalid")
//	}
//
//	// 提交异步任务
//	logs.Infof("batch sops rule online, request bind rs api, request param: %v", converter.PtrToVal(bindRSReqParam))
//	return svc.bindRS(cts, bindRSReqParam)
//}
//
//func convBatchOperationPreviewResultToReq(previewResultList []*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord], accountID string) *cloud.BatchOperationReq[*lblogic.BindRSRecord] {
//	batchOperationReq := new(cloud.BatchOperationReq[*lblogic.BindRSRecord])
//	batchOperationReq.AccountID = accountID
//	for _, previewResult := range previewResultList {
//		batchOpeItem := &cloud.BatchOperation[*lblogic.BindRSRecord]{
//			ClbID:             previewResult.ClbID,
//			ClbName:           previewResult.ClbName,
//			Vip:               previewResult.Vip,
//			NewRsCount:        previewResult.NewRsCount,
//			UpdateWeightCount: previewResult.UpdateWeightCount,
//			Listeners:         previewResult.Listeners,
//		}
//		batchOperationReq.Data = append(batchOperationReq.Data, batchOpeItem)
//	}
//
//	return batchOperationReq
//}
//
//func (svc *lbSvc) convBindRsRecord(bindRSRecordForSops []*cslb.BindRSRecordForSops) []*lblogic.BindRSRecord {
//	bindRsRecord := make([]*lblogic.BindRSRecord, 0)
//	for _, recordSops := range bindRSRecordForSops {
//		record := &lblogic.BindRSRecord{
//			Action:         recordSops.Action,
//			ListenerName:   recordSops.ListenerName,
//			Protocol:       recordSops.Protocol,
//			IPDomainType:   recordSops.IPDomainType,
//			VIP:            recordSops.VIP,
//			VPorts:         recordSops.VPorts,
//			Domain:         recordSops.Domain,
//			URLPath:        recordSops.URLPath,
//			RSIPs:          recordSops.RSIPs,
//			RSPorts:        recordSops.RSPorts,
//			Weight:         recordSops.Weight,
//			Scheduler:      recordSops.Scheduler,
//			SessionExpired: recordSops.SessionExpired,
//			InstType:       recordSops.InstType,
//			ServerCert:     recordSops.ServerCert,
//			ClientCert:     recordSops.ClientCert,
//			RSInfos:        recordSops.RSInfos,
//			HaveEndPort:    recordSops.HaveEndPort,
//		}
//		bindRsRecord = append(bindRsRecord, record)
//	}
//
//	return bindRsRecord
//}
