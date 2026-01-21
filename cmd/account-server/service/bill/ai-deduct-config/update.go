/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package aideductconfig

import (
	"encoding/json"

	actbill "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	gccore "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// UpdateAIDeductConfig 更新AI账单抵扣配置
func (s *aiDeductConfigSvc) UpdateAIDeductConfig(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	req := new(actbill.UpdateAIDeductConfigReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 权限校验
	err := s.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Update}})
	if err != nil {
		return nil, err
	}

	// 获取配置key
	configKey, err := getConfigKeyByVendor(vendor)
	if err != nil {
		logs.Errorf("fail to get config key, err: %v, vendor: %s, rid: %s", err, vendor, cts.Kit.Rid)
		return nil, err
	}
	// 序列化配置值
	configValueBytes, err := json.Marshal(req.MainAccountIDs)
	if err != nil {
		logs.Errorf("fail to marshal main account ids, vendor: %s, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
		return nil, err
	}

	// 先查询配置是否存在
	flt := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeBillAIDeduct),
			tools.RuleEqual("config_key", configKey),
		},
	}
	listReq := &datagconf.ListReq{Filter: flt, Page: core.NewDefaultBasePage()}
	listResp, err := s.client.DataService().Global.GlobalConfig.List(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to list ai deduct config, vendor: %s, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) > 0 {
		// 配置已存在，更新
		updateReq := &datagconf.BatchUpdateReq{Configs: []gccore.GlobalConfig{
			{ID: listResp.Details[0].ID, ConfigValue: json.RawMessage(configValueBytes)},
		}}
		if err = s.client.DataService().Global.GlobalConfig.BatchUpdate(cts.Kit, updateReq); err != nil {
			logs.Errorf("fail to update ai deduct config, vendor: %s, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	}

	// 配置不存在，创建
	createReq := &datagconf.BatchCreateReq{Configs: []gccore.GlobalConfig{
		{
			ConfigKey:   configKey,
			ConfigType:  string(enumor.GlobalConfigTypeBillAIDeduct),
			ConfigValue: json.RawMessage(configValueBytes),
		},
	}}
	if _, err = s.client.DataService().Global.GlobalConfig.BatchCreate(cts.Kit, createReq); err != nil {
		logs.Errorf("fail to create ai deduct config, vendor: %s, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
