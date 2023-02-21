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

package securitygroup

import (
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// BatchDeleteSecurityGroup batch delete security group.
func (svc *securityGroupSvc) BatchDeleteSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.SecurityGroupCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.SecurityGroup,
			Action: meta.Delete, ResourceID: info.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// 已分配业务的资源，不允许操作
	flt := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs}
	err = svc.checkSecurityGroupsInBiz(cts.Kit, flt, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.SecurityGroupAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	successIDs := make([]string, 0)
	for _, id := range req.IDs {
		baseInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
			enumor.SecurityGroupCloudResType, id)
		if err != nil {
			return core.BatchDeleteResp{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    id,
					Error: err.Error(),
				},
			}, errf.NewFromErr(errf.PartialFailed, err)
		}

		switch baseInfo.Vendor {
		case enumor.TCloud:
			err = svc.client.HCService().TCloud.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Aws:
			err = svc.client.HCService().Aws.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.HuaWei:
			err = svc.client.HCService().HuaWei.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Azure:
			err = svc.client.HCService().Azure.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		default:
			return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, baseInfo.Vendor)
		}
		if err != nil {
			return core.BatchDeleteResp{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    id,
					Error: err.Error(),
				},
			}, errf.NewFromErr(errf.PartialFailed, err)
		}

		successIDs = append(successIDs, id)
	}

	return nil, nil
}
