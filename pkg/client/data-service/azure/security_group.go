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
	"context"
	"net/http"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewCloudSecurityGroupClient create a new security group api client.
func NewCloudSecurityGroupClient(client rest.ClientInterface) *SecurityGroupClient {
	return &SecurityGroupClient{
		client: client,
	}
}

// SecurityGroupClient is data service security group api client.
type SecurityGroupClient struct {
	client rest.ClientInterface
}

// CreateSecurityGroup create security group rule.
func (cli *SecurityGroupClient) CreateSecurityGroup(ctx context.Context, h http.Header, request *protocloud.
	SecurityGroupCreateReq[corecloud.AzureSecurityGroupExtension]) (*core.CreateResult, error) {

	resp := new(core.CreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/create").
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

// UpdateSecurityGroup update security group.
func (cli *SecurityGroupClient) UpdateSecurityGroup(ctx context.Context, h http.Header, id string,
	request *protocloud.SecurityGroupUpdateReq[corecloud.AzureSecurityGroupExtension]) error {

	resp := new(core.UpdateResp)

	err := cli.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/%s", id).
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

// GetSecurityGroup get security group.
func (cli *SecurityGroupClient) GetSecurityGroup(ctx context.Context, h http.Header, id string) (
	*corecloud.SecurityGroup[corecloud.AzureSecurityGroupExtension], error) {

	resp := new(protocloud.SecurityGroupGetResp[corecloud.AzureSecurityGroupExtension])

	err := cli.client.Get().
		WithContext(ctx).
		SubResourcef("/security_groups/%s", id).
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

// BatchCreateSecurityGroupRule batch create security group rule.
func (cli *SecurityGroupClient) BatchCreateSecurityGroupRule(ctx context.Context, h http.Header, request *protocloud.
	AzureSGRuleCreateReq, sgID string) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/%s/rules/batch/create", sgID).
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

// BatchUpdateSecurityGroupRule update security group rule.
func (cli *SecurityGroupClient) BatchUpdateSecurityGroupRule(ctx context.Context, h http.Header, request *protocloud.
	AzureSGRuleBatchUpdateReq, sgID string) error {

	resp := new(core.UpdateResp)

	err := cli.client.Put().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/%s/rules/batch", sgID).
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

// ListSecurityGroupRule list security group rule.
func (cli *SecurityGroupClient) ListSecurityGroupRule(ctx context.Context, h http.Header, request *protocloud.
	AzureSGRuleListReq, sgID string) (*protocloud.AzureSGRuleListResult, error) {

	resp := new(protocloud.AzureSGRuleListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/%s/rules/list", sgID).
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

// DeleteSecurityGroupRule delete security group rule.
func (cli *SecurityGroupClient) DeleteSecurityGroupRule(ctx context.Context, h http.Header, request *protocloud.
	AzureSGRuleDeleteReq, sgID string) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/%s/rules/batch", sgID).
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
