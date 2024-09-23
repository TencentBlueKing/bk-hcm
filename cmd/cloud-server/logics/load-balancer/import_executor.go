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
	"encoding/json"
	"fmt"
	"strconv"

	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

// ImportExecutor 导入执行器
type ImportExecutor interface {
	// Execute 导入执行器的唯一入口, 内部执行流程：unmarshalData -> validate -> filter
	// -> buildTaskManagementAndDetails -> buildFlows -> updateTaskManagementAndDetails
	Execute(*kit.Kit, string, json.RawMessage) (string, error)

	// unmarshalData 传入json数据, 反序列化到预览数据结构中
	unmarshalData(json.RawMessage) error
	// validate 校验数据合法性, 主要调用 ImportPreviewExecutor的validate方法
	validate(*kit.Kit) error
	// filter filter existing record
	filter()
	// buildFlows 构建异步任务
	buildFlows(*kit.Kit) ([]string, error)
	// buildTaskManagementAndDetails 创建任务管理&任务详情
	buildTaskManagementAndDetails(kt *kit.Kit, source string) (string, error)
	// updateTaskManagementAndDetails 更新任务管理 flowID & 任务详情关于异步任务的详细数据
	updateTaskManagementAndDetails(kt *kit.Kit, flowIDs []string, taskID string) error
}

// NewImportExecutor ...
func NewImportExecutor(operationType OperationType, dataCli *dataservice.Client,
	taskCli *taskserver.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) (ImportExecutor, error) {

	switch operationType {
	//case CreateLayer4Listener:
	//	return newCreateLayer4ListenerExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case CreateLayer7Listener:
		return newCreateLayer7ListenerExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	//case CreateUrlRule:
	//	return newCreateUrlRuleExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	//case Layer4ListenerBindRs:
	//	return newLayer4ListenerBindRSExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	//case Layer7ListenerBindRs:
	//	return newLayer7ListenerBindRSExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operationType)
	}
}

func unmarshalData(message json.RawMessage, data interface{}) error {
	rawByte, err := message.MarshalJSON()
	if err != nil {
		return err
	}
	err = json.Unmarshal(rawByte, &data)
	if err != nil {
		return err
	}
	return nil
}

func newActionIDGenerator(init, base int) func() (cur string, prev string) {
	current := init
	return func() (cur string, prev string) {
		cur = strconv.FormatInt(int64(current), base)
		if current-1 >= init {
			prev = strconv.FormatInt(int64(current-1), base)
		}
		current++
		return cur, prev
	}
}
