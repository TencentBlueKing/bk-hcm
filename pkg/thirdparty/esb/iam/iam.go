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

// Package iam defines esb client to request iam.
package iam

import (
	"context"

	"hcm/pkg/cc"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/types"
	"hcm/pkg/tools/converter"
)

// Client is an esb client to request iam.
type Client interface {
	RegisterResourceCreatorAction(ctx context.Context, inst *client.InstanceWithCreator) ([]client.CreatorActionPolicy,
		error)
	GetApplyPermUrl(ctx context.Context, opts *meta.IamPermission) (string, error)
}

// NewClient initialize a new iam client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &iam{
		client: client,
		config: config,
	}
}

// iam is an esb client to request iam.
type iam struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

// RegisterResourceCreatorAction register iam resource instance with creator
// returns related actions with policy id that the creator gained
func (i *iam) RegisterResourceCreatorAction(ctx context.Context, inst *client.InstanceWithCreator) (
	[]client.CreatorActionPolicy, error) {

	kt := kit.New()
	kt.Ctx = ctx
	result, err := types.EsbCall[client.InstanceWithCreator, []client.CreatorActionPolicy](i.client, i.config,
		rest.POST, kt,
		inst, "/iam/authorization/resource_creator_action/")
	if err != nil {
		logs.Errorf("fail to register iam resource instance, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return converter.PtrToVal(result), nil
}

// GetApplyPermUrl get iam apply permission url.
func (i *iam) GetApplyPermUrl(ctx context.Context, opts *meta.IamPermission) (string, error) {

	kt := kit.New()
	kt.Ctx = ctx
	result, err := types.EsbCall[meta.IamPermission, GetApplyPermUrlResult](i.client, i.config, rest.POST, kt, opts,
		"/iam/application/")
	if err != nil {
		logs.Errorf("fail to get iam apply permission url, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return result.Url, nil
}
