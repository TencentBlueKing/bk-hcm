/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package biz

import (
	mtypes "hcm/cmd/woa-server/types/meta"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// Logics provides management interface for operations of model and instance and related resources like association
type Logics interface {
	// ListAuthorizedBiz list authorized biz with biz access permission from cmdb.
	ListAuthorizedBiz(kt *kit.Kit) ([]int64, error)
	// GetBizOrgRel get biz org relation.
	GetBizOrgRel(kt *kit.Kit, bkBizID int64) (*mtypes.BizOrgRel, error)
	// ListBizsOrgRel list bizs org rel.
	ListBizsOrgRel(kt *kit.Kit, bkBizIDs []int64) (map[int64]*mtypes.BizOrgRel, error)
	// BatchCheckUserBizAccessAuth batch check user biz access auth.
	BatchCheckUserBizAccessAuth(kt *kit.Kit, bkBizID int64, userNames []string) (map[string]bool, error)
	// GetBkBizMaintainer get biz maintainer.
	GetBkBizMaintainer(kt *kit.Kit, bkBizIDs []int64) (map[int64][]string, error)
	// GetBizNames get biz names.
	GetBizNames(kt *kit.Kit, bkBizIDs []int64) (map[int64]string, error)
	// GetBkBizIDsByOpProductName get biz id by op product name.
	GetBkBizIDsByOpProductName(kt *kit.Kit, opProductNames []string) (map[string][]int64, error)
}

type logics struct {
	cmdbCli    cmdb.Client
	authorizer auth.Authorizer
}

// New create a logics manager
func New(cmdbCli cmdb.Client, authorizer auth.Authorizer) (Logics, error) {
	return &logics{
		cmdbCli:    cmdbCli,
		authorizer: authorizer,
	}, nil
}
