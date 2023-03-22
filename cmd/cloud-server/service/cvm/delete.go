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
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/hooks/handler"
)

// BatchDeleteCvm batch delete cvm.
func (svc *cvmSvc) BatchDeleteCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteCvmSvc(cts, handler.ResValidWithAuth)
}

// BatchDeleteBizCvm batch delete biz cvm.
func (svc *cvmSvc) BatchDeleteBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteCvmSvc(cts, handler.BizValidWithAuth)
}

func (svc *cvmSvc) batchDeleteCvmSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	req := new(proto.BatchDeleteReq)
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
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	delRes, err := svc.cvmLgc.BatchDeleteCvm(cts.Kit, basicInfoMap)
	if err != nil {
		return delRes, err
	}

	return nil, nil
}

func (svc *cvmSvc) deleteCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, string, error) {

	successIDs := make([]string, 0)
	for _, one := range infoMap {
		switch vendor {
		case enumor.Gcp:
			if err := svc.client.HCService().Gcp.Cvm.DeleteCvm(kt.Ctx, kt.Header(), one.ID); err != nil {
				return successIDs, one.ID, err
			}

		case enumor.Azure:
			req := &hcprotocvm.AzureDeleteReq{
				Force: true,
			}
			if err := svc.client.HCService().Azure.Cvm.DeleteCvm(kt.Ctx, kt.Header(), one.ID, req); err != nil {
				return successIDs, one.ID, err
			}

		default:
			return successIDs, one.ID, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
		}
	}

	return successIDs, "", nil
}

// batchDeleteCvm delete cvm.
func (svc *cvmSvc) batchDeleteCvm(kt *kit.Kit, vendor enumor.Vendor, infoMap []types.CloudResourceBasicInfo) (
	[]string, error) {

	cvmMap := classifier.ClassifyBasicInfoByAccount(infoMap)
	successIDs := make([]string, 0)
	for accountID, reginMap := range cvmMap {
		for region, ids := range reginMap {
			switch vendor {
			case enumor.TCloud:
				req := &hcprotocvm.TCloudBatchDeleteReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := svc.client.HCService().TCloud.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.Aws:
				req := &hcprotocvm.AwsBatchDeleteReq{
					AccountID: accountID,
					Region:    region,
					IDs:       ids,
				}
				if err := svc.client.HCService().Aws.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
					return successIDs, err
				}

			case enumor.HuaWei:
				req := &hcprotocvm.HuaWeiBatchDeleteReq{
					AccountID:      accountID,
					Region:         region,
					IDs:            ids,
					DeletePublicIP: true,
					DeleteDisk:     true,
				}
				if err := svc.client.HCService().HuaWei.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), req); err != nil {
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
