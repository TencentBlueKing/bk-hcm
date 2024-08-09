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

// Package audit 对 data-service 提供的审计接口进行二次封装，提供丰富的场景审计能力，方便 account_server 使用。
package audit

import (
	dataservice "hcm/pkg/client/data-service"
)

// Interface define audit interface.
type Interface interface {
	// todo
	// 创建账号审计
	// CreateAccountAudit()
}

var _ Interface = new(audit)

// NewAudit new audit.
func NewAudit(dataCli *dataservice.Client) Interface {
	return &audit{
		dataCli: dataCli,
	}
}

type audit struct {
	dataCli *dataservice.Client
}
