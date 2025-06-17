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
	"errors"
	"fmt"
	"sync"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sdk/operator"
	"hcm/pkg/iam/sys"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/iam"
	"hcm/pkg/tools/converter"
)

// Authorize is the instance for the Authorizer factory.
type Authorize struct {
	// iam client.
	client iam.Client
	// fetch resource if needed
	fetcher ResourceFetcher
}

// Authorize check if a user's operate resource is already authorized or not.
func (a *Authorize) Authorize(kt *kit.Kit, opts *iam.AuthOptions) (*iam.Decision, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// find user's policy with action
	getOpt := iam.GetPolicyOption{
		System:  opts.System,
		Subject: opts.Subject,
		Action:  opts.Action,
		// do not use user's policy, so that we can get all the user's policy.
		Resources: make([]iam.Resource, 0),
	}

	policy, err := a.client.GetUserPolicy(kt, &getOpt)
	if err != nil {
		return nil, err
	}

	authorized, err := a.calculatePolicy(kt, opts.Resources, policy)
	if err != nil {
		return nil, fmt.Errorf("calculate user's auth policy failed, err: %v", err)
	}

	return &iam.Decision{Authorized: authorized}, nil
}

// AuthorizeBatch check if a user's operate resources is authorized or not batch.
// Note: being authorized resources must be the same resource.
func (a *Authorize) AuthorizeBatch(kt *kit.Kit, opts *iam.AuthBatchOptions) ([]*iam.Decision, error) {
	return a.authorizeBatch(kt, opts, true)
}

// AuthorizeAnyBatch check if a user have any authority of the operate actions batch.
func (a *Authorize) AuthorizeAnyBatch(kt *kit.Kit, opts *iam.AuthBatchOptions) ([]*iam.Decision, error) {
	return a.authorizeBatch(kt, opts, false)
}

func (a *Authorize) authorizeBatch(kt *kit.Kit, opts *iam.AuthBatchOptions, exact bool) (
	[]*iam.Decision, error) {

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if len(opts.Batch) == 0 {
		return nil, errors.New("no resource instance need to authorize")
	}

	policies, err := a.listUserPolicyBatchWithCompress(kt, opts)
	if err != nil {
		return nil, fmt.Errorf("list user policy failed, err: %v", err)
	}

	var hitError error
	decisions := make([]*iam.Decision, len(opts.Batch))

	pipe := make(chan struct{}, 50)
	wg := sync.WaitGroup{}
	for idx, b := range opts.Batch {
		wg.Add(1)

		pipe <- struct{}{}
		go func(idx int, resources []iam.Resource, policy *operator.Policy) {
			defer func() {
				wg.Done()
				<-pipe
			}()

			var authorized bool
			var err error
			if exact {
				authorized, err = a.calculatePolicy(kt, resources, policy)
			} else {
				authorized, err = a.calculateAnyPolicy(kt, resources, policy)
			}
			if err != nil {
				hitError = err
				return
			}

			// save the result with index
			decisions[idx] = &iam.Decision{Authorized: authorized}
		}(idx, b.Resources, policies[idx])
	}
	// wait all the policy are calculated
	wg.Wait()

	if hitError != nil {
		return nil, fmt.Errorf("batch calculate policy failed, err: %v", hitError)
	}

	return decisions, nil
}

func (a *Authorize) listUserPolicyBatchWithCompress(kt *kit.Kit,
	opts *iam.AuthBatchOptions) ([]*operator.Policy, error) {

	if len(opts.Batch) == 0 {
		return make([]*operator.Policy, 0), nil
	}

	// because hcm actions are attached to cc biz resource, we need to get policy for each action separately
	// group external resources by action, so that we can cut off the request to iam, and improve the performance.
	actionIDMap := make(map[string]iam.Action)
	actionResMap := make(map[string]map[string]iam.Resource)
	for _, b := range opts.Batch {
		actionIDMap[b.Action.ID] = b.Action
		if _, exists := actionResMap[b.Action.ID]; !exists {
			actionResMap[b.Action.ID] = make(map[string]iam.Resource)
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
		// parse resources in action to iam.ExtResource form
		extResources := make([]iam.ExtResource, 0)
		resMap := make(map[string]map[iam.TypeID][]string)
		for _, resource := range actionResMap[actionID] {
			if _, exists := resMap[resource.System]; !exists {
				resMap[resource.System] = make(map[iam.TypeID][]string)
			}
			resMap[resource.System][resource.Type] = append(resMap[resource.System][resource.Type], resource.ID)
		}

		for system, resTypeMap := range resMap {
			for resType, ids := range resTypeMap {
				extResources = append(extResources, iam.ExtResource{
					System: system,
					Type:   resType,
					IDs:    ids,
				})
			}
		}

		// get action policy by external resources
		getOpts := &iam.GetPolicyByExtResOption{
			AuthOptions: iam.AuthOptions{
				System:  opts.System,
				Subject: opts.Subject,
				Action:  action,
			},
			ExtResources: extResources,
		}

		policyRes, err := a.client.GetUserPolicyByExtRes(kt, getOpts)
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
				rid := kt.Ctx.Value(constant.RidKey)
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
func (a *Authorize) ListAuthorizedInstances(kt *kit.Kit, opts *iam.AuthOptions,
	resourceType iam.TypeID) (*iam.AuthorizeList, error) {

	// find user's policy with action
	getOpts := &iam.GetPolicyByExtResOption{
		AuthOptions: iam.AuthOptions{
			System:  opts.System,
			Subject: opts.Subject,
			Action:  opts.Action,
		},
		ExtResources: make([]iam.ExtResource, 0),
	}

	policyRes, err := a.client.GetUserPolicyByExtRes(kt, getOpts)
	if err != nil {
		return nil, fmt.Errorf("get user policy failed, opts: %#v, err: %v", getOpts, err)
	}

	policy := policyRes.Expression

	if policy == nil || policy.Operator == "" {
		return &iam.AuthorizeList{}, nil
	}
	return a.countPolicy(kt, policy, resourceType)
}

// RegisterResourceCreatorAction registers iam resource instance so that creator will be authorized on related actions
func (a *Authorize) RegisterResourceCreatorAction(kt *kit.Kit, opts *iam.InstanceWithCreator) (
	[]iam.CreatorActionPolicy, error) {

	policies, err := a.client.RegisterResourceCreatorAction(kt, opts)
	if err != nil {
		return nil, err
	}

	return converter.PtrToVal(policies), nil
}

// GetApplyPermUrl get iam apply permission url.
func (a *Authorize) GetApplyPermUrl(kt *kit.Kit, opts *meta.IamPermission) (string, error) {
	url, err := a.client.GetApplyPermUrl(kt, opts)
	if err != nil {
		return "", err
	}

	return url, nil
}
