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

// Package runtimes provides utilities for working with Go runtimes.
package runtimes

import (
	"runtime"
	"strings"
)

// PackageName returns the package name of the calling function.
func PackageName(skip int) string {
	// skip + 1 is for jumping over the PackageName function itself
	fullFuncName := getFullFuncName(skip + 1)
	packageName, _, _ := parseFuncName(fullFuncName)
	return packageName
}

// StructName returns the name of the struct to which the calling method belongs.
func StructName(skip int) string {
	// skip + 1 is for jumping over the StructName function itself
	fullFuncName := getFullFuncName(skip + 1)
	_, structName, _ := parseFuncName(fullFuncName)
	return structName
}

// FuncName returns the name of the calling function or method.
func FuncName(skip int) string {
	// skip + 1 is for jumping over the FuncName function itself
	fullFuncName := getFullFuncName(skip + 1)
	_, _, funcName := parseFuncName(fullFuncName)
	return funcName
}

// getFullFuncName returns the full name of the function, including package and receiver type.
func getFullFuncName(skip int) string {
	// skip + 1 is for jumping over the getFullFuncName function itself
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	return fn.Name()
}

// parseFuncName parses the full function name to extract the package name, struct name, and function name.
func parseFuncName(fullFuncName string) (packageName, structName, funcName string) {
	fullFuncName = strings.ReplaceAll(fullFuncName, "[...]", "")
	parts := strings.Split(fullFuncName, "/")
	lastPart := parts[len(parts)-1]
	segments := strings.Split(lastPart, ".")
	if len(segments) == 1 {
		// Only the function name is present, no package or struct name
		return "unknown", "unknown", segments[0]
	}

	if len(segments) == 2 {
		// Format is packageName.functionName
		return segments[0], "unknown", segments[1]
	}

	// Format is packageName.structName.functionName
	structName = segments[len(segments)-2]
	structName = strings.TrimPrefix(structName, "(")
	structName = strings.TrimPrefix(structName, "*")
	structName = strings.TrimSuffix(structName, ")")
	return strings.Join(segments[:len(segments)-2], "."), structName, segments[len(segments)-1]
}
