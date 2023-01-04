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

	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// CloudClient is data service cloud api client.
type CloudClient struct {
	client rest.ClientInterface
}

// NewCloudClient create a new cloud api client.
func NewCloudClient(client rest.ClientInterface) *CloudClient {
	return &CloudClient{
		client: client,
	}
}

// GetResourceVendor get cloud resource vendor.
func (cli *CloudClient) GetResourceVendor(ctx context.Context, h http.Header, resType enumor.CloudResourceType,
		resID string) (enumor.Vendor, error) {

	resp := new(protocloud.GetResourceVendorResp)

	err := cli.client.Get().
		WithContext(ctx).
		SubResourcef("/cloud/resource/vendor/%s/id/%s", resType, resID).
		WithHeaders(h).
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
