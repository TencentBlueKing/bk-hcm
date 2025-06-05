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

// Package bkbase ...
package bkbase

import (
	"context"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/thirdparty/api-gateway/discovery"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/ssl"
	"hcm/pkg/tools/uuid"

	"github.com/prometheus/client_golang/prometheus"
)

// Client bkbase client
type Client interface {
	QuerySync(ctx context.Context, req *QuerySyncReq) (*QuerySyncResp, error)
}

// NewClient new bkbase client.
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}
	c := &client.Capability{
		Client: cli,
		Discover: &discovery.Discovery{
			Name:    "bkbase",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/v3")
	return &bkbaseCli{
		client: restCli,
		config: cfg,
	}, nil
}

type bkbaseCli struct {
	config *cc.ApiGateway
	client rest.ClientInterface
}

func (h *bkbaseCli) getAuth() string {
	auth := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\", "+
		"\"bk_ticket\":\"%s\"}", h.config.AppCode, h.config.AppSecret, h.config.User, h.config.BkTicket)

	return auth
}

// QuerySync query sync bkbase data
func (h *bkbaseCli) QuerySync(ctx context.Context, req *QuerySyncReq) (*QuerySyncResp, error) {
	resp := new(QuerySyncResp)
	header := http.Header{}
	header.Set(constant.RidKey, uuid.UUID())
	header.Set(constant.BKGWAuthKey, h.getAuth())
	err := h.client.Post().
		SubResourcef("/queryengine/query_sync").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}
	if resp.Result != true || resp.Code != CodeSuccess {
		return nil, fmt.Errorf("query sync bkbase data failed, result: %v, code: %s, msg: %s", resp.Result, resp.Code,
			resp.Message)
	}
	return resp, nil
}

// QuerySql generic sql query method, call QuerySync method of given BKBase client
func QuerySql[T any](bkBaseCli Client, kt *kit.Kit, sql string) ([]T, error) {

	cfg := cc.CloudServer().CloudSelection.BkBase
	req := QuerySyncReq{
		AuthMethod: "token",
		DataToken:  cfg.DataToken,
		AppCode:    cfg.AppCode,
		AppSecret:  cfg.AppSecret,
		Sql:        sql,
	}
	logs.V(3).Infof("querying BKBase, sql: \n%s\n, rid: %s", sql, kt.Rid)
	resp, err := bkBaseCli.QuerySync(kt.Ctx, &req)
	if err != nil {
		logs.Errorf("fail to query bkbase, err: %v, sql: %s, rid: %s", err, sql, kt.Rid)
		return nil, err
	}
	var ret []T
	if err := json.Unmarshal(resp.Data.List, &ret); err != nil {
		logs.Errorf("fail to decode bkbase result, err: %v, resp: %+v, rid: %s", err, resp, kt.Rid)
		return nil, err
	}

	return ret, nil
}
