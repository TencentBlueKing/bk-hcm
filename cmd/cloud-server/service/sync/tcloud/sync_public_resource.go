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

package tcloud

import (
	"hcm/pkg/client"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
)

// SyncPublicResourceOption ...
type SyncPublicResourceOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate SyncPublicResourceOption
func (opt *SyncPublicResourceOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SyncPublicResource ...
func SyncPublicResource(kt *kit.Kit, cliSet *client.ClientSet, opt *SyncPublicResourceOption) error {

	if err := opt.Validate(); err != nil {
		return err
	}

	if err := SyncRegion(kt, cliSet.HCService(), opt.AccountID); err != nil {
		return err
	}

	regions, err := ListRegion(kt, cliSet.DataService())
	if err != nil {
		return err
	}

	if err = SyncZone(kt, cliSet.HCService(), opt.AccountID, regions); err != nil {
		return err
	}

	if err = SyncTCloudImage(kt, cliSet.HCService(), opt.AccountID, regions); err != nil {
		return err
	}

	return nil
}
