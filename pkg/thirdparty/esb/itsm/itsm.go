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

package itsm

import (
	"context"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
)

type Client interface {
	CreateTicket(ctx context.Context, params *CreateTicketParams) (string, error)
	GetTicketResult(ctx context.Context, sn string) (TicketResult, error)
	WithdrawTicket(ctx context.Context, sn string, operator string) error
	VerifyToken(ctx context.Context, token string) (bool, error)
}

// NewClient initialize a new itsm client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &itsm{
		client: client,
		config: config,
	}
}

// itsm is an esb client to request itsm.
type itsm struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}
