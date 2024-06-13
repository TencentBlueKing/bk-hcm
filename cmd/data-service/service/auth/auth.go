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

// Package auth ...
package auth

import (
	"fmt"

	"hcm/cmd/data-service/service/capability"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitAuthService initial the auth used service
func InitAuthService(cap *capability.Capability) {
	svr := &auth{
		dao: cap.Dao,
	}

	h := rest.NewHandler()
	h.Add("ListAuthInstances", "POST", "/list/auth/instances", svr.ListAuthInstances)
	h.Load(cap.WebService)
}

type auth struct {
	dao dao.Set
}

// ListAuthInstances list instances that are used for iam pull resource callback.
func (s *auth) ListAuthInstances(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.ListInstancesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	var tableName table.Name
	switch req.ResourceType {
	case sys.Account:
		tableName = table.AccountTable
	case sys.CloudSelectionScheme:
		tableName = table.CloudSelectionSchemeTable
	case sys.MainAccount:
		tableName = table.MainAccountTable
	default:
		return nil, fmt.Errorf("resource type %s not support", req.ResourceType)
	}

	opts := &types.ListInstancesOption{
		TableName: tableName,
		Filter:    req.Filter,
		Page:      req.Page,
	}

	details, err := s.dao.Auth().ListInstances(cts.Kit, opts)
	if err != nil {
		logs.Errorf("list instances failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	result := &dataservice.ListInstancesResult{
		Count:   details.Count,
		Details: details.Details,
	}

	return result, nil
}
