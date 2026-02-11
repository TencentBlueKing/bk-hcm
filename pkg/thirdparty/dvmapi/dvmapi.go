/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package dvmapi provides ...
package dvmapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
)

// DVMClientInterface docker vm api interface
type DVMClientInterface interface {
	// CreateDvmOrder creates cvm order
	CreateDvmOrder(ctx context.Context, header http.Header, req *OrderCreateReq) (*OrderCreateResp, error)
	// QueryDvmOrders query cvm orders
	QueryDvmOrders(ctx context.Context, header http.Header, orderId string) (*OrderQueryResp, error)
	// ListCluster list docker cluster
	ListCluster(ctx context.Context, header http.Header) ([]*DockerCluster, error)
	// ListHostInCluster list docker host in cluster
	ListHostInCluster(ctx context.Context, header http.Header, req *ListHostReq) ([]*DockerHost, error)
}

// NewDVMClientInterface creates a docker vm api instance
func NewDVMClientInterface(opts cc.DVMCli, reg prometheus.Registerer) (DVMClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "dvm api",
			servers: []string{opts.DvmApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	client := &dvmApi{
		client: rest.NewClient(c, "/"),
		opts:   &opts,
	}

	return client, nil
}

// dvmApi docker vm api interface implementation
type dvmApi struct {
	client rest.ClientInterface
	opts   *cc.DVMCli
}

func (d *dvmApi) sign(rtx string) (tokenString string, err error) {
	// The token content.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  d.opts.SecretID,
		"user": rtx,
		"iat":  time.Now().Add(-1 * time.Hour).Unix(),
	})
	// Sign the token with the specified secret.
	tokenString, err = token.SignedString([]byte(d.opts.SecretKey))
	return
}

// CreateDvmOrder creates dvm order
func (d *dvmApi) CreateDvmOrder(ctx context.Context, header http.Header, req *OrderCreateReq) (*OrderCreateResp,
	error) {

	subPath := "/dockervm/container"

	token, signErr := d.sign(d.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	resp := &struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    *OrderCreateResp `json:"data"`
	}{}
	err := d.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("failed to create dvm order, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, err
}

// QueryDvmOrders query dvm orders
func (d *dvmApi) QueryDvmOrders(ctx context.Context, header http.Header, orderId string) (*OrderQueryResp, error) {
	subPath := "/dockervm/container/bills/%s"

	token, signErr := d.sign(d.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	resp := &struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    *OrderQueryResp `json:"data"`
	}{}
	err := d.client.Get().
		WithContext(ctx).
		SubResourcef(subPath, orderId).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("failed to query dvm order, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, err
}

// ListCluster list docker cluster
func (d *dvmApi) ListCluster(ctx context.Context, header http.Header) ([]*DockerCluster, error) {
	subPath := "/dockervm/misc/set-info"

	token, signErr := d.sign(d.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	resp := &struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    []*DockerCluster `json:"data"`
	}{}
	err := d.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("failed to list cluster, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, err
}

// ListHostInCluster list docker host in cluster
func (d *dvmApi) ListHostInCluster(ctx context.Context, header http.Header, req *ListHostReq) ([]*DockerHost, error) {
	subPath := "/dockervm/misc/minion"

	token, signErr := d.sign(d.opts.Operator)
	if signErr != nil {
		return nil, signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	params := make(map[string]string, 0)
	params["setId"] = req.SetId
	params["deviceClass"] = req.DeviceClass
	params["idle"] = strconv.Itoa(0)
	params["cores"] = strconv.Itoa(req.Cores)
	params["memory"] = strconv.Itoa(req.Memory)
	params["disk"] = strconv.Itoa(req.Disk)
	params["hostRole"] = req.HostRole

	resp := &struct {
		Code    int           `json:"code"`
		Message string        `json:"message"`
		Data    []*DockerHost `json:"data"`
	}{}
	err := d.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(header).
		WithParams(params).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("failed to list host in cluster, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, err
}
