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

package tools

import "fmt"

// BaseRelJoinSqlBuild 因为关联表的join sql部分是一样的，所以该函数提供一种基础的生成原则。
// relTableAlias: 关联表别名
// resTableAlias: 资源表别名
// asIDFieldName: 关联表中映射成id字段的字段名
// saveIDFieldName: 关联表中字段名不变的字段名
// e.g: relTableAlias: rel. resTableAlias: sg. asIDFieldName: security_group_id. saveIDFieldName: cvm_id.
// 生成的SQL: rel.security_group_id as id, rel.cvm_id as cvm_id, sg.creator as creator, sg.created_at as created_at,
// rel.creator as rel_creator, rel.created_at as rel_created_at
func BaseRelJoinSqlBuild(relTableAlias, resTableAlias, asIDFieldName, saveIDFieldName string) string {
	return fmt.Sprintf(`%s.%s as id, IFNULL(%s.%s,"") as %s, %s.creator as creator, %s.created_at as created_at, 
IFNULL(%s.creator,"") as rel_creator, %s.created_at as rel_created_at`, resTableAlias, asIDFieldName, relTableAlias,
		saveIDFieldName, saveIDFieldName, resTableAlias, resTableAlias, relTableAlias, relTableAlias)
}

// BaseRelJoinSqlBuildWithBizID 专用于把关联表的bk_biz_id映射为rel_usage_biz_id，并显式使用资源表的bk_biz_id字段
// relTableAlias: 关联表别名
// resTableAlias: 资源表别名
// asIDFieldName: 关联表中映射成id字段的字段名
// e.g: relTableAlias: rel. resTableAlias: account. asIDFieldName: id.
// 生成的SQL: account.id as id, account.bk_biz_id as bk_biz_id, rel.bk_biz_id as rel_usage_biz_id,
// account.creator as creator, account.created_at as created_at, rel.creator as rel_creator,
// rel.created_at as rel_created_at
func BaseRelJoinSqlBuildWithBizID(relTableAlias, resTableAlias, asIDFieldName string) string {
	return fmt.Sprintf(`%s.%s as id, %s.bk_biz_id as bk_biz_id, IFNULL(%s.bk_biz_id,"") as rel_usage_biz_id, 
%s.creator as creator, %s.created_at as created_at, IFNULL(%s.creator,"") as rel_creator, 
%s.created_at as rel_created_at`, resTableAlias, asIDFieldName, resTableAlias, relTableAlias, resTableAlias,
		resTableAlias, relTableAlias, relTableAlias)
}
