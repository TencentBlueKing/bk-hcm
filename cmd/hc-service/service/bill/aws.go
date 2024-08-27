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

// Package bill defines bill service.
package bill

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"hcm/pkg/adaptor/aws"
	typesBill "hcm/pkg/adaptor/types/bill"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	protobill "hcm/pkg/api/data-service/cloud/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// AwsGetBillList get aws bill list.
func (b bill) AwsGetBillList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbill.AwsBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := b.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("aws bill get cloud client failed, req: %+v, err: %+v", req, err)
		return nil, err
	}

	// 查询aws账单基础表
	billInfo, err := b.GetBillInfo(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("aws bill config get base info db failed, accountID: %s, err: %+v", req.AccountID, err)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "account_id: %s is not found", req.AccountID)
	}

	if billInfo.Status != constant.StatusSuccess {
		return nil, errf.Newf(errf.Aborted, "account_id: %s has not ready yet", req.AccountID)
	}

	opt := &typesBill.AwsBillListOption{
		AccountID: req.AccountID,
		BeginDate: req.BeginDate,
		EndDate:   req.EndDate,
	}
	if req.Page != nil {
		opt.Page = &typesBill.AwsBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		}
	}
	total, list, err := cli.GetBillList(cts.Kit, opt, billInfo)
	if err != nil {
		logs.Errorf("request adaptor list aws bill failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return &hcbill.AwsBillListResult{
		Count:   total,
		Details: list,
	}, nil
}

