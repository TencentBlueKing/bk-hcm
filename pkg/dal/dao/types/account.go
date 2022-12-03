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

package types

//
//import (
//	"hcm/pkg/criteria/errf"
//	"hcm/pkg/dal/table"
//	"hcm/pkg/runtime/filter"
//)
//
//// ListAccountsOption defines options to list accounts.
//type ListAccountsOption struct {
//	Filter *filter.Expression `json:"filter"`
//	Page   *BasePage          `json:"page"`
//}
//
//// Validate the list account options
//func (lao *ListAccountsOption) Validate(po *PageOption) error {
//	if lao.Filter == nil {
//		return errf.New(errf.InvalidParameter, "filter is nil")
//	}
//
//	exprOpt := &filter.ExprOption{
//		// 用来剔除查询语句中手写的查询条件，e.g: 在查询语句中指定了 system = 'hcm'，这里就要把 system 过滤掉，
//		// 且如果filter中传递了这个参数，sql生成的前置校验就会报错。
//		// RuleFields: table.AccountColumns.WithoutColumn("system"),
//		RuleFields: table.AccountColumns.ColumnTypes(),
//	}
//	if err := lao.Filter.Validate(exprOpt); err != nil {
//		return err
//	}
//
//	if lao.Page == nil {
//		return errf.New(errf.InvalidParameter, "page is null")
//	}
//
//	if err := lao.Page.Validate(po); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// ListAccountDetails defines the response details of requested ListAccountDetails
//type ListAccountDetails struct {
//	Count   uint32           `json:"count"`
//	Details []*table.Account `json:"details"`
//}
