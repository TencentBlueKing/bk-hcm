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
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

func addRecyclingFilter(h ListAuthResHandler, cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool,
	error) {
	recycling := filter.AtomRule{Field: "recycle_status", Op: filter.Equal.Factory(), Value: enumor.RecycleStatus}
	if opt.Filter == nil {
		opt.Filter = &filter.Expression{Op: filter.And, Rules: []filter.RuleFactory{recycling}}
	} else {
		bizFilter, err := tools.And(recycling, opt.Filter)
		if err != nil {
			return nil, false, err
		}
		opt.Filter = bizFilter
	}
	return h(cts, opt)
}

// GetRecyclingAuth check get recycled resource permission
func GetRecyclingAuth(cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool, error) {
	return addRecyclingFilter(ListResourceAuthRes, cts, opt)
}

// BizRecyclingAuth validate and authorize cloud resource for biz recycle bin manager handler
func BizRecyclingAuth(cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool, error) {
	return addRecyclingFilter(ListBizAuthRes, cts, opt)
}

// ListResourceRecycleAuthRes list authorized recycled resource for resource manager.
func ListResourceRecycleAuthRes(cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool, error) {
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.RecycleBin, Action: opt.Action}}
	_, authorized, err := opt.Authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, false, err
	}

	if !authorized {
		return nil, true, nil
	}

	return opt.Filter, false, err
}

// ListBizRecycleAuthRes list authorized recycled biz resource for resource manager.
func ListBizRecycleAuthRes(cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool, error) {
	opt.ResType = meta.RecycleBin
	return ListBizAuthRes(cts, opt)
}
