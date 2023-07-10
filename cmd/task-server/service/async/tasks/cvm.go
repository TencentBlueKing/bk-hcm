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

package tasks

import (
	"fmt"
	"sync"

	"hcm/pkg/async"
)

var cvmOnce sync.Once

type cvm struct {
}

func init() {
	cvmOnce.Do(func() {
		cvm := new(cvm)
		async.GetTaskManager().RegisterTask(async.HcmTaskManager, "tcloud_create_cvm", cvm.TCloudCreateCvm)
		async.GetTaskManager().RegisterTask(async.HcmTaskManager, "huawei_create_cvm", cvm.HuaWeiCreateCvm)
		async.GetTaskManager().RegisterTask(async.HcmTaskManager, "check_create_cvm", cvm.CheckCreateCvm)
	})
}

// TCloudCreateCvm test
func (cvm *cvm) TCloudCreateCvm() (bool, error) {
	// TODO: 实际是调用hc-service等接口

	return false, nil
}

// HuaWeiCreateCvm test
func (cvm *cvm) HuaWeiCreateCvm() (bool, error) {
	// TODO: 实际是调用hc-service等接口

	return false, fmt.Errorf("create huawei cvm failed")
}

// CheckCreateCvm test
func (cvm *cvm) CheckCreateCvm(args ...bool) (bool, error) {
	// TODO: 实际是调用hc-service等接口

	ret := true

	for _, arg := range args {
		if !arg {
			ret = false
			break
		}
	}

	return ret, nil
}
