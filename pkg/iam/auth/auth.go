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
	asproto "hcm/pkg/api/auth-server"
	"hcm/pkg/cc"
	authserver "hcm/pkg/client/auth-server"
	"hcm/pkg/client/discovery"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest/client"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"
)

// Authorizer defines all the supported functionalities to do auth operation.
type Authorizer interface {
	// Authorize if user has permission to the resources, returns auth status per resource and for all.
	Authorize(kt *kit.Kit, resources ...meta.ResourceAttribute) ([]meta.Decision, bool, error)
	// AuthorizeAny authorize if user has any permission to the resources, returns auth status per resource and for all.
	AuthorizeAny(kt *kit.Kit, resources ...meta.ResourceAttribute) ([]meta.Decision, error)
	// AuthorizeWithPerm authorize if user has permission, if not, returns unauthorized error.
	AuthorizeWithPerm(kt *kit.Kit, resources ...meta.ResourceAttribute) error
	// ListAuthorizedInstances list authorized instances info.
	ListAuthorizedInstances(kt *kit.Kit, input *meta.ListAuthResInput) (*meta.AuthorizedInstances, error)
	// ListAuthInstWithFilter returns resource filter with authorized instances info & if user has no permission flag.
	ListAuthInstWithFilter(kt *kit.Kit, input *meta.ListAuthResInput, expr *filter.Expression, resIDField string) (
		*filter.Expression, bool, error)
	// RegisterResourceCreatorAction registers iam resource so that creator will be authorized on related actions
	RegisterResourceCreatorAction(kt *kit.Kit, input *meta.RegisterResCreatorActionInst) error
	// GetPermissionToApply get permissions to apply.
	GetPermissionToApply(kt *kit.Kit, res ...meta.ResourceAttribute) (*meta.IamPermission, error)
	// GetApplyPermUrl get iam apply permission url.
	GetApplyPermUrl(kt *kit.Kit, input *meta.IamPermission) (string, error)
}

// NewAuthorizer create an authorizer for iam authorize related operation.
func NewAuthorizer(sd serviced.Discover, tls cc.TLSConfig) (Authorizer, error) {
	var tlsC *ssl.TLSConfig
	if tls.Enable() {
		tlsC = &ssl.TLSConfig{
			InsecureSkipVerify: tls.InsecureSkipVerify,
			CertFile:           tls.CertFile,
			KeyFile:            tls.KeyFile,
			CAFile:             tls.CAFile,
			Password:           tls.Password,
		}
	}

	// initiate auth server api client set.
	cli, err := client.NewClient(tlsC)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client:   cli,
		Discover: discovery.NewAPIDiscovery(cc.AuthServerName, sd),
	}
	authClient := authserver.NewClient(c, "v1")

	return &authorizer{
		authClient: authClient,
	}, nil
}

type authorizer struct {
	// authClient auth server's client api
	authClient *authserver.Client
}

