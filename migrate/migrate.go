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

// hcm-migrate runs the database migrations of HCM.
package main

import (
	"os"

	"hcm/migrate/cli"
	// import migrations, used to register every migration into the build.
	_ "hcm/migrate/migrations"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"
)

func main() {
	// Process logs go to stderr so they stay in the job log after the pod is
	// gone; stdout only carries the plan and the summary.
	logs.InitLogger(logs.LogConfig{ToStdErr: true, LogLineMaxSize: constant.MigrationLogLineMaxKB})

	code := cli.Run(os.Args[1:], os.Stdout, os.Stderr)
	logs.CloseLogs()
	os.Exit(code)
}
