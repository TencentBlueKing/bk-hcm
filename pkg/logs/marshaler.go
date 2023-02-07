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

package logs

import (
	"encoding/json"
	"fmt"

	"hcm/pkg/logs/glog"
)

// ObjectMarshaler allows user-defined types to efficiently add themselves to the
// logging context, and to selectively omit information which shouldn't be
// included in logs (e.g., passwords).
// Note: ObjectMarshaler order to solve the problem of lack of field key information
// if the structure field is a pointer, it is recommended not to use this method
// unless it is necessary!
type ObjectMarshaler interface {
	LogMarshal() string
}

// ObjectEncode is a general object coding function for log print.
func ObjectEncode(object interface{}) string {
	if object == nil {
		return "null"
	}

	marshal, err := json.Marshal(object)
	if err != nil {
		return fmt.Sprintf("%#v", object)
	}

	return string(marshal)
}

// errorJson compared with other log output methods, this method consumes
// performance and should not be used if it is not necessary.
func errorJson(format string, args ...interface{}) {
	params := make([]interface{}, len(args))
	for index, arg := range args {
		if object, ok := arg.(ObjectMarshaler); ok {
			params[index] = object.LogMarshal()
			continue
		}

		params[index] = arg
	}

	glog.ErrorDepthf(1, format, params...)
}
