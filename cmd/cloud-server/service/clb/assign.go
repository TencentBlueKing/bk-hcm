/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package clb

import (
	"hcm/cmd/cloud-server/logics/clb"
	csclb "hcm/pkg/api/cloud-server/clb"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// AssignClbToBiz 分配到业务下
func (svc *clbSvc) AssignClbToBiz(cts *rest.Contexts) (any, error) {
	// 分配关联资源预检
	req := new(csclb.AssignClbToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ClbCloudResType,
		IDs:          req.ClbIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := converter.MapToSlice(basicInfoMap, func(k string, v types.CloudResourceBasicInfo) meta.ResourceAttribute {
		return meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.Clb,
				Action:     meta.Assign,
				ResourceID: v.AccountID,
			},
			BizID: req.BkBizID,
		}
	})

	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	return nil, clb.Assign(cts.Kit, svc.client.DataService(), req.ClbIDs, req.BkBizID)
}
