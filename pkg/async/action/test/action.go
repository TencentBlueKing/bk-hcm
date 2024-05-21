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

package test

import (
	"errors"
	"math/rand"
	"time"

	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/times"
)

var _ action.Action = new(CreateFactory)
var _ action.ParameterAction = new(CreateFactory)

// CreateFactory ...
type CreateFactory struct{}

// TestCreateFactoryParams test create factory params.
type TestCreateFactoryParams struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var _ action.Decoder = new(TestCreateFactoryParams)

// Decode ...
func (t TestCreateFactoryParams) Decode(params types.JsonField, v interface{}) error {
	logs.Infof(" ----------- create factory decode -----------, params: %+v", params)
	return json.UnmarshalFromString(string(params), v)
}

// ParameterNew ...
func (act CreateFactory) ParameterNew() interface{} {
	return new(TestCreateFactoryParams)
}

// Name ...
func (act CreateFactory) Name() enumor.ActionName {
	return enumor.ActionCreateFactoryTest
}

// Run ...
func (act CreateFactory) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	logs.Infof(" ----------- create factory -----------, params: %+v, rid: %s", params, kt.Kit().Rid)
	return nil, kt.ShareData().Set(kt.Kit(), "name", "create_factory")
}

var _ action.Action = new(Produce)

// Produce ...
type Produce struct{}

// Name ...
func (p Produce) Name() enumor.ActionName {
	return enumor.ActionProduceTest
}

// Run ...
func (p Produce) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	logs.Infof(" ----------- Produce -----------, rid: %s", kt.Kit().Rid)
	return nil, nil
}

var _ action.Action = new(Assemble)

// Assemble ...
type Assemble struct{}

// Name ...
func (a Assemble) Name() enumor.ActionName {
	return enumor.ActionAssembleTest
}

// Run ...
func (a Assemble) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {

	logs.Infof(" ----------- Assemble -----------, rid: %s", kt.Kit().Rid)

	name, exist := kt.ShareData().Get("name")
	if !exist {
		return nil, errors.New("name not found from share_name")
	}

	logs.Infof(" ----------- Assemble Get ShareData: %v -----------, rid: %s", name, kt.Kit().Rid)

	return nil, nil
}

var _ action.Action = new(Sleep)
var _ action.ParameterAction = new(Sleep)
var _ action.RollbackAction = new(Sleep)

// Sleep ...
type Sleep struct{}

// Rollback ...
func (s Sleep) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- Sleep Rollback -----------, rid: %s", kt.Kit().Rid)
	return nil
}

// SleepParams define Sleep params.
type SleepParams struct {
	ID       string `json:"id"`
	SleepSec int    `json:"sleep_sec"`
}

// ParameterNew ...
func (s Sleep) ParameterNew() interface{} {
	return new(SleepParams)
}

// Name ...
func (s Sleep) Name() enumor.ActionName {
	return enumor.ActionSleep
}

// Run ...
func (s Sleep) Run(kt run.ExecuteKit, params interface{}) (ret any, err error) {
	p := params.(*SleepParams)
	logs.Infof("[Sleep-%s] %ds start  , time: %v, rid: %s", p.ID, p.SleepSec,
		times.ConvStdTimeNow(), kt.Kit().Rid)
	// failed randomly, quick failed
	shouldFailed := false
	if rand.Float64() > 0.5 {
		shouldFailed = true
		p.SleepSec /= 10
		logs.Infof("[Sleep-%s]: ready to fail ", p.ID)
	}

	timer := time.NewTimer(time.Duration(p.SleepSec) * time.Second)
	defer timer.Stop()
	select {
	case <-kt.Kit().Ctx.Done():
		err = kt.Kit().Ctx.Err()
		if err != nil {
			logs.Infof("[Sleep-%s] got ERROR, err: %v, rid: %s", p.ID, p.SleepSec,
				err, kt.Kit().Rid)
		}

	case <-timer.C:
		logs.Infof("[Sleep-%s] %ds normal  end, time: %v, rid: %s", p.ID, p.SleepSec,
			times.ConvStdTimeNow(), kt.Kit().Rid)
	}

	if shouldFailed {
		err = errors.New("planned failed")
	}
	return nil, err
}
