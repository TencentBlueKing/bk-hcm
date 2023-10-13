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

var _ action.Action = new(StopAction)
var _ action.ParameterAction = new(StopAction)

// StopAction define stop cvm action.
type StopAction struct {
	CvmOperationAction
}

// NewStopAction new stop cvm action.
func NewStopAction() StopAction {
	act := StopAction{
		CvmOperationAction{
			ActionName: enumor.ActionStopCvm,
			TCloudFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.TCloudBatchStopReq{
					AccountID:   opt.AccountID,
					Region:      opt.Region,
					IDs:         opt.IDs,
					StopType:    typecvm.SoftFirst,
					StoppedMode: typecvm.KeepCharging,
				}
				return cli.TCloud.Cvm.BatchStopCvm(kt, req)
			},
			AwsFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.AwsBatchStopReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
					Force:     true,
					Hibernate: false,
				}
				return cli.Aws.Cvm.BatchStopCvm(kt, req)
			},
			HuaWeiFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.HuaWeiBatchStopReq{
					AccountID: opt.AccountID,
					Region:    opt.Region,
					IDs:       opt.IDs,
					Force:     true,
				}
				return cli.HuaWei.Cvm.BatchStopCvm(kt, req)
			},
			GcpFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				return cli.Gcp.Cvm.StopCvm(kt, opt.IDs[0])
			},
			AzureFunc: func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error {
				req := &hcprotocvm.AzureStopReq{
					SkipShutdown: false,
				}
				return cli.Azure.Cvm.StopCvm(kt, opt.IDs[0], req)
			},
		},
	}

	return act
}
