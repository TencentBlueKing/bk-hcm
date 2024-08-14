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

package errf

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const mysqlDuplicatedNumber = 1062

// GetTypedError 尝试转换为指定类型的错误
func GetTypedError[T error](err error) *T {
	var terr T
	if errors.As(err, &terr) {
		return &terr
	}
	return nil
}

// GetMySQLDuplicated return mysql.MySQLError if error is a mysql duplicated error
func GetMySQLDuplicated(err error) (merr *mysql.MySQLError) {
	if errors.As(err, &merr) && merr.Number == mysqlDuplicatedNumber {
		return merr
	}
	return nil
}

// IsRecordNotFound return true if error is a not found error
func IsRecordNotFound(err error) bool {
	if err == nil {
		return false
	}
	var ef *ErrorF
	if errors.As(err, &ef) {
		return ef.Code == RecordNotFound
	}
	return false
}

// IsContextCanceled return true if error contains string "context canceled"
func IsContextCanceled(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "context canceled")
}

// IsDuplicated return true if error is a duplicated error
func IsDuplicated(err error) bool {
	var merr *mysql.MySQLError
	if errors.As(err, &merr) {
		return merr.Number == mysqlDuplicatedNumber
	}
	var ef *ErrorF
	if errors.As(err, &ef) {
		return ef.Code == RecordDuplicated
	}
	return false
}

// Error try to convert the error to ErrorF if possible.
// it is used by the RPC client to wrap the response error response
// by the RPC server to the ErrorF, user can use this ErrorF to test
// if an error is returned or not, if yes, then use the ErrorF to
// response with error code and message.
func Error(err error) *ErrorF {
	if err == nil {
		return nil
	}

	ef, ok := err.(*ErrorF)
	if ok {
		return ef
	}

	s := err.Error()

	// test if the error is a json error,
	// if not, then this is an error without error code.
	if !strings.HasPrefix(s, "{") {
		return &ErrorF{
			Code:    Unknown,
			Message: s,
		}
	}

	// this is a json error, try decoding it to standard error directly.
	ef = new(ErrorF)
	if err := json.Unmarshal([]byte(s), ef); err != nil {
		return &ErrorF{
			Code:    Unknown,
			Message: s,
		}
	}

	if ef.Code == 0 {
		return &ErrorF{
			Code:    Unknown,
			Message: s,
		}
	}

	return ef
}
