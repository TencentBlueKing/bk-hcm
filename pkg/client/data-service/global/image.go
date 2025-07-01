/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
)

// ListImage 查询公共镜像列表
func (rc *restClient) ListImage(kt *kit.Kit, request *core.ListReq) (
	*dataproto.ListResult, error) {

	resp := new(core.BaseResp[*dataproto.ListResult])

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

// DeleteImage 删除公共镜像记录
func (rc *restClient) DeleteImage(kt *kit.Kit, request *dataproto.DeleteReq) error {

	resp := new(core.DeleteResp)
	err := rc.client.Delete().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/images/batch").
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
