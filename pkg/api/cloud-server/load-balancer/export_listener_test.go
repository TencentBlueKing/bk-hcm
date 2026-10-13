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

package cslb

import (
	"fmt"
	"testing"

	"hcm/pkg/criteria/constant"

	"github.com/stretchr/testify/assert"
)

// genExportListeners generate listeners param, every element has lbCount unique lb id with lblCount listener id
func genExportListeners(lbCount, lblCount int) []ExportListener {
	listeners := make([]ExportListener, 0, lbCount)
	for i := 0; i < lbCount; i++ {
		lblIDs := make([]string, 0, lblCount)
		for j := 0; j < lblCount; j++ {
			lblIDs = append(lblIDs, fmt.Sprintf("lbl-%d-%d", i, j))
		}
		listeners = append(listeners, ExportListener{LbID: fmt.Sprintf("lb-%d", i), LblIDs: lblIDs})
	}

	return listeners
}

func TestExportListenerReq_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		listeners []ExportListener
		wantErr   bool
	}{
		{
			name:      "listeners为空",
			listeners: nil,
			wantErr:   true,
		},
		{
			name:      "跳过数量限制时lb_id为空",
			listeners: []ExportListener{{LblIDs: []string{"lbl-1"}}},
			wantErr:   true,
		},
		{
			name:      "不跳过数量限制时lb_id为空",
			listeners: append(genExportListeners(constant.ExportSkipLimitLbCount+1, 1), ExportListener{}),
			wantErr:   true,
		},
		{
			name:      "达到阈值的负载均衡数量，单个元素监听器数量超过批量上限",
			listeners: genExportListeners(constant.ExportSkipLimitLbCount, 3000),
			wantErr:   false,
		},
		{
			name:      "超过阈值的负载均衡数量，单个元素监听器数量超过批量上限",
			listeners: genExportListeners(constant.ExportSkipLimitLbCount+1, constant.BatchOperationMaxLimit+1),
			wantErr:   true,
		},
		{
			name:      "超过阈值的负载均衡数量，监听器数量总数超过参数限制",
			listeners: genExportListeners(constant.ExportListenerParamLimit/50+1, 50),
			wantErr:   true,
		},
		{
			name:      "超过阈值的负载均衡数量，各项数量未超过限制",
			listeners: genExportListeners(constant.ExportSkipLimitLbCount+1, constant.BatchOperationMaxLimit),
			wantErr:   false,
		},
		{
			name:      "超过阈值的负载均衡数量，元素数量超过参数限制",
			listeners: genExportListeners(constant.ExportListenerParamLimit+1, 0),
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExportListenerReq{Listeners: tc.listeners}
			err := req.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestExportListenerReq_SkipCountLimit(t *testing.T) {
	testCases := []struct {
		name      string
		listeners []ExportListener
		expected  bool
	}{
		{
			name:      "勾选1个负载均衡",
			listeners: genExportListeners(1, 1),
			expected:  true,
		},
		{
			name:      "勾选数量达到阈值",
			listeners: genExportListeners(constant.ExportSkipLimitLbCount, 1),
			expected:  true,
		},
		{
			name:      "勾选数量超过阈值",
			listeners: genExportListeners(constant.ExportSkipLimitLbCount+1, 1),
			expected:  false,
		},
		{
			name: "同一负载均衡拆分为多个元素，去重后不超过阈值",
			listeners: []ExportListener{
				{LbID: "lb-1", LblIDs: []string{"lbl-1"}},
				{LbID: "lb-1", LblIDs: []string{"lbl-2"}},
				{LbID: "lb-1", LblIDs: []string{"lbl-3"}},
				{LbID: "lb-1", LblIDs: []string{"lbl-4"}},
				{LbID: "lb-1", LblIDs: []string{"lbl-5"}},
				{LbID: "lb-2", LblIDs: []string{"lbl-6"}},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExportListenerReq{Listeners: tc.listeners}
			assert.Equal(t, tc.expected, req.SkipCountLimit())
		})
	}
}

func TestExportListenerReq_GetAllLbIDs(t *testing.T) {
	testCases := []struct {
		name      string
		listeners []ExportListener
		expected  []string
	}{
		{
			name: "同一负载均衡拆分为多个元素时去重",
			listeners: []ExportListener{
				{LbID: "lb-1", LblIDs: []string{"lbl-1"}},
				{LbID: "lb-1", LblIDs: []string{"lbl-2"}},
			},
			expected: []string{"lb-1"},
		},
		{
			name: "不同负载均衡各自保留",
			listeners: []ExportListener{
				{LbID: "lb-1"},
				{LbID: "lb-2", LblIDs: []string{"lbl-1"}},
			},
			expected: []string{"lb-1", "lb-2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExportListenerReq{Listeners: tc.listeners}
			assert.ElementsMatch(t, tc.expected, req.GetAllLbIDs())
		})
	}
}
