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
	corecloud "hcm/pkg/api/core/cloud"
	proto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewCloudSGCvmRelClient create a new security group api client.
func NewCloudSGCvmRelClient(client rest.ClientInterface) *SGCvmRelClient {
	return &SGCvmRelClient{
		client: client,
	}
}

// SGCvmRelClient is data service security group cvm rel api client.
type SGCvmRelClient struct {
	client rest.ClientInterface
}

// BatchCreateSgCvmRels security group cvm rels.
// Deprecated: use SGCommonRelClient's BatchCreateSgCommonRels instead.
func (cli *SGCvmRelClient) BatchCreateSgCvmRels(ctx context.Context, h http.Header,
	request *protocloud.SGCvmRelBatchCreateReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_group_cvm_rels/batch/create").
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

// BatchDeleteSgCvmRels security group cvm rels.
// Deprecated: use SGCommonRelClient's BatchDeleteSgCommonRels instead.
func (cli *SGCvmRelClient) BatchDeleteSgCvmRels(ctx context.Context, h http.Header, request *proto.BatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_group_cvm_rels/batch").
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

// ListSgCvmRels security group cvm rels.
// Deprecated: use SGCommonRelClient's ListSgCommonRels instead.
func (cli *SGCvmRelClient) ListSgCvmRels(ctx context.Context, h http.Header, request *core.ListReq) (
	*protocloud.SGCvmRelListResult, error) {

	resp := new(protocloud.SGCvmRelListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_group_cvm_rels/list").
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

// ListWithSecurityGroup security group cvm rels with security group.
// Deprecated: use SGCommonRelClient's ListWithSecurityGroup instead.
func (cli *SGCvmRelClient) ListWithSecurityGroup(ctx context.Context, h http.Header,
	request *protocloud.SGCvmRelWithSecurityGroupListReq) ([]corecloud.SGCvmRelWithBaseSecurityGroup, error) {

	resp := new(protocloud.SGCvmRelWithSGListResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/security_group_cvm_rels/with/security_group/list").
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
