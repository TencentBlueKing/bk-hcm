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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ResValidWithAuth validate and authorize cloud resource for resource manager.
func ResValidWithAuth(cts *rest.Contexts, opt *ValidWithAuthOption) error {
	// authorize one resource
	if opt.BasicInfo != nil {
		// validate if resource is not in biz for write operation
		if opt.Action != meta.Find && opt.BasicInfo.BkBizID != constant.UnassignedBiz && opt.BasicInfo.BkBizID != 0 {
			return errf.Newf(errf.InvalidParameter, "resource %s is already assigned", opt.BasicInfo.ID)
		}

		authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: opt.ResType, Action: opt.Action,
			ResourceID: opt.BasicInfo.AccountID}}
		return opt.Authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	}

	// batch authorize resource
	authRes := make([]meta.ResourceAttribute, 0, len(opt.BasicInfos))
	assignedIDs := make([]string, 0)
	for id, info := range opt.BasicInfos {
		// validate if resource is not in biz for write operation
		if opt.Action != meta.Find && info.BkBizID != constant.UnassignedBiz && info.BkBizID != 0 {
			assignedIDs = append(assignedIDs, id)
		}

		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: opt.ResType, Action: opt.Action,
			ResourceID: info.AccountID}})
	}

	if len(assignedIDs) > 0 {
		return errf.Newf(errf.InvalidParameter, "resources(ids: %+v) are already assigned", assignedIDs)
	}

	return opt.Authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
}

// ListResourceAuthRes list authorized cloud resource for resource manager.
func ListResourceAuthRes(cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool, error) {
	authOpt := &meta.ListAuthResInput{Type: opt.ResType, Action: opt.Action}
	return opt.Authorizer.ListAuthInstWithFilter(cts.Kit, authOpt, opt.Filter, "account_id")
}
