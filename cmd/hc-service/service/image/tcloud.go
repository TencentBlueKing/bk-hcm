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
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/api/hc-service/image"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

func (svc *imageSvc) initTCloudImageService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("ListImage", http.MethodPost, "/vendors/tcloud/images/list", svc.ListImage)

	h.Load(cap.WebService)
}

// ListImage ...
func (svc *imageSvc) ListImage(cts *rest.Contexts) (interface{}, error) {

	req := new(image.TCloudImageListOption)
	err := cts.DecodeInto(req)
	if err != nil {
		return nil, err
	}
	err = req.Validate()
	if err != nil {
		return nil, err
	}
	cli, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	result, err := cli.ListImage(cts.Kit, req.TCloudImageListOption)
	if err != nil {
		logs.Errorf("list images failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
