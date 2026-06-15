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

package tcloud

import (
	"context"
	"net/http"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core/cloud"
	hsaccount "hcm/pkg/api/hc-service/account"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// AccountClient is hc service account api client.
type AccountClient struct {
	client rest.ClientInterface
}

// NewAccountClient create a new account api client.
func NewAccountClient(client rest.ClientInterface) *AccountClient {
	return &AccountClient{
		client: client,
	}
}

// SyncSubAccount sync sub account
func (a *AccountClient) SyncSubAccount(kt *kit.Kit, req *sync.TCloudGlobalSyncReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/sub_accounts/sync").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// SyncPermissionTemplate sync permission templates for the given account.
func (a *AccountClient) SyncPermissionTemplate(kt *kit.Kit, req *sync.TCloudGlobalSyncReq) error {
	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/permission_templates/sync").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// Check 联通性和云上字段匹配校验
func (a *AccountClient) Check(ctx context.Context, h http.Header, request *hsaccount.TCloudAccountCheckReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/check").
		WithHeaders(h).
		Do().
		Into(resp)

	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// GetZoneQuota get account zone quota.
func (a *AccountClient) GetZoneQuota(ctx context.Context, h http.Header,
	request *hsaccount.GetTCloudAccountZoneQuotaReq) (*typeaccount.TCloudAccountQuota, error) {

	resp := new(hsaccount.GetTCloudAccountZoneQuotaResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/zones/quotas").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetBySecret get account info by secret
func (a *AccountClient) GetBySecret(ctx context.Context, h http.Header,
	request *cloud.TCloudSecret) (*cloud.TCloudInfoBySecret, error) {

	resp := new(account.BySecretResp[cloud.TCloudInfoBySecret])
	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/accounts/secret").
		WithHeaders(h).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// ListAuthPolicy list auth policy
func (a *AccountClient) ListAuthPolicy(kt *kit.Kit, req *hsaccount.ListTCloudAuthPolicyReq) (
	[]*cam.ListGrantServiceAccessNode, error) {

	resp := &struct {
		rest.BaseResp `json:",inline"`
		Data          []*cam.ListGrantServiceAccessNode `json:"data"`
	}{}

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/accounts/auth_policies/list").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetResCountBySecret get account res count by secret
func (a *AccountClient) GetResCountBySecret(kt *kit.Kit, request *cloud.TCloudSecret) (*hsaccount.ResCount, error) {

	resp := new(hsaccount.ResCountResp)

	err := a.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/accounts/res_counts/by_secrets").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetNetworkAccountType ...
func (a *AccountClient) GetNetworkAccountType(kt *kit.Kit, accountID string) (
	*vpc.DescribeNetworkAccountTypeResponseParams, error) {

	return common.Request[common.Empty, vpc.DescribeNetworkAccountTypeResponseParams](
		a.client, http.MethodGet, kt, nil, "accounts/%s/network_type", accountID)

}

// CreateSubAccount create subaccount via TCloud CAM AddUser.
func (a *AccountClient) CreateSubAccount(kt *kit.Kit, req *hssubaccount.TCloudCreateSubAccountReq,
) (*hssubaccount.TCloudCreateSubAccountResult, error) {

	return common.Request[hssubaccount.TCloudCreateSubAccountReq, hssubaccount.TCloudCreateSubAccountResult](
		a.client, http.MethodPost, kt, req, "/sub_accounts/create")
}

// UpdateSubAccount update subaccount via TCloud CAM UpdateUser.
func (a *AccountClient) UpdateSubAccount(kt *kit.Kit, req *hssubaccount.TCloudUpdateSubAccountReq) error {
	return common.RequestNoResp[hssubaccount.TCloudUpdateSubAccountReq](
		a.client, http.MethodPost, kt, req, "/sub_accounts/update")
}

// DeleteSubAccount delete subaccount via TCloud CAM DeleteUser.
func (a *AccountClient) DeleteSubAccount(kt *kit.Kit, req *hssubaccount.TCloudDeleteSubAccountReq) error {
	return common.RequestNoResp[hssubaccount.TCloudDeleteSubAccountReq](
		a.client, http.MethodPost, kt, req, "/sub_accounts/delete")
}

// DescribeSafeAuthFlagColl get subaccount safe auth flag settings via TCloud CAM DescribeSafeAuthFlagColl.
func (a *AccountClient) DescribeSafeAuthFlagColl(kt *kit.Kit, req *hssubaccount.TCloudDescribeSafeAuthFlagCollReq,
) ([]hssubaccount.TCloudDescribeSafeAuthFlagCollResult, error) {

	resp, err := common.Request[hssubaccount.TCloudDescribeSafeAuthFlagCollReq,
		[]hssubaccount.TCloudDescribeSafeAuthFlagCollResult](
		a.client, http.MethodPost, kt, req, "/sub_accounts/safe_auth_flag")
	if err != nil {
		return nil, err
	}

	return *resp, nil
}

// SetMfaFlag set subaccount login protection and sensitive operation protection via TCloud CAM SetMfaFlag.
func (a *AccountClient) SetMfaFlag(kt *kit.Kit, req *hssubaccount.TCloudSetMfaFlagReq) error {
	return common.RequestNoResp[hssubaccount.TCloudSetMfaFlagReq](
		a.client, http.MethodPost, kt, req, "/sub_accounts/set_mfa_flag")
}

// DescribeSubAccounts query subaccounts by UIN list via TCloud CAM DescribeSubAccounts.
func (a *AccountClient) DescribeSubAccounts(kt *kit.Kit, req *hssubaccount.TCloudDescribeSubAccountsReq,
) ([]typeaccount.TCloudSubAccountUser, error) {

	resp, err := common.Request[hssubaccount.TCloudDescribeSubAccountsReq, []typeaccount.TCloudSubAccountUser](
		a.client, http.MethodPost, kt, req, "/sub_accounts/describe")
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

// CreateAccessKey create access key for TCloud CAM sub-user.
func (a *AccountClient) CreateAccessKey(kt *kit.Kit, req *hssubaccount.TCloudCreateAccessKeyReq,
) (*hssubaccount.TCloudCreateAccessKeyResult, error) {

	return common.Request[hssubaccount.TCloudCreateAccessKeyReq, hssubaccount.TCloudCreateAccessKeyResult](
		a.client, http.MethodPost, kt, req, "/sub_accounts/secrets/create")
}

// DeleteAccessKey delete access key for TCloud CAM sub-user.
func (a *AccountClient) DeleteAccessKey(kt *kit.Kit, req *hssubaccount.TCloudDeleteAccessKeyReq) error {
	return common.RequestNoResp[hssubaccount.TCloudDeleteAccessKeyReq](
		a.client, http.MethodPost, kt, req, "/sub_accounts/secrets/delete")
}

// UpdateAccessKey update access key status for TCloud CAM sub-user.
func (a *AccountClient) UpdateAccessKey(kt *kit.Kit, req *hssubaccount.TCloudUpdateAccessKeyReq) error {
	return common.RequestNoResp[hssubaccount.TCloudUpdateAccessKeyReq](
		a.client, http.MethodPost, kt, req, "/sub_accounts/secrets/update")
}

// AttachUserPolicies batch attaches CAM policies to a TCloud sub-user.
func (a *AccountClient) AttachUserPolicies(kt *kit.Kit, req *hssubaccount.TCloudAttachUserPoliciesReq) error {
	return common.RequestNoResp[hssubaccount.TCloudAttachUserPoliciesReq](
		a.client, http.MethodPost, kt, req, "/sub_accounts/attach_user_policies")
}

// DetachUserPolicies batch detaches CAM policies from a TCloud sub-user.
func (a *AccountClient) DetachUserPolicies(kt *kit.Kit, req *hssubaccount.TCloudDetachUserPoliciesReq) error {
	return common.RequestNoResp[hssubaccount.TCloudDetachUserPoliciesReq](
		a.client, http.MethodPost, kt, req, "/sub_accounts/detach_user_policies")
}
