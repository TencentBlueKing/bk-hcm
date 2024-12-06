/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package monthtask

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	billcore "hcm/pkg/api/core/bill"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
	"github.com/shopspring/decimal"
)

func newHuaweiRunner(taskType enumor.MonthTaskType) (MonthTaskRunner, error) {
	switch taskType {
	case enumor.HuaweiSupportMonthTask:
		return &HuaweiSupportMonthTask{}, nil
	default:
		return nil, errors.New("not support task type of huawei: " + string(taskType))
	}
}

type huaweiMonthTaskBaseRunner struct {
	excludeAccountCloudIds []string
}

func (a *huaweiMonthTaskBaseRunner) initExtension(opt *MonthTaskActionOption) {
	if opt.Extension == nil {
		return
	}

	if opt.Extension[constant.HuaweiCommonExpenseExcludeCloudIDKey] != "" {
		excludeCloudIDStr := opt.Extension[constant.HuaweiCommonExpenseExcludeCloudIDKey]
		excluded := strings.Split(excludeCloudIDStr, ",")
		a.excludeAccountCloudIds = excluded
	}
}

// listMainAccount rootAsMainAccount 作为二级账号存在的根账号，将分摊后的账单抵冲该账号支出
func (a huaweiMonthTaskBaseRunner) listMainAccount(kt *kit.Kit, rootAccount *dataproto.HuaweiRootAccount) (
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAsMainAccount *protocore.BaseMainAccount, err error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("parent_account_id", rootAccount.ID)),
		Page:   core.NewDefaultBasePage(),
	}
	mainAccountsResp, err := actcli.GetDataService().Global.MainAccount.List(kt, listReq)
	if err != nil {
		logs.Errorf("failt to list main account for %s month task, err: %v, rid: %s",
			enumor.HuaWei, err, kt.Rid)
		return nil, nil, err
	}
	mainAccountMap = make(map[string]*protocore.BaseMainAccount, len(mainAccountsResp.Details))
	for _, account := range mainAccountsResp.Details {
		mainAccountMap[account.ID] = account
		// 查找作为主账号录入的根账号
		if account.CloudID == rootAccount.CloudID {
			rootAsMainAccount = account
		}
	}
	if rootAsMainAccount == nil {
		return nil, nil, errors.New("can not found root as main account " + rootAccount.CloudID)
	}

	return mainAccountMap, rootAsMainAccount, nil
}

func convHuaweiBillItemExtension(productName string, opt *MonthTaskActionOption, mainAccountCloudID string,
	cost decimal.Decimal) ([]byte, error) {

	record := model.ResFeeRecordV2{
		BillDate:             cvt.ValToPtr(fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, 1)),
		BillType:             cvt.ValToPtr(constant.HuaweiBillTypePurchase),
		CustomerId:           cvt.ValToPtr(mainAccountCloudID),
		CloudServiceType:     cvt.ValToPtr(productName),
		ResourceType:         cvt.ValToPtr(productName),
		CloudServiceTypeName: cvt.ValToPtr(productName),
		ResourceTypeName:     cvt.ValToPtr(productName),
		ChargeMode:           cvt.ValToPtr(constant.HuaweiBillChargeModeMonthlyYearly),
		DebtAmount:           cvt.ValToPtr(cost.InexactFloat64()),
	}
	ext := billcore.HuaweiRawBillItem{ResFeeRecordV2: record}
	return json.Marshal(ext)
}