// AwsBillPipeline aws bill pipeline
func (b bill) AwsBillPipeline(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbill.BillPipelineReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	billInfo, hasDone, err := b.CheckBillInfo(cts.Kit, req)
	if err != nil {
		logs.Errorf("aws bill pipeline get db bill info failed, req: %+v, err: %+v", req, err)
		return nil, err
	}
	if hasDone {
		logs.V(3).Infof("aws bill pipeline has process done, req: %+v, billExt: %+v",
			req, billInfo.Extension)
		return nil, nil
	}

	// 查询账号扩展信息
	account, err := b.cs.DataService().Aws.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), req.AccountID)
	if err != nil {
		logs.Errorf("aws bill pipeline get db account info failed, req: %+v, err: %+v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}
	if account.Extension == nil {
		logs.Errorf("aws bill pipeline get db account extension is empty, req: %+v, rid: %s", req, cts.Kit.Rid)
		return nil, errf.Newf(errf.Aborted, "account_id: %s extension is empty", req.AccountID)
	}

	// 创建S3存储桶
	err = b.AwsBucketCreate(cts.Kit, req.AccountID, billInfo, account)
	if err != nil {
		logs.Errorf("aws bill pipeline create bucket failed, req: %+v, billInfo: %+v, billExt: %+v, "+
			"err: %+v, rid: %s", req, billInfo, billInfo.Extension, err, cts.Kit.Rid)
		return nil, err
	}

	// 对S3存储桶设置需要的权限
	err = b.AwsPutBucketPolicy(cts.Kit, req.AccountID, billInfo)
	if err != nil {
		logs.Errorf("aws bill pipeline put bucket policy failed, req: %+v, billInfo: %+v, billExt: %+v, "+
			"err: %+v, rid: %s", req, billInfo, billInfo.Extension, err, cts.Kit.Rid)
		return nil, err
	}

	// 创建成本和使用率报告
	err = b.AwsPutReportDefinition(cts.Kit, req.AccountID, billInfo)
	if err != nil {
		logs.Errorf("aws bill pipeline put report definition failed, req: %+v, billInfo: %+v, billExt: %+v, "+
			"err: %+v, rid: %s", req, billInfo, billInfo.Extension, err, cts.Kit.Rid)
		return nil, err
	}

	// 检查yml文件
	err = b.CheckCrawlerCfnYml(cts.Kit, req.AccountID, billInfo)
	if err != nil {
		logs.Errorf("aws bill pipeline check yml file failed, req: %+v, billExt: %+v, err: %+v, rid: %s",
			req, billInfo.Extension, err, cts.Kit.Rid)
		return nil, err
	}

	// 创建CloudFormation模版，用于同步数据到S3存储桶
	err = b.AwsCreateStack(cts.Kit, req.AccountID, billInfo)
	if err != nil {
		logs.Errorf("aws bill pipeline create stack failed, req: %+v, billInfo: %+v, billExt: %+v, err: %+v, "+
			"rid: %s", req, billInfo, billInfo.Extension, err, cts.Kit.Rid)
		return nil, err
	}

	// 检查Stack状态
	err = b.CheckStackStatus(cts.Kit, req.AccountID, billInfo)
	if err != nil {
		logs.Errorf("aws bill pipeline check stack status failed, req: %+v, billInfo: %+v, billExt: %+v, err: %+v, "+
			"rid: %s", req, billInfo, billInfo.Extension, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetBillInfo get bill info.
func (b bill) GetBillInfo(kt *kit.Kit, accountID string) (
	*cloud.AccountBillConfig[cloud.AwsBillConfigExtension], error) {

	// 查询aws账单基础表
	billList, err := b.cs.DataService().Aws.Bill.List(kt.Ctx, kt.Header(), &core.ListReq{
		Filter: tools.EqualExpression("account_id", accountID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	})
	if err != nil {
		logs.Errorf("aws get base info from db failed, accountID: %s, err: %+v", accountID, err)
		return nil, err
	}
	if len(billList.Details) == 0 {
		return nil, nil
	}

	return &billList.Details[0], nil
}

// CheckBillInfo check bill info.
func (b bill) CheckBillInfo(kt *kit.Kit, req *hcbill.BillPipelineReq) (
	*cloud.AccountBillConfig[cloud.AwsBillConfigExtension], bool, error) {

	billInfo, err := b.GetBillInfo(kt, req.AccountID)
	if err != nil {
		logs.Errorf("aws bill pipeline get base info db failed, accountID: %s, err: %+v", req.AccountID, err)
		return nil, false, err
	}

	if billInfo == nil {
		billInfo = &cloud.AccountBillConfig[cloud.AwsBillConfigExtension]{
			BaseAccountBillConfig: cloud.BaseAccountBillConfig{
				Vendor:    enumor.Aws,
				AccountID: req.AccountID,
				// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版100:正常)
				Status: constant.StatusDefault,
			},
			Extension: &cloud.AwsBillConfigExtension{Region: aws.BucketRegion},
		}

		billReq := &protobill.AccountBillConfigBatchCreateReq[cloud.AwsBillConfigExtension]{
			Bills: []protobill.AccountBillConfigReq[cloud.AwsBillConfigExtension]{
				{
					Vendor:    billInfo.Vendor,
					AccountID: billInfo.AccountID,
					Status:    billInfo.Status,
				},
			},
		}
		billResp, err := b.cs.DataService().Aws.Bill.BatchCreate(kt.Ctx, kt.Header(), billReq)
		if err != nil {
			logs.Errorf("aws bill pipeline bucket db create of bill failed, req: %+v, err: %+v", req, err)
			return nil, false, err
		}
		billInfo.ID = billResp.IDs[0]
	} else {
		if billInfo.Status == constant.StatusSuccess {
			return billInfo, true, nil
		}

		resTime, err := time.Parse(constant.TimeStdFormat, billInfo.CreatedAt)
		if err != nil {
			return nil, false, err
		}

		nowTime := time.Now()
		hourDur := nowTime.Sub(resTime).Hours()
		if billInfo.Status != constant.StatusSuccess && hourDur > aws.BucketTimeOut {
			logs.Errorf("aws bill pipeline is timeout, accountID: %s, CreatedAt: %s, now: %v, hourDur: %f, "+
				"rid: %s", req.AccountID, billInfo.CreatedAt, nowTime.Local(), hourDur, kt.Rid)
			return billInfo, false, errf.New(errf.PartialFailed, "aws bill config pipeline has timeout")
		}
		if billInfo.Extension != nil && billInfo.Extension.Region == "" {
			billInfo.Extension.Region = aws.BucketRegion
		}
	}

	return billInfo, false, nil
}

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

// AwsPutReportDefinition aws put report definition.
func (b bill) AwsPutReportDefinition(kt *kit.Kit, accountID string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) error {

	if billInfo.Status != constant.StatusSetBucketPolicy {
		return nil
	}

	cli, err := b.ad.Aws(kt, accountID)
	if err != nil {
		logs.Errorf("aws put report definition client failed, accountID: %s, err: %+v", accountID, err)
		return err
	}

	opt := &typesBill.AwsBillPutReportDefinitionReq{
		Bucket:           billInfo.Extension.Bucket,
		Region:           billInfo.Extension.Region,
		CurName:          aws.CurName,
		CurPrefix:        aws.CurPrefix,
		Format:           aws.CurFormat,
		TimeUnit:         aws.CurTimeUnit,
		Compression:      aws.CurCompression,
		SchemaElements:   []*string{converter.ValToPtr(aws.ResourceSchemaElement)},
		Artifacts:        []*string{converter.ValToPtr(aws.AthenaArtifact)},
		ReportVersioning: aws.ReportVersioning,
	}
	err = cli.PutReportDefinition(kt, opt)
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
			logs.Errorf("aws put report definition failed and update db failed, accountID: %s, billID: %s, "+
				"updateErr: %+v, err: %+v, rid: %s", accountID, billInfo.ID, updateErr, err, kt.Rid)
		}
		return err
	}

	// 更新
	savePath := strings.ReplaceAll(aws.AthenaSavePath, "{BucketName}", opt.Bucket)
	savePath = strings.ReplaceAll(savePath, "{CurPrefix}", opt.CurPrefix)
	savePath = strings.ReplaceAll(savePath, "{CurName}", opt.CurName)
	ymlURL := strings.ReplaceAll(aws.YmlURL, "{BucketName}", opt.Bucket)
	ymlURL = strings.ReplaceAll(ymlURL, "{CurPrefix}", opt.CurPrefix)
	ymlURL = strings.ReplaceAll(ymlURL, "{CurName}", opt.CurName)
	billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
		Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
			{
				ID: billInfo.ID,
				// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版100:正常)
				Status:            constant.StatusCreateCur,
				ErrMsg:            []string{},
				CloudDatabaseName: fmt.Sprintf("%s_%s", aws.DatabaseNamePrefix, opt.CurName),
				CloudTableName:    opt.CurName,
				Extension: &cloud.AwsBillConfigExtension{
					Bucket:    billInfo.Extension.Bucket,
					Region:    billInfo.Extension.Region,
					CurName:   opt.CurName,
					CurPrefix: opt.CurPrefix,
					SavePath:  savePath,
					YmlURL:    ymlURL,
				},
			},
		},
	}
	err = b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
	if err != nil {
		return err
	}

	billInfo.Extension = billReq.Bills[0].Extension
	billInfo.Status = billReq.Bills[0].Status
	billInfo.CloudDatabaseName = billReq.Bills[0].CloudDatabaseName
	billInfo.CloudTableName = billReq.Bills[0].CloudTableName
	return nil
}

// CheckCrawlerCfnYml check crawler-cfn.yml.
func (b bill) CheckCrawlerCfnYml(kt *kit.Kit, accountID string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) error {

	if billInfo.Status != constant.StatusCreateCur {
		return nil
	}

	cli, err := b.ad.Aws(kt, accountID)
	if err != nil {
		logs.Errorf("aws check crawler-cfn.yml client failed, accountID: %s, err: %+v", accountID, err)
		return err
	}

	if billInfo.CloudDatabaseName == "" || billInfo.CloudTableName == "" || billInfo.Extension == nil ||
		billInfo.Extension.Bucket == "" || billInfo.Extension.Region == "" || billInfo.Extension.CurPrefix == "" ||
		billInfo.Extension.CurName == "" || billInfo.Extension.SavePath == "" {
		return errf.Newf(errf.Aborted, "account_id: %s has not ready yet", accountID)
	}

	fileKey := fmt.Sprintf(aws.CrawlerCfnFileKey, billInfo.Extension.CurPrefix, billInfo.Extension.CurName)
	opt := &typesBill.AwsBillGetObjectReq{
		Bucket: billInfo.Extension.Bucket,
		Region: billInfo.Extension.Region,
		Key:    fileKey,
	}
	_, err = cli.GetObject(kt, opt)
	if err != nil {
		ymlFile := fmt.Sprintf("s3://%s%s", opt.Bucket, fileKey)
		billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
			Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
				{
					ID:     billInfo.ID,
					ErrMsg: []string{ymlFile, err.Error()},
				},
			},
		}
		updateErr := b.cs.DataService().Aws.Bill.BatchUpdate(kt.Ctx, kt.Header(), billReq)
		if updateErr != nil {
			logs.Errorf("aws check crawler-cfn.yml failed and update db failed, accountID: %s, billID: %s, "+
				"updateErr: %+v, err: %+v, rid: %s", accountID, billInfo.ID, updateErr, err, kt.Rid)
		}
		if strings.Contains(err.Error(), "NoSuchKey") {
			return errf.Newf(errf.RecordNotFound, "%s does not exist", ymlFile)
		}
		return err
	}

	billReq := &protobill.AccountBillConfigBatchUpdateReq[cloud.AwsBillConfigExtension]{
		Bills: []protobill.AccountBillConfigUpdateReq[cloud.AwsBillConfigExtension]{
			{
				ID:     billInfo.ID,
				ErrMsg: []string{},
				Status: constant.StatusCheckYml,
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
		Capabilities: []*string{converter.ValToPtr(aws.CapabilitiesIam)},
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
		stackCreateTime := converter.PtrToVal(stackInfo.CreationTime)
		if converter.PtrToVal(stackInfo.StackStatus) == "CREATE_COMPLETE" ||
			nowTime.Sub(stackCreateTime).Seconds() > aws.StackTimeOut {
			break
		}

		time.Sleep(duration)
	}

	var stackStatus = converter.PtrToVal(stackInfo.StackStatus)
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

// AwsBillConfigDelete aws bill config delete
func (b bill) AwsBillConfigDelete(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("id").String()

	billInfo, err := b.GetBillInfo(cts.Kit, accountID)
	if err != nil {
		logs.Errorf("aws delete stack get db base info failed, accountID: %s, err: %+v", accountID, err)
		return nil, err
	}

	if billInfo == nil || billInfo.Extension == nil || billInfo.Extension.Region == "" {
		return nil, errf.Newf(errf.RecordNotFound, "account_id: %s is not found", accountID)
	}

	cli, err := b.ad.Aws(cts.Kit, accountID)
	if err != nil {
		logs.Errorf("aws delete stack client failed, accountID: %s, err: %+v", accountID, err)
		return nil, err
	}

	if billInfo.Extension.StackID != "" {
		opt := &typesBill.AwsDeleteStackReq{
			Region:  billInfo.Extension.Region,
			StackID: billInfo.Extension.StackID,
		}
		err = cli.DeleteStack(cts.Kit, opt)
		if err != nil {
			logs.Errorf("aws delete stack cloud failed, accountID: %s, opt: %+v, err: %+v", accountID, opt, err)
			return nil, err
		}
	}

	if billInfo.Extension.CurName != "" {
		reportOpt := &typesBill.AwsBillDeleteReportDefinitionReq{
			ReportName: billInfo.Extension.CurName,
			Region:     billInfo.Extension.Region,
		}
		err = cli.DeleteReportDefinition(cts.Kit, reportOpt)
		if err != nil {
			logs.Errorf("aws delete cur report failed, accountID: %s, opt: %+v, err: %+v",
				accountID, reportOpt, err)
			return nil, err
		}
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", billInfo.ID),
	}
	err = b.cs.DataService().Global.Bill.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		logs.Errorf("aws delete bill config db failed, accountID: %s, billID: %s, err: %+v",
			accountID, billInfo.ID, err)
		return nil, err
	}

	return nil, nil
}

// sanitizeString 匹配任何非[中划线、小写字母、英文点号和数字]的字符
func sanitizeString(str string) string {
	reg := regexp.MustCompile(`[^a-z0-9.\-]`)
	return reg.ReplaceAllString(strings.ToLower(str), "")
}

// AwsGetRootAccountBillList get aws bill record list
func (b bill) AwsGetRootAccountBillList(cts *rest.Contexts) (any, error) {

	req := new(hcbill.AwsRootBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Page == nil {
		req.Page = &hcbill.AwsBillListPage{Offset: 0, Limit: adcore.AwsQueryLimit}
	}

	// 查询aws账单基础表
	billInfo, err := getRootAccountBillConfigInfo[billcore.AwsBillConfigExtension](
		cts.Kit, req.RootAccountID, b.cs.DataService())
	if err != nil {
		logs.Errorf("aws root account bill config get base info db failed, main_account_cloud_id: %s, err: %+v,rid: %s",
			req.MainAccountCloudID, err, cts.Kit.Rid)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "bill config for root_account_id: %s is not found",
			req.RootAccountID)
	}

	cli, err := b.ad.AwsRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("aws request adaptor client err, req: %+v, err: %+v,rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typesBill.AwsMainBillListOption{
		CloudAccountID: req.MainAccountCloudID,
		BeginDate:      req.BeginDate,
		EndDate:        req.EndDate,
		Page: &typesBill.AwsBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		},
	}
	count, resp, err := cli.GetMainAccountBillList(cts.Kit, opt, billInfo)
	if err != nil {
		logs.Errorf("fail to list main account bill for aws, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &hcbill.AwsBillListResult{
		Count:   count,
		Details: resp,
	}, nil
}

// AwsGetRootAccountSpTotalUsage ...
func (b bill) AwsGetRootAccountSpTotalUsage(cts *rest.Contexts) (any, error) {

	req := new(hcbill.AwsRootSpUsageTotalReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rootAccount, err := b.cs.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("fait to find root account, err: %+v,rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 查询aws账单基础表
	billInfo, err := getRootAccountBillConfigInfo[billcore.AwsBillConfigExtension](
		cts.Kit, req.RootAccountID, b.cs.DataService())
	if err != nil {
		logs.Errorf("aws get root account(id: %s) bill config for aws sp usage total failed, err: %+v, rid: %s",
			req.RootAccountID, err, cts.Kit.Rid)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "bill config for root account: %s is not found",
			req.RootAccountID)
	}

	cli, err := b.ad.AwsRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("aws request adaptor client err, req: %+v, err: %+v,rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}
	opt := &typesBill.AwsRootSpUsageOption{
		PayerCloudID:  rootAccount.CloudID,
		UsageCloudIDs: req.SpUsageAccountCloudIds,
		SpArnPrefix:   req.SpArnPrefix,
		Year:          req.Year,
		Month:         req.Month,
		StartDay:      req.StartDay,
		EndDay:        req.EndDay,
	}
	usage, err := cli.GetRootSpTotalUsage(cts.Kit, billInfo, opt)
	if err != nil {
		logs.Errorf("fail to get root account sp total usage, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	result := hcbill.AwsSpUsageTotalResult{
		UnblendedCost: usage.UnblendedCost,
		SPCost:        usage.SpCost,
		SPNetCost:     usage.SpNetCost,
		AccountCount:  usage.AccountCount,
	}
	return result, nil
}
