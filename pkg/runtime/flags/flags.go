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

// Package flags ...
package flags

import (
	"strings"

	"hcm/pkg/cc"

	"github.com/spf13/pflag"
)

// wordSepNormalizeFunc changes all flags that contain "_" separators
func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}

// SysFlags normalizes and parses the command line flags
func SysFlags(fs *pflag.FlagSet) *cc.SysOption {
	opt := new(cc.SysOption)
	fs.SetNormalizeFunc(wordSepNormalizeFunc)

	fs.StringVarP(&opt.ConfigFile, "config-file", "c", "", "the absolute path of the configuration file")
	fs.IPVarP(&opt.BindIP, "bind-ip", "h", []byte{}, "which IP the server is listen to")
	fs.BoolVarP(&opt.Versioned, "version", "v", false, "show version")

	fs.StringVarP(&opt.Environment, "env", "e", "", "the environment of the server")
	fs.StringSliceVarP(&opt.Labels, "label", "l", nil, "the service labels")
	fs.BoolVarP(&opt.DisableElection, "disable-election", "d", false, "disable leader election")
	return opt
}
