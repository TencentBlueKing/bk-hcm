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

// Package tfruntime 定义 tf_provider 插件进程管理与执行器的对接契约。
//
// 本包目前只有接口、结构体与返回 ErrNotImplemented 的桩方法，不含任何协议实现：
// go-plugin 建连、握手、gRPC 收发、DynamicValue 编解码、日志落盘等由对接方在此骨架上填充，
// 进程池等上游只依赖这里的类型。
package tfruntime

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotImplemented 契约中的方法尚未由对接方实现。
var ErrNotImplemented = errors.New("tfruntime: not implemented")

// DiagnosticsError 插件以 Error 级 Diagnostic 拒绝了一次调用，属于业务错误。
//
// 调用方用 errors.As 区分：能取到 *DiagnosticsError 的是业务错误（换进程重试也会失败），
// 其余（超时、连接断开、进程退出等）是传输错误。
// Error() 只包含 Error 级 Diagnostic 的 Summary 与属性路径，不含 Detail，更不含请求里的 Config。
type DiagnosticsError struct {
	Diagnostics []Diagnostic
}

// Error 实现 error 接口。
func (e *DiagnosticsError) Error() string {
	if e == nil {
		return "tfruntime: nil diagnostics error"
	}

	msgs := make([]string, 0, len(e.Diagnostics))
	for _, d := range e.Diagnostics {
		if d.Severity != SeverityError {
			continue
		}
		if len(d.Attribute) > 0 {
			msgs = append(msgs, fmt.Sprintf("%s (%s)", d.Summary, d.Attribute.String()))
			continue
		}
		msgs = append(msgs, d.Summary)
	}
	return "tfruntime: provider returned error diagnostics: " + strings.Join(msgs, "; ")
}

// HasError 是否存在 Error 级 Diagnostic。Warning 不使调用失败。
func HasError(diags []Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// DiagnosticsToError 存在 Error 级 Diagnostic 时返回 *DiagnosticsError，否则返回 nil。
func DiagnosticsToError(diags []Diagnostic) error {
	if !HasError(diags) {
		return nil
	}
	return &DiagnosticsError{Diagnostics: diags}
}
