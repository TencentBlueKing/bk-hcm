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

package scheduler

import (
	"testing"

	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/thirdparty/cvmapi"

	"github.com/stretchr/testify/assert"
)

// TestOrderToUnifyOrder_ChargeTypeDefault tests charge_type default value filling logic
// TAPD: 129995406 - 测试API申领未传计费模式时的默认值填充
func TestOrderToUnifyOrder_ChargeTypeDefault(t *testing.T) {
	tests := []struct {
		name           string
		inputOrders    []*types.ApplyOrder
		expectedCharge cvmapi.ChargeType
		description    string
	}{
		{
			name: "charge_type为空字符串-填充PREPAID",
			inputOrders: []*types.ApplyOrder{
				{
					SubOrderId: "test-001",
					Spec: &types.ResourceSpec{
						ChargeType:   "",  // 空字符串
						ChargeMonths: 36,
						DeviceType:   "SA2",
					},
				},
			},
			expectedCharge: cvmapi.ChargeTypePrePaid,
			description:    "API申领未传计费模式时，应自动填充为PREPAID（包年包月）",
		},
		{
			name: "charge_type为POSTPAID_BY_HOUR-保持不变",
			inputOrders: []*types.ApplyOrder{
				{
					SubOrderId: "test-002",
					Spec: &types.ResourceSpec{
						ChargeType: cvmapi.ChargeTypePostPaidByHour,
						DeviceType: "SA2",
					},
				},
			},
			expectedCharge: cvmapi.ChargeTypePostPaidByHour,
			description:    "用户明确选择按量计费时，应保持原值",
		},
		{
			name: "charge_type为PREPAID-保持不变",
			inputOrders: []*types.ApplyOrder{
				{
					SubOrderId: "test-003",
					Spec: &types.ResourceSpec{
						ChargeType:   cvmapi.ChargeTypePrePaid,
						ChargeMonths: 12,
						DeviceType:   "SA2",
					},
				},
			},
			expectedCharge: cvmapi.ChargeTypePrePaid,
			description:    "用户明确选择包年包月时，应保持原值",
		},
		{
			name: "Spec为nil-不崩溃",
			inputOrders: []*types.ApplyOrder{
				{
					SubOrderId: "test-004",
					Spec:       nil,
				},
			},
			expectedCharge: "",
			description:    "Spec为nil时不应崩溃，跳过处理",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建最小化的scheduler实例（仅用于测试orderToUnifyOrder方法）
			s := &scheduler{}

			// 执行转换
			result := s.orderToUnifyOrder(nil, tt.inputOrders, false)

			// 验证结果
			assert.Equal(t, 1, len(result), "应该返回1个UnifyOrder")

			// 验证charge_type
			if tt.inputOrders[0].Spec != nil {
				actual := result[0].Spec.ChargeType
				assert.Equal(t, tt.expectedCharge, actual,
					"charge_type不符合预期: %s", tt.description)
			} else {
				// Spec为nil时，不检查ChargeType
				assert.Nil(t, result[0].Spec, "Spec应该为nil")
			}
		})
	}
}

// TestOrderToUnifyOrder_ChargeTypeBatch tests charge_type default filling in batch scenarios
func TestOrderToUnifyOrder_ChargeTypeBatch(t *testing.T) {
	// 准备测试数据：100个申领单，其中50个charge_type为空，50个有值
	orders := make([]*types.ApplyOrder, 100)
	for i := 0; i < 100; i++ {
		spec := &types.ResourceSpec{
			DeviceType:   "SA2",
			ChargeMonths: 12,
		}
		
		// 前50个charge_type为空，后50个有值
		if i < 50 {
			spec.ChargeType = ""
		} else if i < 75 {
			spec.ChargeType = cvmapi.ChargeTypePrePaid
		} else {
			spec.ChargeType = cvmapi.ChargeTypePostPaidByHour
		}
		
		orders[i] = &types.ApplyOrder{
			SubOrderId: "batch-test-" + string(rune(i)),
			Spec:       spec,
		}
	}

	// 执行转换
	s := &scheduler{}
	result := s.orderToUnifyOrder(nil, orders, false)

	// 验证结果
	assert.Equal(t, 100, len(result), "应该返回100个UnifyOrder")

	// 验证前50个（原本为空）都被填充为PREPAID
	for i := 0; i < 50; i++ {
		assert.Equal(t, cvmapi.ChargeTypePrePaid, result[i].Spec.ChargeType,
			"空值应被填充为PREPAID: index=%d", i)
	}

	// 验证第51-75个（原本为PREPAID）保持不变
	for i := 50; i < 75; i++ {
		assert.Equal(t, cvmapi.ChargeTypePrePaid, result[i].Spec.ChargeType,
			"PREPAID应保持不变: index=%d", i)
	}

	// 验证第76-100个（原本为POSTPAID_BY_HOUR）保持不变
	for i := 75; i < 100; i++ {
		assert.Equal(t, cvmapi.ChargeTypePostPaidByHour, result[i].Spec.ChargeType,
			"POSTPAID_BY_HOUR应保持不变: index=%d", i)
	}
}

// TestOrderToUnifyOrder_ChargeTypeEdgeCases tests edge cases for charge_type handling
func TestOrderToUnifyOrder_ChargeTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		inputOrder  *types.ApplyOrder
		expectPanic bool
		description string
	}{
		{
			name: "空Order列表",
			inputOrder: nil,
			expectPanic: false,
			description: "输入空列表应返回空结果，不崩溃",
		},
		{
			name: "Order为nil-不应panic",
			inputOrder: &types.ApplyOrder{
				SubOrderId: "test-nil",
				Spec:       nil,
			},
			expectPanic: false,
			description: "Order不为nil但Spec为nil时不应崩溃",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scheduler{}

			defer func() {
				r := recover()
				if tt.expectPanic {
					assert.NotNil(t, r, "应该panic: %s", tt.description)
				} else {
					assert.Nil(t, r, "不应该panic: %s", tt.description)
				}
			}()

			var orders []*types.ApplyOrder
			if tt.inputOrder != nil {
				orders = []*types.ApplyOrder{tt.inputOrder}
			}

			result := s.orderToUnifyOrder(nil, orders, false)

			if tt.inputOrder != nil {
				assert.Equal(t, 1, len(result), "应该返回1个结果")
			} else {
				assert.Equal(t, 0, len(result), "空输入应返回空结果")
			}
		})
	}
}
