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

package sgcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListSgCvmRels rels.
func (svc *sgCvmRelSvc) ListSgCvmRels(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.SGCvmRel().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list security group cvm rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list security group cvm rels failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.SGCvmRelListResult{Count: result.Count}, nil
	}

	details := make([]corecloud.SecurityGroupCvmRel, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.SecurityGroupCvmRel{
			ID:              one.ID,
			CvmID:           one.CvmID,
			SecurityGroupID: one.SecurityGroupID,
			Creator:         one.Creator,
			CreatedAt:       one.CreatedAt.String(),
		})
	}

	return &protocloud.SGCvmRelListResult{Details: details}, nil
}
