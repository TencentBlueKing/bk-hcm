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
	coreaudit "hcm/pkg/api/core/audit"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// AuditClient is data service audit api client.
type AuditClient struct {
	client rest.ClientInterface
}

// NewAuditClient create a new audit api client.
func NewAuditClient(client rest.ClientInterface) *AuditClient {
	return &AuditClient{
		client: client,
	}
}

// CloudResourceUpdateAudit cloud resource update audit.
func (a *AuditClient) CloudResourceUpdateAudit(ctx context.Context, h http.Header,
	request *protoaudit.CloudResourceUpdateAuditReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/resources/update_audits/create").
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

// CloudResourceDeleteAudit cloud resource delete audit.
func (a *AuditClient) CloudResourceDeleteAudit(ctx context.Context, h http.Header,
	request *protoaudit.CloudResourceDeleteAuditReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/resources/delete_audits/create").
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

// CloudResourceAssignAudit cloud resource assign audit.
func (a *AuditClient) CloudResourceAssignAudit(ctx context.Context, h http.Header,
	request *protoaudit.CloudResourceAssignAuditReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/resources/assign_audits/create").
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

// CloudResourceOperationAudit cloud resource operation audit.
func (a *AuditClient) CloudResourceOperationAudit(ctx context.Context, h http.Header,
	request *protoaudit.CloudResourceOperationAuditReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cloud/resources/operation_audits/create").
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

// CloudResourceRecycleAudit create cloud resource recycle/recover audit.
func (a *AuditClient) CloudResourceRecycleAudit(ctx context.Context, h http.Header,
	req *protoaudit.CloudResourceRecycleAuditReq) error {

	resp := new(rest.BaseResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/cloud/resources/recycle_audits/create").
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

// ListAudit list audit.
func (a *AuditClient) ListAudit(ctx context.Context, h http.Header, request *core.ListReq) (
	*protoaudit.ListResult, error) {

	resp := new(protoaudit.ListResp)

	err := a.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/audits/list").
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

// GetAudit get audit.
func (a *AuditClient) GetAudit(ctx context.Context, h http.Header, id uint64) (
	*coreaudit.Audit, error) {

	resp := new(protoaudit.GetResp)

	err := a.client.Get().
		WithContext(ctx).
		SubResourcef("/audits/%d", id).
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

// GetAuditRaw  get audit with raw detail
func (a *AuditClient) GetAuditRaw(kt *kit.Kit, id uint64) (
	*coreaudit.RawAudit, error) {

	return common.Request[common.Empty, coreaudit.RawAudit](a.client, rest.GET, kt, nil,
		"/audits/%d", id)
}
