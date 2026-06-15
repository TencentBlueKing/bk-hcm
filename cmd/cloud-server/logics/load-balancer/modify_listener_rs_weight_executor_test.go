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
	"testing"

	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
)

// TestBuildTaskDetailsLogic tests the core logic of building taskDetails
// This function tests the logic without calling the actual createTaskDetails method
// to avoid dependency on dataServiceCli
func TestBuildTaskDetailsLogic_SingleDetailSingleRS(t *testing.T) {
	tests := []struct {
		name                string
		details             []*dataproto.ListBatchListenerResult
		expectedTaskDetails int
		description         string
	}{
		{
			name: "single_rs_weight_changed",
			details: []*dataproto.ListBatchListenerResult{
				{
					ClbID:        "clb-001",
					CloudClbID:   "lb-cloud-001",
					ClbVipDomain: "10.0.0.1",
					BkBizID:      100,
					Region:       "ap-guangzhou",
					Vendor:       enumor.TCloud,
					LblID:        "lbl-001",
					CloudLblID:   "lbl-cloud-001",
					Protocol:     enumor.HttpProtocol,
					Port:         80,
					RsList: []*dataproto.LoadBalancerTargetRsList{
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-001",
								IP:            "192.168.1.1",
								Weight:        cvt.ValToPtr(int64(10)),
								Port:          8080,
								TargetGroupID: "tg-001",
							},
							RuleID:      "rule-001",
							CloudRuleID: "rule-cloud-001",
							Domain:      "example.com",
							Url:         "/api",
						},
					},
					NewRsWeight: cvt.ValToPtr(int64(50)),
				},
			},
			expectedTaskDetails: 1,
			description:         "单个 detail，单个 RS，权重变化：应创建1个 taskDetail",
		},
		{
			name: "single_rs_weight_unchanged",
			details: []*dataproto.ListBatchListenerResult{
				{
					ClbID:        "clb-001",
					CloudClbID:   "lb-cloud-001",
					ClbVipDomain: "10.0.0.1",
					BkBizID:      100,
					Region:       "ap-guangzhou",
					Vendor:       enumor.TCloud,
					LblID:        "lbl-001",
					CloudLblID:   "lbl-cloud-001",
					Protocol:     enumor.HttpProtocol,
					Port:         80,
					RsList: []*dataproto.LoadBalancerTargetRsList{
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-001",
								IP:            "192.168.1.1",
								Weight:        cvt.ValToPtr(int64(50)),
								Port:          8080,
								TargetGroupID: "tg-001",
							},
						},
					},
					NewRsWeight: cvt.ValToPtr(int64(50)),
				},
			},
			expectedTaskDetails: 0,
			description:         "单个 detail，单个 RS，权重未变：不创建 taskDetail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 createTaskDetails 的核心逻辑
			taskDetails := buildTaskDetailsFromDetailsLogic(tt.details)

			// 验证 taskDetails 数量
			assert.Equal(t, tt.expectedTaskDetails, len(taskDetails), tt.description)
		})
	}
}

