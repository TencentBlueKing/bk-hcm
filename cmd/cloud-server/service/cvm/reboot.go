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
	typecvm "hcm/pkg/adaptor/types/cvm"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// BatchRebootCvm ...
func (svc *cvmSvc) BatchRebootCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.BatchRebootCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm,
			Action: meta.Reboot, ResourceID: info.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// 校验资源是否已经分配，已分配资源不允许进行操作
	flt := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs}
	err = CheckCvmsInBiz(cts.Kit, svc.client, flt, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResBaseOperationAudit(cts.Kit, enumor.CvmAuditResType, protoaudit.Reboot, req.IDs); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmVendorMap := cvmClassificationByVendor(basicInfoMap)
	successIDs := make([]string, 0)
	for vendor, infos := range cvmVendorMap {
		switch vendor {
		case enumor.TCloud, enumor.Aws, enumor.HuaWei:
			ids, err := svc.batchRebootCvm(cts.Kit, vendor, infos)
			successIDs = append(successIDs, ids...)
			if err != nil {
				return core.BatchDeleteResp{
					Succeeded: successIDs,
					Failed: &core.FailedInfo{
						Error: err.Error(),
					},
				}, errf.NewFromErr(errf.PartialFailed, err)
			}

		case enumor.Azure, enumor.Gcp:
			ids, failedID, err := svc.rebootCvm(cts.Kit, vendor, infos)
			successIDs = append(successIDs, ids...)
			if err != nil {
				return core.BatchDeleteResp{
					Succeeded: successIDs,
					Failed: &core.FailedInfo{
						ID:    failedID,
						Error: err.Error(),
					},
				}, errf.NewFromErr(errf.PartialFailed, err)
			}

		default:
			return core.BatchDeleteResp{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    infos[0].ID,
					Error: errf.Newf(errf.Unknown, "vendor: %s not support", vendor).Error(),
				},
			}, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}

	}

	return nil, nil
}

func (svc *cvmSvc) rebootCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, string, error) {

	successIDs := make([]string, 0)
	for _, one := range infoMap {
		switch vendor {
		case enumor.Gcp:
			if err := svc.client.HCService().Gcp.Cvm.RebootCvm(kt.Ctx, kt.Header(), one.ID); err != nil {
				return successIDs, one.ID, err
			}

		case enumor.Azure:
			if err := svc.client.HCService().Azure.Cvm.RebootCvm(kt.Ctx, kt.Header(), one.ID); err != nil {
				return successIDs, one.ID, err
			}

		default:
			return successIDs, one.ID, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}
	}

	return successIDs, "", nil
}

// batchRebootCvm reboot cvm.
func (svc *cvmSvc) batchRebootCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, error) {

	cvmMap := cvmClassification(infoMap)
	successIDs := make([]string, 0)
	for accountID, reginMap := range cvmMap {
		for region, ids := range reginMap {
			switch vendor {
			case enumor.TCloud:
				req := &hcprotocvm.TCloudBatchRebootReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
					StopType:  typecvm.SoftFirst,
				}
				if err := svc.client.HCService().TCloud.Cvm.BatchRebootCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.Aws:
				req := &hcprotocvm.AwsBatchRebootReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := svc.client.HCService().Aws.Cvm.BatchRebootCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.HuaWei:
				req := &hcprotocvm.HuaWeiBatchRebootReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
					Force:     true,
				}
				if err := svc.client.HCService().HuaWei.Cvm.BatchRebootCvm(kt.Ctx, kt.Header(), req); err != nil {
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
