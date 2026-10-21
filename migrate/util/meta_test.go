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

package util

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fillDestFromRows populates dest (a pointer to a slice of struct, tagged
// with `db:"..."`) from rows, mimicking what sqlx.StructScan would do, so
// ListTables can be tested without a real database.
func fillDestFromRows(t *testing.T, dest interface{}, rows []map[string]string) {
	t.Helper()

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		t.Fatalf("fillDestFromRows: dest must be pointer to slice, got %T", dest)
	}

	sliceVal := v.Elem()
	elemType := sliceVal.Type().Elem()
	for _, row := range rows {
		elem := reflect.New(elemType).Elem()
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			if val, ok := row[field.Tag.Get("db")]; ok {
				elem.Field(i).SetString(val)
			}
		}
		sliceVal.Set(reflect.Append(sliceVal, elem))
	}
}

func TestMetaOrm_Has(t *testing.T) {
	testCases := []struct {
		name     string
		call     func(m MetaOrm) (bool, error)
		count    uint64
		countErr error
		want     bool
		wantErr  bool
	}{
		{
			name:  "has table true",
			call:  func(m MetaOrm) (bool, error) { return m.HasTable(context.Background(), "cvm") },
			count: 1,
			want:  true,
		},
		{
			name:  "has table false",
			call:  func(m MetaOrm) (bool, error) { return m.HasTable(context.Background(), "cvm") },
			count: 0,
			want:  false,
		},
		{
			name:  "has view true",
			call:  func(m MetaOrm) (bool, error) { return m.HasView(context.Background(), "v_cvm") },
			count: 1,
			want:  true,
		},
		{
			name:  "has column true",
			call:  func(m MetaOrm) (bool, error) { return m.HasColumn(context.Background(), "cvm", "bk_asset_id") },
			count: 1,
			want:  true,
		},
		{
			name:  "has index false",
			call:  func(m MetaOrm) (bool, error) { return m.HasIndex(context.Background(), "cvm", "idx_asset") },
			count: 0,
			want:  false,
		},
		{
			name: "has constraint true",
			call: func(m MetaOrm) (bool, error) {
				return m.HasConstraint(context.Background(), "cvm", "fk_account")
			},
			count: 1,
			want:  true,
		},
		{
			name:     "query error propagates",
			call:     func(m MetaOrm) (bool, error) { return m.HasTable(context.Background(), "cvm") },
			countErr: errors.New("db down"),
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			do := &fakeDo{
				countFn: func(_ string, _ map[string]interface{}) (uint64, error) {
					return tc.count, tc.countErr
				},
			}
			m := NewMetaOrm(newFakeOrm(do))

			got, err := tc.call(m)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMetaOrm_ListTables(t *testing.T) {
	do := &fakeDo{
		selectFn: func(dest interface{}, _ string, _ map[string]interface{}) error {
			fillDestFromRows(t, dest, []map[string]string{
				{"table_name": "cvm"},
				{"table_name": "disk"},
			})
			return nil
		},
	}

	tables, err := NewMetaOrm(newFakeOrm(do)).ListTables(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"cvm", "disk"}, tables)
}

func TestMetaOrm_ListTables_Error(t *testing.T) {
	do := &fakeDo{
		selectFn: func(_ interface{}, _ string, _ map[string]interface{}) error {
			return errors.New("db down")
		},
	}

	_, err := NewMetaOrm(newFakeOrm(do)).ListTables(context.Background())
	assert.Error(t, err)
}
