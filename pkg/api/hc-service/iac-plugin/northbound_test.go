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

package iacplugin

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
)

func validCommonReq() CommonInstanceReq {
	return CommonInstanceReq{
		Plugin:           "bkcmdb",
		PluginInstanceID: "00000002",
		TypeName:         "bkcmdb_host",
	}
}

func TestHealthAndGetSchemaReq_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		fn      func() error
		wantErr bool
	}{
		{
			name:    "valid health req",
			fn:      func() error { return (&HealthReq{Plugin: "bkcmdb"}).Validate() },
			wantErr: false,
		},
		{
			name:    "nil health req",
			fn:      func() error { var r *HealthReq; return r.Validate() },
			wantErr: true,
		},
		{
			name:    "invalid plugin name in health req",
			fn:      func() error { return (&HealthReq{Plugin: "bad/name"}).Validate() },
			wantErr: true,
		},
		{
			name:    "valid get schema req without version",
			fn:      func() error { return (&GetSchemaReq{Plugin: "bkcmdb"}).Validate() },
			wantErr: false,
		},
		{
			name:    "valid get schema req with version",
			fn:      func() error { return (&GetSchemaReq{Plugin: "bkcmdb", PluginVersion: "1.2.0"}).Validate() },
			wantErr: false,
		},
		{
			name:    "invalid get schema req version",
			fn:      func() error { return (&GetSchemaReq{Plugin: "bkcmdb", PluginVersion: "1.2"}).Validate() },
			wantErr: true,
		},
		{
			name:    "nil get schema req",
			fn:      func() error { var r *GetSchemaReq; return r.Validate() },
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestExecuteAndPlanReq_Validate(t *testing.T) {
	validB64 := base64.StdEncoding.EncodeToString([]byte("private-bytes"))
	testCases := []struct {
		name       string
		priorState map[string]any
		config     map[string]any
		recompute  []string
		private    *string
		wantErr    bool
	}{
		{
			name:       "create action",
			priorState: nil,
			config:     map[string]any{"bk_biz_id": 213},
		},
		{
			name:       "update action with recompute and base64 private",
			priorState: map[string]any{"bk_biz_id": 213},
			config:     map[string]any{"bk_biz_id": 213},
			recompute:  []string{"bk_module_id"},
			private:    cvt.ValToPtr(validB64),
		},
		{
			name:       "delete action",
			priorState: map[string]any{"bk_biz_id": 213},
			config:     nil,
		},
		{
			name:       "both prior_state and config nil",
			priorState: nil,
			config:     nil,
			wantErr:    true,
		},
		{
			name:       "recompute contains empty string",
			priorState: map[string]any{"bk_biz_id": 213},
			config:     map[string]any{"bk_biz_id": 213},
			recompute:  []string{"bk_module_id", "  "},
			wantErr:    true,
		},
		{
			name:       "invalid base64 prior_private",
			priorState: map[string]any{"bk_biz_id": 213},
			config:     map[string]any{"bk_biz_id": 213},
			private:    cvt.ValToPtr("not-valid-base64!!!"),
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertExecuteAndPlanCase(t, tc.priorState, tc.config, tc.recompute, tc.private, tc.wantErr)
		})
	}
}

func assertExecuteAndPlanCase(t *testing.T, priorState, config map[string]any, recompute []string,
	private *string, wantErr bool) {

	t.Helper()
	execReq := &ExecuteReq{
		CommonInstanceReq: validCommonReq(),
		PriorState:        priorState,
		Config:            config,
		Recompute:         recompute,
		PriorPrivate:      private,
	}
	planReq := &PlanReq{
		CommonInstanceReq: validCommonReq(),
		PriorState:        priorState,
		Config:            config,
		Recompute:         recompute,
		PriorPrivate:      private,
	}
	if wantErr {
		assert.Error(t, execReq.Validate())
		assert.Error(t, planReq.Validate())
		return
	}
	assert.NoError(t, execReq.Validate())
	assert.NoError(t, planReq.Validate())
}

func TestRemainingRuntimeReqs_Validate(t *testing.T) {
	c := validCommonReq()
	testCases := []struct {
		name    string
		fn      func() error
		wantErr bool
	}{
		{
			name:    "valid result",
			fn:      func() error { return (&GetResultReq{CommonInstanceReq: c, ApplyID: "ap-1"}).Validate() },
			wantErr: false,
		},
		{
			name:    "empty apply_id in result",
			fn:      func() error { return (&GetResultReq{CommonInstanceReq: c, ApplyID: " "}).Validate() },
			wantErr: true,
		},
		{
			name:    "valid stop",
			fn:      func() error { return (&StopReq{CommonInstanceReq: c, ApplyID: "ap-1"}).Validate() },
			wantErr: false,
		},
		{
			name:    "empty apply_id in stop",
			fn:      func() error { return (&StopReq{CommonInstanceReq: c, ApplyID: ""}).Validate() },
			wantErr: true,
		},
		{
			name: "valid validate resource",
			fn: func() error {
				return (&ValidateReq{
					CommonInstanceReq: c,
					ConfigType:        ConfigTypeResource,
					Config:            map[string]any{},
				}).Validate()
			},
			wantErr: false,
		},
		{
			name: "valid validate data source",
			fn: func() error {
				return (&ValidateReq{
					CommonInstanceReq: c,
					ConfigType:        ConfigTypeDataSource,
					Config:            map[string]any{},
				}).Validate()
			},
			wantErr: false,
		},
		{
			name: "missing config_type in validate",
			fn: func() error {
				return (&ValidateReq{CommonInstanceReq: c, Config: map[string]any{}}).Validate()
			},
			wantErr: true,
		},
		{
			name: "invalid config_type in validate",
			fn: func() error {
				return (&ValidateReq{
					CommonInstanceReq: c,
					ConfigType:        "provider",
					Config:            map[string]any{},
				}).Validate()
			},
			wantErr: true,
		},
		{
			name: "nil config in validate",
			fn: func() error {
				return (&ValidateReq{CommonInstanceReq: c, ConfigType: ConfigTypeResource}).Validate()
			},
			wantErr: true,
		},
		{
			name:    "valid read_data",
			fn:      func() error { return (&ReadDataReq{CommonInstanceReq: c, Config: map[string]any{}}).Validate() },
			wantErr: false,
		},
		{
			name:    "nil config in read_data",
			fn:      func() error { return (&ReadDataReq{CommonInstanceReq: c}).Validate() },
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantErr {
				assert.Error(t, tc.fn())
				return
			}
			assert.NoError(t, tc.fn())
		})
	}
}

