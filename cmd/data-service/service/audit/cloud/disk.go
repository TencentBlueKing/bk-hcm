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
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

func (ad Audit) diskAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error,
) {
	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idDiskMap, err := ad.listDisk(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		diskData, exist := idDiskMap[one.ResID]
		if !exist {
			continue
		}

		var action enumor.AuditAction
		switch one.AssignedResType {
		case enumor.BizAuditAssignedResType:
			action = enumor.Assign
		case enumor.DeliverAssignedResType:
			action = enumor.Deliver
		default:
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		changed := map[string]int64{"bk_biz_id": one.AssignedResID}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: diskData.CloudID,
			ResName:    converter.PtrToVal(&diskData.Name),
			ResType:    enumor.DiskAuditResType,
			Action:     action,
			BkBizID:    diskData.BkBizID,
			Vendor:     enumor.Vendor(diskData.Vendor),
			AccountID:  diskData.AccountID,
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

func (ad Audit) diskOperationAuditBuild(kt *kit.Kit, ops []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error,
) {
	assCvmOps := make([]protoaudit.CloudResourceOperationInfo, 0)

	for _, op := range ops {
		switch op.Action {
		case protoaudit.Associate, protoaudit.Disassociate:
			switch op.AssociatedResType {
			case enumor.CvmAuditResType:
				assCvmOps = append(assCvmOps, op)
			default:
				return nil, fmt.Errorf("audit associated resource type: %s not support", op.AssociatedResType)
			}
		default:
			return nil, fmt.Errorf("audit action: %s not support", op.Action)
		}
	}

	audits := make([]*tableaudit.AuditTable, 0, len(ops))
	if len(ops) != 0 {
		audit, err := ad.diskAssCvmOperationAuditBuild(kt, assCvmOps)
		if err != nil {
			return nil, err
		}
		audits = append(audits, audit...)
	}
	return audits, nil
}

func (ad Audit) diskAssCvmOperationAuditBuild(
	kt *kit.Kit,
	ops []protoaudit.CloudResourceOperationInfo,
) ([]*tableaudit.AuditTable, error) {
	diskIDs := make([]string, 0)
	cvmIDs := make([]string, 0)
	for _, one := range ops {
		diskIDs = append(diskIDs, one.ResID)
		cvmIDs = append(cvmIDs, one.AssociatedResID)
	}

	diskIDMap, err := ad.listDisk(kt, diskIDs)
	if err != nil {
		return nil, err
	}

	cvmIDMap, err := cvm.ListCvm(kt, ad.dao, cvmIDs)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0)
	for _, one := range ops {

		diskData, exist := diskIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "disk: %s not found", one.ResID)
		}

		cvmData, exist := cvmIDMap[one.AssociatedResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "cvm: %s not found", one.AssociatedResID)
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      diskData.ID,
			CloudResID: diskData.CloudID,
			ResName:    diskData.Name,
			ResType:    enumor.DiskAuditResType,
			Action:     action,
			BkBizID:    diskData.BkBizID,
			Vendor:     enumor.Vendor(diskData.Vendor),
			AccountID:  diskData.AccountID,
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

func (ad Audit) diskDeleteAuditBuild(
	kt *kit.Kit,
	deletes []protoaudit.CloudResourceDeleteInfo,
) ([]*tableaudit.AuditTable, error) {
	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	diskIDMap, err := ad.listDisk(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0)
	for _, one := range deletes {
		diskData, exist := diskIDMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: diskData.CloudID,
			ResName:    diskData.Name,
			ResType:    enumor.DiskAuditResType,
			Action:     enumor.Delete,
			BkBizID:    diskData.BkBizID,
			Vendor:     enumor.Vendor(diskData.Vendor),
			AccountID:  diskData.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: diskData,
			},
		})
	}

	return audits, nil
}

func (ad Audit) listDisk(kt *kit.Kit, ids []string) (map[string]*disk.DiskModel, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := ad.dao.Disk().List(kt, opt)
	if err != nil {
		logs.Errorf("list disk failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]*disk.DiskModel, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}
