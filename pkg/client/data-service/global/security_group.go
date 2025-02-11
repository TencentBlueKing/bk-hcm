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

package global

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
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

// ListSecurityGroup security group.
func (cli *SecurityGroupClient) ListSecurityGroup(ctx context.Context, h http.Header, request *protocloud.
	SecurityGroupListReq) (*protocloud.SecurityGroupListResult, error) {

	resp := new(protocloud.SecurityGroupListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
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

// BatchDeleteSecurityGroup batch delete security group.
func (cli *SecurityGroupClient) BatchDeleteSecurityGroup(ctx context.Context, h http.Header, request *protocloud.
	SecurityGroupBatchDeleteReq) error {

	resp := new(core.DeleteResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/batch").
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

// BatchUpdateSecurityGroupCommonInfo batch update security group common info.
func (cli *SecurityGroupClient) BatchUpdateSecurityGroupCommonInfo(ctx context.Context, h http.Header,
	request *protocloud.SecurityGroupCommonInfoBatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/common/info/batch/update").
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

func (cli *SecurityGroupClient) BatchUpdateSecurityGroupMgmtAttr(ctx context.Context, h http.Header,
	request *protocloud.BatchUpdateSecurityGroupMgmtAttrReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_groups/mgmt/attr/batch/update").
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
