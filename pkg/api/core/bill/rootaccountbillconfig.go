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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
)

// RootAccountBillConfig defines account bill config info.
type RootAccountBillConfig[T RootAccountBillConfigExtension] struct {
	BaseRootAccountBillConfig `json:",inline"`
	Extension                 *T `json:"extension"`
}

// BaseRootAccountBillConfig define account bill config.
type BaseRootAccountBillConfig struct {
	ID                string          `json:"id"`
	Vendor            enumor.Vendor   `json:"vendor"`
	RootAccountID     string          `json:"root_account_id"`
	CloudDatabaseName string          `json:"cloud_database_name"`
	CloudTableName    string          `json:"cloud_table_name"`
	Status            int64           `json:"status"`
	ErrMsg            []string        `json:"err_msg"`
	Extension         types.JsonField `json:"extension"`
	*core.Revision    `json:",inline"`
}

// RootAccountBillConfigExtension defines account bill config extensional info.
type RootAccountBillConfigExtension interface {
	AwsBillConfigExtension | GcpBillConfigExtension
}

// AwsBillConfigExtension define aws bill config extension.
type AwsBillConfigExtension struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	CurName   string `json:"cur_name"`
	CurPrefix string `json:"cur_prefix"`
	YmlURL    string `json:"yml_url"`
	SavePath  string `json:"save_path"`
	StackID   string `json:"stack_id"`
	StackName string `json:"stack_name"`
}

// GcpBillConfigExtension define gcp bill config extension.
type GcpBillConfigExtension struct {
}
