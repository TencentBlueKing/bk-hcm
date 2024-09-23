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

	"github.com/shopspring/decimal"
)

func newAwsRunner(taskType enumor.MonthTaskType) (MonthTaskRunner, error) {
	switch taskType {
	case enumor.AwsSupportMonthTask:
		return &AwsSupportMonthTask{}, nil
	case enumor.AwsSavingsPlansMonthTask:
		return &AwsSavingsPlanMonthTask{}, nil
	default:
		return nil, errors.New("not support task type of aws: " + string(taskType))
	}
}

type awsMonthTaskBaseRunner struct {
	spArnPrefix            string
	spMainAccountCloudID   string
	excludeAccountCloudIds []string
}

func (a *awsMonthTaskBaseRunner) initExtension(opt *MonthTaskActionOption) {
	if opt.Extension == nil {
		return
	}

	a.spArnPrefix = opt.Extension[constant.AwsSavingsPlanARNPrefixKey]
	a.spMainAccountCloudID = opt.Extension[constant.AwsSavingsPlanAccountCloudIDKey]
	if opt.Extension[constant.AwsCommonExpenseExcludeCloudIDKey] != "" {
		excludeCloudIDStr := opt.Extension[constant.AwsCommonExpenseExcludeCloudIDKey]
		excluded := strings.Split(excludeCloudIDStr, ",")
		a.excludeAccountCloudIds = excluded
	}
}

// listMainAccount rootAsMainAccount 作为二级账号存在的根账号，将分摊后的账单抵冲该账号支出
func (a awsMonthTaskBaseRunner) listMainAccount(kt *kit.Kit, rootAccount *dataproto.AwsRootAccount) (
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAsMainAccount *protocore.BaseMainAccount, err error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("parent_account_id", rootAccount.ID)),
		Page:   core.NewDefaultBasePage(),
	}
	mainAccountsResp, err := actcli.GetDataService().Global.MainAccount.List(kt, listReq)
	if err != nil {
		logs.Errorf("failt to list main account for %s month task, err: %v, rid: %s",
			enumor.Aws, err, kt.Rid)
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

	return mainAccountMap, rootAsMainAccount, nil
}

func getDecimal(dict map[string]string, key string) (*decimal.Decimal, error) {
	val, ok := dict[key]
	if !ok {
		return nil, errors.New("key not found: " + key)
	}
	d, err := decimal.NewFromString(val)
	if err != nil {
		return nil, fmt.Errorf("fail to convert to decimal, key: %s, value: %s, err: %v", key, val, err)
	}
	return &d, nil
}

func convAwsBillItemExtension(productName string, opt *MonthTaskActionOption, rootAccountCloudID string,
	mainAccountCloudID string, currencyCode enumor.CurrencyCode, cost decimal.Decimal) ([]byte, error) {

	ext := billcore.AwsRawBillItem{
		Year:                     fmt.Sprintf("%4d", opt.BillYear),
		Month:                    fmt.Sprintf("%02d", opt.BillMonth),
		BillPayerAccountId:       rootAccountCloudID,
		LineItemUsageAccountId:   mainAccountCloudID,
		LineItemCurrencyCode:     string(currencyCode),
		LineItemNetUnblendedCost: cost.String(),
		LineItemProductCode:      productName,
		ProductProductName:       productName,
		PricingCurrency:          string(currencyCode),
	}
	return json.Marshal(ext)
}
