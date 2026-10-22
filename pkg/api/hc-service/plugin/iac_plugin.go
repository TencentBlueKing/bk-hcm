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

// Package plugin defines http iac plugin register / upgrade / update request and response.
// cloud-server 对外接口与 hc-service 内部路由共用同一组结构：cloud-server 只校验、鉴权后原样转发。
package plugin

import (
	"errors"
	"regexp"

	coreplugin "hcm/pkg/api/core/plugin"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/iacplugin"
	cvt "hcm/pkg/tools/converter"
)

// pluginNameRegexp 插件名：字母或数字开头，只含字母、数字、_、-、.，最长 64。插件名也是锁 key 的一部分，不能含 /。
var pluginNameRegexp = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)

func validatePluginName(name string) error {
	if !pluginNameRegexp.MatchString(name) {
		return errors.New("plugin_name should start with a letter or digit, contain only letters, digits, " +
			"'_', '-', '.', and be at most 64 characters")
	}

	return nil
}

// HTTPRegisterReq 注册 http 插件请求。插件版本取自探活结果，不由请求传入。
type HTTPRegisterReq struct {
	PluginName   string                       `json:"plugin_name" validate:"required"`
	APIPath      string                       `json:"api_path" validate:"required,max=255"`
	CryptoType   string                       `json:"crypto_type" validate:"required"`
	CryptoConfig *coreplugin.HTTPCryptoConfig `json:"crypto_config" validate:"required"`
	Memo         *string                      `json:"memo" validate:"omitempty,max=255"`
}

// Validate HTTPRegisterReq.
func (req *HTTPRegisterReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := validatePluginName(req.PluginName); err != nil {
		return err
	}

	if err := iacplugin.ValidateAPIPath(req.APIPath); err != nil {
		return err
	}

	return iacplugin.ValidateCrypto(req.CryptoType, req.CryptoConfig)
}

// HTTPUpgradeReq 升级 http 插件请求。按插件名找到版本最高的记录再探活。
type HTTPUpgradeReq struct {
	PluginName string `json:"plugin_name" validate:"required"`
}

// Validate HTTPUpgradeReq.
func (req *HTTPUpgradeReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	return validatePluginName(req.PluginName)
}

// HTTPUpdateReq 修改 http 插件请求，作用于该插件名下版本最高的记录。除 plugin_name 外至少再传一个字段，未传的保持不变。
// 加密配置的合法性要与记录里的已有值合并后才能判断，在 hc-service 合并后校验。
type HTTPUpdateReq struct {
	PluginName   string                       `json:"plugin_name" validate:"required"`
	APIPath      *string                      `json:"api_path" validate:"omitempty,max=255"`
	CryptoType   *string                      `json:"crypto_type" validate:"omitempty"`
	CryptoConfig *coreplugin.HTTPCryptoConfig `json:"crypto_config" validate:"omitempty"`
	Memo         *string                      `json:"memo" validate:"omitempty,max=255"`
}

// Validate HTTPUpdateReq.
func (req *HTTPUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := validatePluginName(req.PluginName); err != nil {
		return err
	}

	if req.APIPath == nil && req.CryptoType == nil && req.CryptoConfig == nil && req.Memo == nil {
		return errors.New("at least one of api_path, crypto_type, crypto_config, memo is required")
	}

	if req.APIPath != nil {
		return iacplugin.ValidateAPIPath(cvt.PtrToVal(req.APIPath))
	}

	return nil
}

// HTTPPluginResult 注册 / 升级 / 修改结果：该插件名下版本最高记录的当前状态。
type HTTPPluginResult struct {
	ID            string                `json:"id"`
	PluginName    string                `json:"plugin_name"`
	PluginVersion string                `json:"plugin_version"`
	State         enumor.IaCPluginState `json:"state"`
}
