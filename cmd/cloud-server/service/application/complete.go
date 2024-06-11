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

	createmainaccount "hcm/cmd/cloud-server/service/application/handlers/main-account/create-main-account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// CompleteForCreateMainAccount 申请单完成流程
func (a *applicationSvc) CompleteForCreateMainAccount(cts *rest.Contexts) (interface{}, error) {
	// 获取请求参数
	completeReq := new(proto.MainAccountCompleteReq)
	if err := cts.DecodeInto(completeReq); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := completeReq.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询单据
	application, err := a.getApplicationBySN(cts, completeReq.SN)
	if err != nil {
		logs.Errorf("get application by sn failed, sn: %s, err: %s, rid: %s", completeReq.SN, err, cts.Kit.Rid)
		return nil, err
	}

	if application.Status != enumor.Delivering {
		logs.Errorf("application status is not delivering, sn: %s, status: %s, rid: %s", completeReq.SN, application.Status, cts.Kit.Rid)
		return nil, fmt.Errorf("application status is not delivering")
	}

	// 将执行人设置为申请人
	cts.Kit.User = application.Applicant

	// 除非交付成功，否则都属于交付失败状态
	deliverStatus := enumor.DeliverError
	deliveryDetailStr := `{"error": "unknown deliver error"}`
	defer func() {
		err := a.updateStatusWithDetail(cts, application.ID, deliverStatus, deliveryDetailStr)
		if err != nil {
			logs.Errorf("%s execute application[id=%s] delivery of %s failed, updateStatusWithDetail err: %s, rid: %s",
				constant.ApplicationDeliverFailed, application.ID, application.Type, err, cts.Kit.Rid)
			return
		}
	}()

	// 获取handler
	opt := a.getHandlerOption(cts)
	originReq, err := parseReqFromApplicationContent[proto.MainAccountCreateReq](application.Content)
	if err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	handler := createmainaccount.NewApplicationOfCreateMainAccount(opt, a.authorizer, originReq, completeReq)
	// complete
	status, deliveryDetail, err := handler.Complete()
	if err != nil {
		return nil, err
	}

	// 更新状态
	deliverStatus = status
	deliveryDetailStr, err = json.MarshalToString(deliveryDetail)
	if err != nil {
		logs.Errorf("marshal deliver detail failed, err: %v, detail: %+v, rid: %s", err, deliveryDetail, cts.Kit.Rid)

		deliverStatus = enumor.DeliverError
		deliveryDetailStr = `{"error": "marshal deliver detail failed"}`
		return nil, err
	}
	return &core.CreateResult{
		ID: application.ID,
	}, nil
}
