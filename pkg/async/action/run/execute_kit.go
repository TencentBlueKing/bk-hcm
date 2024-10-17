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

package run

import (
	"hcm/pkg/kit"
	"hcm/pkg/tools/uuid"
)

// ExecuteKit is a kit using by action
type ExecuteKit interface {
	Kit() *kit.Kit
	AsyncKit() *kit.Kit
	KitWithNewRid() *kit.Kit
	ShareData() ShareDataOperator
}

// ShareDataOperator used to operate share data
type ShareDataOperator interface {
	Get(key string) (string, bool)
	Set(kt *kit.Kit, key string, val string) error
	AppendIDs(kt *kit.Kit, key string, ids ...string) error
}

// NewExecuteContext new execute context for task exec.
func NewExecuteContext(kt *kit.Kit, shareData ShareDataOperator) ExecuteKit {
	return &DefExecuteContext{
		kit:       kt,
		shareData: shareData,
	}
}

// DefExecuteContext default execute context.
type DefExecuteContext struct {
	kit       *kit.Kit
	shareData ShareDataOperator
}

// KitWithNewRid return kit with new rid.
func (ctx *DefExecuteContext) KitWithNewRid() *kit.Kit {
	return ctx.kit.NewSubKitWithRid(uuid.UUID())
}

// AsyncKit Kit with async request source.
func (ctx *DefExecuteContext) AsyncKit() *kit.Kit {
	return ctx.kit.WithAsyncSource()
}

// Kit return kit.
func (ctx *DefExecuteContext) Kit() *kit.Kit {
	return ctx.kit
}

// ShareData return share data.
func (ctx *DefExecuteContext) ShareData() ShareDataOperator {
	return ctx.shareData
}
