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

package gcp

import (
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncImageOption ...
type SyncImageOption struct {
}

// Validate ...
func (opt SyncImageOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Image ...
func (cli *client) Image(kt *kit.Kit, params *SyncBaseParams, opt *SyncImageOption) (*SyncResult, error) {
	// TODO implement me
	panic("implement me")
}

func (cli *client) RemoveImageDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
	// TODO implement me
	panic("implement me")
}

func (cli *client) listImageFromDBForCvm(kt *kit.Kit, params *ListBySelfLinkOption) (
	[]*dataproto.ImageResult, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &dataproto.ImageListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "extension.self_link",
					Op:    filter.JSONIn.Factory(),
					Value: params.SelfLink,
				},
			},
		},
		Page: core.DefaultBasePage,
	}
	result, err := cli.dbCli.Global.ListImage(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list image from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.TCloud, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}
