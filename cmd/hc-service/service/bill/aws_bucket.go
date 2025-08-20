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
	"fmt"

	"hcm/pkg/adaptor/aws"
	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	protobill "hcm/pkg/api/data-service/cloud/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// AwsBucketCreate create aws bucket.
func (b bill) AwsBucketCreate(kt *kit.Kit, accountID string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension],
	account *protocloud.AccountGetResult[cloud.AwsAccountExtension]) error {

	if billInfo.Status != constant.StatusDefault {
		return nil
	}

	cli, err := b.ad.Aws(kt, accountID)
	if err != nil {
		logs.Errorf("aws bucket create client failed, accountID: %s, err: %+v", accountID, err)
		return err
	}

	// 存储桶名称，需要全球唯一
	bucketName := fmt.Sprintf(aws.BucketNameDefault, account.Extension.CloudAccountID,
		sanitizeString(account.Extension.CloudIamUsername))

	opt := &typesBill.AwsBillBucketCreateReq{
		Bucket: bucketName,
		Region: billInfo.Extension.Region,
	}
	_, err = cli.CreateBucket(kt, opt)
	if err != nil {
		billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
			Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
				{
					ID:     billInfo.ID,
					ErrMsg: []string{err.Error()},
					Extension: &cloud.AwsBillConfigExtension{
						Region: billInfo.Extension.Region,
					},
				},
			},
		}
		updateErr := b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
		if updateErr != nil {
			logs.Errorf("aws bucket create failed and update db failed, accountID: %s, billID: %s, "+
				"updateErr: %+v, err: %+v, rid: %s", billInfo.ID, accountID, updateErr, err, kt.Rid)
		}
		return err
	}

	// 更新
	billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
		Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
			{
				ID: billInfo.ID,
				// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版100:正常)
				Status: constant.StatusCreateBucket,
				ErrMsg: []string{},
				Extension: &cloud.AwsBillConfigExtension{
					Bucket: opt.Bucket,
					Region: billInfo.Extension.Region,
				},
			},
		},
	}
	err = b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
	if err != nil {
		return err
	}

	billInfo.Extension.Bucket = opt.Bucket
	billInfo.Status = billReq.Bills[0].Status
	return nil
}

// AwsPutBucketPolicy aws put bucket policy.
func (b bill) AwsPutBucketPolicy(kt *kit.Kit, accountID string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) error {

	if billInfo.Status != constant.StatusCreateBucket {
		return nil
	}

	cli, err := b.ad.Aws(kt, accountID)
	if err != nil {
		logs.Errorf("aws put bucket policy client failed, accountID: %s, err: %+v", accountID, err)
		return err
	}

	opt := &typesBill.AwsBillBucketPolicyReq{
		AccountID: accountID,
		Bucket:    billInfo.Extension.Bucket,
		Region:    billInfo.Extension.Region,
	}
	err = cli.PutBucketPolicy(kt, opt)
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
			logs.Errorf("aws put bucket policy failed and update db failed, accountID: %s, billID: %s, "+
				"updateErr: %+v, err: %+v, rid: %s", accountID, billInfo.ID, updateErr, err, kt.Rid)
		}
		return err
	}

	// 更新
	billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
		Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
			{
				ID: billInfo.ID,
				// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版100:正常)
				Status: constant.StatusSetBucketPolicy,
				ErrMsg: []string{},
			},
		},
	}
	err = b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
	if err != nil {
		return err
	}

	billInfo.Status = billReq.Bills[0].Status
	return nil
}
