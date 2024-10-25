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

package application

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// CancelApplication ...
func (a *applicationSvc) CancelApplication(cts *rest.Contexts) (interface{}, error) {
	applicationID := cts.PathParameter("application_id").String()
	application, err := a.client.DataService().Global.Application.GetApplication(
		cts.Kit.Ctx, cts.Kit.Header(), applicationID)
	if err != nil {
		return nil, err
	}

	// 只能查询自己的申请单
	if application.Applicant != cts.Kit.User {
		return nil, errf.NewFromErr(
			errf.PermissionDenied, fmt.Errorf("you can not operate other people's application"),
		)
	}

	// 根据SN调用ITSM接口撤销单据
	err = a.itsmCli.WithdrawTicket(cts.Kit, application.SN, cts.Kit.User)
	if err != nil {
		return nil, fmt.Errorf("call itsm cancel ticket api failed, err: %v", err)
	}

	// 更新状态
	err = a.updateStatusWithDetail(cts, applicationID, enumor.Cancelled, "")
	if err != nil {
		return nil, err
	}

	return nil, nil
}
