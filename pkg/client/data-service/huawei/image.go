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
	"hcm/pkg/api/core"
	coreimage "hcm/pkg/api/core/cloud/image"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
)

// BatchCreateImage 批量创建公共镜像
func (rc *restClient) BatchCreateImage(kt *kit.Kit, request *dataproto.BatchCreateReq[coreimage.HuaWeiExtension]) (
	*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := rc.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/images/batch/create").
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

// GetImage 查询单个公共镜像详情
func (rc *restClient) GetImage(kt *kit.Kit, imageID string) (*coreimage.Image[coreimage.HuaWeiExtension], error) {

	resp := new(core.BaseResp[*coreimage.Image[coreimage.HuaWeiExtension]])

	err := rc.client.Get().
		WithContext(kt.Ctx).
		SubResourcef("/images/%s", imageID).
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

// ListImage 查询公共镜像列表(带 extension 字段)
func (rc *restClient) ListImage(kt *kit.Kit, request *core.ListReq) (
	*dataproto.ListExtResult[coreimage.HuaWeiExtension], error) {

	resp := new(core.BaseResp[*dataproto.ListExtResult[coreimage.HuaWeiExtension]])

	err := rc.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/images/list").
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

// BatchUpdateImage 更新公共镜像(带 extension 字段)
func (rc *restClient) BatchUpdateImage(kt *kit.Kit, request *dataproto.BatchUpdateReq[coreimage.HuaWeiExtension]) (
	interface{}, error) {

	resp := new(core.UpdateResp)
	err := rc.client.Patch().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/images").
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
