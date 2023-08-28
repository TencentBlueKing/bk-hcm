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

package adaptormock

import (
	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// LogReporter  for gomock controller
type LogReporter struct {
}

// Fatalf default do not panic
func (r *LogReporter) Fatalf(fmt string, args ...any) {
	logs.Fatalf(fmt, args...)
}

// Errorf default do not panic
func (r *LogReporter) Errorf(fmt string, args ...any) {
	logs.Errorf(fmt, args...)
}

// NewCloudResStore returns new store, keyGetter use for get key from value. For example get id from value.
func NewCloudResStore[CloudType common.CloudResType](
	items ...CloudType) *Store[string, CloudType] {

	s := Store[string, CloudType]{keyGetter: CloudType.GetCloudID}
	s.Init(items...)

	return &s
}

// Store a general purpose kv store
type Store[K comparable, V any] struct {
	keyGetter func(V) K
	dict      map[K]V
}

// Init replace inside storage with new one
func (st *Store[K, V]) Init(items ...V) {
	st.dict = make(map[K]V, len(items))
	st.AddItems(items...)
}

func (st *Store[K, V]) Get(key K) (val V, exists bool) {
	val, exists = st.dict[key]
	return
}

// GetByKeys 如果输入参数为空，则返回全部数据; 如果参数非空，则返回可以找到的Value
func (st *Store[K, V]) GetByKeys(keys ...K) (values []V) {
	if len(keys) == 0 {
		return st.ListAll()
	}
	for _, k := range keys {
		if val, exists := st.dict[k]; exists {
			values = append(values, val)
		}
	}
	return values
}

// Update item must exits
func (st *Store[K, V]) Update(key K, val V) error {
	if _, exits := st.dict[key]; !exits {
		return errf.Newf(errf.RecordNotFound, "not found in mock store: %v", key)
	}
	st.dict[key] = val
	return nil
}

// Set does not care key exist or not
func (st *Store[K, V]) Set(key K, val V) {
	st.dict[key] = val
}

// AddItems batch add new items using keyGetter func, return false when any item exists
func (st *Store[K, V]) AddItems(val ...V) (ok bool) {
	for _, v := range val {
		if !st.Add(st.keyGetter(v), v) {
			return false
		}
	}
	return true
}

// Add return false when item exists
func (st *Store[K, V]) Add(key K, val V) (ok bool) {
	if _, exists := st.dict[key]; exists {
		return false
	}
	logs.V(3).Infof("Adding  %+v", val)
	st.dict[key] = val
	return true
}

// ListAll returns all value in a slice
func (st *Store[K, V]) ListAll() (valueSlice []V) {
	return converter.MapValueToSlice(st.dict)
}

// Filter return values match given func
func (st *Store[K, V]) Filter(match func(V) bool) (valueSlice []V) {
	for _, val := range st.dict {
		if match(val) {
			valueSlice = append(valueSlice, val)
		}
	}
	return
}

// Remove existing item
func (st *Store[K, V]) Remove(key K) error {
	if _, exits := st.dict[key]; !exits {
		return errf.Newf(errf.RecordNotFound, "not found in mock store: %v", key)
	}
	delete(st.dict, key)
	return nil
}
