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

package cloud

//import (
//	"time"
//
//	"hcm/pkg/dal/dao/types"
//	"hcm/pkg/dal/table"
//	"hcm/pkg/runtime/filter"
//)
//
//type AccountBizRelModel struct {
//	// 账号自增 ID
//	ID uint64 `db:"id"`
//	// 蓝鲸业务 ID
//	BkBizID uint64 `db:"bk_biz_id"`
//	// 云账号主键 ID
//	AccountID uint64 `db:"account_id"`
//	// 创建者
//	Creator string `db:"creator"`
//	// 更新者
//	Reviser string `db:"reviser"`
//	// 创建时间
//	CreatedAt *time.Time `db:"created_at"`
//	// 更新时间
//	UpdatedAt *time.Time `db:"updated_at"`
//	// model manager
//	ModelManager *table.ModelManager
//}
//
//func (m *AccountBizRelModel) TableName() string {
//	return "account_biz_rel"
//}
//
//// GenerateInsertSQL ...
//func (m *AccountBizRelModel) GenerateInsertSQL() string {
//	return m.ModelManager.GenerateInsertSQL(m)
//}
//
//// GenerateInsertSQL ...
//func (m *AccountBizRelModel) GenerateUpdateSQL(expr *filter.Expression) (string, error) {
//	return m.ModelManager.GenerateUpdateSQL(m, expr)
//}
//
//// GenerateUpdateFieldKV ...
//func (m *AccountBizRelModel) GenerateUpdateFieldKV() map[string]interface{} {
//	return m.ModelManager.GenerateUpdateFieldKV(m)
//}
//
//// GenerateListSQL ...
//func (m *AccountBizRelModel) GenerateListSQL(opt *types.ListOption) (string, error) {
//	return m.ModelManager.GenerateListSQL(m, opt)
//}
//
//func (m *AccountBizRelModel) GenerateDeleteSQL(expr *filter.Expression) (string, error) {
//	return m.ModelManager.GenerateDeleteSQL(m, expr)
//}
