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

package lblogic

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/enumor"
)

func validateSession(session int) error {
	if session > 0 && (session < 30 || session > 3600) {
		return errors.New("session_expire must be '0' or between `30` and `3600`")
	}
	return nil
}

func validateScheduler(scheduler enumor.Scheduler) error {
	if scheduler != enumor.WRR && scheduler != enumor.LEAST_CONN {
		return errors.New("负载均衡算法错误")
	}
	return nil
}

func validatePort(ports []int) error {
	if len(ports) > 2 || len(ports) == 0 {
		return errors.New("端口数量错误")
	}
	for _, port := range ports {
		if port < 0 || port > 65535 {
			return fmt.Errorf("端口范围错误: %d ", port)
		}
	}
	return nil
}

func validateInstType(instType enumor.InstType) error {
	if instType != enumor.CvmInstType && instType != enumor.EniInstType {
		return errors.New("实例类型错误")
	}
	return nil
}

func validateWeight(weight int) error {
	if weight < 0 || weight > 100 {
		return errors.New("权重范围错误")
	}
	return nil
}

func validateEndPort(listenerPort, rsPort []int) error {
	if len(listenerPort) != len(rsPort) {
		return errors.New("监听器端口和RS端口数量不一致")
	}

	if len(listenerPort) == 2 && listenerPort[1]-listenerPort[0] != rsPort[1]-rsPort[0] {
		return errors.New("监听器端口和RS端口 端口段长度不一致")
	}

	return nil
}
