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

package config

import (
	"encoding/json"
	"fmt"

	"hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	tablegconf "hcm/pkg/dal/table/global-config"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
)

// SpringResPoolIf provides management interface for operations of spring resource pool config
type SpringResPoolIf interface {
	// GetChargeType get charge type config by biz id
	GetChargeType(kt *kit.Kit, bizID int64) (cvmapi.ChargeType, error)
	// UpsertChargeType upsert charge type config
	UpsertChargeType(kt *kit.Kit, req *config.UpsertSpringResPoolChargeTypeReq) error
	// DeleteChargeType delete charge type config
	DeleteChargeType(kt *kit.Kit, bizID *int64) error
}

// NewSpringResPoolOp creates a spring resource pool interface
func NewSpringResPoolOp(cli *client.ClientSet) SpringResPoolIf {
	return &springResPool{
		client: cli,
	}
}

type springResPool struct {
	client *client.ClientSet
}

// GetChargeType get charge type config by biz id
func (s *springResPool) GetChargeType(kt *kit.Kit, bizID int64) (cvmapi.ChargeType, error) {
	// 1. 先查询业务专属配置
	bizConfigKey := enumor.GetBizSpringResPoolChargeTypeKey(bizID)
	bizConfig, exist, err := s.getConfigByKey(kt, bizConfigKey)
	if err != nil {
		logs.Errorf("failed to get biz spring res pool charge type config, bizID: %d, err: %v, rid: %s",
			bizID, err, kt.Rid)
		return "", err
	}
	if exist {
		return s.parseChargeType(bizConfig.ConfigValue)
	}

	// 2. 如果业务专属配置不存在，查询全局配置
	globalConfig, exist, err := s.getConfigByKey(kt, string(enumor.GlobalConfigKeySpringResPoolChargeTypeGlobal))
	if err != nil {
		logs.Errorf("failed to get spring res pool charge type config, bizID: %d, err: %v, rid: %s", bizID, err, kt.Rid)
		return "", err
	}
	if !exist {
		return cvmapi.ChargeTypePostPaidByHour, nil
	}

	return s.parseChargeType(globalConfig.ConfigValue)
}

// UpsertChargeType upsert charge type config
func (s *springResPool) UpsertChargeType(kt *kit.Kit, req *config.UpsertSpringResPoolChargeTypeReq) error {
	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	configKey := string(enumor.GlobalConfigKeySpringResPoolChargeTypeGlobal)
	if req.BizID != nil {
		configKey = enumor.GetBizSpringResPoolChargeTypeKey(cvt.PtrToVal(req.BizID))
	}

	// 查询配置是否存在
	existingConfig, exist, err := s.getConfigByKey(kt, configKey)
	if err != nil {
		logs.Errorf("failed to get spring res pool charge type config, key: %s, err: %v, rid: %s",
			configKey, err, kt.Rid)
		return err
	}

	if exist {
		// 更新配置
		updateReq := &datagconf.BatchUpdateReq{
			Configs: []cgconf.GlobalConfig{{ID: existingConfig.ID, ConfigValue: req.ChargeType}},
		}
		if err = s.client.DataService().Global.GlobalConfig.BatchUpdate(kt, updateReq); err != nil {
			logs.Errorf("failed to update spring res pool charge type config, key: %s, err: %v, rid: %s",
				configKey, err, kt.Rid)
			return err
		}
		return nil
	}

	// 创建配置
	createReq := &datagconf.BatchCreateReq{
		Configs: []cgconf.GlobalConfig{
			{
				ConfigType:  string(enumor.GlobalConfigTypeSpringResPool),
				ConfigKey:   configKey,
				ConfigValue: req.ChargeType,
			},
		},
	}
	if _, err = s.client.DataService().Global.GlobalConfig.BatchCreate(kt, createReq); err != nil {
		logs.Errorf("failed to create spring res pool charge type config, key: %s, err: %v, rid: %s",
			configKey, err, kt.Rid)
		return err
	}

	return nil
}

// DeleteChargeType delete charge type config
func (s *springResPool) DeleteChargeType(kt *kit.Kit, bizID *int64) error {
	configKey := string(enumor.GlobalConfigKeySpringResPoolChargeTypeGlobal)
	if bizID != nil {
		configKey = enumor.GetBizSpringResPoolChargeTypeKey(cvt.PtrToVal(bizID))
	}

	// 查询配置是否存在
	existingConfig, exist, err := s.getConfigByKey(kt, configKey)
	if err != nil {
		logs.Errorf("failed to get existing config, key: %s, err: %v, rid: %s", configKey, err, kt.Rid)
		return err
	}
	if !exist {
		logs.Errorf("spring res pool charge type config not found, key: %s, rid: %s", configKey, kt.Rid)
		return fmt.Errorf("spring res pool charge type config not found, key: %s, rid: %s", configKey, kt.Rid)
	}

	// 删除配置
	deleteReq := &datagconf.BatchDeleteReq{
		BatchDeleteReq: core.BatchDeleteReq{
			IDs: []string{existingConfig.ID},
		},
	}
	if err = s.client.DataService().Global.GlobalConfig.BatchDelete(kt, deleteReq); err != nil {
		logs.Errorf("failed to delete spring res pool charge type config, key: %s, err: %v, rid: %s",
			configKey, err, kt.Rid)
		return err
	}

	return nil
}

// getConfigByKey get config by config key
func (s *springResPool) getConfigByKey(kt *kit.Kit, configKey string) (*tablegconf.GlobalConfigTable, bool, error) {
	filter := tools.ExpressionAnd(
		tools.RuleEqual("config_type", string(enumor.GlobalConfigTypeSpringResPool)),
		tools.RuleEqual("config_key", configKey),
	)

	dataReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewDefaultBasePage(),
	}
	dataResp, err := s.client.DataService().Global.GlobalConfig.List(kt, dataReq)
	if err != nil {
		logs.Errorf("failed to get config, key: %s, err: %v, rid: %s", configKey, err, kt.Rid)
		return nil, false, err
	}

	if len(dataResp.Details) == 0 {
		return nil, false, nil
	}

	return &dataResp.Details[0], true, nil
}

// parseChargeType parse charge type from config value
func (s *springResPool) parseChargeType(configValue tabletypes.JsonField) (cvmapi.ChargeType, error) {
	var chargeType cvmapi.ChargeType
	if err := json.Unmarshal([]byte(configValue), &chargeType); err != nil {
		return "", err
	}
	if err := chargeType.Validate(); err != nil {
		return "", err
	}

	return chargeType, nil
}
