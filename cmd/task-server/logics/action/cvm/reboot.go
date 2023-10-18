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
	typecvm "hcm/pkg/adaptor/types/cvm"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/async/action"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
)

var _ action.Action = new(RebootAction)
var _ action.ParameterAction = new(RebootAction)

// RebootAction define reboot cvm action.
type RebootAction struct {
	CvmOperationAction
}

// NewRebootAction new reboot cvm action.
func NewRebootAction() RebootAction {
	act := RebootAction{
		CvmOperationAction{
			ActionName: enumor.ActionRebootCvm,
			TCloudFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.TCloudBatchRebootReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
					StopType:  typecvm.SoftFirst,
				}
				return cli.TCloud.Cvm.BatchRebootCvm(kt, req)
			},
			AwsFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.AwsBatchRebootReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
				}
				return cli.Aws.Cvm.BatchRebootCvm(kt, req)
			},
			HuaWeiFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.HuaWeiBatchRebootReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
					Force:     true,
				}
				return cli.HuaWei.Cvm.BatchRebootCvm(kt, req)
			},
			GcpFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				return cli.Gcp.Cvm.RebootCvm(kt, opt.IDs[0])
			},
			AzureFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				return cli.Azure.Cvm.RebootCvm(kt, opt.IDs[0])
			},
		},
	}

	return act
}