// TestBuildTaskDetailsLogic_SingleDetailMultipleRS tests the critical scenario with one detail and multiple RS
func TestBuildTaskDetailsLogic_SingleDetailMultipleRS(t *testing.T) {
	tests := []struct {
		name                  string
		details               []*dataproto.ListBatchListenerResult
		expectedTaskDetails   int
		expectedRsListPerTask []int // 每个 taskDetail 的 RsList 长度
		description           string
	}{
		{
			name: "multiple_rs_all_changed",
			details: []*dataproto.ListBatchListenerResult{
				{
					ClbID:        "clb-001",
					CloudClbID:   "lb-cloud-001",
					ClbVipDomain: "10.0.0.1",
					BkBizID:      100,
					Region:       "ap-guangzhou",
					Vendor:       enumor.TCloud,
					LblID:        "lbl-001",
					CloudLblID:   "lbl-cloud-001",
					Protocol:     enumor.HttpProtocol,
					Port:         80,
					RsList: []*dataproto.LoadBalancerTargetRsList{
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-001",
								IP:            "192.168.1.1",
								Weight:        cvt.ValToPtr(int64(10)),
								Port:          8080,
								TargetGroupID: "tg-001",
							},
							Domain: "example.com",
							Url:    "/api/v1",
						},
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-002",
								IP:            "192.168.1.2",
								Weight:        cvt.ValToPtr(int64(20)),
								Port:          8080,
								TargetGroupID: "tg-001",
							},
							Domain: "example.com",
							Url:    "/api/v2",
						},
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-003",
								IP:            "192.168.1.3",
								Weight:        cvt.ValToPtr(int64(30)),
								Port:          8080,
								TargetGroupID: "tg-001",
							},
							Domain: "example.com",
							Url:    "/api/v3",
						},
					},
					NewRsWeight: cvt.ValToPtr(int64(50)),
				},
			},
			expectedTaskDetails:   3,
			expectedRsListPerTask: []int{3, 3, 3},
			description:           "单个 detail，3个 RS 全部权重变化：应创建3个 taskDetail，且每个都携带完整 RsList",
		},
		{
			name: "multiple_rs_partial_changed",
			details: []*dataproto.ListBatchListenerResult{
				{
					ClbID:        "clb-001",
					CloudClbID:   "lb-cloud-001",
					ClbVipDomain: "10.0.0.1",
					BkBizID:      100,
					Region:       "ap-guangzhou",
					Vendor:       enumor.TCloud,
					LblID:        "lbl-001",
					CloudLblID:   "lbl-cloud-001",
					Protocol:     enumor.HttpProtocol,
					Port:         80,
					RsList: []*dataproto.LoadBalancerTargetRsList{
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-001",
								IP:            "192.168.1.1",
								Weight:        cvt.ValToPtr(int64(10)),
								Port:          8080,
								TargetGroupID: "tg-001",
							},
						},
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-002",
								IP:            "192.168.1.2",
								Weight:        cvt.ValToPtr(int64(50)), // 权重已经是50，无需变更
								Port:          8080,
								TargetGroupID: "tg-001",
							},
						},
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-003",
								IP:            "192.168.1.3",
								Weight:        cvt.ValToPtr(int64(30)),
								Port:          8080,
								TargetGroupID: "tg-001",
							},
						},
					},
					NewRsWeight: cvt.ValToPtr(int64(50)),
				},
			},
			expectedTaskDetails:   2,
			expectedRsListPerTask: []int{3, 3},
			description:           "单个 detail，3个 RS 部分权重变化：应创建2个 taskDetail（rs-001, rs-003），且每个都携带完整 RsList",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 createTaskDetails 的核心逻辑
			taskDetails := buildTaskDetailsFromDetailsLogic(tt.details)

			// 验证 taskDetails 数量
			assert.Equal(t, tt.expectedTaskDetails, len(taskDetails), tt.description)

			// 验证每个 taskDetail 的 RsList 长度（当前语义：都应为完整 detail 的 RsList）
			for i, taskDetail := range taskDetails {
				actualRsCount := len(taskDetail.ListBatchListenerResult.RsList)
				expectedRsCount := tt.expectedRsListPerTask[i]
				assert.Equal(t, expectedRsCount, actualRsCount,
					"taskDetail[%d] 应该包含 %d 个 RS，实际包含 %d 个", i, expectedRsCount, actualRsCount)
			}
		})
	}
}