// Authorize if user has permission to the resources, returns auth status per resource and for all.
func (a authorizer) Authorize(kt *kit.Kit, resources ...meta.ResourceAttribute) ([]meta.Decision, bool, error) {
	userInfo := &meta.UserInfo{UserName: kt.User}

	req := &asproto.AuthorizeBatchReq{
		User:      userInfo,
		Resources: resources,
	}

	decisions, err := a.authClient.AuthorizeBatch(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("authorize failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, false, err
	}

	authorized := true
	for _, decision := range decisions {
		if !decision.Authorized {
			authorized = false
			break
		}
	}

	return decisions, authorized, nil
}

// AuthorizeAny if user has any permission to the resources, returns auth status per resource and for all.
func (a authorizer) AuthorizeAny(kt *kit.Kit, resources ...meta.ResourceAttribute) ([]meta.Decision, error) {
	userInfo := &meta.UserInfo{UserName: kt.User}

	req := &asproto.AuthorizeBatchReq{
		User:      userInfo,
		Resources: resources,
	}

	decisions, err := a.authClient.AuthorizeAnyBatch(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("authorize any failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	return decisions, nil
}

// AuthorizeWithPerm authorize if user has permission, if not, returns unauthorized error.
func (a authorizer) AuthorizeWithPerm(kt *kit.Kit, resources ...meta.ResourceAttribute) error {
	_, authorized, err := a.Authorize(kt, resources...)
	if err != nil {
		return errf.New(errf.DoAuthorizeFailed, "authorize failed")
	}

	if !authorized {
		permission, err := a.GetPermissionToApply(kt, resources...)
		if err != nil {
			logs.Errorf("get permission to apply failed, resources: %#v, err: %v, rid: %s", resources, err, kt.Rid)
			return errf.New(errf.DoAuthorizeFailed, "get permission to apply failed")
		}

		return errf.NewWithPerm(errf.PermissionDenied, "no permission", permission)
	}

	return nil
}

// ListAuthorizedInstances list authorized instances info.
func (a authorizer) ListAuthorizedInstances(kt *kit.Kit, input *meta.ListAuthResInput) (*meta.AuthorizedInstances,
	error) {

	if input == nil || len(input.Action) == 0 || len(input.Type) == 0 {
		return nil, errf.New(errf.InvalidParameter, "list authorized instances input is invalid")
	}

	userInfo := &meta.UserInfo{UserName: kt.User}

	req := &asproto.ListAuthorizedInstancesReq{
		User:   userInfo,
		Type:   input.Type,
		Action: input.Action,
	}

	resources, err := a.authClient.ListAuthorizedInstances(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("list authorized instances failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	return resources, nil
}

// ListAuthInstWithFilter returns resource filter with authorized instances info & if user has no permission flag.
func (a authorizer) ListAuthInstWithFilter(kt *kit.Kit, input *meta.ListAuthResInput, expr *filter.Expression,
	resIDField string) (*filter.Expression, bool, error) {

	authInst, err := a.ListAuthorizedInstances(kt, input)
	if err != nil {
		return nil, false, err
	}

	if authInst.IsAny {
		return expr, false, err
	}

	if len(authInst.IDs) == 0 {
		return nil, true, nil
	}

	authRule := filter.AtomRule{Field: resIDField, Op: filter.In.Factory(), Value: authInst.IDs}

	if expr == nil {
		return &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{authRule},
		}, false, nil
	}

	filterExpr, err := tools.And(authRule, expr)
	if err != nil {
		return nil, false, err
	}
	return filterExpr, false, nil
}

// RegisterResourceCreatorAction registers iam resource instance so that creator will be authorized on related actions
func (a authorizer) RegisterResourceCreatorAction(kt *kit.Kit, input *meta.RegisterResCreatorActionInst) error {
	if input == nil || len(input.Type) == 0 || len(input.ID) == 0 || len(input.Name) == 0 {
		return errf.New(errf.InvalidParameter, "register resource creator action input is invalid")
	}

	req := &asproto.RegisterResourceCreatorActionReq{
		Creator:  kt.User,
		Instance: input,
	}

	_, err := a.authClient.RegisterResourceCreatorAction(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("register resource creator action failed, err: %v, req: %#v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}

// GetApplyPermUrl get iam apply permission url.
func (a authorizer) GetApplyPermUrl(kt *kit.Kit, input *meta.IamPermission) (string, error) {
	if input == nil {
		return "", errf.New(errf.InvalidParameter, "get iam apply permission url input is nil")
	}

	url, err := a.authClient.GetApplyPermUrl(kt.Ctx, kt.Header(), input)
	if err != nil {
		logs.Errorf("get iam apply permission url failed, err: %v, input: %#v, rid: %s", err, input, kt.Rid)
		return "", err
	}

	return url, nil
}

// GetPermissionToApply get permissions to apply.
func (a authorizer) GetPermissionToApply(kt *kit.Kit, res ...meta.ResourceAttribute) (*meta.IamPermission, error) {
	req := &asproto.GetPermissionToApplyReq{
		Resources: res,
	}

	permission, err := a.authClient.GetPermissionToApply(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("get permission to apply failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, errf.New(errf.DoAuthorizeFailed, "get permission to apply failed")
	}

	return permission, nil

}
