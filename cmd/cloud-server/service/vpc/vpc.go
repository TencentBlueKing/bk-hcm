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

package vpc

import (
	"fmt"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitVpcService initialize the vpc service.
func InitVpcService(c *capability.Capability) {
	svc := &vpcSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("BatchDeleteVpc", "DELETE", "/vpcs/batch", svc.BatchDeleteVpc)

	h.Load(c.WebService)
}

type vpcSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// BatchDeleteVpc batch delete vpcs.
func (svc *vpcSvc) BatchDeleteVpc(cts *rest.Contexts) (interface{}, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.VpcCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	succeeded := make([]string, 0)
	for _, id := range req.IDs {
		basicInfo, exists := basicInfoMap[id]
		if !exists {
			return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("id %s has no corresponding vendor", id))
		}

		switch basicInfo.Vendor {
		case enumor.TCloud:
			err := svc.client.HCService().TCloud.Vpc.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
			if err != nil {
				return core.BatchDeleteResp{
					Succeeded: succeeded,
					Failed: &core.FailedInfo{
						ID:    id,
						Error: err.Error(),
					},
				}, errf.NewFromErr(errf.PartialFailed, err)
			}
		}

		succeeded = append(succeeded, id)
	}

	return core.BatchDeleteResp{Succeeded: succeeded}, nil
}
