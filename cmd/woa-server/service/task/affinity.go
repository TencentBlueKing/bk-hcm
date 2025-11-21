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

// Package task affinity check service
package task

import (
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetAffinityMatchDetail 获取亲和性匹配详情
func (s *service) GetAffinityMatchDetail(cts *rest.Contexts) (any, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	err = s.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bkBizID,
	})
	if err != nil {
		logs.Errorf("failed to check biz affinity match permission, bizID: %d, err: %v, rid: %s",
			bkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	input := new(types.AffinityMatchReq)
	if err = cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get affinity match detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	input.BkBizID = bkBizID

	if err = input.Validate(); err != nil {
		logs.Errorf("failed to get affinity match detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(pkg.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Scheduler().GetAffinityMatchDetail(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get affinity match detail, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
