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

package handler

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// BizValidWithAuth validate and authorize cloud resource for biz maintainer handler
func BizValidWithAuth(cts *rest.Contexts, opt *ValidWithAuthOption) error {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return err
	}

	if bizID <= 0 {
		return errf.New(errf.InvalidParameter, "resource is not in biz")
	}

	// authorize one resource
	if opt.BasicInfo != nil {
		if !opt.DisableBizIDEqual && opt.BasicInfo.BkBizID != bizID {
			return errf.New(errf.InvalidParameter, "resource biz not matches url biz")
		}

		if opt.BasicInfo.RecycleStatus == enumor.RecycleStatus {
			return errf.New(errf.InvalidParameter, "resource is in recycle bin")
		}

		authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: opt.ResType, Action: opt.Action},
			BizID: bizID}
		return opt.Authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	}

	// batch authorize resource
	authRes := make([]meta.ResourceAttribute, 0, len(opt.BasicInfos))
	notMatchedIDs, recycledIDs := make([]string, 0), make([]string, 0)
	for id, info := range opt.BasicInfos {
		if info.BkBizID != bizID {
			notMatchedIDs = append(notMatchedIDs, id)
		}

		if info.RecycleStatus == enumor.RecycleStatus {
			recycledIDs = append(recycledIDs, id)
		}

		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: opt.ResType, Action: opt.Action},
			BizID: bizID})
	}

	if !opt.DisableBizIDEqual && len(notMatchedIDs) > 0 {
		return errf.Newf(errf.InvalidParameter, "resources(ids: %+v) not matches url biz", notMatchedIDs)
	}

	if len(recycledIDs) > 0 {
		return errf.Newf(errf.InvalidParameter, "resources(ids: %+v) is in recycle bin", recycledIDs)
	}

	return opt.Authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
}

// ListBizAuthRes list authorized cloud resource for resource manager.
func ListBizAuthRes(cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, false, err
	}

	if bizID <= 0 {
		return nil, false, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: opt.ResType, Action: opt.Action}, BizID: bizID}
	_, authorized, err := opt.Authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, false, err
	}

	if !authorized {
		return nil, true, nil
	}

	bizFilter, err := tools.And(filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bizID},
		opt.Filter)
	if err != nil {
		return nil, false, err
	}

	return bizFilter, false, err
}
