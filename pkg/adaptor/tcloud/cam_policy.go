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

package tcloud

import (
	"fmt"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
)

// CreatePolicy creates a CAM policy.
// reference: https://cloud.tencent.com/document/product/598/34578
func (t *TCloudImpl) CreatePolicy(kt *kit.Kit, opt *typeaccount.TCloudCreatePolicyOption) (
	*typeaccount.TCloudCreatePolicyResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	region := opt.Region
	if region == "" {
		region = constant.TCloudDefaultRegion
	}

	client, err := t.clientSet.CamServiceClient(region)
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewCreatePolicyRequest()
	req.PolicyName = converter.ValToPtr(opt.PolicyName)
	req.PolicyDocument = converter.ValToPtr(opt.PolicyDocument)
	if opt.Description != "" {
		req.Description = converter.ValToPtr(opt.Description)
	}

	resp, err := client.CreatePolicyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create cam policy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &typeaccount.TCloudCreatePolicyResult{
		PolicyID: converter.PtrToVal(resp.Response.PolicyId),
	}, nil
}
