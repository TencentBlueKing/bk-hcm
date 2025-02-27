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

// Package aws adaptor
package aws

import (
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
)

// NewAws new aws.
func NewAws(s *types.BaseSecret, cloudAccountID string, site enumor.AccountSiteType) (*Aws, error) {
	if err := validateSecret(s); err != nil {
		return nil, err
	}

	return &Aws{clientSet: newClientSet(s), cloudAccountID: cloudAccountID, site: site}, nil
}

// Aws is aws operator.
type Aws struct {
	clientSet      *clientSet
	cloudAccountID string
	site           enumor.AccountSiteType
}

func validateSecret(s *types.BaseSecret) error {
	if s == nil {
		return errf.New(errf.InvalidParameter, "secret is required")
	}

	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}

// CloudAccountID return cloud account id.
func (a *Aws) CloudAccountID() string {
	return a.cloudAccountID
}

// IsChinaSite is china site.
func (a *Aws) IsChinaSite() bool {
	return a.site == enumor.ChinaSite
}

// DefaultRegion return default region.
func (a *Aws) DefaultRegion() string {
	if a.IsChinaSite() {
		return "cn-north-1"
	}
	return "ap-northeast-1"
}
