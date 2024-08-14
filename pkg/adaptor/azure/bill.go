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

package azure

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"hcm/pkg/adaptor/types"
	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/tools/json"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/consumption/armconsumption"
)

const (
	ManageServerURL     = "https://management.azure.com/"
	LoginServerURL      = "https://login.microsoftonline.com/"
	AuthHeader          = "Authorization"
	APIVersion          = "2023-03-01"
	MeterDetailsExpand  = "meterDetails,additionalInfo"
	ClientCredGrantType = "client_credentials"
)

type billDiscovery struct {
	server string
}

// GetServers ...
func (s *billDiscovery) GetServers() ([]string, error) {
	if len(s.server) == 0 {
		return []string{}, errors.New("there is no bill discovery server can be used")
	}
	return []string{s.server}, nil
}

type billClient struct {
	// http client instance
	client     rest.ClientInterface
	LoginToken *LoginTokenProto `json:"login_token"`
}

// LoginTokenProto ...
type LoginTokenProto struct {
	SubscriptionID string `json:"subscription_id"`
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	ExpiresIn      string `json:"expires_in"`
	ExpiresOn      string `json:"expires_on"`
	Resource       string `json:"resource"`
}

// UsageCommonError ...
type UsageCommonError struct {
	Error *UsageRespError `json:"error"`
}

// UsageRespError ...
type UsageRespError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newBillClient(server string, token *LoginTokenProto) (*billClient, error) {
	// 生成Client
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &billDiscovery{
			server: server,
		},
	}
	restCli := rest.NewClient(c, "")

	return &billClient{client: restCli, LoginToken: token}, nil
}

func getToken(kt *kit.Kit, credential *types.AzureCredential) (*LoginTokenProto, error) {
	cli, err := newBillClient(LoginServerURL, nil)
	if err != nil {
		return nil, err
	}

	token := &LoginTokenProto{SubscriptionID: credential.CloudSubscriptionID}
	h := http.Header{}
	err = cli.client.Post().
		WithContext(kt.Ctx).
		WithHeaders(h).
		WithContentType("application/x-www-form-urlencoded").
		WithParam("grant_type", ClientCredGrantType).
		WithParam("client_id", credential.CloudApplicationID).
		WithParam("client_secret", credential.CloudClientSecretKey).
		WithParam("resource", ManageServerURL).
		SubResourcef("%s/oauth2/token", credential.CloudTenantID).
		Do().Into(token)

	return token, nil
}

func (b *billClient) processError(result *rest.Result) error {
	if result.Err != nil {
		return result.Err
	}

	respErr := new(UsageCommonError)
	err := json.Unmarshal(result.Body, &respErr)
	if err != nil {
		logs.Errorf("azure invalid response body, unmarshal json failed, reply: %s, err: %v", result.Body, err)
		return fmt.Errorf("azure invalid http response, unmarshal json failed, err: %+v", err)
	}

	if respErr != nil && respErr.Error != nil && respErr.Error.Code != "" {
		logs.Errorf("azure error response body, err: %+v", respErr)
		return fmt.Errorf("azure error http response, err: %s(%s)", respErr.Error.Message, respErr.Error.Code)
	}

	return nil
}

// GetUsageDetail get usage detail
// https://management.azure.com/subscriptions/7C99B444-456A-4EEF-A083-40F6AB39CEAA/providers/
// Microsoft.Consumption/usageDetails
func (b *billClient) GetUsageDetail(kt *kit.Kit, opt *typesBill.AzureBillListOption) (
	*armconsumption.UsageDetailsListResult, error) {

	h := http.Header{}
	// 获取AccessToken
	h.Set(AuthHeader, "Bearer "+b.LoginToken.AccessToken)

	apiPath := fmt.Sprintf("subscriptions/%s/providers/Microsoft.Consumption/usageDetails",
		opt.SubscriptionID)
	usageClient := b.client.Get().
		WithContext(kt.Ctx).
		WithHeaders(h).
		WithParam("api-version", APIVersion).
		WithParam("$expand", MeterDetailsExpand).
		SubResourcef(apiPath)
	if opt.Page != nil && opt.Page.NextLink != "" {
		nextLink := strings.ReplaceAll(opt.Page.NextLink, ManageServerURL+apiPath+"?", "")
		URLParams, err := url.ParseQuery(nextLink)
		if err != nil {
			return nil, err
		}

		if _, ok := URLParams["$filter"]; ok && len(URLParams["$filter"]) > 0 {
			usageClient = usageClient.WithParam("$filter", fmt.Sprintf("%s", URLParams["$filter"][0]))
		}
		if _, ok := URLParams["$top"]; ok && len(URLParams["$top"]) > 0 {
			usageClient = usageClient.WithParam("$top", fmt.Sprintf("%s", URLParams["$top"][0]))
		}
		if _, ok := URLParams["$skiptoken"]; ok && len(URLParams["$skiptoken"]) > 0 {
			usageClient = usageClient.WithParam("$skiptoken", fmt.Sprintf("%s", URLParams["$skiptoken"][0]))
		}
		if _, ok := URLParams["id"]; ok && len(URLParams["id"]) > 0 {
			usageClient = usageClient.WithParam("id", fmt.Sprintf("%s", URLParams["id"][0]))
		}
	} else {
		// 格式：properties/usageEnd+ge+'2023-04-05'+AND+properties/usageEnd+le+'2023-04-05'
		if opt.BeginDate != "" && opt.EndDate != "" {
			usageClient = usageClient.WithParam("$filter", fmt.Sprintf(
				"properties/usageStart ge '%s' AND properties/usageEnd le '%s'", opt.BeginDate, opt.EndDate))
		}
		if opt.Page != nil && opt.Page.Limit > 0 {
			usageClient = usageClient.WithParam("$top", fmt.Sprintf("%d", opt.Page.Limit))
		}
	}

	result := usageClient.Do()
	err := b.processError(result)
	if err != nil {
		return nil, err
	}

	resp := new(armconsumption.UsageDetailsListResult)
	err = result.Into(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetBillList get bill list.
// reference: https://learn.microsoft.com/zh-cn/rest/api/consumption/usage-details/list?tabs=HTTP#usagedetailslistresult
func (az *Azure) GetBillList(kt *kit.Kit, opt *typesBill.AzureBillListOption) (
	*armconsumption.UsageDetailsListResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := az.clientSet.usageDetailClient(kt)
	if err != nil {
		return nil, fmt.Errorf("new usage detail client failed, opt: %+v, err: %v", opt, err)
	}

	return cli.GetUsageDetail(kt, opt)
}
