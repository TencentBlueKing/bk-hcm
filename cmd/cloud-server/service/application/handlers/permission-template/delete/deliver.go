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

package deletepermtemplate

import (
	"fmt"
	"strconv"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	hspermissiontemplate "hcm/pkg/api/hc-service/permission-template"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
)

// GenerateApplicationContent generates the content to be stored in DB for this application.
func (a *ApplicationOfDeletePermTemplate) GenerateApplicationContent() interface{} {
	return a.content
}

// Deliver executes the actual resource deletion after approval.
func (a *ApplicationOfDeletePermTemplate) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	switch a.Vendor() {
	case enumor.TCloud:
		return a.deleteTCloud()
	default:
		err := errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", a.Vendor())
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}
}

func (a *ApplicationOfDeletePermTemplate) deleteTCloud() (enumor.ApplicationStatus, map[string]interface{}, error) {
	kt := a.Cts.Kit

	req := &protocloud.PermissionTemplateExtListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.content.ID)),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := a.Client.DataService().TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)
	if err != nil {
		logs.Errorf("deliver: get tcloud permission template failed, id: %s, err: %v, rid: %s",
			a.content.ID, err, kt.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error": fmt.Sprintf("get permission template failed, err: %v", err),
		}, err
	}

	if len(result.Details) == 0 {
		err = fmt.Errorf("permission template(%s) not found", a.content.ID)
		return enumor.DeliverError, map[string]interface{}{"error": err.Error()}, err
	}

	tmpl := result.Details[0]
	cloudPolicyID, err := strconv.ParseUint(tmpl.CloudID, 10, 64)
	if err != nil {
		logs.Errorf("deliver: parse cloud policy id failed, cloudID: %s, err: %v, rid: %s", tmpl.CloudID, err, kt.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error": fmt.Sprintf("parse cloud policy id failed, err: %v", err),
		}, err
	}

	cloudReq := &hspermissiontemplate.DeleteCAMPolicyReq{AccountID: tmpl.AccountID, PolicyIDs: []uint64{cloudPolicyID}}
	if err = a.Client.HCService().TCloud.PermissionTemplate.DeleteCAMPolicy(kt, cloudReq); err != nil {
		logs.Errorf("deliver: delete cam policy failed, templateID: %s, policyID: %d, err: %v, rid: %s", a.content.ID,
			cloudPolicyID, err, kt.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error": fmt.Sprintf("delete cam policy failed, err: %v", err),
		}, err
	}

	if err = a.Audit.ResDeleteAudit(kt, enumor.PermissionTemplateAuditResType, []string{a.content.ID}); err != nil {
		logs.Errorf("deliver: create delete permission template audit failed, id: %s, err: %v, rid: %s",
			a.content.ID, err, kt.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error": fmt.Sprintf("create delete permission template audit failed, err: %v", err),
		}, err
	}

	localReq := &protocloud.PermissionTemplateBatchDeleteReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("id", a.content.ID)),
	}
	if err = a.Client.DataService().Global.PermissionTemplate.BatchDelete(kt, localReq); err != nil {
		logs.Errorf("deliver: delete local permission template failed, templateID: %s, err: %v, rid: %s", a.content.ID,
			err, kt.Rid)
		return enumor.DeliverError, map[string]interface{}{
			"error": fmt.Sprintf("delete local permission template failed, err: %v", err),
		}, err
	}

	return enumor.Completed, map[string]interface{}{"id": a.content.ID}, nil
}
