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

package permissionpolicylibrary

import (
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// DeletePermissionPolicyLibrary delete permission policy library.
func (svc *svc) DeletePermissionPolicyLibrary(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.PermissionPolicyLibrary,
			Action: meta.Delete,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("delete permission policy library auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	listReq := &protocloud.PermissionPolicyLibraryListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("get permission policy library failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}
	if result == nil || len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "permission policy library %s is not found", id)
	}
	if result.Details[0].Vendor != vendor {
		return nil, errf.Newf(errf.InvalidParameter, "vendor mismatch: path vendor %s, record vendor %s",
			vendor, result.Details[0].Vendor)
	}

	// TODO: 待云权限模板功能实现后补充关联检查，若关联了云权限模板则不允许删除

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.PermissionPolicyLibraryAuditResType, []string{id}); err != nil {
		logs.Errorf("create delete audit failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	deleteReq := &protocloud.PermissionPolicyLibraryBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = svc.client.DataService().Global.PermissionPolicyLibrary.BatchDelete(cts.Kit, deleteReq); err != nil {
		logs.Errorf("delete permission policy library failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
