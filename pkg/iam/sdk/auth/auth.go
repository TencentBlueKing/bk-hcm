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
	"context"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/thirdparty/esb"
)

// Authorizer defines all the supported functionalities to do auth operation.
type Authorizer interface {
	// Authorize check if a user's operate resource is already authorized or not.
	Authorize(ctx context.Context, opts *client.AuthOptions) (*client.Decision, error)

	// AuthorizeBatch check if a user's operate resources is authorized or not batch.
	// Note: being authorized resources must be the same resource.
	AuthorizeBatch(ctx context.Context, opts *client.AuthBatchOptions) ([]*client.Decision, error)

	// AuthorizeAnyBatch check if a user have any authority of the operate actions batch.
	AuthorizeAnyBatch(ctx context.Context, opts *client.AuthBatchOptions) ([]*client.Decision, error)

	// ListAuthorizedInstances list a user's all the authorized resource instance list with an action.
	// Note: opts.Resources are not required.
	// the returned list may be huge, we do not do result paging
	ListAuthorizedInstances(ctx context.Context, opts *client.AuthOptions, resourceType client.TypeID) (
		*client.AuthorizeList, error)

	// RegisterResourceCreatorAction registers iam resource so that creator will be authorized on related actions
	RegisterResourceCreatorAction(ctx context.Context, opts *client.InstanceWithCreator) (
		[]client.CreatorActionPolicy, error)

	GetApplyPermUrl(ctx context.Context, opts *meta.IamPermission) (string, error)
}

// ResourceFetcher defines all the supported operations for iam to fetch resources from hcm
type ResourceFetcher interface {
	// ListInstancesWithAttributes get "same" resource instances with attributes
	// returned with the resource's instance id list matched with options.
	ListInstancesWithAttributes(ctx context.Context, opts *client.ListWithAttributes) (idList []string, err error)
}

// NewAuth initialize an authorizer
func NewAuth(c *client.Client, fetcher ResourceFetcher, esbClient esb.Client) (Authorizer, error) {
	if c == nil {
		return nil, errf.New(errf.InvalidParameter, "client is nil")
	}

	if fetcher == nil {
		return nil, errf.New(errf.InvalidParameter, "fetcher is nil")
	}

	if esbClient == nil {
		return nil, errf.New(errf.InvalidParameter, "esb client is nil")
	}

	return &Authorize{
		client:    c,
		fetcher:   fetcher,
		esbClient: esbClient,
	}, nil
}
