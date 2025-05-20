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

package hcservice

import (
	"fmt"

	"hcm/pkg/client/hc-service/aws"
	"hcm/pkg/client/hc-service/azure"
	"hcm/pkg/client/hc-service/gcp"
	"hcm/pkg/client/hc-service/huawei"
	"hcm/pkg/client/hc-service/other"
	"hcm/pkg/client/hc-service/tcloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
)

// Client is hc-service api client.
type Client struct {
	TCloud *tcloud.Client
	Aws    *aws.Client
	HuaWei *huawei.Client
	Gcp    *gcp.Client
	Azure  *azure.Client
	Other  *other.Client
}

// NewClient create a new hc-service api client.
func NewClient(c *client.Capability, version string) *Client {
	prefixPath := fmt.Sprintf("/api/%s/hc/vendors", version)
	return &Client{
		// Note: 对于Global Client，主要是用于无vendor区分即全局或跨多个云的请求，HC服务应该没有全局的API
		TCloud: tcloud.NewClient(
			rest.NewClient(c, fmt.Sprintf("%s/%s", prefixPath, enumor.TCloud)),
		),
		Aws: aws.NewClient(
			rest.NewClient(c, fmt.Sprintf("%s/%s", prefixPath, enumor.Aws)),
		),
		HuaWei: huawei.NewClient(
			rest.NewClient(c, fmt.Sprintf("%s/%s", prefixPath, enumor.HuaWei)),
		),
		Gcp: gcp.NewClient(
			rest.NewClient(c, fmt.Sprintf("%s/%s", prefixPath, enumor.Gcp)),
		),
		Azure: azure.NewClient(
			rest.NewClient(c, fmt.Sprintf("%s/%s", prefixPath, enumor.Azure)),
		),
		Other: other.NewClient(
			rest.NewClient(c, fmt.Sprintf("%s/%s", prefixPath, enumor.Other)),
		),
	}
}
