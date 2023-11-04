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

	cloudproto "hcm/pkg/api/cloud-server/eip"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// AssociateEip associate eip.
func (svc *eipSvc) AssociateEip(cts *rest.Contexts) (interface{}, error) {
	return svc.associateEip(cts, handler.ResOperateAuth)
}

// AssociateBizEip associate biz eip.
func (svc *eipSvc) AssociateBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.associateEip(cts, handler.BizOperateAuth)
}

func (svc *eipSvc) associateEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	req := new(cloudproto.AssociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 鉴权和校验资源分配状态和回收状态
	basicInfos, err := svc.associateValidate(cts, validHandler, req)
	if err != nil {
		logs.Errorf("associate validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 创建Eip主机关联审计
	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.EipAuditResType,
		ResID:             req.EipID,
		Action:            protoaudit.Associate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   req.CvmID,
	}
	if err := svc.audit.ResOperationAudit(cts.Kit, operationInfo); err != nil {
		logs.Errorf("create associate eip audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	var vendor enumor.Vendor
	var accountID string
	for _, info := range basicInfos {
		vendor = info.Vendor
		accountID = info.AccountID
		break
	}

	switch vendor {
	case enumor.TCloud:
		return svc.tcloud.AssociateEip(cts, accountID, req)
	case enumor.Aws:
		return svc.aws.AssociateEip(cts, accountID, req)
	case enumor.HuaWei:
		return svc.huawei.AssociateEip(cts, accountID, req)
	case enumor.Gcp:
		return svc.gcp.AssociateEip(cts, accountID, req)
	case enumor.Azure:
		return svc.azure.AssociateEip(cts, accountID, req)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (svc *eipSvc) associateValidate(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler,
	req *cloudproto.AssociateReq) (map[string]types.CloudResourceBasicInfo, error) {

	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.EipCloudResType, IDs: []string{req.EipID}, Fields: types.ResWithRecycleBasicFields},
		},
	}
	if len(req.CvmID) != 0 {
		basicReq.Items = append(basicReq.Items, cloud.ListResourceBasicInfoReq{
			ResourceType: enumor.CvmCloudResType, IDs: []string{req.CvmID}, Fields: types.ResWithRecycleBasicFields})
	}

	if len(req.NetworkInterfaceID) != 0 {
		basicReq.Items = append(basicReq.Items, cloud.ListResourceBasicInfoReq{
			ResourceType: enumor.NetworkInterfaceCloudResType, IDs: []string{req.NetworkInterfaceID},
			Fields: types.ResWithRecycleBasicFields})
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Eip,
		Action: meta.Associate, BasicInfos: basicInfos})
	if err != nil {
		return nil, err
	}

	return basicInfos, nil
}
