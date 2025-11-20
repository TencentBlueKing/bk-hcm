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

// Package config config
package config

import (
	"encoding/json"
	"errors"
	"time"

	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	globalconf "hcm/pkg/dal/table/global-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Config provides interface for operations of dissolve config.
type Config interface {
	GetDissolveHostApplyTime(kt *kit.Kit) (*time.Time, error)
	UpsertDissolveHostApplyTime(kt *kit.Kit, time *time.Time) error
	GetApprovalLimit(kt *kit.Kit) (*float64, error)
	UpsertApprovalLimit(kt *kit.Kit, approveLimit *float64) error
}

type logics struct {
	cliSet *client.ClientSet
}

// New create dissolve config logics.
func New(client *client.ClientSet) Config {
	return &logics{
		cliSet: client,
	}
}

// GetDissolveHostApplyTime get dissolve host apply time.
func (l *logics) GetDissolveHostApplyTime(kt *kit.Kit) (*time.Time, error) {
	config, exist, err := l.getDissolveConfigByKey(kt, enumor.GlobalConfigDissolveHostApplyTime)
	if err != nil {
		logs.Errorf("failed to get dissolve host apply time config, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if !exist {
		logs.Errorf("dissolve host apply time config not exist, rid: %s", kt.Rid)
		return nil, errors.New("dissolve host apply time config not exist")
	}

	applyTime := new(time.Time)
	if err = json.Unmarshal([]byte(config.ConfigValue), &applyTime); err != nil {
		logs.Errorf("failed to unmarshal config value, err: %v, value: %s, rid: %s", err, config.ConfigValue, kt.Rid)
		return nil, err
	}

	return applyTime, nil
}

func (l *logics) getDissolveConfigByKey(kt *kit.Kit, key enumor.GlobalConfigResDissolveKey) (
	*globalconf.GlobalConfigTable, bool, error) {

	req := core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigResDissolve),
			tools.RuleJSONEqual("config_key", key),
		),
		Page: core.NewDefaultBasePage(),
	}
	cfgResp, err := l.cliSet.DataService().Global.GlobalConfig.List(kt, &req)
	if err != nil {
		logs.Errorf("failed to list global config, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, false, err
	}
	if len(cfgResp.Details) == 0 {
		return nil, false, nil
	}

	return &cfgResp.Details[0], true, nil
}

// UpsertDissolveHostApplyTime upsert dissolve host apply time.
func (l *logics) UpsertDissolveHostApplyTime(kt *kit.Kit, time *time.Time) error {
	return l.upsertDissolveConfig(kt, enumor.GlobalConfigDissolveHostApplyTime, time)
}

func (l *logics) upsertDissolveConfig(kt *kit.Kit, key enumor.GlobalConfigResDissolveKey, value interface{}) error {
	if value == nil {
		logs.Errorf("value is nil, key: %s, rid: %s", key, kt.Rid)
		return errors.New("value is nil")
	}

	oldConf, exist, err := l.getDissolveConfigByKey(kt, key)
	if err != nil {
		logs.Errorf("failed to get dissolve config, err: %v, key: %s, rid: %s", err, key, kt.Rid)
		return err
	}

	conf := cgconf.GlobalConfigT[any]{
		ConfigType:  string(enumor.GlobalConfigResDissolve),
		ConfigKey:   string(key),
		ConfigValue: value,
	}
	if !exist {
		createReq := datagconf.BatchCreateReqT[any]{Configs: []cgconf.GlobalConfigT[any]{conf}}
		if _, err = l.cliSet.DataService().Global.GlobalConfig.BatchCreate(kt, &createReq); err != nil {
			logs.Errorf("failed to create dissolve global config, err: %v, req: %+v, rid: %s", err, createReq, kt.Rid)
			return err
		}
		return nil
	}

	conf.ID = oldConf.ID
	updateReq := datagconf.BatchUpdateReq{Configs: []cgconf.GlobalConfigT[any]{conf}}
	if err = l.cliSet.DataService().Global.GlobalConfig.BatchUpdate(kt, &updateReq); err != nil {
		logs.Errorf("failed to update dissolve global config, err: %v, req: %+v, rid: %s", err, updateReq, kt.Rid)
		return err
	}

	return nil
}

// GetApprovalLimit get approval limit.
func (l *logics) GetApprovalLimit(kt *kit.Kit) (*float64, error) {
	config, exist, err := l.getDissolveConfigByKey(kt, enumor.GlobalConfigDissolveApprovalLimit)
	if err != nil {
		logs.Errorf("failed to get dissolve approval limit config, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if !exist {
		logs.Errorf("dissolve approval limit config not exist, rid: %s", kt.Rid)
		return nil, errors.New("dissolve approval limit config not exist")
	}

	approvalLimit := new(float64)
	if err = json.Unmarshal([]byte(config.ConfigValue), &approvalLimit); err != nil {
		logs.Errorf("failed to unmarshal config value, err: %v, value: %s, rid: %s", err, config.ConfigValue, kt.Rid)
		return nil, err
	}

	return approvalLimit, nil
}

// UpsertApprovalLimit upsert approval limit.
func (l *logics) UpsertApprovalLimit(kt *kit.Kit, approvalLimit *float64) error {
	return l.upsertDissolveConfig(kt, enumor.GlobalConfigDissolveApprovalLimit, approvalLimit)
}
