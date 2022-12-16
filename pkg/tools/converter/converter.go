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

// ValToPtr convert one value to pointer.
func ValToPtr[T any](val T) *T {
	return &val
}

// PtrToVal convert pointer to one value.
func PtrToVal[T any](ptr *T) T {
	var value T
	if ptr != nil {
		value = *ptr
	}
	return value
}

// SliceToPtr convert slice to pointer.
func SliceToPtr[T any](slice []T) []*T {
	ptrArr := make([]*T, len(slice))
	for idx, val := range slice {
		ptrArr[idx] = &val
	}
	return ptrArr
}

// PtrToSlice convert pointer to slice.
func PtrToSlice[T any](slice []*T) []T {
	ptrArr := make([]T, len(slice))
	for idx, ptr := range slice {
		if ptr != nil {
			ptrArr[idx] = *ptr
		}
	}
	return ptrArr
}
