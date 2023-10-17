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

package actioncvm

import (
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/async/action"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

var _ action.Action = new(StartAction)
var _ action.ParameterAction = new(StartAction)

// StartAction define start cvm action.
type StartAction struct {
	CvmOperationAction
}

// NewStartAction new start cvm action.
func NewStartAction() StartAction {
	act := StartAction{
		CvmOperationAction{
			ActionName: enumor.ActionStartCvm,
			TCloudFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.TCloudBatchStartReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
				}
				return cli.TCloud.Cvm.BatchStartCvm(kt, req)
			},
			AwsFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.AwsBatchStartReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
				}
				return cli.Aws.Cvm.BatchStartCvm(kt, req)
			},
			HuaWeiFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.HuaWeiBatchStartReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
				}
				return cli.HuaWei.Cvm.BatchStartCvm(kt, req)
			},
			GcpFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				return cli.Gcp.Cvm.StartCvm(kt, opt.IDs[0])
			},
			AzureFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				return cli.Azure.Cvm.StartCvm(kt, opt.IDs[0])
			},
		},
	}

	return act
}
