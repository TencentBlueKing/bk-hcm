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

// Package options ...
package options

import (
	"hcm/pkg/cc"
	"hcm/pkg/runtime/flags"

	"github.com/spf13/pflag"
)

// Option defines the app's runtime flag options.
type Option struct {
	Sys *cc.SysOption
	// PublicKey used to api gateway jwt token.
	PublicKey string
	// DisableJWT whether to enable blueking api-gateway jwt parser.if disable-tgw = false, api-service
	// parse request will api-gateway parser, requests from other parties will not be parsed.if
	// disable-tgw = true, api-service parse requests for direct access that have not been processed by
	// the gateway. Parse rule details：pkg/runtime/parser/parser.go
	DisableJWT bool
}

// InitOptions init api server's options from command flags.
func InitOptions() *Option {
	fs := pflag.CommandLine
	sysOpt := flags.SysFlags(fs)
	opt := &Option{Sys: sysOpt}

	fs.StringVarP(&opt.PublicKey, "public-key", "", "", "the api gateway public key path")
	fs.BoolVarP(&opt.DisableJWT, "disable-jwt", "", false, "to disable jwt authorize for "+
		"all the incoming request. Note: disable jwt authorize may cause security problems.")

	// parses the command-line flags from os.Args[1:]. must be called after all flags are defined
	// and before flags are accessed by the program.
	pflag.Parse()

	// check if the command-line flag is show current version info cmd.
	sysOpt.CheckV()

	return opt
}
