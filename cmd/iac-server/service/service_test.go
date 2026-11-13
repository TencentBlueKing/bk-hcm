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

package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/handler"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/ctl"
	"hcm/pkg/runtime/ctl/cmd"
	"hcm/pkg/runtime/shutdown"

	"github.com/stretchr/testify/assert"
)

func buildTestMux(svc *Service) *http.ServeMux {
	_ = ctl.LoadCtl(cmd.WithLog())
	root := http.NewServeMux()
	root.HandleFunc("/", svc.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", svc.Healthz)
	root.HandleFunc("/alivez", svc.Alivez)
	handler.SetCommonHandler(root)
	return root
}

// TestService_ProbesAndMetricsHealthy tests /alivez, /healthz, /metrics and 404 route during normal runtime (AC-005, AC-006, AC-008).
func TestService_ProbesAndMetricsHealthy(t *testing.T) {
	cc.InitService(cc.IacServerName)
	cc.InitRuntime(&cc.IacServerSetting{})
	metrics.InitMetrics("127.0.0.1:9605")

	origCheck := healthzCheck
	defer func() {
		healthzCheck = origCheck
	}()

	svc := &Service{}
	mux := buildTestMux(svc)

	testCases := []struct {
		name           string
		path           string
		mockHealthErr  error
		expectedStatus int
		expectedCode   int32
		expectedMsgSub string
	}{
		{
			name:           "alivez returns 200 and alive",
			path:           "/alivez",
			expectedStatus: http.StatusOK,
			expectedCode:   errf.OK,
			expectedMsgSub: "alive",
		},
		{
			name:           "healthz returns 200 and healthy when etcd is healthy",
			path:           "/healthz",
			mockHealthErr:  nil,
			expectedStatus: http.StatusOK,
			expectedCode:   errf.OK,
			expectedMsgSub: "healthy",
		},
		{
			name:           "healthz returns UnHealthy when etcd check fails",
			path:           "/healthz",
			mockHealthErr:  errors.New("context deadline exceeded"),
			expectedStatus: http.StatusOK,
			expectedCode:   errf.UnHealthy,
			expectedMsgSub: "serviced healthz error, context deadline exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			healthzCheck = func(_ context.Context, _ cc.Service) error {
				return tc.mockHealthErr
			}

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			resp := new(rest.BaseResp)
			err := json.Unmarshal(rec.Body.Bytes(), resp)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCode, resp.Code)
			assert.True(t, strings.Contains(resp.Message, tc.expectedMsgSub),
				"message %q should contain %q", resp.Message, tc.expectedMsgSub)
		})
	}

	t.Run("metrics endpoint returns 200 and prometheus text", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, strings.Contains(rec.Body.String(), "go_goroutines"))
	})

	t.Run("unregistered route under /api/v1/iac safely returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/iac/non-existent-path", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// TestService_ProbesDuringShutdown tests that /alivez and /healthz return 503 and UnHealthy when shutting down (AC-006).
func TestService_ProbesDuringShutdown(t *testing.T) {
	cc.InitService(cc.IacServerName)
	svc := &Service{}
	mux := buildTestMux(svc)

	shutdown.SignalShutdownGracefully()
	assert.Eventually(t, func() bool {
		return shutdown.IsShuttingDown()
	}, time.Second, 10*time.Millisecond)

	testCases := []struct {
		name string
		path string
	}{
		{name: "alivez returns 503 when shutting down", path: "/alivez"},
		{name: "healthz returns 503 when shutting down", path: "/healthz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
			resp := new(rest.BaseResp)
			err := json.Unmarshal(rec.Body.Bytes(), resp)
			assert.NoError(t, err)
			assert.Equal(t, int32(errf.UnHealthy), resp.Code)
		})
	}
}
