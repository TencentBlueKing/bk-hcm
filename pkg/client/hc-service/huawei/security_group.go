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

package huawei

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
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

// CreateSecurityGroup create security group.
func (cli *SecurityGroupClient) CreateSecurityGroup(ctx context.Context, h http.Header,
	request *proto.HuaWeiSecurityGroupCreateReq) (*core.CreateResult, error) {

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

// SyncSecurityGroup security group.
func (cli *SecurityGroupClient) SyncSecurityGroup(ctx context.Context, h http.Header,
	request *sync.HuaWeiSyncReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/sync").
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

// UpdateSecurityGroup update security group rule.
func (cli *SecurityGroupClient) UpdateSecurityGroup(ctx context.Context, h http.Header, id string,
	request *proto.SecurityGroupUpdateReq) error {

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

// DeleteSecurityGroup delete security group.
func (cli *SecurityGroupClient) DeleteSecurityGroup(kt *kit.Kit, id string) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(kt.Ctx).
		SubResourcef("/security_groups/%s", id).
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

// CreateSecurityGroupRule create security group rule.
func (cli *SecurityGroupClient) CreateSecurityGroupRule(kt *kit.Kit, sgID string,
	request *proto.HuaWeiSGRuleCreateReq) (*core.CreateResult, error) {

	resp := new(core.CreateResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/security_groups/%s/rules/create", sgID).
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

// DeleteSecurityGroupRule delete security group rule.
func (cli *SecurityGroupClient) DeleteSecurityGroupRule(ctx context.Context, h http.Header, sgID, id string) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(ctx).
		SubResourcef("/security_groups/%s/rules/%s", sgID, id).
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

// AssociateCvm ...
func (cli *SecurityGroupClient) AssociateCvm(ctx context.Context, h http.Header,
	req *proto.SecurityGroupAssociateCvmReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/security_groups/associate/cvms").
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

// DisassociateCvm ...
func (cli *SecurityGroupClient) DisassociateCvm(ctx context.Context, h http.Header,
	req *proto.SecurityGroupAssociateCvmReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/security_groups/disassociate/cvms").
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

// ListSecurityGroupStatistic 查询安全组关联的云上资源数量
func (cli *SecurityGroupClient) ListSecurityGroupStatistic(kt *kit.Kit, req *proto.ListSecurityGroupStatisticReq) (
	*proto.ListSecurityGroupStatisticResp, error) {

	return common.Request[proto.ListSecurityGroupStatisticReq, proto.ListSecurityGroupStatisticResp](
		cli.client, rest.POST, kt, req, "/security_groups/statistic")
}

// SyncSecurityGroupUsageBizRel ...
func (cli *SecurityGroupClient) SyncSecurityGroupUsageBizRel(kt *kit.Kit, req *sync.HuaWeiSyncReq) error {
	return common.RequestNoResp[sync.HuaWeiSyncReq](cli.client, rest.POST, kt, req,
		"/security_groups/usage_biz_rels/sync")
}
