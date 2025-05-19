/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * r Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain r copy of the License at http://opensource.org/licenses/MIT
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

package resusagebizrel

import (
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListResUsageBizRel list res usage biz relation.
func (r *service) ListResUsageBizRel(cts *rest.Contexts) (any, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}

	data, err := r.dao.ResUsageBizRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list resUsage biz relations failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &core.ListResult{Count: data.Count}, nil
	}

	details := make([]corecloud.ResUsageBizRel, len(data.Details))
	for idx, table := range data.Details {
		details[idx] = corecloud.ResUsageBizRel{
			RelID:        table.RelID,
			ResType:      table.ResType,
			ResID:        table.ResID,
			ResCloudID:   table.ResCloudID,
			UsageBizID:   table.UsageBizID,
			RelCreator:   table.RelCreator,
			RelCreatedAt: table.RelCreatedAt.String(),
		}
	}

	return &protocloud.ListResUsageBizRelResult{Details: details}, nil
}
