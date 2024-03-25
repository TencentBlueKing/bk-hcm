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

// Package validator ...
package validator

import (
	"errors"
	"reflect"
	"strings"

	gvalidator "github.com/go-playground/validator/v10"
)

var Validate = gvalidator.New()

// Interface ...
type Interface interface {
	Validate() error
}

// ValidateTool validate tool.
func ValidateTool(opts ...Interface) error {
	if len(opts) == 0 {
		return nil
	}

	for _, opt := range opts {
		if opt == nil {
			return errors.New("option can not nil")
		}

		if err := opt.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	// 返回json tag 名称
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}
