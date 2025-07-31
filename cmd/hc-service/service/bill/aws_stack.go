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
	"time"

	"hcm/pkg/adaptor/aws"
	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/api/core/cloud"
	protobill "hcm/pkg/api/data-service/cloud/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// AwsCreateStack aws create stack.
func (b bill) AwsCreateStack(kt *kit.Kit, accountID string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) error {

	if billInfo.Status != constant.StatusCheckYml {
		return nil
	}

	cli, err := b.ad.Aws(kt, accountID)
	if err != nil {
		logs.Errorf("aws create stack client failed, accountID: %s, err: %+v", accountID, err)
		return err
	}

	if billInfo.Extension == nil || billInfo.Extension.CurName == "" || billInfo.Extension.YmlURL == "" {
		return errf.Newf(errf.Aborted, "account_id: %s has not ready yet", accountID)
	}

	opt := &typesBill.AwsCreateStackReq{
		Region:       billInfo.Extension.Region,
		StackName:    billInfo.Extension.CurName,
		TemplateURL:  billInfo.Extension.YmlURL,
		Capabilities: []*string{cvt.ValToPtr(aws.CapabilitiesIam)},
	}
	stackID, err := cli.CreateStack(kt, opt)
	if err != nil {
		billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
			Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
				{
					ID:     billInfo.ID,
					ErrMsg: []string{err.Error()},
				},
			},
		}
		updateErr := b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
		if updateErr != nil {
			logs.Errorf("aws create stack failed and update db failed, accountID: %s, billID: %s, opt: %+v, "+
				"updateErr: %+v, err: %+v, rid: %s", accountID, billInfo.ID, opt, updateErr, err, kt.Rid)
		}
		return err
	}

	billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
		Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
			{
				ID:     billInfo.ID,
				ErrMsg: []string{},
				Status: constant.StatusCreateCloudFormation,
				Extension: &cloud.AwsBillConfigExtension{
					Bucket:    billInfo.Extension.Bucket,
					Region:    billInfo.Extension.Region,
					CurName:   billInfo.Extension.CurName,
					CurPrefix: billInfo.Extension.CurPrefix,
					YmlURL:    billInfo.Extension.YmlURL,
					SavePath:  billInfo.Extension.SavePath,
					StackID:   stackID,
					StackName: opt.StackName,
				},
			},
		},
	}
	err = b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
	if err != nil {
		return err
	}

	billInfo.Status = billReq.Bills[0].Status
	billInfo.Extension.StackID = billReq.Bills[0].Extension.StackID
	billInfo.Extension.StackName = billReq.Bills[0].Extension.StackName
	return nil
}

// CheckStackStatus check stack status
func (b bill) CheckStackStatus(kt *kit.Kit, accountID string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) error {

	if billInfo.Status != constant.StatusCreateCloudFormation {
		return nil
	}

	cli, err := b.ad.Aws(kt, accountID)
	if err != nil {
		logs.Errorf("aws check stack status client failed, accountID: %s, err: %+v", accountID, err)
		return err
	}

	if billInfo.Extension == nil || billInfo.Extension.CurName == "" || billInfo.Extension.YmlURL == "" {
		return errf.Newf(errf.Aborted, "account_id: %s has not ready yet", accountID)
	}

	opt := &typesBill.AwsDeleteStackReq{
		Region:  billInfo.Extension.Region,
		StackID: billInfo.Extension.StackID,
	}

	var (
		stackInfo *cloudformation.Stack
		nowTime   = time.Now()
		duration  = time.Duration(10) * time.Second // Pause for 10 seconds
	)
	for {
		stackList, err := cli.DescribeStack(kt, opt)
		if err != nil || len(stackList) == 0 {
			billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
				Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
					{
						ID:     billInfo.ID,
						ErrMsg: []string{err.Error()},
					},
				},
			}
			updateErr := b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
			if updateErr != nil {
				logs.Errorf("aws check stack status failed and update db failed, accountID: %s, opt: %+v, "+
					"updateErr: %+v, err: %+v, rid: %s", accountID, opt, updateErr, err, kt.Rid)
			}
			return errf.Newf(errf.Aborted, "account_id: %s check stack status failed, err: %+v", accountID, err)
		}

		stackInfo = stackList[0]
		stackCreateTime := cvt.PtrToVal(stackInfo.CreationTime)
		if cvt.PtrToVal(stackInfo.StackStatus) == "CREATE_COMPLETE" ||
			nowTime.Sub(stackCreateTime).Seconds() > aws.StackTimeOut {
			break
		}

		time.Sleep(duration)
	}

	var stackStatus = cvt.PtrToVal(stackInfo.StackStatus)
	var updateReq = protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{ID: billInfo.ID}
	if stackStatus != "CREATE_COMPLETE" {
		updateReq.ErrMsg = append(updateReq.ErrMsg, "stack status is "+stackStatus)
	} else {
		updateReq.ErrMsg = []string{}
		updateReq.Status = constant.StatusSuccess
	}

	billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
		Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{updateReq},
	}
	err = b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
	if err != nil {
		return err
	}

	return nil
}
