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

package converter

import (
	"testing"
)

func TestValToPtr(t *testing.T) {
	strPtr := ValToPtr("test")
	if strPtr == nil || *strPtr != "test" {
		t.Errorf("test string value to pointer failed, got: %+v", strPtr)
		return
	}

	uintPtr := ValToPtr(uint64(100))
	if uintPtr == nil || *uintPtr != uint64(100) {
		t.Errorf("test uint64 value to pointer failed, got: %+v", uintPtr)
		return
	}

	val := 1
	ptrPtr := ValToPtr(&val)
	if ptrPtr == nil || *ptrPtr == nil || **ptrPtr != val {
		t.Errorf("test pointer value to pointer failed, got: %+v", ptrPtr)
		return
	}

	var nilVal map[string]interface{} = nil
	nilPtr := ValToPtr(nilVal)
	if nilPtr == nil || *nilPtr != nil {
		t.Errorf("test pointer value to pointer failed, got: %+v", nilPtr)
		return
	}
}

func TestPtrToVal(t *testing.T) {
	strV := "test"
	strVal := PtrToVal(&strV)
	if strVal != strV {
		t.Errorf("test string pointer to value failed, got: %+v", strVal)
		return
	}

	uintV := uint64(100)
	uintVal := PtrToVal(&uintV)
	if uintVal != uintV {
		t.Errorf("test string pointer to value failed, got: %+v", uintVal)
		return
	}

	ptr := &strV
	ptrVal := PtrToVal(&ptr)
	if ptrVal == nil || *ptrVal != strVal {
		t.Errorf("test pointer pointer to value failed, got: %+v", ptrVal)
		return
	}

	var nilPtr *string = nil
	nilPtrVal := PtrToVal(nilPtr)
	if nilPtrVal != "" {
		t.Errorf("test pointer value to pointer failed, got: %+v", nilPtrVal)
		return
	}

	var nilVal map[string]interface{} = nil
	nilValPtrVal := PtrToVal(&nilVal)
	if nilValPtrVal != nil {
		t.Errorf("test pointer value to pointer failed, got: %+v", nilValPtrVal)
		return
	}
}

func TestSliceToPtr(t *testing.T) {
	strPtr := SliceToPtr([]string{"test"})
	if len(strPtr) != 1 || *strPtr[0] != "test" {
		t.Errorf("test string value to pointer failed, got: %+v", strPtr)
		return
	}

	uintPtr := SliceToPtr([]uint64{100})
	if len(uintPtr) != 1 || *uintPtr[0] != uint64(100) {
		t.Errorf("test uint64 value to pointer failed, got: %+v", uintPtr)
		return
	}

	val := 1
	ptrPtr := SliceToPtr([]*int{&val})
	if len(ptrPtr) != 1 || ptrPtr[0] == nil || *ptrPtr[0] == nil || **ptrPtr[0] != val {
		t.Errorf("test pointer value to pointer failed, got: %+v", ptrPtr)
		return
	}

	nilPtr := SliceToPtr([]*int{nil})
	if len(nilPtr) != 1 || nilPtr[0] == nil || *nilPtr[0] != nil {
		t.Errorf("test pointer value to pointer failed, got: %+v", nilPtr)
		return
	}
}

func TestPtrToSlice(t *testing.T) {
	strPtr := PtrToSlice([]*string{ValToPtr("test")})
	if len(strPtr) != 1 || strPtr[0] != "test" {
		t.Errorf("test string value to pointer failed, got: %+v", strPtr)
		return
	}

	uintPtr := PtrToSlice([]*uint64{ValToPtr(uint64(100))})
	if len(uintPtr) != 1 || uintPtr[0] != uint64(100) {
		t.Errorf("test uint64 value to pointer failed, got: %+v", uintPtr)
		return
	}

	val := 1
	ptrPtr := PtrToSlice([]**int{ValToPtr(&val)})
	if len(ptrPtr) != 1 || ptrPtr[0] == nil || *ptrPtr[0] != val {
		t.Errorf("test pointer value to pointer failed, got: %+v", ptrPtr)
		return
	}

	nilPtr := PtrToSlice([]*int{nil})
	if len(nilPtr) != 1 || nilPtr[0] != 0 {
		t.Errorf("test pointer value to pointer failed, got: %+v", nilPtr)
		return
	}
}