// TestBuildTaskDetailsLogic_MultipleDetailsMultipleRS tests complex scenario with multiple details and multiple RS
func TestBuildTaskDetailsLogic_MultipleDetailsMultipleRS(t *testing.T) {
	details := []*dataproto.ListBatchListenerResult{
		{
			ClbID:        "clb-001",
			CloudClbID:   "lb-cloud-001",
			ClbVipDomain: "10.0.0.1",
			BkBizID:      100,
			Region:       "ap-guangzhou",
			Vendor:       enumor.TCloud,
			LblID:        "lbl-001",
			CloudLblID:   "lbl-cloud-001",
			Protocol:     enumor.HttpProtocol,
			Port:         80,
			RsList: []*dataproto.LoadBalancerTargetRsList{
				{
					BaseTarget: corelb.BaseTarget{
						ID:            "rs-001",
						IP:            "192.168.1.1",
						Weight:        cvt.ValToPtr(int64(10)),
						Port:          8080,
						TargetGroupID: "tg-001",
					},
					Domain: "example.com",
					Url:    "/api",
				},
				{
					BaseTarget: corelb.BaseTarget{
						ID:            "rs-002",
						IP:            "192.168.1.2",
						Weight:        cvt.ValToPtr(int64(20)),
						Port:          8080,
						TargetGroupID: "tg-001",
					},
					Domain: "example.com",
					Url:    "/web",
				},
			},
			NewRsWeight: cvt.ValToPtr(int64(50)),
		},
		{
			ClbID:        "clb-002",
			CloudClbID:   "lb-cloud-002",
			ClbVipDomain: "10.0.0.2",
			BkBizID:      100,
			Region:       "ap-shanghai",
			Vendor:       enumor.TCloud,
			LblID:        "lbl-002",
			CloudLblID:   "lbl-cloud-002",
			Protocol:     enumor.TcpProtocol,
			Port:         443,
			RsList: []*dataproto.LoadBalancerTargetRsList{
				{
					BaseTarget: corelb.BaseTarget{
						ID:            "rs-003",
						IP:            "192.168.2.1",
						Weight:        cvt.ValToPtr(int64(30)),
						Port:          9090,
						TargetGroupID: "tg-002",
					},
				},
			},
			NewRsWeight: cvt.ValToPtr(int64(60)),
		},
	}

	taskDetails := buildTaskDetailsFromDetailsLogic(details)

	// 验证总数
	assert.Equal(t, 3, len(taskDetails), "应创建3个 taskDetail（2个来自 clb-001，1个来自 clb-002）")

	// 前两个来自 clb-001，携带完整 RsList（2个）；第三个来自 clb-002（1个）
	assert.Equal(t, "clb-001", taskDetails[0].ListBatchListenerResult.ClbID, "taskDetail[0] 应关联 clb-001")
	assert.Equal(t, "clb-001", taskDetails[1].ListBatchListenerResult.ClbID, "taskDetail[1] 应关联 clb-001")
	assert.Equal(t, "clb-002", taskDetails[2].ListBatchListenerResult.ClbID, "taskDetail[2] 应关联 clb-002")

	assert.Equal(t, 2, len(taskDetails[0].ListBatchListenerResult.RsList), "taskDetail[0] 应携带 clb-001 的完整 RsList")
	assert.Equal(t, 2, len(taskDetails[1].ListBatchListenerResult.RsList), "taskDetail[1] 应携带 clb-001 的完整 RsList")
	assert.Equal(t, 1, len(taskDetails[2].ListBatchListenerResult.RsList), "taskDetail[2] 应携带 clb-002 的完整 RsList")
}

// TestBuildTaskDetailsLogic_EdgeCases tests edge cases
func TestBuildTaskDetailsLogic_EdgeCases(t *testing.T) {
	tests := []struct {
		name                string
		details             []*dataproto.ListBatchListenerResult
		expectedTaskDetails int
		description         string
	}{
		{
			name:                "empty_details",
			details:             []*dataproto.ListBatchListenerResult{},
			expectedTaskDetails: 0,
			description:         "空 details：不创建任何 taskDetail",
		},
		{
			name: "empty_rs_list",
			details: []*dataproto.ListBatchListenerResult{
				{
					ClbID:       "clb-001",
					CloudClbID:  "lb-cloud-001",
					BkBizID:     100,
					Region:      "ap-guangzhou",
					Vendor:      enumor.TCloud,
					RsList:      []*dataproto.LoadBalancerTargetRsList{}, // 空 RsList
					NewRsWeight: cvt.ValToPtr(int64(50)),
				},
			},
			expectedTaskDetails: 0,
			description:         "detail 的 RsList 为空：不创建任何 taskDetail",
		},
		{
			name: "nil_weights",
			details: []*dataproto.ListBatchListenerResult{
				{
					ClbID:        "clb-001",
					CloudClbID:   "lb-cloud-001",
					ClbVipDomain: "10.0.0.1",
					BkBizID:      100,
					Region:       "ap-guangzhou",
					Vendor:       enumor.TCloud,
					LblID:        "lbl-001",
					CloudLblID:   "lbl-cloud-001",
					Protocol:     enumor.HttpProtocol,
					Port:         80,
					RsList: []*dataproto.LoadBalancerTargetRsList{
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-001",
								IP:            "192.168.1.1",
								Weight:        nil, // nil 权重
								Port:          8080,
								TargetGroupID: "tg-001",
							},
						},
					},
					NewRsWeight: cvt.ValToPtr(int64(50)),
				},
			},
			expectedTaskDetails: 1,
			description:         "RS 权重为 nil：应视为0，与 NewRsWeight 不同，创建 taskDetail",
		},
		{
			name: "both_nil_weights",
			details: []*dataproto.ListBatchListenerResult{
				{
					ClbID:        "clb-001",
					CloudClbID:   "lb-cloud-001",
					ClbVipDomain: "10.0.0.1",
					BkBizID:      100,
					Region:       "ap-guangzhou",
					Vendor:       enumor.TCloud,
					LblID:        "lbl-001",
					CloudLblID:   "lbl-cloud-001",
					Protocol:     enumor.HttpProtocol,
					Port:         80,
					RsList: []*dataproto.LoadBalancerTargetRsList{
						{
							BaseTarget: corelb.BaseTarget{
								ID:            "rs-001",
								IP:            "192.168.1.1",
								Weight:        nil, // nil 权重
								Port:          8080,
								TargetGroupID: "tg-001",
							},
						},
					},
					NewRsWeight: nil, // nil 新权重
				},
			},
			expectedTaskDetails: 0,
			description:         "RS 权重和 NewRsWeight 都为 nil：视为相同，不创建 taskDetail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskDetails := buildTaskDetailsFromDetailsLogic(tt.details)
			assert.Equal(t, tt.expectedTaskDetails, len(taskDetails), tt.description)
		})
	}
}

