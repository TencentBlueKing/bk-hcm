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

// Package handler ...
package handler

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ValidWithAuthHandler 资源除查询外操作合法性校验
type ValidWithAuthHandler func(cts *rest.Contexts, opt *ValidWithAuthOption) error

// ValidWithAuthOption authorize cloud resource options.
type ValidWithAuthOption struct {
	Authorizer        auth.Authorizer
	ResType           meta.ResourceType
	Action            meta.Action
	BasicInfo         *types.CloudResourceBasicInfo
	BasicInfos        map[string]types.CloudResourceBasicInfo
	DisableBizIDEqual bool
}

// Validate ValidWithAuthOption
func (opt *ValidWithAuthOption) Validate() error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "validate with auth option must be set")
	}

	if opt.Authorizer == nil {
		return errf.New(errf.InvalidParameter, "authorizer must be set")
	}

	if len(opt.ResType) == 0 {
		return errf.New(errf.InvalidParameter, "authorize resource type must be set")
	}

	if len(opt.Action) == 0 {
		return errf.New(errf.InvalidParameter, "authorize action must be set")
	}

	if opt.BasicInfo == nil && len(opt.BasicInfos) == 0 {
		return errf.New(errf.InvalidParameter, "one of resource basic info and resource basic infos must be set")
	}

	if opt.BasicInfo != nil && len(opt.BasicInfos) != 0 {
		return errf.New(errf.InvalidParameter, "only one of resource basic info and resource basic infos can be set")
	}

	return nil
}

// ListAuthResHandler 资源查询操作合法性校验
type ListAuthResHandler func(cts *rest.Contexts, opt *ListAuthResOption) (
	filterExp *filter.Expression, noPerm bool, err error)

// ListAuthResOption list authorized cloud resource options.
type ListAuthResOption struct {
	Authorizer auth.Authorizer
	ResType    meta.ResourceType
	Action     meta.Action
	Filter     *filter.Expression
}

// Validate ListAuthResOption
func (opt *ListAuthResOption) Validate() error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "validate with auth option must be set")
	}

	if opt.Authorizer == nil {
		return errf.New(errf.InvalidParameter, "authorizer must be set")
	}

	if len(opt.ResType) == 0 {
		return errf.New(errf.InvalidParameter, "authorize resource type must be set")
	}

	if len(opt.Action) == 0 {
		return errf.New(errf.InvalidParameter, "authorize action must be set")
	}

	return nil
}

// ListAuthManagerHandler 系统管理员查询操作合法性校验
type ListAuthManagerHandler func (cts *rest.Contexts, opt *ListAuthManagerOption) (noPerm bool, err error)

// ListAuthManagerOption list authorized manager operations options
type ListAuthManagerOption struct {
	Authorizer auth.Authorizer
	ResType    meta.ResourceType
	Action     meta.Action
}

// Validate ListAuthManagerOption
func (opt *ListAuthManagerOption) Validate() error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "validate with auth option must be set")
	}

	if opt.Authorizer == nil {
		return errf.New(errf.InvalidParameter, "authorizer must be set")
	}

	if len(opt.ResType) == 0 {
		return errf.New(errf.InvalidParameter, "authorize resource type must be set")
	}

	if len(opt.Action) == 0 {
		return errf.New(errf.InvalidParameter, "authorize action must be set")
	}

	return nil
}
