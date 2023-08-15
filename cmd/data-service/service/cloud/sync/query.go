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

package sync

import (
	"fmt"

	"hcm/pkg/api/core"
	coresync "hcm/pkg/api/core/cloud/sync"
	dssync "hcm/pkg/api/data-service/cloud/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListAccountSyncDetail list account sync detail.
func (svc *service) ListAccountSyncDetail(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	daoResp, err := svc.dao.AccountSyncDetail().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account sync detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list account sync detail failed, err: %v", err)
	}
	if req.Page.Count {
		return &dssync.ListResult{Count: daoResp.Count}, nil
	}

	details := make([]coresync.AccountSyncDetailTable, 0, len(daoResp.Details))
	for _, one := range daoResp.Details {
		details = append(details, coresync.AccountSyncDetailTable(one))
	}

	return &dssync.ListResult{Details: details}, nil
}
