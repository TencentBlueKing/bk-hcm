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

package iam

import (
	"fmt"
	"strconv"
	"strings"

	"hcm/pkg/cc"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sdk/operator"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/thirdparty/api-gateway/bkuser"
	"hcm/pkg/thirdparty/api-gateway/discovery"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client ...
type Client interface {
	RegisterResourceCreatorAction(kt *kit.Kit, opts *InstanceWithCreator) (*[]CreatorActionPolicy, error)
	GetApplyPermUrl(kt *kit.Kit, opts *meta.IamPermission) (string, error)

	RegisterSystem(kt *kit.Kit, sys *System) error
	GetSystemInfo(kt *kit.Kit, fields []SystemQueryField) (*SystemResp, error)
	UpdateSystemConfig(kt *kit.Kit, sys System) error
	RegisterResourcesTypes(kt *kit.Kit, resTypes []ResourceType) error
	UpdateResourcesType(kt *kit.Kit, resType ResourceType) error
	DeleteResourcesTypes(kt *kit.Kit, resTypeIDs []TypeID) error
	RegisterActions(kt *kit.Kit, actions []ResourceAction) error
	UpdateAction(kt *kit.Kit, action ResourceAction) error
	DeleteActions(kt *kit.Kit, actionIDs []ActionID) error
	RegisterActionGroups(kt *kit.Kit, actionGroups []ActionGroup) error
	UpdateActionGroups(kt *kit.Kit, actionGroups []ActionGroup) error
	RegisterInstanceSelections(kt *kit.Kit, instanceSelections []InstanceSelection) error
	UpdateInstanceSelection(kt *kit.Kit, instanceSelection InstanceSelection) error
	DeleteInstanceSelections(kt *kit.Kit, instanceSelectionIDs []InstanceSelectionID) error
	RegisterResourceCreatorActions(kt *kit.Kit, resourceCreatorActions ResourceCreatorActions) error
	UpdateResourceCreatorActions(kt *kit.Kit, resourceCreatorActions ResourceCreatorActions) error
	RegisterCommonActions(kt *kit.Kit, commonActions []CommonAction) error
	UpdateCommonActions(kt *kit.Kit, commonActions []CommonAction) error
	DeleteActionPolicies(kt *kit.Kit, actionID ActionID) error
	ListPolicies(kt *kit.Kit, params *ListPoliciesParams) (*ListPoliciesData, error)
	GetSystemToken(kt *kit.Kit) (string, error)
	GetUserPolicy(kt *kit.Kit, opt *GetPolicyOption) (*operator.Policy, error)
	ListUserPolicies(kt *kit.Kit, opts *ListPolicyOptions) ([]ActionPolicy, error)
	GetUserPolicyByExtRes(kt *kit.Kit, opts *GetPolicyByExtResOption) (*GetPolicyByExtResResult, error)
}

type iam struct {
	config    *cc.ApiGateway
	client    rest.ClientInterface
	systemID  string
	bkUserCli bkuser.Client
}

// NewClient ...
func NewClient(systemID string, cfg *cc.ApiGateway, bkUserCli bkuser.Client, reg prometheus.Registerer) (Client,
	error) {

	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &discovery.Discovery{
			Name:    "iam",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	return &iam{
		config:    cfg,
		client:    rest.NewClient(c, "/"),
		systemID:  systemID,
		bkUserCli: bkUserCli,
	}, nil
}

// RegisterResourceCreatorAction register iam resource instance with creator, returns related actions with policy id
// that the creator gained.
func (i *iam) RegisterResourceCreatorAction(kt *kit.Kit, opts *InstanceWithCreator) (*[]CreatorActionPolicy, error) {
	return apigateway.ApiGatewayCall[InstanceWithCreator, []CreatorActionPolicy](i.client, i.bkUserCli, i.config,
		rest.POST, kt, opts, "/api/v1/open/authorization/resource_creator_action/")
}

// GetApplyPermUrl get iam apply permission url.
func (i *iam) GetApplyPermUrl(kt *kit.Kit, opts *meta.IamPermission) (string, error) {
	result, err := apigateway.ApiGatewayCall[meta.IamPermission, GetApplyPermUrlResult](i.client, i.bkUserCli, i.config,
		rest.POST, kt, opts, "/api/v1/open/application/")
	if err != nil {
		return "", err
	}

	return result.Url, nil
}

// RegisterSystem register a system in IAM.
func (i *iam) RegisterSystem(kt *kit.Kit, sys *System) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Post().
		SubResourcef("/api/v1/model/systems").
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(sys).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("register system failed, code: %d, msg: %s", resp.Code, resp.Message),
		}
	}

	return nil
}

