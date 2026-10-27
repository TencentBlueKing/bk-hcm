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

// Package query tf_provider 与 http 插件共用的注册表操作：读取、落 init、条件更新、删除、标记失败，以及 Schema 是否已落库。
// 状态推进统一走 expect_state 条件更新，审计由 data-service 在同一事务内写入。
package query

import (
	"encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
)

// List 按过滤条件翻页读出全部插件记录。
func List(kt *kit.Kit, cli *dataservice.Client, flt *filter.Expression) ([]coreplugin.IaCPlugin, error) {
	var records []coreplugin.IaCPlugin
	page := core.NewDefaultBasePage()
	for {
		res, err := cli.Global.IaCPlugin.List(kt, &core.ListReq{Filter: flt, Page: page})
		if err != nil {
			return nil, fmt.Errorf("list plugin records failed: %w", err)
		}

		records = append(records, res.Details...)
		if uint(len(res.Details)) < page.Limit {
			return records, nil
		}
		page.Start += uint32(page.Limit)
	}
}

// Insert 插入一条 init 记录，返回记录 ID。
func Insert(kt *kit.Kit, cli *dataservice.Client, req *dataplugin.IaCPluginCreateReq) (string, error) {
	if req.State == "" {
		req.State = enumor.IaCPluginStateInit
	}

	created, err := cli.Global.IaCPlugin.Create(kt, req)
	if err != nil {
		return "", fmt.Errorf("create plugin %s@%s record failed: %w", req.PluginName, req.PluginVersion, err)
	}
	if len(created.IDs) != 1 {
		return "", fmt.Errorf("create plugin %s@%s got %d ids, expect 1", req.PluginName, req.PluginVersion,
			len(created.IDs))
	}

	return created.IDs[0], nil
}

// Update 按 expect 条件更新记录，当前状态与 expect 不一致时更新失败。
//   - 更新状态为 state 并清空 message（原 message 已为空时不会计入变更或审计）；
//   - ext 非空时覆盖 extension；
//   - memo 非 nil 时覆盖备注。
func Update(kt *kit.Kit, cli *dataservice.Client, id string, expect, state enumor.IaCPluginState,
	ext json.RawMessage, memo *string) error {

	return cli.Global.IaCPlugin.Update(kt, &dataplugin.IaCPluginUpdateReq{
		ID: id, ExpectState: expect, State: state, Message: cvt.ValToPtr(""), Extension: ext, Memo: memo,
	})
}

// Delete 按过滤条件删除插件记录。
func Delete(kt *kit.Kit, cli *dataservice.Client, flt *filter.Expression) error {
	return cli.Global.IaCPlugin.BatchDelete(kt, &dataplugin.IaCPluginBatchDeleteReq{Filter: flt})
}

// MarkFailed 把当前状态为 expectState 的记录置为 failed 并写入按列宽截断的失败原因。
func MarkFailed(kt *kit.Kit, cli *dataservice.Client, id string, expectState enumor.IaCPluginState,
	cause error) error {

	msg := Truncate(cause.Error(), dataplugin.IaCPluginMessageMaxLen)
	return cli.Global.IaCPlugin.Update(kt, &dataplugin.IaCPluginUpdateReq{
		ID: id, ExpectState: expectState, State: enumor.IaCPluginStateFailed, Message: cvt.ValToPtr(msg),
	})
}

// Truncate 按字符截断，避免截断在多字节字符中间。
func Truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}

	return string(r[:max])
}

// SchemaExists 按插件名和版本查询 Schema 是否已落库。
func SchemaExists(kt *kit.Kit, cli *dataservice.Client, pluginName, pluginVersion string) (bool, error) {
	resp, err := cli.Global.IacPluginSchema.List(kt, &core.ListReq{
		Fields: []string{"id"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("plugin_name", pluginName),
			tools.RuleEqual("plugin_version", pluginVersion),
		),
		Page: &core.BasePage{Limit: 1},
	})
	if err != nil {
		return false, fmt.Errorf("list plugin %s@%s schema failed: %w", pluginName, pluginVersion, err)
	}

	return resp != nil && len(resp.Details) > 0, nil
}
