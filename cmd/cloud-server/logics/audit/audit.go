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

// Package audit 对 data-service 提供的审计接口进行二次封装，提供丰富的场景审计能力，方便 cloud_server 使用。
package audit

import (
	protoaudit "hcm/pkg/api/data-service/audit"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Interface define audit interface.
type Interface interface {
	// ResDeleteAudit 资源删除审计
	ResDeleteAudit(kt *kit.Kit, resType enumor.AuditResourceType, ids []string) error
	// ResUpdateAudit 资源更新审计
	ResUpdateAudit(kt *kit.Kit, resType enumor.AuditResourceType, id string, updateFields map[string]interface{}) error
	// ResAssignAudit 资源分配审计
	ResAssignAudit(kt *kit.Kit, resType enumor.AuditResourceType, resIDs []string, bizID int64) error
}

var _ Interface = new(audit)

// NewAudit new audit.
func NewAudit(dataCli *dataservice.Client) Interface {
	return &audit{
		dataCli: dataCli,
	}
}

type audit struct {
	dataCli *dataservice.Client
}

// ResAssignAudit resource assign audit.
func (a audit) ResAssignAudit(kt *kit.Kit, resType enumor.AuditResourceType, resIDs []string, bizID int64) error {
	req := &protoaudit.CloudResourceAssignAuditReq{
		Assigns: make([]protoaudit.CloudResourceAssignInfo, 0, len(resIDs)),
	}

	for _, resID := range resIDs {
		req.Assigns = append(req.Assigns, protoaudit.CloudResourceAssignInfo{
			ResType: resType,
			ResID:   resID,
			BkBizID: bizID,
		})
	}
	if err := a.dataCli.Global.Audit.CloudResourceAssignAudit(kt.Ctx, kt.Header(), req); err != nil {
		logs.Errorf("request dataservice CloudResourceAssignAudit failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}

// ResDeleteAudit resource delete audit.
func (a audit) ResDeleteAudit(kt *kit.Kit, resType enumor.AuditResourceType, ids []string) error {
	req := &protoaudit.CloudResourceDeleteAuditReq{
		Deletes: make([]protoaudit.CloudResourceDeleteInfo, 0, len(ids)),
	}

	for _, id := range ids {
		req.Deletes = append(req.Deletes, protoaudit.CloudResourceDeleteInfo{
			ResType: resType,
			ResID:   id,
		})
	}
	if err := a.dataCli.Global.Audit.CloudResourceDeleteAudit(kt.Ctx, kt.Header(), req); err != nil {
		logs.Errorf("request dataservice CloudResourceDeleteAudit failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}

// ResUpdateAudit resource update audit.
func (a audit) ResUpdateAudit(kt *kit.Kit, resType enumor.AuditResourceType, id string,
	updateFields map[string]interface{}) error {

	req := &protoaudit.CloudResourceUpdateAuditReq{
		Updates: []protoaudit.CloudResourceUpdateInfo{
			{
				ResType:      resType,
				ResID:        id,
				UpdateFields: updateFields,
			},
		},
	}
	if err := a.dataCli.Global.Audit.CloudResourceUpdateAudit(kt.Ctx, kt.Header(), req); err != nil {
		logs.Errorf("request dataservice CloudResourceUpdateAudit failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}
