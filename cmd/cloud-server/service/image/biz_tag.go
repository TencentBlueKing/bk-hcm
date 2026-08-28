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

package image

import (
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// UpdateImageBizTag 更新镜像业务标签
func (svc *imageSvc) UpdateImageBizTag(cts *rest.Contexts) (interface{}, error) {
	imageID := cts.PathParameter("image_id").String()
	if imageID == "" {
		return nil, errf.New(errf.InvalidParameter, "image_id is required")
	}

	req := new(dataproto.UpdateImageBizTagReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode update image biz tag request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("validate update image biz tag request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Image, Action: meta.Update},
	}); err != nil {
		logs.Errorf("update image biz tag auth failed, bkBizID: %d, err: %v, rid: %s",
			req.BkBizID, err, cts.Kit.Rid)
		return nil, err
	}

	if err := svc.client.DataService().Global.UpdateImageBizTag(cts.Kit, imageID, req); err != nil {
		logs.Errorf("update image biz tag failed, imageID: %s, req: %+v, err: %v, rid: %s",
			imageID, req, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
