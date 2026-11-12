/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package service

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFilterRequest(t *testing.T, method, requestURI string) *restful.Request {
	t.Helper()

	parsed, err := url.ParseRequestURI(requestURI)
	require.NoError(t, err)

	httpReq := &http.Request{
		Method:     method,
		URL:        parsed,
		RequestURI: requestURI,
		Header:     make(http.Header),
	}
	return restful.NewRequest(httpReq)
}

func TestIsSystemAdminRequestUsesDecodedPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		requestURI string
		want       bool
	}{
		{
			name:       "literal admin path",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/admin/system/init",
			want:       true,
		},
		{
			name:       "encoded slash still matches decoded path",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/admin%2Fsystem%2Finit?x=/api/v1/cloud/applications/approve",
			want:       true,
		},
		{
			name:       "dot segments normalize into admin prefix",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/public/../admin/system/init",
			want:       true,
		},
		{
			name:       "query string does not count as path",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/accounts/list?x=/api/v1/cloud/admin/system/init",
			want:       false,
		},
		{
			name:       "adjacent path is not the prefix",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/admin/system-extra",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := newFilterRequest(t, tt.method, tt.requestURI)
			assert.Equal(t, tt.want, isSystemAdminRequest(req))
		})
	}
}

func TestIsITSMCallbackRequestIgnoresRawRequestURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		requestURI string
		want       bool
	}{
		{
			name:       "exact approve callback",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/applications/approve",
			want:       true,
		},
		{
			name:       "query string cannot impersonate callback",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/admin%2Fsystem%2Finit?x=/api/v1/cloud/applications/approve",
			want:       false,
		},
		{
			name:       "non-post method",
			method:     http.MethodGet,
			requestURI: "/api/v1/cloud/applications/approve",
			want:       false,
		},
		{
			name:       "longer path is not the callback",
			method:     http.MethodPost,
			requestURI: "/api/v1/cloud/applications/approve/extra",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := newFilterRequest(t, tt.method, tt.requestURI)
			assert.Equal(t, tt.want, isITSMCallbackRequest(req))
		})
	}
}
