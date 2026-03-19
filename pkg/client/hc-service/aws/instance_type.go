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
	"encoding/json"

	instancetype "hcm/pkg/api/hc-service/instance-type"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// InstanceTypeClient ...
type InstanceTypeClient struct {
	client rest.ClientInterface
}

// NewInstanceTypeClient ...
func NewInstanceTypeClient(client rest.ClientInterface) *InstanceTypeClient {
	return &InstanceTypeClient{
		client: client,
	}
}

// List ...
func (c *InstanceTypeClient) List(kt *kit.Kit, request *instancetype.AwsInstanceTypeListReq) (
	[]*instancetype.AwsInstanceTypeResp, error) {

	resp := new(instancetype.AwsInstanceTypeListResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/instance_types/list").
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

// ListAssumeRoleInstanceType lists instance types via AssumeRole cross-account access.
func (c *InstanceTypeClient) ListAssumeRoleInstanceType(kt *kit.Kit,
	request *instancetype.AwsAssumeRoleInstanceTypeListReq) ([]*instancetype.AwsInstanceTypeResp, error) {

	resp := new(instancetype.AwsAssumeRoleInstanceTypeListResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/assume_role/instance_types/list").
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

// ListAssumeRoleInstance lists EC2 instances via AssumeRole cross-account access.
// Returns raw AWS EC2 Instance JSON for transparent pass-through.
func (c *InstanceTypeClient) ListAssumeRoleInstance(kt *kit.Kit,
	request *instancetype.AwsAssumeRoleInstanceListReq) (json.RawMessage, error) {

	resp := new(instancetype.AwsAssumeRoleInstanceListResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/assume_role/instances/list").
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

// ListAssumeRoleMetricData queries CloudWatch metric data via AssumeRole cross-account access.
func (c *InstanceTypeClient) ListAssumeRoleMetricData(kt *kit.Kit,
	request *instancetype.AwsAssumeRoleGetMetricDataReq) ([]*instancetype.MetricDataResultItem, error) {

	resp := new(instancetype.AwsAssumeRoleGetMetricDataResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/assume_role/cloudwatch/metric_data/list").
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

// ListAssumeRoleMetrics lists available CloudWatch metrics via AssumeRole cross-account access.
// Returns raw AWS CloudWatch Metric JSON for transparent pass-through.
func (c *InstanceTypeClient) ListAssumeRoleMetrics(kt *kit.Kit,
	request *instancetype.AwsAssumeRoleListMetricsReq) (json.RawMessage, error) {

	resp := new(instancetype.AwsAssumeRoleListMetricsResp)

	err := c.client.Post().
		WithContext(kt.Ctx).
		Body(request).
		SubResourcef("/assume_role/cloudwatch/metrics/list").
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
