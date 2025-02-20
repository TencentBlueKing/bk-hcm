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
	"testing"
)

func TestPackageName(t *testing.T) {
	cases := []struct {
		skip     int
		expected string
	}{
		{0, "runtimes"},
		{1, "testing"},
		{2, "runtime"},
	}

	for _, c := range cases {
		if got := PackageName(c.skip); got != c.expected {
			t.Errorf("PackageName(%d) = %s, want %s", c.skip, got, c.expected)
		}
	}
}

func TestStructName(t *testing.T) {
	cases := []struct {
		skip     int
		expected string
	}{
		{0, "testStruct"},
		{1, "testStruct"},
		{2, "testStruct"},
		{3, "unknown"},
	}

	s := &testStruct{}

	for _, c := range cases {
		if got := s.testStruct1(c.skip); got != c.expected {
			t.Errorf("s.testStruct1(%d) = %s, want %s", c.skip, got, c.expected)
		}
	}
}

func TestFuncName(t *testing.T) {
	cases := []struct {
		skip     int
		expected string
	}{
		{0, "testFunc3"},
		{1, "testFunc2"},
		{2, "testFunc1"},
		{3, "TestFuncName"},
	}

	s := &testStruct{}

	for _, c := range cases {
		if got := s.testFunc1(c.skip); got != c.expected {
			t.Errorf("s.testFunc1(%d) = %s, want %s", c.skip, got, c.expected)
		}
	}
}

type testStruct struct {
}

func (t *testStruct) testStruct1(skip int) string {
	return t.testStruct2(skip)
}

func (t *testStruct) testStruct2(skip int) string {
	return t.testStruct3(skip)
}

func (t *testStruct) testStruct3(skip int) string {
	return StructName(skip)
}

func (t *testStruct) testFunc1(skip int) string {
	return t.testFunc2(skip)
}

func (t *testStruct) testFunc2(skip int) string {
	return t.testFunc3(skip)
}

func (t *testStruct) testFunc3(skip int) string {
	return FuncName(skip)
}