// GetSystemInfo get a system info from IAM, if fields is empty, find all system info.
func (i *iam) GetSystemInfo(kt *kit.Kit, fields []SystemQueryField) (*SystemResp, error) {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	fieldsStr := ""
	if len(fields) > 0 {
		fieldArr := make([]string, len(fields))
		for idx, field := range fields {
			fieldArr[idx] = string(field)
		}
		fieldsStr = strings.Join(fieldArr, ",")
	}

	subPath := "/api/v1/model/systems/%s/query"
	resp := new(SystemResp)
	result := i.client.Get().
		SubResourcef(subPath, i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		WithParam("fields", fieldsStr).
		Body(nil).
		Do()

	if err := result.Into(resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		if resp.Code == codeNotFound {
			return resp, ErrNotFound
		}
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp, nil
}

// UpdateSystemConfig update system config in IAM
// Note: can only update provider_config.host field.
func (i *iam) UpdateSystemConfig(kt *kit.Kit, sys System) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Put().
		SubResourcef("/api/v1/model/systems/%s", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(sys).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("update system config failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterResourcesTypes register resource types in IAM.
func (i *iam) RegisterResourcesTypes(kt *kit.Kit, resTypes []ResourceType) error {
	if len(resTypes) == 0 {
		return nil
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Post().
		SubResourcef("/api/v1/model/systems/%s/resource-types", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(resTypes).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("register system failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return nil

}

// UpdateResourcesType update resource type in IAM.
func (i *iam) UpdateResourcesType(kt *kit.Kit, resType ResourceType) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Put().
		SubResourcef("/api/v1/model/systems/%s/resource-types/%s", i.systemID, resType.ID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(resType).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("udpate resource type %s failed, code: %d, msg:%s", resType.ID, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// DeleteResourcesTypes delete resource types in IAM.
func (i *iam) DeleteResourcesTypes(kt *kit.Kit, resTypeIDs []TypeID) error {
	if len(resTypeIDs) == 0 {
		return nil
	}

	ids := make([]struct {
		ID TypeID `json:"id"`
	}, len(resTypeIDs))
	for idx := range resTypeIDs {
		ids[idx].ID = resTypeIDs[idx]
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/resource-types", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(ids).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("delete resource type %v failed, code: %d, msg:%s", resTypeIDs, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// RegisterActions register actions in IAM.
func (i *iam) RegisterActions(kt *kit.Kit, actions []ResourceAction) error {
	if len(actions) == 0 {
		return nil
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Post().
		SubResourcef("/api/v1/model/systems/%s/actions", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(actions).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("add resource actions %v failed, code: %d, msg:%s", actions, resp.Code, resp.Message),
		}
	}

	return nil
}

// UpdateAction update action in IAM.
func (i *iam) UpdateAction(kt *kit.Kit, action ResourceAction) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Put().
		SubResourcef("/api/v1/model/systems/%s/actions/%s", i.systemID, action.ID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(action).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("udpate resource action %v failed, code: %d, msg:%s", action, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// DeleteActions delete actions in IAM.
func (i *iam) DeleteActions(kt *kit.Kit, actionIDs []ActionID) error {
	ids := make([]struct {
		ID ActionID `json:"id"`
	}, len(actionIDs))
	for idx := range actionIDs {
		ids[idx].ID = actionIDs[idx]
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/actions", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(ids).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("delete resource actions %v failed, code: %d, msg:%s", actionIDs, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// RegisterActionGroups register action groups in IAM.
func (i *iam) RegisterActionGroups(kt *kit.Kit, actionGroups []ActionGroup) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Post().
		SubResourcef("/api/v1/model/systems/%s/configs/action_groups", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(actionGroups).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("register action groups %v failed, code: %d, msg:%s", actionGroups, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// UpdateActionGroups update action groups in IAM.
func (i *iam) UpdateActionGroups(kt *kit.Kit, actionGroups []ActionGroup) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Put().
		SubResourcef("/api/v1/model/systems/%s/configs/action_groups", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(actionGroups).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("update action groups %v failed, code: %d, msg:%s", actionGroups, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// RegisterInstanceSelections register instance selections.
func (i *iam) RegisterInstanceSelections(kt *kit.Kit, instanceSelections []InstanceSelection) error {
	if len(instanceSelections) == 0 {
		return nil
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Post().
		SubResourcef("/api/v1/model/systems/%s/instance-selections", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(instanceSelections).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("add instance selections %v failed, code: %d, msg:%s", instanceSelections,
				resp.Code, resp.Message),
		}
	}

	return nil
}

// UpdateInstanceSelection update instance selection in IAM.
func (i *iam) UpdateInstanceSelection(kt *kit.Kit, instanceSelection InstanceSelection) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Put().
		SubResourcef("/api/v1/model/systems/%s/instance-selections/%s", i.systemID, instanceSelection.ID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(instanceSelection).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("udpate instance selections %v failed, code: %d, msg:%s", instanceSelection,
				resp.Code, resp.Message),
		}
	}

	return nil
}

// DeleteInstanceSelections delete instance selections in IAM.
func (i *iam) DeleteInstanceSelections(kt *kit.Kit, instanceSelectionIDs []InstanceSelectionID) error {
	if len(instanceSelectionIDs) == 0 {
		return nil
	}

	ids := make([]struct {
		ID InstanceSelectionID `json:"id"`
	}, len(instanceSelectionIDs))
	for idx := range instanceSelectionIDs {
		ids[idx].ID = instanceSelectionIDs[idx]
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/instance-selections", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(ids).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("delete instance selections %v failed, code: %d, msg:%s", instanceSelectionIDs,
				resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterResourceCreatorActions regitser resource creator actions in IAM.
func (i *iam) RegisterResourceCreatorActions(kt *kit.Kit, resourceCreatorActions ResourceCreatorActions) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Post().
		SubResourcef("/api/v1/model/systems/%s/configs/resource_creator_actions", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(resourceCreatorActions).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("register resource creator actions %v failed, code: %d, msg:%s",
				resourceCreatorActions, resp.Code, resp.Message),
		}
	}

	return nil
}

// UpdateResourceCreatorActions update resource creator actions in IAM.
func (i *iam) UpdateResourceCreatorActions(kt *kit.Kit, resourceCreatorActions ResourceCreatorActions) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Put().
		SubResourcef("/api/v1/model/systems/%s/configs/resource_creator_actions", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(resourceCreatorActions).Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("update resource creator actions %v failed, code: %d, msg:%s",
				resourceCreatorActions, resp.Code, resp.Message),
		}
	}

	return nil
}

// RegisterCommonActions register common actions in IAM.
func (i *iam) RegisterCommonActions(kt *kit.Kit, commonActions []CommonAction) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Post().
		SubResourcef("/api/v1/model/systems/%s/configs/common_actions", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(commonActions).Do()

	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("register common actions %v failed, code: %d, msg: %s", commonActions, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// UpdateCommonActions update common actions in IAM.
func (i *iam) UpdateCommonActions(kt *kit.Kit, commonActions []CommonAction) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Put().
		SubResourcef("/api/v1/model/systems/%s/configs/common_actions", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(commonActions).Do()

	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("update common actions %v failed, code: %d, msg: %s", commonActions, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// DeleteActionPolicies delete action policies in IAM.
func (i *iam) DeleteActionPolicies(kt *kit.Kit, actionID ActionID) error {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(BaseResponse)
	result := i.client.Delete().
		SubResourcef("/api/v1/model/systems/%s/actions/%s/policies", i.systemID, actionID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Do()
	err := result.Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason: fmt.Errorf("delete action %s policies failed, code: %d, msg: %s", actionID, resp.Code,
				resp.Message),
		}
	}

	return nil
}

// ListPolicies list iam policies.
func (i *iam) ListPolicies(kt *kit.Kit, params *ListPoliciesParams) (*ListPoliciesData, error) {
	parsedParams := map[string]string{"action_id": string(params.ActionID)}
	if params.Page != 0 {
		parsedParams["page"] = strconv.FormatInt(params.Page, 10)
	}
	if params.PageSize != 0 {
		parsedParams["page_size"] = strconv.FormatInt(params.PageSize, 10)
	}
	if params.Timestamp != 0 {
		parsedParams["timestamp"] = strconv.FormatInt(params.Timestamp, 10)
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(ListPoliciesResp)
	result := i.client.Get().
		SubResourcef("/api/v1/systems/%s/policies", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		WithParams(parsedParams).
		Body(nil).Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}

// GetSystemToken get system token from iam, used to validate if request is from iam.
func (i *iam) GetSystemToken(kt *kit.Kit) (string, error) {
	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	resp := new(struct {
		BaseResponse
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	})
	result := i.client.Get().
		SubResourcef("/api/v1/model/systems/%s/token", i.systemID).
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(nil).Do()
	err := result.Into(resp)
	if err != nil {
		return "", err
	}

	if resp.Code != 0 {
		return "", &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp.Data.Token, nil
}

// GetUserPolicy get a user's policy with a action and resources
func (i *iam) GetUserPolicy(kt *kit.Kit, opt *GetPolicyOption) (*operator.Policy, error) {
	resp := new(GetPolicyResp)

	// iam requires resources to be set
	if opt.Resources == nil {
		opt.Resources = make([]Resource, 0)
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	result := i.client.Post().
		SubResourcef("/api/v1/policy/query").
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(opt).
		Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get system info failed, code: %d, msg:%s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}

// ListUserPolicies get a user's policy with multiple actions and resources.
func (i *iam) ListUserPolicies(kt *kit.Kit, opts *ListPolicyOptions) ([]ActionPolicy, error) {

	resp := new(ListPolicyResp)

	// iam requires resources to be set
	if opts.Resources == nil {
		opts.Resources = make([]Resource, 0)
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	result := i.client.Post().
		SubResourcef("/api/v1/policy/query_by_actions").
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(opts).
		Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("list user policies failed, code: %d, msg: %s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}

// GetUserPolicyByExtRes get a user's policy by external resource.
func (i *iam) GetUserPolicyByExtRes(kt *kit.Kit, opts *GetPolicyByExtResOption) (*GetPolicyByExtResResult, error) {
	resp := new(GetPolicyByExtResResp)

	// iam requires resources to be set
	if opts.Resources == nil {
		opts.Resources = make([]Resource, 0)
	}

	header := apigateway.GetCommonHeader(kt, i.bkUserCli, i.config)
	result := i.client.Post().
		SubResourcef("/api/v1/policy/query_by_ext_resources").
		WithContext(kt.Ctx).
		WithHeaders(header).
		Body(opts).
		Do()

	err := result.Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, &AuthError{
			RequestID: result.Header.Get(RequestIDHeader),
			Reason:    fmt.Errorf("get policy by external resource failed, code: %d, msg: %s", resp.Code, resp.Message),
		}
	}

	return resp.Data, nil
}
