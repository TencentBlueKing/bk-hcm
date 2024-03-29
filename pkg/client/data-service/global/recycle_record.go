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
	"hcm/pkg/api/core"
	rr "hcm/pkg/api/core/recycle-record"
	proto "hcm/pkg/api/data-service/recycle-record"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// RecycleRecordClient is data service recycle record api client.
type RecycleRecordClient struct {
	client rest.ClientInterface
}

// NewRecycleRecordClient create a new recycle record api client.
func NewRecycleRecordClient(client rest.ClientInterface) *RecycleRecordClient {
	return &RecycleRecordClient{
		client: client,
	}
}

// BatchRecycleCloudRes 将资源批量加入回收站
func (r *RecycleRecordClient) BatchRecycleCloudRes(kt *kit.Kit, req *proto.BatchRecycleReq) (
	string, error) {

	resp := new(proto.RecycleResp)

	err := r.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/cloud/resources/batch/recycle").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return "", err
	}

	if resp.Code != errf.OK {
		return "", errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// BatchRecoverCloudResource batch recover cloud resource.
func (r *RecycleRecordClient) BatchRecoverCloudResource(kt *kit.Kit, request *proto.BatchRecoverReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/cloud/resources/batch/recover").
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

// ListRecycleRecord list recycle record.
func (r *RecycleRecordClient) ListRecycleRecord(kt *kit.Kit, request *core.ListReq) (
	*proto.ListResult, error) {

	resp := new(proto.ListResp)

	err := r.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/recycle_records/list").
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

// ListCvmRecycleRecord list cvm recycle record.
func (r *RecycleRecordClient) ListCvmRecycleRecord(kt *kit.Kit, request *core.ListReq) (
	*core.ListResultT[rr.CvmRecycleRecord], error) {

	return common.Request[core.ListReq, core.ListResultT[rr.CvmRecycleRecord]](r.client, rest.POST, kt, request,
		"/recycle_records/list")
}

// BatchUpdateRecycleRecord batch update recycle record.
func (r *RecycleRecordClient) BatchUpdateRecycleRecord(kt *kit.Kit, request *proto.BatchUpdateReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Patch().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/recycle_records/batch").
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

// BatchUpdateRecycleStatus update recycle_status of resources itself(disk,cvm)
func (r *RecycleRecordClient) BatchUpdateRecycleStatus(kt *kit.Kit, request *proto.BatchUpdateRecycleStatusReq) error {

	resp := new(rest.BaseResp)

	err := r.client.Patch().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/recycle_records/recycle_status/batch").
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
