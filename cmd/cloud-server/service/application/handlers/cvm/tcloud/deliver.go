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

package tcloud

import (
	"errors"

	"hcm/pkg/criteria/enumor"
)

// Deliver 执行资源交付
func (a *ApplicationOfCreateTCloudCvm) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// 创建主机
	result, err := a.Client.HCService().TCloud.Cvm.BatchCreateCvm(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		a.toHcProtoTCloudBatchCreateReq(false),
	)
	if err != nil || result == nil {
		return enumor.DeliverError, map[string]interface{}{"error": err}, err
	}

	// 全部失败
	if len(result.FailedCloudIDs) == int(a.req.RequiredCount) {
		return enumor.DeliverError, map[string]interface{}{
			"error": result.FailedMessage,
		}, errors.New(result.FailedMessage)
	}

	status := enumor.Completed
	// 部分成功
	if len(result.SuccessCloudIDs) != int(a.req.RequiredCount) {
		status = enumor.DeliverPartial
	}
	deliverDetail := map[string]interface{}{
		"success_cloud_ids": result.SuccessCloudIDs,
		"failed_cloud_ids":  result.FailedCloudIDs,
		"error":             errors.New(result.FailedMessage),
	}
	return status, deliverDetail, nil

	// TODO: 云ID查询主机
	// TODO: 主机分配给业务，同时更新主机备注
	// TODO: 主机关联资源分配给业务
}
