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

package global

import (
	"fmt"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/client/common"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
)

// IaCPluginClient is data service iac plugin api client.
type IaCPluginClient struct {
	client rest.ClientInterface
}

// NewIaCPluginClient create a new iac plugin api client.
func NewIaCPluginClient(client rest.ClientInterface) *IaCPluginClient {
	return &IaCPluginClient{client: client}
}

// Create create iac plugin record.
func (c *IaCPluginClient) Create(kt *kit.Kit, req *dataplugin.IaCPluginCreateReq) (*core.BatchCreateResult, error) {
	return common.Request[dataplugin.IaCPluginCreateReq, core.BatchCreateResult](
		c.client, rest.POST, kt, req, "/iac_plugins/create")
}

// BatchDelete 按过滤条件删除 iac plugin 记录。
func (c *IaCPluginClient) BatchDelete(kt *kit.Kit, req *dataplugin.IaCPluginBatchDeleteReq) error {
	return common.RequestNoResp[dataplugin.IaCPluginBatchDeleteReq](
		c.client, rest.DELETE, kt, req, "/iac_plugins/batch")
}

// Update update iac plugin record. Set req.CurrentState to update only when the stored state equals it.
func (c *IaCPluginClient) Update(kt *kit.Kit, req *dataplugin.IaCPluginUpdateReq) error {
	return common.RequestNoResp[dataplugin.IaCPluginUpdateReq](
		c.client, rest.PATCH, kt, req, "/iac_plugins/update")
}

// List list iac plugin records.
func (c *IaCPluginClient) List(kt *kit.Kit, req *core.ListReq) (*dataplugin.IaCPluginListResult, error) {
	return common.Request[core.ListReq, dataplugin.IaCPluginListResult](
		c.client, rest.POST, kt, req, "/iac_plugins/list")
}

// ListAll 翻页查询满足过滤条件的全部 iac plugin 记录。
func (c *IaCPluginClient) ListAll(kt *kit.Kit, expr *filter.Expression,
	fields []string) ([]coreplugin.IaCPlugin, error) {

	details := make([]coreplugin.IaCPlugin, 0)
	page := core.NewDefaultBasePage()
	for {
		result, err := c.List(kt, &core.ListReq{Filter: expr, Fields: fields, Page: page})
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, fmt.Errorf("list iac plugin got empty response")
		}

		details = append(details, result.Details...)
		if uint(len(result.Details)) < page.Limit {
			return details, nil
		}
		page.Start += uint32(page.Limit)
	}
}

// CreateInit 插入一条 init 状态的插件记录并返回其记录 ID。
func (c *IaCPluginClient) CreateInit(kt *kit.Kit, req *dataplugin.IaCPluginCreateReq) (string, error) {
	if req.State == "" {
		req.State = enumor.IaCPluginStateInit
	}

	created, err := c.Create(kt, req)
	if err != nil {
		return "", fmt.Errorf("create plugin %s@%s record failed: %w", req.PluginName, req.PluginVersion, err)
	}
	if len(created.IDs) != 1 {
		return "", fmt.Errorf("create plugin %s@%s got %d ids, expect 1", req.PluginName, req.PluginVersion,
			len(created.IDs))
	}

	return created.IDs[0], nil
}

// MarkFailed 将当前状态为 expectState 的插件记录置为 failed，并写入按列宽截断的失败原因。
func (c *IaCPluginClient) MarkFailed(kt *kit.Kit, id string, expectState enumor.IaCPluginState,
	cause error) error {

	msg := cause.Error()
	if r := []rune(msg); len(r) > constant.IaCPluginMessageMaxLen {
		msg = string(r[:constant.IaCPluginMessageMaxLen])
	}
	return c.Update(kt, &dataplugin.IaCPluginUpdateReq{
		ID:           id,
		CurrentState: expectState,
		State:        enumor.IaCPluginStateFailed,
		Message:      cvt.ValToPtr(msg),
	})
}