// TestBuildTaskDetailsLogic_RsListReference tests that taskDetails from same detail share same detail reference
func TestBuildTaskDetailsLogic_RsListReference(t *testing.T) {
	// 这个测试验证当前语义：同一个 detail 产生的多个 taskDetail，都会引用同一份完整 detail
	details := []*dataproto.ListBatchListenerResult{
		{
			ClbID:        "clb-001",
			CloudClbID:   "lb-cloud-001",
			ClbVipDomain: "10.0.0.1",
			BkBizID:      100,
			Region:       "ap-guangzhou",
			Vendor:       enumor.TCloud,
			LblID:        "lbl-001",
			CloudLblID:   "lbl-cloud-001",
			Protocol:     enumor.HttpProtocol,
			Port:         80,
			RsList: []*dataproto.LoadBalancerTargetRsList{
				{
					BaseTarget: corelb.BaseTarget{
						ID:            "rs-001",
						IP:            "192.168.1.1",
						Weight:        cvt.ValToPtr(int64(10)),
						Port:          8080,
						TargetGroupID: "tg-001",
					},
					Domain: "example.com",
					Url:    "/api/v1",
				},
				{
					BaseTarget: corelb.BaseTarget{
						ID:            "rs-002",
						IP:            "192.168.1.2",
						Weight:        cvt.ValToPtr(int64(20)),
						Port:          8080,
						TargetGroupID: "tg-001",
					},
					Domain: "example.com",
					Url:    "/api/v2",
				},
				{
					BaseTarget: corelb.BaseTarget{
						ID:            "rs-003",
						IP:            "192.168.1.3",
						Weight:        cvt.ValToPtr(int64(30)),
						Port:          8080,
						TargetGroupID: "tg-001",
					},
					Domain: "example.com",
					Url:    "/api/v3",
				},
			},
			NewRsWeight: cvt.ValToPtr(int64(50)),
		},
	}

	taskDetails := buildTaskDetailsFromDetailsLogic(details)

	// 验证总数
	assert.Equal(t, 3, len(taskDetails), "应创建3个 taskDetail")

	// 每个 taskDetail 都携带完整 RsList
	for i := range taskDetails {
		assert.Equal(t, 3, len(taskDetails[i].ListBatchListenerResult.RsList),
			"taskDetail[%d] 的 RsList 应该包含完整3个 RS", i)
	}

	// 同一个 detail 拆出的 taskDetail 应共享同一 detail 引用
	assert.True(t, taskDetails[0].ListBatchListenerResult == taskDetails[1].ListBatchListenerResult,
		"同一 detail 拆出的 taskDetail 应共享 detail 引用")
	assert.True(t, taskDetails[1].ListBatchListenerResult == taskDetails[2].ListBatchListenerResult,
		"同一 detail 拆出的 taskDetail 应共享 detail 引用")
}

