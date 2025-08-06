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

// Package cos cos.
package cos

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
)

// CosExtension extension.
type CosExtension interface {
}

// BaseCos base cos.
type BaseCos struct {
	ID        string        `json:"id"`
	CloudID   string        `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Name      string        `db:"name" validate:"lte=255" json:"name"`
	Vendor    enumor.Vendor `db:"vendor" validate:"lte=16"  json:"vendor"`
	AccountID string        `db:"account_id" validate:"lte=64" json:"account_id"`
	BkBizID   int64         `db:"bk_biz_id" json:"bk_biz_id"`
	Region    string        `db:"region" validate:"lte=20" json:"region"`

	ACL                       string          `db:"acl" json:"acl"`
	GrantFullControl          string          `db:"grant_full_control" json:"grant_full_control"`
	GrantRead                 string          `db:"grant_read" json:"grant_read"`
	GrantWrite                string          `db:"grant_write" json:"grant_write"`
	GrantReadACP              string          `db:"grant_read_acp" json:"grant_read_acp"`
	GrantWriteACP             string          `db:"grant_write_acp" json:"grant_write_acp"`
	CreateBucketConfiguration types.JsonField `db:"create_bucket_configuration" json:"create_bucket_configuration"`

	Domain           string      `db:"domain" json:"domain"`
	Status           string      `db:"status" json:"status"`
	CloudCreatedTime string      `db:"cloud_created_time" json:"cloud_created_time"`
	CloudStatusTime  string      `db:"cloud_status_time" json:"cloud_status_time"`
	CloudExpiredTime string      `db:"cloud_expired_time" json:"cloud_expired_time"`
	SyncTime         string      `db:"sync_time" json:"sync_time"`
	Tags             core.TagMap `db:"tags" json:"tags"`

	*core.Revision `json:",inline"`
}

// Cos cos.
type Cos[T CosExtension] struct {
	*BaseCos  `json:",inline"`
	Extension *T `json:"extension"`
}
