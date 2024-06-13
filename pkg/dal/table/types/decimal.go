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

package types

import (
	"fmt"

	"database/sql/driver"

	"github.com/shopspring/decimal"
)

// Decimal is wrapper for
type Decimal struct {
	decimal.Decimal
}

// Scan is used to decode raw message which is read from db into
func (d *Decimal) Scan(raw interface{}) error {
	if raw == nil {
		return nil
	}
	data := ""
	switch v := raw.(type) {
	case []byte:
		data = string(v)
	case string:
		data = v
	default:
		return fmt.Errorf("unsupported decimal raw type: %T", v)
	}
	internalDecimal, err := decimal.NewFromString(data)
	if err != nil {
		return fmt.Errorf("parse decimal %s failed, err %s", data, err.Error())
	}
	d.Decimal = internalDecimal
	return nil
}

// Value encode the Decimal to a json raw, so that it can be stored to db with json raw.
func (d *Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}
