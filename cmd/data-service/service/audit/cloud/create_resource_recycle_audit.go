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
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CloudResourceRecycleAudit cloud resource recycle audit.
func (ad Audit) CloudResourceRecycleAudit(cts *rest.Contexts) (interface{}, error) {
	req := new(protoaudit.CloudResourceRecycleAuditReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	action, err := req.Action.ConvAuditAction()
	if err != nil {
		return nil, err
	}

	resType, exists := enumor.RecycleAuditResTypeMap[req.ResType]
	if !exists {
		return nil, errf.Newf(errf.InvalidParameter, "recycle resource type %s is invalid", req.ResType)
	}

	resIDs := make([]string, 0, len(req.Infos))
	for _, info := range req.Infos {
		resIDs = append(resIDs, info.ResID)
	}
	infos, err := ad.dao.RecycleRecord().ListResourceInfo(cts.Kit, resType, resIDs)
	if err != nil {
		return nil, err
	}

	if len(infos) != len(req.Infos) {
		return nil, errf.Newf(errf.InvalidParameter, "audit recycle resource count is invalid")
	}

	audits := make([]*tableaudit.AuditTable, 0, len(infos))
	for idx, info := range infos {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      info.ID,
			CloudResID: info.CloudID,
			ResName:    info.Name,
			ResType:    req.ResType,
			Action:     action,
			BkBizID:    info.BkBizID,
			Vendor:     info.Vendor,
			AccountID:  info.AccountID,
			Operator:   cts.Kit.User,
			Source:     cts.Kit.GetRequestSource(),
			Rid:        cts.Kit.Rid,
			AppCode:    cts.Kit.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: req.Infos[idx].Data,
			},
		})
	}

	if err := ad.dao.Audit().BatchCreate(cts.Kit, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
