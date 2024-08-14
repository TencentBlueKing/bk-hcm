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

package rootaccount

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// List list root account with options
func (s *service) List(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListWithoutFieldReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验用户有一级账号管理权限
	if err := s.checkPermission(cts, meta.RootAccount, meta.Find); err != nil {
		return nil, err
	}

	listReq := &core.ListReq{Filter: req.Filter, Page: req.Page}
	accounts, err := s.client.DataService().Global.RootAccount.List(cts.Kit, listReq)
	if err != nil {
		return nil, err
	}

	return accounts, nil

}
