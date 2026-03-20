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

package constant

// Excel 校验消息常量，面向前端用户展示的中文校验结果（非 Go error），用于 Excel 导入及表单字段校验场景。
// 格式串中第一个 %s 为列头名称。
const (
	// ExcelValidateRequiredEmpty 必填项为空时的提示
	ExcelValidateRequiredEmpty = "%s: 必填项不能为空"
	// ExcelValidateMustBeInt 值必须为整数时的提示
	ExcelValidateMustBeInt = "%s: 必须为整数"
	// ExcelValidateMustBeNumber 值必须为数字时的提示
	ExcelValidateMustBeNumber = "%s: 必须为数字"
	// ExcelValidateMustBeString 值必须为字符串时的提示
	ExcelValidateMustBeString = "%s: 必须为字符串"
	// ExcelValidateEnumTypeMismatch 枚举值类型不匹配（期望数字）时的提示
	ExcelValidateEnumTypeMismatch = "%s: 值'%v'类型不匹配，应为数字"
	// ExcelValidateEnumNotInRange 枚举值不在允许范围内时的提示
	ExcelValidateEnumNotInRange = "%s: 值'%s'不在允许范围%s内"
	// ExcelValidateFormulaMismatch 单元格公式与 schema 定义不符时的提示
	ExcelValidateFormulaMismatch = "%s: 单元格公式与模板不符，期望：%s，实际：%s"

	// ExcelValidateIntGT 整数值不满足严格大于约束时的提示
	ExcelValidateIntGT = "%s: 值%d必须大于%s"
	// ExcelValidateIntGTE 整数值不满足大于等于约束时的提示
	ExcelValidateIntGTE = "%s: 值%d不能小于%s"
	// ExcelValidateIntLT 整数值不满足严格小于约束时的提示
	ExcelValidateIntLT = "%s: 值%d必须小于%s"
	// ExcelValidateIntLTE 整数值不满足小于等于约束时的提示
	ExcelValidateIntLTE = "%s: 值%d不能大于%s"

	// ExcelValidateFloatGT 浮点值不满足严格大于约束时的提示
	ExcelValidateFloatGT = "%s: 值%s必须大于%s"
	// ExcelValidateFloatGTE 浮点值不满足大于等于约束时的提示
	ExcelValidateFloatGTE = "%s: 值%s不能小于%s"
	// ExcelValidateFloatLT 浮点值不满足严格小于约束时的提示
	ExcelValidateFloatLT = "%s: 值%s必须小于%s"
	// ExcelValidateFloatLTE 浮点值不满足小于等于约束时的提示
	ExcelValidateFloatLTE = "%s: 值%s不能大于%s"

	// ExcelValidateStrLenGT 字符串长度不满足严格大于约束时的提示
	ExcelValidateStrLenGT = "%s: 长度%d必须大于%s"
	// ExcelValidateStrLenGTE 字符串长度不满足大于等于约束时的提示
	ExcelValidateStrLenGTE = "%s: 长度%d不能小于%s"
	// ExcelValidateStrLenLT 字符串长度不满足严格小于约束时的提示
	ExcelValidateStrLenLT = "%s: 长度%d必须小于%s"
	// ExcelValidateStrLenLTE 字符串长度不满足小于等于约束时的提示
	ExcelValidateStrLenLTE = "%s: 长度%d不能大于%s"
)

// MaxExcelFileSize defines the max size of excel file 
const  MaxExcelFileSize int64 = 1024 * 1024 * 50 // 50MB