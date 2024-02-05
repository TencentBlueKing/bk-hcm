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

// BatchCreateSecurityGroup batch create security group rule.
func (cli *SecurityGroupClient) BatchCreateSecurityGroup(ctx context.Context, h http.Header, request *protocloud.
	SecurityGroupBatchCreateReq[corecloud.TCloudSecurityGroupExtension]) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/batch/create").
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

// BatchUpdateSecurityGroup batch update security group.
func (cli *SecurityGroupClient) BatchUpdateSecurityGroup(ctx context.Context, h http.Header,
	request *protocloud.SecurityGroupBatchUpdateReq[corecloud.TCloudSecurityGroupExtension]) error {

	resp := new(rest.BaseResp)

	err := cli.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/batch/update").
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
	*corecloud.SecurityGroup[corecloud.TCloudSecurityGroupExtension], error) {

	resp := new(protocloud.SecurityGroupGetResp[corecloud.TCloudSecurityGroupExtension])

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

// ListSecurityGroupExt list security group with extension.
func (cli *SecurityGroupClient) ListSecurityGroupExt(ctx context.Context, h http.Header, req *core.ListReq) (
	*protocloud.SecurityGroupExtListResult[corecloud.TCloudSecurityGroupExtension], error) {

	resp := new(protocloud.SecurityGroupExtListResp[corecloud.TCloudSecurityGroupExtension])

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/security_groups/list").
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
	TCloudSGRuleCreateReq, sgID string) (*core.BatchCreateResult, error) {

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
	TCloudSGRuleBatchUpdateReq, sgID string) error {

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
	TCloudSGRuleListReq, sgID string) (*protocloud.TCloudSGRuleListResult, error) {

	resp := new(protocloud.TCloudSGRuleListResp)

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

// ListSecurityGroupRuleExt list security group rule ext.
func (cli *SecurityGroupClient) ListSecurityGroupRuleExt(ctx context.Context, h http.Header, request *protocloud.
	TCloudSGRuleListReq) (*protocloud.TCloudSGRuleListExtResult, error) {

	resp := new(protocloud.TCloudSGRuleListExtResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/rules/list").
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

// BatchDeleteSecurityGroupRule delete security group rule.
func (cli *SecurityGroupClient) BatchDeleteSecurityGroupRule(ctx context.Context, h http.Header, request *protocloud.
	TCloudSGRuleBatchDeleteReq, sgID string) error {

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
