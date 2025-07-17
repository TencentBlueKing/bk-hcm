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

package loadbalancer

import (
	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	"hcm/cmd/cloud-server/service/common"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AssignLbToBiz 分配到业务下
func (svc *lbSvc) AssignLbToBiz(cts *rest.Contexts) (any, error) {
	// 分配关联资源预检
	req := new(cslb.AssignLbToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := common.ValidateTargetBizID(cts.Kit, svc.client.DataService(), enumor.LoadBalancerCloudResType, req.LbIDs,
		req.BkBizID)
	if err != nil {
		return nil, err
	}

	// 权限校验
	clbInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.LoadBalancerCloudResType,
		IDs:          req.LbIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, clbInfoReq)
	if err != nil {
		logs.Errorf("list clb info failed, err: %s, clb_ids: %v, rid: %s", err, req.LbIDs, cts.Kit.Rid)
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.LoadBalancer,
				Action:     meta.Assign,
				ResourceID: info.AccountID,
			},
			BizID: req.BkBizID,
		})
	}

	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	return nil, lblogic.AssignTCloud(cts.Kit, svc.client.DataService(), req.LbIDs, req.BkBizID)
}
