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
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// Authorizer defines all the supported functionalities to do auth operation.
type Authorizer interface {
	// Authorize check if a user's operate resource is already authorized or not.
	Authorize(kt *kit.Kit, opts *iam.AuthOptions) (*iam.Decision, error)

	// AuthorizeBatch check if a user's operate resources is authorized or not batch.
	// Note: being authorized resources must be the same resource.
	AuthorizeBatch(kt *kit.Kit, opts *iam.AuthBatchOptions) ([]*iam.Decision, error)

	// AuthorizeAnyBatch check if a user have any authority of the operate actions batch.
	AuthorizeAnyBatch(kt *kit.Kit, opts *iam.AuthBatchOptions) ([]*iam.Decision, error)

	// ListAuthorizedInstances list a user's all the authorized resource instance list with an action.
	// Note: opts.Resources are not required.
	// the returned list may be huge, we do not do result paging
	ListAuthorizedInstances(kt *kit.Kit, opts *iam.AuthOptions, resourceType iam.TypeID) (
		*iam.AuthorizeList, error)

	// RegisterResourceCreatorAction registers iam resource so that creator will be authorized on related actions
	RegisterResourceCreatorAction(kt *kit.Kit, opts *iam.InstanceWithCreator) ([]iam.CreatorActionPolicy, error)

	GetApplyPermUrl(kt *kit.Kit, opts *meta.IamPermission) (string, error)
}

// ResourceFetcher defines all the supported operations for iam to fetch resources from hcm
type ResourceFetcher interface {
	// ListInstancesWithAttributes get "same" resource instances with attributes
	// returned with the resource's instance id list matched with options.
	ListInstancesWithAttributes(kt *kit.Kit, opts *iam.ListWithAttributes) (idList []string, err error)
}

// NewAuth initialize an authorizer
func NewAuth(iamCli iam.Client, fetcher ResourceFetcher) (Authorizer, error) {
	if iamCli == nil {
		return nil, errf.New(errf.InvalidParameter, "iam client is nil")
	}

	return &Authorize{
		client:  iamCli,
		fetcher: fetcher,
	}, nil
}