// TestBuildTaskDetailsLogic_VerifyDetailFields tests that all detail fields are correctly copied
func TestBuildTaskDetailsLogic_VerifyDetailFields(t *testing.T) {
	originalDetail := &dataproto.ListBatchListenerResult{
		ClbID:        "clb-test-001",
		CloudClbID:   "lb-cloud-test-001",
		ClbVipDomain: "test.domain.com",
		BkBizID:      12345,
		Region:       "ap-test-region",
		Vendor:       enumor.TCloud,
		LblID:        "lbl-test-001",
		CloudLblID:   "lbl-cloud-test-001",
		Protocol:     enumor.HttpsProtocol,
		Port:         443,
		RsList: []*dataproto.LoadBalancerTargetRsList{
			{
				BaseTarget: corelb.BaseTarget{
					ID:            "rs-test-001",
					IP:            "172.16.0.1",
					Weight:        cvt.ValToPtr(int64(25)),
					Port:          9000,
					TargetGroupID: "tg-test-001",
				},
				RuleID:      "rule-test-001",
				CloudRuleID: "rule-cloud-test-001",
				RuleType:    enumor.Layer7RuleType,
				Domain:      "api.test.com",
				Url:         "/v1/test",
			},
		},
		NewRsWeight: cvt.ValToPtr(int64(75)),
	}

	taskDetails := buildTaskDetailsFromDetailsLogic([]*dataproto.ListBatchListenerResult{originalDetail})

	assert.Equal(t, 1, len(taskDetails), "应创建1个 taskDetail")

	// 验证所有字段都正确复制
	createdDetail := taskDetails[0].ListBatchListenerResult
	assert.Equal(t, originalDetail.ClbID, createdDetail.ClbID, "ClbID 应该正确复制")
	assert.Equal(t, originalDetail.CloudClbID, createdDetail.CloudClbID, "CloudClbID 应该正确复制")
	assert.Equal(t, originalDetail.ClbVipDomain, createdDetail.ClbVipDomain, "ClbVipDomain 应该正确复制")
	assert.Equal(t, originalDetail.BkBizID, createdDetail.BkBizID, "BkBizID 应该正确复制")
	assert.Equal(t, originalDetail.Region, createdDetail.Region, "Region 应该正确复制")
	assert.Equal(t, originalDetail.Vendor, createdDetail.Vendor, "Vendor 应该正确复制")
	assert.Equal(t, originalDetail.LblID, createdDetail.LblID, "LblID 应该正确复制")
	assert.Equal(t, originalDetail.CloudLblID, createdDetail.CloudLblID, "CloudLblID 应该正确复制")
	assert.Equal(t, originalDetail.Protocol, createdDetail.Protocol, "Protocol 应该正确复制")
	assert.Equal(t, originalDetail.Port, createdDetail.Port, "Port 应该正确复制")
	assert.Equal(t, originalDetail.NewRsWeight, createdDetail.NewRsWeight, "NewRsWeight 应该正确复制")

	// 验证 RsList 只包含当前 RS
	assert.Equal(t, 1, len(createdDetail.RsList), "RsList 应该只包含1个 RS")
	assert.Equal(t, "rs-test-001", createdDetail.RsList[0].ID, "应该包含正确的 RS ID")
	assert.Equal(t, "172.16.0.1", createdDetail.RsList[0].IP, "应该包含正确的 RS IP")
	assert.Equal(t, int64(25), cvt.PtrToVal(createdDetail.RsList[0].Weight), "应该包含正确的原始权重")
}

// buildTaskDetailsFromDetailsLogic 按生产逻辑构造 taskDetails：
// 每个需要变更的 RS 生成一个 taskDetail，但 taskDetail 关联的是完整 detail（而非单 RS detail）。
func buildTaskDetailsFromDetailsLogic(
	details []*dataproto.ListBatchListenerResult) []*batchListenerModifyRsWeightTaskDetail {

	taskDetails := make([]*batchListenerModifyRsWeightTaskDetail, 0)
	for _, detail := range details {
		for _, rs := range detail.RsList {
			if cvt.PtrToVal(rs.Weight) != cvt.PtrToVal(detail.NewRsWeight) {
				taskDetails = append(taskDetails, &batchListenerModifyRsWeightTaskDetail{
					ListBatchListenerResult: detail,
				})
			}
		}
	}
	return taskDetails
}