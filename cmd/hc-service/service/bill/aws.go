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
	"hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	protobill "hcm/pkg/api/data-service/cloud/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

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

// AwsPutReportDefinition aws put report definition.
func (b bill) AwsPutReportDefinition(kt *kit.Kit, accountID string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) error {

	if billInfo.Status != constant.StatusSetBucketPolicy && billInfo.Status != constant.StatusReCurReport {
		return nil
	}

	cli, err := b.ad.Aws(kt, accountID)
	if err != nil {
		logs.Errorf("aws put report definition client failed, accountID: %s, err: %+v", accountID, err)
		return err
	}

	// 获取cur名称
	curName := aws.CurName
	if billInfo.Status == constant.StatusReCurReport {
		curName += fmt.Sprintf("%s", time.Now().Format("20060102150405"))
	}

	opt := newAwsBillPutReportDefinitionReq(curName)
	opt.Bucket = billInfo.Extension.Bucket
	opt.Region = billInfo.Extension.Region
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
	// 记录日志方便排查问题
	if billInfo.Status == constant.StatusReCurReport {
		logs.Infof("aws put report definition regen cur report success, accountID: %s, billID: %s, billInfo: %+v, "+
			"opt: %+v, rid: %s", accountID, billInfo.ID, cvt.PtrToVal(billInfo), cvt.PtrToVal(opt), kt.Rid)
	}

	// 更新
	savePath := fmt.Sprintf(aws.AthenaSavePath, opt.Bucket, opt.CurPrefix, opt.CurName)
	ymlURL := fmt.Sprintf(aws.YmlURL, opt.Bucket, opt.CurPrefix, opt.CurName)
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

func newAwsBillPutReportDefinitionReq(curName string) *typesBill.AwsBillPutReportDefinitionReq {
	return &typesBill.AwsBillPutReportDefinitionReq{
		CurName:          curName,
		CurPrefix:        aws.CurPrefix,
		Format:           aws.CurFormat,
		TimeUnit:         aws.CurTimeUnit,
		Compression:      aws.CurCompression,
		SchemaElements:   []*string{cvt.ValToPtr(aws.ResourceSchemaElement)},
		Artifacts:        []*string{cvt.ValToPtr(aws.AthenaArtifact)},
		ReportVersioning: aws.ReportVersioning,
	}
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
