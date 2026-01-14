/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package ziyan

import (
	"testing"
)

func TestExtractAreaAndCityName(t *testing.T) {
	tests := []struct {
		name         string
		regionName   string
		expectedArea string
		expectedCity string
	}{
		{
			name:         "正常情况：有括号",
			regionName:   "华南地区(广州)",
			expectedArea: "华南地区",
			expectedCity: "广州",
		},
		{
			name:         "正常情况：华东地区",
			regionName:   "华东地区(上海)",
			expectedArea: "华东地区",
			expectedCity: "上海",
		},
		{
			name:         "没有括号：原始字符串作为city_name",
			regionName:   "广州",
			expectedArea: "",
			expectedCity: "广州",
		},
		{
			name:         "空字符串",
			regionName:   "",
			expectedArea: "",
			expectedCity: "",
		},
		{
			name:         "只有左括号没有右括号",
			regionName:   "华南地区(广州",
			expectedArea: "",
			expectedCity: "华南地区(广州",
		},
		{
			name:         "只有右括号没有左括号",
			regionName:   "华南地区广州)",
			expectedArea: "",
			expectedCity: "华南地区广州)",
		},
		{
			name:         "括号顺序错误：右括号在左括号前面",
			regionName:   "华南地区)广州(",
			expectedArea: "",
			expectedCity: "华南地区)广州(",
		},
		{
			name:         "多个括号：取第一个括号对",
			regionName:   "华南地区(广州)(深圳)",
			expectedArea: "华南地区",
			expectedCity: "广州",
		},
		{
			name:         "括号内为空",
			regionName:   "华南地区()",
			expectedArea: "华南地区",
			expectedCity: "",
		},
		{
			name:         "括号前为空：左括号在开头",
			regionName:   "(广州)",
			expectedArea: "",
			expectedCity: "广州",
		},
		{
			name:         "包含前后空格",
			regionName:   " 华南地区(广州) ",
			expectedArea: "华南地区",
			expectedCity: "广州",
		},
		{
			name:         "括号内包含空格",
			regionName:   "华南地区( 广州 )",
			expectedArea: "华南地区",
			expectedCity: "广州",
		},
		{
			name:         "括号前包含空格",
			regionName:   "华南地区 (广州)",
			expectedArea: "华南地区",
			expectedCity: "广州",
		},
		{
			name:         "括号内只有空格",
			regionName:   "华南地区(   )",
			expectedArea: "华南地区",
			expectedCity: "",
		},
		{
			name:         "嵌套括号：取最外层",
			regionName:   "华南地区(广州(天河))",
			expectedArea: "华南地区",
			expectedCity: "广州",
		},
		{
			name:         "左括号紧跟在字符后",
			regionName:   "华南(广州)",
			expectedArea: "华南",
			expectedCity: "广州",
		},
		{
			name:         "右括号紧跟在字符前",
			regionName:   "华南地区(广州)区",
			expectedArea: "华南地区",
			expectedCity: "广州",
		},
		{
			name:         "包含中文括号",
			regionName:   "华南地区（广州）",
			expectedArea: "",
			expectedCity: "华南地区（广州）",
		},
		{
			name:         "括号内包含特殊字符",
			regionName:   "华南地区(广州-天河)",
			expectedArea: "华南地区",
			expectedCity: "广州-天河",
		},
		{
			name:         "括号内包含数字",
			regionName:   "华南地区(广州1区)",
			expectedArea: "华南地区",
			expectedCity: "广州1区",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			areaName, cityName := extractAreaAndCityName(tt.regionName)
			if areaName != tt.expectedArea {
				t.Errorf("extractAreaAndCityName() areaName = %v, want %v", areaName, tt.expectedArea)
			}
			if cityName != tt.expectedCity {
				t.Errorf("extractAreaAndCityName() cityName = %v, want %v", cityName, tt.expectedCity)
			}
		})
	}
}
