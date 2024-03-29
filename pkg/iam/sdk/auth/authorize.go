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

package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sdk/operator"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb"
)

// Authorize is the instance for the Authorizer factory.
type Authorize struct {
	// iam client.
	client *client.Client
	// fetch resource if needed
	fetcher ResourceFetcher
	// esb client.
	esbClient esb.Client
}

// Authorize check if a user's operate resource is already authorized or not.
func (a *Authorize) Authorize(ctx context.Context, opts *client.AuthOptions) (*client.Decision, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// find user's policy with action
	getOpt := client.GetPolicyOption{
		System:  opts.System,
		Subject: opts.Subject,
		Action:  opts.Action,
		// do not use user's policy, so that we can get all the user's policy.
		Resources: make([]client.Resource, 0),
	}

	policy, err := a.client.GetUserPolicy(ctx, &getOpt)
	if err != nil {
		return nil, err
	}

	authorized, err := a.calculatePolicy(ctx, opts.Resources, policy)
	if err != nil {
		return nil, fmt.Errorf("calculate user's auth policy failed, err: %v", err)
	}

	return &client.Decision{Authorized: authorized}, nil
}

// AuthorizeBatch check if a user's operate resources is authorized or not batch.
// Note: being authorized resources must be the same resource.
func (a *Authorize) AuthorizeBatch(ctx context.Context, opts *client.AuthBatchOptions) ([]*client.Decision, error) {
	return a.authorizeBatch(ctx, opts, true)
}

// AuthorizeAnyBatch check if a user have any authority of the operate actions batch.
func (a *Authorize) AuthorizeAnyBatch(ctx context.Context, opts *client.AuthBatchOptions) ([]*client.Decision, error) {
	return a.authorizeBatch(ctx, opts, false)
}

func (a *Authorize) authorizeBatch(ctx context.Context, opts *client.AuthBatchOptions, exact bool) (
	[]*client.Decision, error) {

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if len(opts.Batch) == 0 {
		return nil, errors.New("no resource instance need to authorize")
	}

	policies, err := a.listUserPolicyBatchWithCompress(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list user policy failed, err: %v", err)
	}

	var hitError error
	decisions := make([]*client.Decision, len(opts.Batch))

	pipe := make(chan struct{}, 50)
	wg := sync.WaitGroup{}
	for idx, b := range opts.Batch {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, resources []client.Resource, policy *operator.Policy) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			var authorized bool
			var err error
			if exact {
				authorized, err = a.calculatePolicy(ctx, resources, policy)
			} else {
				authorized, err = a.calculateAnyPolicy(ctx, resources, policy)
			}
			if err != nil {
				hitError = err
				return
			}

			// save the result with index
			decisions[idx] = &client.Decision{Authorized: authorized}
		}(idx, b.Resources, policies[idx])
	}
	// wait all the policy are calculated
	wg.Wait()

	if hitError != nil {
		return nil, fmt.Errorf("batch calculate policy failed, err: %v", hitError)
	}

	return decisions, nil
}

func (a *Authorize) listUserPolicyBatchWithCompress(ctx context.Context,
	opts *client.AuthBatchOptions) ([]*operator.Policy, error) {

	if len(opts.Batch) == 0 {
		return make([]*operator.Policy, 0), nil
	}

	// because hcm actions are attached to cc biz resource, we need to get policy for each action separately
	// group external resources by action, so that we can cut off the request to iam, and improve the performance.
	actionIDMap := make(map[string]client.Action)
	actionResMap := make(map[string]map[string]client.Resource)
	for _, b := range opts.Batch {
		actionIDMap[b.Action.ID] = b.Action
		if _, exists := actionResMap[b.Action.ID]; !exists {
			actionResMap[b.Action.ID] = make(map[string]client.Resource)
		}

		for _, resource := range b.Resources {
			if resource.System != sys.SystemIDHCM {
				actionResMap[b.Action.ID][resource.ID] = resource
			}
		}
	}

	// query user policy by actions
	policyMap := make(map[string]*operator.Policy)
	for actionID, action := range actionIDMap {
		// parse resources in action to client.ExtResource form
		extResources := make([]client.ExtResource, 0)
		resMap := make(map[string]map[client.TypeID][]string)
		for _, resource := range actionResMap[actionID] {
			if _, exists := resMap[resource.System]; !exists {
				resMap[resource.System] = make(map[client.TypeID][]string)
			}
			resMap[resource.System][resource.Type] = append(resMap[resource.System][resource.Type], resource.ID)
		}

		for system, resTypeMap := range resMap {
			for resType, ids := range resTypeMap {
				extResources = append(extResources, client.ExtResource{
					System: system,
					Type:   resType,
					IDs:    ids,
				})
			}
		}

		// get action policy by external resources
		getOpts := &client.GetPolicyByExtResOption{
			AuthOptions: client.AuthOptions{
				System:  opts.System,
				Subject: opts.Subject,
				Action:  action,
			},
			ExtResources: extResources,
		}

		policyRes, err := a.client.GetUserPolicyByExtRes(ctx, getOpts)
		if err != nil {
			return nil, fmt.Errorf("get user policy failed, opts: %#v, err: %v", getOpts, err)
		}

		policyMap[actionID] = policyRes.Expression
	}

	allPolicies := make([]*operator.Policy, len(opts.Batch))
	for idx, b := range opts.Batch {
		policy, exist := policyMap[b.Action.ID]
		if !exist {
			// when user has no permission to this action, iam would return an empty policy
			if logs.V(2) {
				rid := ctx.Value(constant.RidKey)
				logs.Infof("list user's policy, but can not find action id %s in response, rid: %s", b.Action.ID, rid)
			}
			continue
		}
		allPolicies[idx] = policy
	}

	return allPolicies, nil
}

// ListAuthorizedInstances list a user's all the authorized resource instance list with an action.
// Note: opts.Resources are not required.
// the returned list may be huge, we do not do result paging
func (a *Authorize) ListAuthorizedInstances(ctx context.Context, opts *client.AuthOptions,
	resourceType client.TypeID) (*client.AuthorizeList, error) {

	// find user's policy with action
	getOpts := &client.GetPolicyByExtResOption{
		AuthOptions: client.AuthOptions{
			System:  opts.System,
			Subject: opts.Subject,
			Action:  opts.Action,
		},
		ExtResources: make([]client.ExtResource, 0),
	}

	policyRes, err := a.client.GetUserPolicyByExtRes(ctx, getOpts)
	if err != nil {
		return nil, fmt.Errorf("get user policy failed, opts: %#v, err: %v", getOpts, err)
	}

	policy := policyRes.Expression

	if policy == nil || policy.Operator == "" {
		return &client.AuthorizeList{}, nil
	}
	return a.countPolicy(ctx, policy, resourceType)
}

// RegisterResourceCreatorAction registers iam resource instance so that creator will be authorized on related actions
func (a *Authorize) RegisterResourceCreatorAction(ctx context.Context, opts *client.InstanceWithCreator) (
	[]client.CreatorActionPolicy, error) {

	policies, err := a.esbClient.Iam().RegisterResourceCreatorAction(ctx, opts)
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// GetApplyPermUrl get iam apply permission url.
func (a *Authorize) GetApplyPermUrl(ctx context.Context, opts *meta.IamPermission) (string, error) {
	url, err := a.esbClient.Iam().GetApplyPermUrl(ctx, opts)
	if err != nil {
		return "", err
	}

	return url, nil
}
