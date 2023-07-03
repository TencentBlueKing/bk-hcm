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

package cvm

import (
	proto "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/hooks/handler"
)

// BatchStartCvm batch start cvm.
func (svc *cvmSvc) BatchStartCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchStartCvmSvc(cts, handler.ResValidWithAuth)
}

// BatchStartBizCvm batch start biz cvm.
func (svc *cvmSvc) BatchStartBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchStartCvmSvc(cts, handler.BizValidWithAuth)
}

func (svc *cvmSvc) batchStartCvmSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	req := new(proto.BatchStartCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          req.IDs,
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Start, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResBaseOperationAudit(cts.Kit, enumor.CvmAuditResType, protoaudit.Start, req.IDs); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmVendorMap := classifier.ClassifyBasicInfoByVendor(basicInfoMap)
	successIDs := make([]string, 0)
	for vendor, infos := range cvmVendorMap {
		switch vendor {
		case enumor.TCloud, enumor.Aws, enumor.HuaWei:
			ids, err := svc.batchStartCvm(cts.Kit, vendor, infos)
			successIDs = append(successIDs, ids...)
			if err != nil {
				return core.BatchOperateResult{
					Succeeded: successIDs,
					Failed: &core.FailedInfo{
						Error: err,
					},
				}, errf.NewFromErr(errf.PartialFailed, err)
			}

		case enumor.Azure, enumor.Gcp:
			ids, failedID, err := svc.startCvm(cts.Kit, vendor, infos)
			successIDs = append(successIDs, ids...)
			if err != nil {
				return core.BatchOperateResult{
					Succeeded: successIDs,
					Failed: &core.FailedInfo{
						ID:    failedID,
						Error: err,
					},
				}, errf.NewFromErr(errf.PartialFailed, err)
			}

		default:
			return core.BatchOperateResult{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    infos[0].ID,
					Error: errf.Newf(errf.Unknown, "vendor: %s not support", vendor),
				},
			}, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}

	}

	return nil, nil
}

func (svc *cvmSvc) startCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, string, error) {

	successIDs := make([]string, 0)
	for _, one := range infoMap {
		switch vendor {
		case enumor.Gcp:
			if err := svc.client.HCService().Gcp.Cvm.StartCvm(kt.Ctx, kt.Header(), one.ID); err != nil {
				return successIDs, one.ID, err
			}

		case enumor.Azure:
			if err := svc.client.HCService().Azure.Cvm.StartCvm(kt.Ctx, kt.Header(), one.ID); err != nil {
				return successIDs, one.ID, err
			}

		default:
			return successIDs, one.ID, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}
	}

	return successIDs, "", nil
}

// batchStartCvm start cvm.
func (svc *cvmSvc) batchStartCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, error) {

	cvmMap := classifier.ClassifyBasicInfoByAccount(infoMap)
	successIDs := make([]string, 0)
	for accountID, reginMap := range cvmMap {
		for region, ids := range reginMap {
			switch vendor {
			case enumor.TCloud:
				req := &hcprotocvm.TCloudBatchStartReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := svc.client.HCService().TCloud.Cvm.BatchStartCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.Aws:
				req := &hcprotocvm.AwsBatchStartReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := svc.client.HCService().Aws.Cvm.BatchStartCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.HuaWei:
				req := &hcprotocvm.HuaWeiBatchStartReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := svc.client.HCService().HuaWei.Cvm.BatchStartCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			default:
				return successIDs, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
			}

			successIDs = append(successIDs, ids...)
		}
	}

	return successIDs, nil
}