func TestImportRefreshAndListReqs_Validate(t *testing.T) {
	c := validCommonReq()
	testCases := []struct {
		name    string
		fn      func() error
		wantErr bool
	}{
		{
			name:    "valid import",
			fn:      func() error { return (&ImportReq{CommonInstanceReq: c, ImportID: "imp-1"}).Validate() },
			wantErr: false,
		},
		{
			name:    "empty import_id in import",
			fn:      func() error { return (&ImportReq{CommonInstanceReq: c}).Validate() },
			wantErr: true,
		},
		{
			name:    "valid refresh",
			fn:      func() error { return (&RefreshReq{CommonInstanceReq: c, State: map[string]any{"id": 1}}).Validate() },
			wantErr: false,
		},
		{
			name:    "empty state in refresh",
			fn:      func() error { return (&RefreshReq{CommonInstanceReq: c, State: map[string]any{}}).Validate() },
			wantErr: true,
		},
		{
			name:    "valid list",
			fn:      func() error { return (&ListReq{CommonInstanceReq: c, Limit: 500}).Validate() },
			wantErr: false,
		},
		{
			name:    "list limit 0",
			fn:      func() error { return (&ListReq{CommonInstanceReq: c, Limit: 0}).Validate() },
			wantErr: true,
		},
		{
			name:    "list limit 501",
			fn:      func() error { return (&ListReq{CommonInstanceReq: c, Limit: 501}).Validate() },
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantErr {
				assert.Error(t, tc.fn())
				return
			}
			assert.NoError(t, tc.fn())
		})
	}
}

func TestNorthboundRequestsZeroSecretFields(t *testing.T) {
	forbiddenJSONNames := map[string]struct{}{
		"launch_config": {},
		"secret_attrs":  {},
		"secret_key":    {},
		"secret_id":     {},
		"access_token":  {},
		"properties":    {},
	}

	reqTypes := []reflect.Type{
		reflect.TypeOf(HealthReq{}),
		reflect.TypeOf(GetSchemaReq{}),
		reflect.TypeOf(CommonInstanceReq{}),
		reflect.TypeOf(ExecuteReq{}),
		reflect.TypeOf(GetResultReq{}),
		reflect.TypeOf(StopReq{}),
		reflect.TypeOf(PlanReq{}),
		reflect.TypeOf(ValidateReq{}),
		reflect.TypeOf(ImportReq{}),
		reflect.TypeOf(RefreshReq{}),
		reflect.TypeOf(ReadDataReq{}),
		reflect.TypeOf(ListReq{}),
	}

	for _, rt := range reqTypes {
		assertNoForbiddenFields(t, rt, forbiddenJSONNames)
	}
}

func assertNoForbiddenFields(t *testing.T, rt reflect.Type, forbidden map[string]struct{}) {
	t.Helper()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			assertNoForbiddenFields(t, field.Type, forbidden)
			continue
		}
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		_, isForbidden := forbidden[jsonTag]
		assert.Falsef(t, isForbidden, "struct %s must not contain forbidden field %q", rt.Name(), jsonTag)
	}
}
