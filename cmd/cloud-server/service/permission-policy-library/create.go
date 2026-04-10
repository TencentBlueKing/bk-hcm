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
	"fmt"

	proto "hcm/pkg/api/cloud-server"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// CreatePermissionPolicyLibrary create permission policy library.
func (svc *svc) CreatePermissionPolicyLibrary(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.PermissionPolicyLibraryCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.PermissionPolicyLibrary,
			Action: meta.Create,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create permission policy library auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return svc.createForTCloud(cts, req)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (svc *svc) createForTCloud(cts *rest.Contexts, req *proto.PermissionPolicyLibraryCreateReq) (interface{}, error) {
	dsReq := &protocloud.PermissionPolicyLibraryBatchCreateReq{
		PermissionPolicyLibraries: []protocloud.PermissionPolicyLibraryCreate{
			{
				Name:           req.Name,
				PolicyDocument: req.PolicyDocument,
				BkBizIDs:       req.BkBizIDs,
				Memo:           cvt.ValToPtr(req.Memo),
			},
		},
	}

	result, err := svc.client.DataService().TCloud.PermissionPolicyLibrary.BatchCreate(cts.Kit, dsReq)
	if err != nil {
		logs.Errorf("create permission policy library failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if result == nil || len(result.IDs) == 0 {
		return nil, errf.New(errf.Aborted, "create returned empty result")
	}

	return &proto.PermissionPolicyLibraryCreateResult{ID: result.IDs[0]}, nil
}
