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

package lblogic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/rest/client"
)

// staticServerDiscovery 测试用静态服务发现，始终返回同一个 httptest 地址。
type staticServerDiscovery struct {
	addr string
}

// GetServers ...
func (d staticServerDiscovery) GetServers() ([]string, error) {
	return []string{d.addr}, nil
}

// newTestDataServiceClient 启动一个 httptest server 承载给定 handler，返回一个指向该 server 的
// data-service 客户端；测试结束时自动关闭 server。
func newTestDataServiceClient(t *testing.T, handler http.Handler) *dataservice.Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cap := &client.Capability{
		Client:   srv.Client(),
		Discover: staticServerDiscovery{addr: srv.URL},
	}
	return dataservice.NewClient(cap, "v1")
}

// writeOKResp 写出一个 code=0 的通用响应体。
func writeOKResp(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	body := map[string]any{"code": 0, "message": "", "data": data}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("encode test response failed: %v", err)
	}
}
