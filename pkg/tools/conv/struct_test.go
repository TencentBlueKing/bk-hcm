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

package conv_test

import (
	"hcm/pkg/tools/conv"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Struct", func() {
	emptyMap := map[string]interface{}{}
	DescribeTable("StructToMap cases", func(value interface{}, expected map[string]interface{}, willError bool) {
		data, err := conv.StructToMap(value)

		if willError {
			assert.Error(GinkgoT(), err)
		} else {
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, data)
		}
	},
		Entry("a nil pointer", nil, emptyMap, true),
		Entry("not a struct", 0, emptyMap, true),
		Entry("json marshal error", "1", emptyMap, true),
		Entry("a struct", struct {
			A string `json:"a"`
			B int64  `json:"b"`
		}{A: "a", B: 1}, map[string]interface{}{"a": "a", "b": 1}, false),
	)
})
