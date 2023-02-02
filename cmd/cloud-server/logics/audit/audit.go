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
	// ChildResDeleteAudit 子资源删除审计
	ChildResDeleteAudit(kt *kit.Kit, resType enumor.AuditResourceType, parentID string, ids []string) error
	// ResUpdateAudit 资源更新审计
	ResUpdateAudit(kt *kit.Kit, resType enumor.AuditResourceType, id string, updateFields map[string]interface{}) error
	// ChildResUpdateAudit 子资源更新审计
	ChildResUpdateAudit(kt *kit.Kit, resType enumor.AuditResourceType, parentID, id string,
		updateFields map[string]interface{}) error
	// ResBizAssignAudit 资源分配到业务审计
	ResBizAssignAudit(kt *kit.Kit, resType enumor.AuditResourceType, resIDs []string, bizID int64) error
	// ResCloudAreaAssignAudit 资源分配到云区域审计
	ResCloudAreaAssignAudit(kt *kit.Kit, resType enumor.AuditResourceType, opt []ResCloudAreaAssignOption) error
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

// ResBizAssignAudit resource assign to biz audit.
func (a audit) ResBizAssignAudit(kt *kit.Kit, resType enumor.AuditResourceType, resIDs []string, bizID int64) error {
	req := &protoaudit.CloudResourceAssignAuditReq{
		Assigns: make([]protoaudit.CloudResourceAssignInfo, 0, len(resIDs)),
	}

	for _, resID := range resIDs {
		req.Assigns = append(req.Assigns, protoaudit.CloudResourceAssignInfo{
			ResType:         resType,
			ResID:           resID,
			AssignedResType: enumor.BizAuditAssignedResType,
			AssignedResID:   bizID,
		})
	}
	if err := a.dataCli.Global.Audit.CloudResourceAssignAudit(kt.Ctx, kt.Header(), req); err != nil {
		logs.Errorf("request dataservice CloudResourceAssignAudit failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}

// ResCloudAreaAssignOption resource assign to cloud area option.
type ResCloudAreaAssignOption struct {
	ResID   string
	CloudID int64
}

// ResCloudAreaAssignAudit resource assign to cloud area audit.
func (a audit) ResCloudAreaAssignAudit(kt *kit.Kit, resType enumor.AuditResourceType,
	opt []ResCloudAreaAssignOption) error {

	req := &protoaudit.CloudResourceAssignAuditReq{
		Assigns: make([]protoaudit.CloudResourceAssignInfo, 0, len(opt)),
	}

	for _, op := range opt {
		req.Assigns = append(req.Assigns, protoaudit.CloudResourceAssignInfo{
			ResType:         resType,
			ResID:           op.ResID,
			AssignedResType: enumor.CloudAreaAuditAssignedResType,
			AssignedResID:   op.CloudID,
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

// ChildResDeleteAudit child resource delete audit.
func (a audit) ChildResDeleteAudit(kt *kit.Kit, resType enumor.AuditResourceType, parentID string, ids []string) error {
	req := &protoaudit.CloudResourceDeleteAuditReq{
		ParentID: parentID,
		Deletes:  make([]protoaudit.CloudResourceDeleteInfo, 0, len(ids)),
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

// ChildResUpdateAudit child resource update audit.
func (a audit) ChildResUpdateAudit(kt *kit.Kit, resType enumor.AuditResourceType, parentID, id string,
	updateFields map[string]interface{}) error {

	req := &protoaudit.CloudResourceUpdateAuditReq{
		ParentID: parentID,
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
