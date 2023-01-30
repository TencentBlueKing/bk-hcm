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

package azure

import (
	"context"
	"net/http"

	"hcm/pkg/adaptor/types"
<<<<<<< HEAD
	"hcm/pkg/criteria/enumor"
=======
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// SubnetClient is hc service azure subnet api client.
type SubnetClient struct {
	client rest.ClientInterface
}

// NewSubnetClient create a new subnet api client.
func NewSubnetClient(client rest.ClientInterface) *SubnetClient {
	return &SubnetClient{
		client: client,
	}
}

// Update subnet.
func (v *SubnetClient) Update(ctx context.Context, h http.Header, id string, req *types.AzureSubnetUpdateOption) error {
	resp := new(rest.BaseResp)

	err := v.client.Patch().
		WithContext(ctx).
		Body(req).
<<<<<<< HEAD
		SubResourcef("/vendors/%s/subnets/%s", enumor.Azure, id).
=======
		SubResourcef("/subnets/%s", id).
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
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

// Delete subnet.
func (v *SubnetClient) Delete(ctx context.Context, h http.Header, id string) error {
	resp := new(rest.BaseResp)

	err := v.client.Delete().
		WithContext(ctx).
		Body(nil).
<<<<<<< HEAD
		SubResourcef("/vendors/%s/subnets/%s", enumor.Azure, id).
=======
		SubResourcef("/subnets/%s", id).
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
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
