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

package aws

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCvmClient create a new cvm api client.
func NewCvmClient(client rest.ClientInterface) *CvmClient {
	return &CvmClient{
		client: client,
	}
}

// CvmClient is hc service cvm api client.
type CvmClient struct {
	client rest.ClientInterface
}

// SyncCvmWithRelResource sync cvm with rel resource.
func (cli *CvmClient) SyncCvmWithRelResource(ctx context.Context, h http.Header, request *sync.AwsSyncReq) error {

	resp := new(core.SyncResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/cvms/with/relation_resources/sync").
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

// BatchStartCvm ....
func (cli *CvmClient) BatchStartCvm(kt *kit.Kit, request *protocvm.AwsBatchStartReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/cvms/batch/start").
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

// BatchStopCvm ....
func (cli *CvmClient) BatchStopCvm(kt *kit.Kit, request *protocvm.AwsBatchStopReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/cvms/batch/stop").
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

// BatchRebootCvm ....
func (cli *CvmClient) BatchRebootCvm(kt *kit.Kit, request *protocvm.AwsBatchRebootReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/cvms/batch/reboot").
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

// BatchDeleteCvm ....
func (cli *CvmClient) BatchDeleteCvm(kt *kit.Kit, request *protocvm.AwsBatchDeleteReq) error {

	resp := new(rest.BaseResp)

	err := cli.client.Delete().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/cvms/batch").
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

// BatchCreateCvm ....
func (cli *CvmClient) BatchCreateCvm(kt *kit.Kit, request *protocvm.AwsBatchCreateReq) (
	*protocvm.BatchCreateResult, error) {

	resp := new(protocvm.BatchCreateResp)

	err := cli.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/cvms/batch/create").
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

// BatchAssociateSecurityGroup ....
func (cli *CvmClient) BatchAssociateSecurityGroup(kt *kit.Kit,
	request *protocvm.AwsCvmBatchAssociateSecurityGroupReq) error {
	return common.RequestNoResp[protocvm.AwsCvmBatchAssociateSecurityGroupReq](
		cli.client, http.MethodPost, kt, request, "/cvms/security_groups/batch/associate")
}

// ListCvmNetworkInterface ....
func (cli *CvmClient) ListCvmNetworkInterface(kt *kit.Kit, request *protocvm.ListCvmNetworkInterfaceReq) (
	*map[string]*protocvm.ListCvmNetworkInterfaceRespItem, error) {

	return common.Request[protocvm.ListCvmNetworkInterfaceReq, map[string]*protocvm.ListCvmNetworkInterfaceRespItem](
		cli.client, rest.POST, kt, request, "/cvms/network_interfaces/list")
}

// SyncCCInfo ...
func (cli *CvmClient) SyncCCInfo(kt *kit.Kit, req *sync.AwsSyncReq) error {
	return common.RequestNoResp[sync.AwsSyncReq](cli.client, rest.POST, kt, req, "/cvms/cc_info/sync")
}

// SyncCCInfoByCond ...
func (cli *CvmClient) SyncCCInfoByCond(kt *kit.Kit, req *sync.SyncCvmByCondReq) error {
	return common.RequestNoResp[sync.SyncCvmByCondReq](cli.client, rest.POST, kt, req,
		"/cvms/cc_info/by_condition/sync")
}
