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
	asproto "hcm/pkg/api/auth-server"
	"hcm/pkg/cc"
	"hcm/pkg/client/auth-server"
	"hcm/pkg/client/discovery"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest/client"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/ssl"
)

// Authorizer defines all the supported functionalities to do auth operation.
type Authorizer interface {
	// Authorize if user has permission to the resources, returns auth status per resource and for all.
	Authorize(kt *kit.Kit, resources ...meta.ResourceAttribute) ([]meta.Decision, bool, error)
	// AuthorizeWithPerm authorize if user has permission, if not, returns unauthorized error.
	AuthorizeWithPerm(kt *kit.Kit, resources ...meta.ResourceAttribute) error
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

// AuthorizeWithPerm authorize if user has permission, if not, returns unauthorized error.
func (a authorizer) AuthorizeWithPerm(kt *kit.Kit, resources ...meta.ResourceAttribute) error {
	_, authorized, err := a.Authorize(kt, resources...)
	if err != nil {
		return errf.New(errf.DoAuthorizeFailed, "authorize failed")
	}

	if !authorized {
		req := &asproto.GetPermissionToApplyReq{
			Resources: resources,
		}

		permission, err := a.authClient.GetPermissionToApply(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("get permission to apply failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
			return errf.New(errf.DoAuthorizeFailed, "get permission to apply failed")
		}

		return errf.NewWithPerm(errf.PermissionDenied, "no permission", permission)
	}

	return nil
}
