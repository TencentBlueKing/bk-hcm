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
	corecloud "hcm/pkg/api/core/cloud/zone"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// NewZoneClient create a new zone api client.
func NewZoneClient(client rest.ClientInterface) *ZoneClient {
	return &ZoneClient{
		client: client,
	}
}

// ZoneClient is data service zone api client.
type ZoneClient struct {
	client rest.ClientInterface
}

// BatchCreateZone batch create zone.
func (cli *ZoneClient) BatchCreateZone(ctx context.Context, h http.Header, request *protocloud.
	ZoneBatchCreateReq[corecloud.AwsZoneExtension]) (*core.BatchCreateResult, error) {

	resp := new(core.BatchCreateResp)

	err := cli.client.Post().
		WithContext(ctx).
		Body(request).
		SubResourcef("/zones/batch/create").
		WithHeaders(h).
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

// BatchUpdateZone batch update zone.
func (cli *ZoneClient) BatchUpdateZone(ctx context.Context, h http.Header,
	request *protocloud.ZoneBatchUpdateReq[corecloud.AwsZoneExtension]) error {

	resp := new(rest.BaseResp)

	err := cli.client.Patch().
		WithContext(ctx).
		Body(request).
		SubResourcef("/zones/batch/update").
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
