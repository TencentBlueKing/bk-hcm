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

package bill

import (
	"time"

	hcbillservice "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/client"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// BillConfigOption bill config option.
type BillConfigOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate BillConfigOption
func (opt *BillConfigOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AccountBillConfig account bill config.
func AccountBillConfig(kt *kit.Kit, cliSet *client.ClientSet, opt *BillConfigOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	start := time.Now()
	logs.V(3).Infof("aws account[%s] bill config start, time: %v, opt: %v, rid: %s", opt.AccountID,
		start, opt, kt.Rid)

	var hitErr error
	defer func() {
		if hitErr != nil {
			logs.Errorf("%s: account bill config failed, err: %v, account: %s, rid: %s",
				constant.AccountBillConfigFailed, hitErr, opt.AccountID, kt.Rid)
			return
		}
		logs.V(3).Infof("aws account[%s] bill config end, cost: %v, opt: %+v, rid: %s", opt.AccountID,
			time.Since(start), opt, kt.Rid)
	}()

	if hitErr = AwsBillPipeline(kt, cliSet.HCService(), opt.AccountID); hitErr != nil {
		return hitErr
	}

	return nil
}

// AwsBillPipeline aws bill pipeline
func AwsBillPipeline(kt *kit.Kit, service *hcservice.Client, accountID string) error {
	start := time.Now()
	logs.V(3).Infof("aws account[%s] bill pipeline start, time: %v, rid: %s", accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("aws account[%s] bill pipeline end, cost: %v, rid: %s",
			accountID, time.Since(start), kt.Rid)
	}()

	req := &hcbillservice.BillPipelineReq{
		AccountID: accountID,
	}
	if err := service.Aws.Bill.BillPipeline(kt.Ctx, kt.Header(), req); err != nil {
		logs.Errorf("aws account[%s] bill pipeline failed, req: %+v, err: %+v, rid: %s",
			accountID, req, err, kt.Rid)
		return err
	}

	return nil
}
