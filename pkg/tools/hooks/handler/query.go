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
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListBizAuthRes 校验用户是否有业务下对应资源查看权限, 如果有传入filter, 给filter 加上业务id 过滤条件
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
	if opt.Filter == nil {
		return nil, false, nil
	}

	bizFilter, err := tools.And(filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: bizID},
		opt.Filter)
	if err != nil {
		return nil, false, err
	}
	return bizFilter, false, err
}

// ListResourceAuthRes 资源下 查询校验
func ListResourceAuthRes(cts *rest.Contexts, opt *ListAuthResOption) (*filter.Expression, bool, error) {
	authOpt := &meta.ListAuthResInput{Type: opt.ResType, Action: opt.Action}
	return opt.Authorizer.ListAuthInstWithFilter(cts.Kit, authOpt, opt.Filter, "account_id")
}
