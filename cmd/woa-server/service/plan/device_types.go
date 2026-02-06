/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package plan ...
package plan

import (
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateDeviceTypePhysicalRel create device type physical rel.
func (s *service) CreateDeviceTypePhysicalRel(cts *rest.Contexts) (any, error) {
	req := new(rpproto.WoaDeviceTypePhysicalRelBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to create device type phy rel, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create device type parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GlobalConfig, Action: meta.Create}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, err
	}

	result, err := s.client.DataService().Global.ResourcePlan.BatchCreateWoaDeviceTypePhysicalRel(cts.Kit, req)
	if err != nil {
		logs.Errorf("failed to create device type, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	return result.IDs, nil
}
