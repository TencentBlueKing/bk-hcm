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

package image

import (
	"encoding/json"
	"testing"

	csimage "hcm/pkg/api/cloud-server/image"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/runtime/filter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBizImageVisibilityFilter(t *testing.T) {
	t.Parallel()

	expr := buildBizImageVisibilityFilter(100)
	require.Equal(t, filter.Or, expr.Op)
	require.Len(t, expr.Rules, 2)

	publicSharedRule, ok := expr.Rules[0].(*filter.AtomRule)
	require.True(t, ok)
	assert.Equal(t, "type", publicSharedRule.Field)
	assert.Equal(t, filter.In.Factory(), publicSharedRule.Op)
	assert.Equal(t, []enumor.ImageTypeNormalized{enumor.ImageTypePublic, enumor.ImageTypeShared},
		publicSharedRule.Value)

	privateExpr, ok := expr.Rules[1].(*filter.Expression)
	require.True(t, ok)
	assert.Equal(t, filter.And, privateExpr.Op)
	require.Len(t, privateExpr.Rules, 2)
	assert.Equal(t, enumor.ImageTypePrivate, privateExpr.Rules[0].(*filter.AtomRule).Value)
	assert.Equal(t, int64(100), privateExpr.Rules[1].(*filter.AtomRule).Value)
}

func TestBuildBizImageListFilter_OtherVendorNoEnableCvm(t *testing.T) {
	t.Parallel()

	expr, err := buildBizImageListFilter(100, enumor.Aws, &csimage.BizImageListReq{
		Page: core.NewDefaultBasePage(),
	})
	require.NoError(t, err)

	raw, err := json.Marshal(expr)
	require.NoError(t, err)
	body := string(raw)
	assert.Contains(t, body, string(enumor.Aws))
	assert.NotContains(t, body, "extension.enable_cvm")
}

func TestBuildBizImageListFilter_PublicType(t *testing.T) {
	t.Parallel()

	expr, err := buildBizImageListFilter(200, enumor.Aws, &csimage.BizImageListReq{
		Type: enumor.ImageTypePublic,
		Page: core.NewDefaultBasePage(),
	})
	require.NoError(t, err)

	var found bool
	for _, rule := range expr.Rules {
		atom, ok := rule.(*filter.AtomRule)
		if ok && atom.Field == "type" && atom.Value == enumor.ImageTypePublic {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestBuildBizImageListFilter_SharedType(t *testing.T) {
	t.Parallel()

	expr, err := buildBizImageListFilter(200, enumor.TCloud, &csimage.BizImageListReq{
		Type: enumor.ImageTypeShared,
		Page: core.NewDefaultBasePage(),
	})
	require.NoError(t, err)

	raw, err := json.Marshal(expr)
	require.NoError(t, err)
	assert.Contains(t, string(raw), string(enumor.ImageTypeShared))
	assert.NotContains(t, string(raw), string(enumor.TCloudSharedImage))
}

func TestBuildBizImageListFilter_EmptyFilters(t *testing.T) {
	t.Parallel()

	expr, err := buildBizImageListFilter(100, enumor.TCloud, &csimage.BizImageListReq{
		Page: core.NewDefaultBasePage(),
	})
	require.NoError(t, err)
	require.Equal(t, filter.And, expr.Op)
	require.Len(t, expr.Rules, 2)
}

func TestBizImageListReq_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     *csimage.BizImageListReq
		wantErr bool
	}{
		{
			name:    "missing page",
			req:     &csimage.BizImageListReq{},
			wantErr: true,
		},
		{
			name: "invalid type",
			req: &csimage.BizImageListReq{
				Type: "invalid",
				Page: core.NewDefaultBasePage(),
			},
			wantErr: true,
		},
		{
			name: "legacy cloud-side shared type",
			req: &csimage.BizImageListReq{
				Type: enumor.ImageTypeNormalized(enumor.TCloudSharedImage),
				Page: core.NewDefaultBasePage(),
			},
			wantErr: true,
		},
		{
			name: "empty type is allowed",
			req: &csimage.BizImageListReq{
				Page: core.NewDefaultBasePage(),
			},
			wantErr: false,
		},
		{
			name: "private type",
			req: &csimage.BizImageListReq{
				Type: enumor.ImageTypePrivate,
				Page: core.NewDefaultBasePage(),
			},
			wantErr: false,
		},
		{
			name: "shared type",
			req: &csimage.BizImageListReq{
				Type: enumor.ImageTypeShared,
				Page: core.NewDefaultBasePage(),
			},
			wantErr: false,
		},
		{
			name: "public type",
			req: &csimage.BizImageListReq{
				Type: enumor.ImageTypePublic,
				Page: core.NewDefaultBasePage(),
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.req.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
