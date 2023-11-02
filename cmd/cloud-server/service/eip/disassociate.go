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

	"github.com/tidwall/gjson"

	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// DisassociateEip disassociate eip.
func (svc *eipSvc) DisassociateEip(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateEip(cts, handler.ResValidWithAuth)
}

// DisassociateBizEip disassociate biz eip.
func (svc *eipSvc) DisassociateBizEip(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateEip(cts, handler.BizValidWithAuth)
}

func (svc *eipSvc) disassociateEip(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	// 获取Eip和Cvm的ID
	eipID, cvmID, err := svc.getDisassociateParam(cts)
	if err != nil {
		return nil, err
	}

	// 鉴权和校验资源分配状态和回收状态
	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.EipCloudResType, IDs: []string{eipID}, Fields: types.ResWithRecycleBasicFields},
			{ResourceType: enumor.CvmCloudResType, IDs: []string{cvmID}, Fields: types.ResWithRecycleBasicFields},
		},
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

	// 创建审计
	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.EipAuditResType,
		ResID:             eipID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   cvmID,
	}
	if err := svc.audit.ResOperationAudit(cts.Kit, operationInfo); err != nil {
		logs.Errorf("create disassociate eip audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		return svc.tcloud.DisassociateEip(cts, accountID, eipID, cvmID)
	case enumor.Aws:
		return svc.aws.DisassociateEip(cts, accountID, eipID, cvmID)
	case enumor.HuaWei:
		return svc.huawei.DisassociateEip(cts, accountID, eipID, cvmID)
	case enumor.Gcp:
		return svc.gcp.DisassociateEip(cts, accountID, eipID, cvmID)
	case enumor.Azure:
		return svc.azure.DisassociateEip(cts, accountID, eipID, cvmID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (svc *eipSvc) getDisassociateParam(cts *rest.Contexts) (string, string, error) {
	body, err := cts.RequestBody()
	if err != nil {
		return "", "", err
	}

	eipID := gjson.GetBytes(body, "eip_id").String()
	if len(eipID) == 0 {
		return "", "", fmt.Errorf("eip_id is required")
	}

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("eip_id", eipID),
		Page:   core.NewDefaultBasePage(),
	}
	rel, err := svc.client.DataService().Global.ListEipCvmRel(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list eip cvm rel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", "", err
	}

	if len(rel.Details) == 0 {
		return "", "", fmt.Errorf("eip_cvm_rel(eip: %s) not found", eipID)
	}

	return eipID, rel.Details[0].CvmID, nil
}
