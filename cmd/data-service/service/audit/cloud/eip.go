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

package cloud

import (
	"fmt"

	"hcm/cmd/data-service/service/audit/cloud/cvm"
	networkinterface "hcm/cmd/data-service/service/audit/cloud/network-interface"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/dal/table/cloud/eip"
	tableni "hcm/pkg/dal/table/cloud/network-interface"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

func (ad Audit) eipAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error,
) {
	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idEipMap, err := ad.listEip(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		eipData, exist := idEipMap[one.ResID]
		if !exist {
			continue
		}

		changed := make(map[string]interface{})
		if one.AssignedResType != enumor.BizAuditAssignedResType {
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}
		changed["bk_biz_id"] = one.AssignedResID

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: eipData.CloudID,
			ResName:    converter.PtrToVal(eipData.Name),
			ResType:    enumor.EipAuditResType,
			Action:     enumor.Assign,
			BkBizID:    eipData.BkBizID,
			Vendor:     enumor.Vendor(eipData.Vendor),
			AccountID:  eipData.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Changed: changed,
			},
		})
	}

	return audits, nil
}

func (ad Audit) eipOperationAuditBuild(kt *kit.Kit, ops []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error,
) {
	assCvmOps := make([]protoaudit.CloudResourceOperationInfo, 0)
	assNetworkOps := make([]protoaudit.CloudResourceOperationInfo, 0)

	for _, op := range ops {
		switch op.Action {
		case protoaudit.Associate, protoaudit.Disassociate:
			switch op.AssociatedResType {
			case enumor.CvmAuditResType:
				assCvmOps = append(assCvmOps, op)
			case enumor.NetworkInterfaceAuditResType:
				assNetworkOps = append(assNetworkOps, op)
			default:
				return nil, fmt.Errorf("audit associated resource type: %s not support", op.AssociatedResType)
			}

		default:
			return nil, fmt.Errorf("audit action: %s not support", op.Action)
		}
	}

	audits := make([]*tableaudit.AuditTable, 0, len(ops))
	if len(assCvmOps) != 0 {
		audit, err := ad.eipAssCvmOperationAuditBuild(kt, assCvmOps)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	if len(assNetworkOps) != 0 {
		audit, err := ad.eipAssNetworkOperationAuditBuild(kt, assNetworkOps)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	return audits, nil
}

func (ad Audit) listEip(kt *kit.Kit, ids []string) (map[string]*eip.EipModel, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := ad.dao.Eip().List(kt, opt)
	if err != nil {
		logs.Errorf("list eip failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]*eip.EipModel, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

func (ad Audit) eipAssCvmOperationAuditBuild(
	kt *kit.Kit,
	ops []protoaudit.CloudResourceOperationInfo,
) ([]*tableaudit.AuditTable, error) {
	eipIDs := make([]string, 0)
	cvmIDs := make([]string, 0)
	for _, one := range ops {
		eipIDs = append(eipIDs, one.ResID)
		cvmIDs = append(cvmIDs, one.AssociatedResID)
	}

	eipIDMap, err := ad.listEip(kt, eipIDs)
	if err != nil {
		return nil, err
	}

	cvmIDMap, err := cvm.ListCvm(kt, ad.dao, cvmIDs)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0)
	for _, one := range ops {

		eipData, exist := eipIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "eip: %s not found", one.ResID)
		}

		cvmData, exist := cvmIDMap[one.AssociatedResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "cvm: %s not found", one.AssociatedResID)
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		var resName string
		if eipData.Name != nil {
			resName = *eipData.Name
		}
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      eipData.ID,
			CloudResID: eipData.CloudID,
			ResName:    resName,
			ResType:    enumor.EipAuditResType,
			Action:     action,
			BkBizID:    eipData.BkBizID,
			Vendor:     enumor.Vendor(eipData.Vendor),
			AccountID:  eipData.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.AssociatedOperationAudit{
					AssResType:    enumor.CvmAuditResType,
					AssResID:      cvmData.ID,
					AssResCloudID: cvmData.CloudID,
					AssResName:    cvmData.Name,
				},
			},
		})
	}
	return audits, nil
}

func (ad Audit) eipAssNetworkOperationAuditBuild(
	kt *kit.Kit,
	ops []protoaudit.CloudResourceOperationInfo,
) ([]*tableaudit.AuditTable, error) {
	eipIDs := make([]string, 0)
	networkIDs := make([]string, 0)
	for _, one := range ops {
		eipIDs = append(eipIDs, one.ResID)
		// 统计有效 ID
		if one.AssociatedResID != "" {
			networkIDs = append(networkIDs, one.AssociatedResID)
		}
	}

	eipIDMap, err := ad.listEip(kt, eipIDs)
	if err != nil {
		return nil, err
	}

	var networkIDMap map[string]tableni.NetworkInterfaceTable

	if len(networkIDs) > 0 {
		networkIDMap, err = networkinterface.ListNetworkInterface(kt, ad.dao, networkIDs)
		if err != nil {
			return nil, err
		}
	}

	audits := make([]*tableaudit.AuditTable, 0)
	for _, one := range ops {
		eipData, exist := eipIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "eip: %s not found", one.ResID)
		}

		basicDetail := new(tableaudit.AssociatedOperationAudit)
		if one.AssociatedResID != "" {
			networkData, exist := networkIDMap[one.AssociatedResID]
			if !exist {
				return nil, errf.Newf(errf.RecordNotFound, "network interface: %s not found", one.AssociatedResID)
			}
			basicDetail.AssResType = enumor.NetworkInterfaceAuditResType
			basicDetail.AssResID = networkData.ID
			basicDetail.AssResCloudID = networkData.CloudID
			basicDetail.AssResName = networkData.Name
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      eipData.ID,
			CloudResID: eipData.CloudID,
			ResName:    converter.PtrToVal(eipData.Name),
			ResType:    enumor.EipAuditResType,
			Action:     action,
			BkBizID:    eipData.BkBizID,
			Vendor:     enumor.Vendor(eipData.Vendor),
			AccountID:  eipData.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: basicDetail,
			},
		})
	}

	return audits, nil
}

func (ad Audit) eipDeleteAuditBuild(
	kt *kit.Kit,
	deletes []protoaudit.CloudResourceDeleteInfo,
) ([]*tableaudit.AuditTable, error) {
	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	eipIDMap, err := ad.listEip(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0)
	for _, one := range deletes {
		eipData, exist := eipIDMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: eipData.CloudID,
			ResName:    *eipData.Name,
			ResType:    enumor.EipAuditResType,
			Action:     enumor.Delete,
			BkBizID:    eipData.BkBizID,
			Vendor:     enumor.Vendor(eipData.Vendor),
			AccountID:  eipData.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: eipData,
			},
		})
	}

	return audits, nil
}
