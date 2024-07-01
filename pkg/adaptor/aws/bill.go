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

package aws

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	typesBill "hcm/pkg/adaptor/types/bill"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/math"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	curservice "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// QueryBillSQL 查询云账单的SQL
	QueryBillSQL = "SELECT %s FROM %s.%s %s"
	// QueryBillTotalSQL 查询云账单总数量的SQL
	QueryBillTotalSQL = "SELECT COUNT(*) FROM %s.%s %s"
	BucketNameDefault = "hcm-bill-%s-%s"
	BucketTimeOut     = 12  // 12小时
	StackTimeOut      = 120 // 120秒
	BucketPolicy      = `{"Version":"2008-10-17","Id":"Policy{RandomNum}","Statement":[{"Sid":"Stmta{RandomNum}",
"Effect":"Allow","Principal":{"Service":"billingreports.amazonaws.com"},"Action":["s3:GetBucketAcl",
"s3:GetBucketPolicy","s3:ListBucket"],"Resource":"arn:aws:s3:::{BucketName}",
"Condition":{"StringEquals":{"aws:SourceArn":"arn:aws:cur:{BucketRegion}:{AccountID}:definition/*",
"aws:SourceAccount":"{AccountID}"}}},{"Sid":"Stmtb{RandomNum}","Effect":"Allow",
"Principal":{"Service":"billingreports.amazonaws.com"},"Action":["s3:PutObject","s3:PutObjectAcl"],
"Resource":"arn:aws:s3:::{BucketName}/*","Condition":{"StringEquals":{"aws:SourceArn":
"arn:aws:cur:{BucketRegion}:{AccountID}:definition/*","aws:SourceAccount":"{AccountID}"}}}]}`
	// CUR配置
	CurName               = "hcmbillingreport"
	CurPrefix             = "cur"
	CurTimeUnit           = "HOURLY"
	CurFormat             = "Parquet"
	CurCompression        = "Parquet"
	ResourceSchemaElement = "RESOURCES"
	AthenaArtifact        = "ATHENA"
	ReportVersioning      = "OVERWRITE_REPORT"
	DatabaseNamePrefix    = "athenacurcfn"
	CapabilitiesIam       = "CAPABILITY_IAM"
	BucketRegion          = endpoints.UsEast1RegionID
	AthenaSavePath        = "s3://{BucketName}/{CurPrefix}/{CurName}/QueryLog"
	CrawlerCfnFileKey     = "/%s/%s/crawler-cfn.yml"
	YmlURL                = "https://{BucketName}.s3.amazonaws.com/{CurPrefix}/{CurName}/crawler-cfn.yml"
)

// GetBillList get bill list
func (a *Aws) GetBillList(kt *kit.Kit, opt *typesBill.AwsBillListOption,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) (int64, []map[string]string, error) {

	where, err := parseCondition(opt)
	if err != nil {
		return 0, nil, err
	}

	// 只有第一页时才返回数量
	var total = int64(0)
	if opt.Page != nil && opt.Page.Offset == 0 {
		// get bill total
		total, err = a.GetBillTotal(kt, where, billInfo)
		if err != nil {
			return 0, nil, err
		}
		if total == 0 {
			return 0, nil, nil
		}
	}

	sql := fmt.Sprintf(QueryBillSQL, "*", billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	if opt.Page != nil {
		sql += fmt.Sprintf(" OFFSET %d LIMIT %d", opt.Page.Offset, opt.Page.Limit)
	}
	list, err := a.GetAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

// GetBillTotal get bill total num
func (a *Aws) GetBillTotal(kt *kit.Kit, where string, billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) (
	int64, error) {

	sql := fmt.Sprintf(QueryBillTotalSQL, billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	cloudList, err := a.GetAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		return 0, err
	}

	total, err := strconv.ParseInt(cloudList[0]["_col0"], 10, 64)
	if err != nil {
		return 0, errf.Newf(errf.InvalidParameter, "get bill total parse id %d failed, err: %v", total, err)
	}

	return total, nil
}

func (a *Aws) GetAwsAthenaQuery(kt *kit.Kit, query string,
	billInfo *cloud.AccountBillConfig[cloud.AwsBillConfigExtension]) ([]map[string]string, error) {

	client, err := a.clientSet.athenaClient(billInfo.Extension.Region)
	if err != nil {
		return nil, err
	}

	var s athena.StartQueryExecutionInput
	s.SetQueryString(query)

	var r athena.ResultConfiguration
	r.SetOutputLocation(billInfo.Extension.SavePath)
	s.SetResultConfiguration(&r)

	result, err := client.StartQueryExecution(&s)
	if err != nil {
		logs.Errorf("aws athena start query error, billInfo: %+v, err: %v, rid: %s", billInfo, err, kt.Rid)
		return nil, err
	}

	var qri athena.GetQueryExecutionInput
	qri.SetQueryExecutionId(*result.QueryExecutionId)

	var qrop *athena.GetQueryExecutionOutput
	duration := time.Duration(100) * time.Millisecond

	for {
		qrop, err = client.GetQueryExecution(&qri)
		if err != nil {
			logs.Errorf("aws cloud athena get query loop err, queryExecutionId: %s, err: %v, rid: %s",
				*result.QueryExecutionId, err, kt.Rid)
			return nil, err
		}

		if *qrop.QueryExecution.Status.State != "RUNNING" && *qrop.QueryExecution.Status.State != "QUEUED" {
			break
		}
		time.Sleep(duration)
	}

	if *qrop.QueryExecution.Status.State == "SUCCEEDED" {
		var ip athena.GetQueryResultsInput
		ip.SetQueryExecutionId(*result.QueryExecutionId)

		op, err := client.GetQueryResults(&ip)
		if err != nil {
			logs.Errorf("aws cloud athena get query result err, queryExecutionId: %s, err: %v, rid: %s",
				*result.QueryExecutionId, err, kt.Rid)
			return nil, err
		}

		list := make([]map[string]string, 0)
		resultMap := make([]string, 0)
		for index, row := range op.ResultSet.Rows {
			// parse table field
			if index == 0 {
				for _, column := range row.Data {
					tmpField := converter.PtrToVal(column.VarCharValue)
					resultMap = append(resultMap, tmpField)
				}
			} else {
				tmpMap := make(map[string]string, 0)
				for colKey, column := range row.Data {
					tmpValue := converter.PtrToVal(column.VarCharValue)
					if tmpValue == "" || strings.IndexAny(tmpValue, "Ee") == -1 {
						tmpMap[resultMap[colKey]] = tmpValue
						continue
					}

					decimalNum, err := math.NewDecimalFromString(tmpValue)
					if err != nil {
						tmpMap[resultMap[colKey]] = tmpValue
						continue
					}
					tmpMap[resultMap[colKey]] = decimalNum.ToString()
				}
				list = append(list, tmpMap)
			}
		}

		return list, nil
	}

	var errMsg = *qrop.QueryExecution.Status.State
	if qrop.QueryExecution.Status.StateChangeReason != nil {
		errMsg = *qrop.QueryExecution.Status.StateChangeReason
	}

	if strings.Contains(errMsg, fmt.Sprintf("%s does not exist", billInfo.CloudDatabaseName)) {
		return nil, errf.Newf(errf.RecordNotFound, "accountID: %s bill record is not found", billInfo.AccountID)
	}

	return nil, errf.Newf(errf.DecodeRequestFailed, "Aws Athena Query Failed(%s)", errMsg)
}

func parseCondition(opt *typesBill.AwsBillListOption) (string, error) {
	var condition string
	if opt.BeginDate != "" && opt.EndDate != "" {
		searchDate, err := time.Parse(constant.DateLayout, opt.BeginDate)
		if err != nil {
			return "", fmt.Errorf("conv search date failed, err: %v", err)
		}
		condition = fmt.Sprintf("WHERE year = '%d' AND month = '%d' AND "+
			"date(line_item_usage_start_date) >= date '%s' AND date(line_item_usage_start_date) <= date '%s'",
			searchDate.Year(), searchDate.Month(), opt.BeginDate, opt.EndDate)
	}
	return condition, nil
}

// CreateBucket create s3 bucket.
// reference: https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/API_CreateBucket.html
func (a *Aws) CreateBucket(kt *kit.Kit, opt *typesBill.AwsBillBucketCreateReq) (*string, error) {
	client, err := a.clientSet.s3Client(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor s3 bucket client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	req := &s3.CreateBucketInput{Bucket: converter.ValToPtr(opt.Bucket)}

	resp, err := client.CreateBucketWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor s3 create bucket failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	return resp.Location, nil
}

// DeleteBucket delete s3 bucket.
// reference: https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/API_DeleteBucket.html
func (a *Aws) DeleteBucket(kt *kit.Kit, opt *typesBill.AwsBillBucketDeleteReq) error {
	client, err := a.clientSet.s3Client(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor s3 delete bucket client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	req := &s3.DeleteBucketInput{Bucket: converter.ValToPtr(opt.Bucket)}
	_, err = client.DeleteBucketWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor s3 delete bucket failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	return nil
}

// ListBucket list bucket.
// reference: https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/API_ListBuckets.html
func (a *Aws) ListBucket(kt *kit.Kit, region string) ([]*s3.Bucket, error) {
	client, err := a.clientSet.s3Client(region)
	if err != nil {
		logs.Errorf("aws adaptor bill bucket list client failed, region: %s, err: %v, rid: %s",
			region, err, kt.Rid)
		return nil, err
	}

	req := &s3.ListBucketsInput{}
	resp, err := client.ListBucketsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor bill bucket list failed, region: %s, err: %v, rid: %s", region, err, kt.Rid)
		return nil, err
	}

	return resp.Buckets, nil
}

// GetObject get object.
// reference: https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/API_GetObject.html
func (a *Aws) GetObject(kt *kit.Kit, opt *typesBill.AwsBillGetObjectReq) (*s3.GetObjectOutput, error) {
	client, err := a.clientSet.s3Client(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor bill get object client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	req := &s3.GetObjectInput{
		Bucket: converter.ValToPtr(opt.Bucket),
		Key:    converter.ValToPtr(opt.Key),
	}
	resp, err := client.GetObjectWithContext(kt.Ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetBucketPolicy get bucket policy.
// reference: https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/API_GetBucketPolicy.html
func (a *Aws) GetBucketPolicy(kt *kit.Kit, opt *typesBill.AwsBillBucketPolicyReq) (*string, error) {
	client, err := a.clientSet.s3Client(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor get bucket policy client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	req := &s3.GetBucketPolicyInput{
		Bucket: converter.ValToPtr(opt.Bucket),
	}
	resp, err := client.GetBucketPolicyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor get bucket policy failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	return resp.Policy, nil
}

// PutBucketPolicy put bucket policy.
// reference: https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/API_PutBucketPolicy.html
func (a *Aws) PutBucketPolicy(kt *kit.Kit, opt *typesBill.AwsBillBucketPolicyReq) error {
	client, err := a.clientSet.s3Client(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor put bucket policy client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	randomNum := time.Now().UnixMilli()
	policy := strings.ReplaceAll(BucketPolicy, "{RandomNum}", strconv.FormatInt(randomNum, 10))
	policy = strings.ReplaceAll(policy, "{BucketName}", opt.Bucket)
	policy = strings.ReplaceAll(policy, "{BucketRegion}", opt.Region)
	policy = strings.ReplaceAll(policy, "{AccountID}", a.CloudAccountID())

	req := &s3.PutBucketPolicyInput{
		Bucket: converter.ValToPtr(opt.Bucket),
		Policy: converter.ValToPtr(policy),
	}
	_, err = client.PutBucketPolicyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor put bucket policy failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	return nil
}

// PutReportDefinition put report definition.
// reference: https://docs.aws.amazon.com/zh_cn/aws-cost-management/latest/APIReference/API_cur_PutReportDefinition.html
func (a *Aws) PutReportDefinition(kt *kit.Kit, opt *typesBill.AwsBillPutReportDefinitionReq) error {
	client, err := a.clientSet.costAndUsageReportClient(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor cur put report definition client failed, opt: %v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return err
	}

	req := &curservice.PutReportDefinitionInput{
		ReportDefinition: &curservice.ReportDefinition{
			S3Bucket:                 converter.ValToPtr(opt.Bucket),
			ReportName:               converter.ValToPtr(opt.CurName),
			S3Prefix:                 converter.ValToPtr(opt.CurPrefix),
			S3Region:                 converter.ValToPtr(opt.Region),
			Format:                   converter.ValToPtr(opt.Format),
			TimeUnit:                 converter.ValToPtr(opt.TimeUnit),
			Compression:              converter.ValToPtr(opt.Compression),
			AdditionalSchemaElements: opt.SchemaElements,
			AdditionalArtifacts:      opt.Artifacts,
			ReportVersioning:         converter.ValToPtr(opt.ReportVersioning),
		},
	}

	_, err = client.PutReportDefinitionWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor cur put report definition error, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	return nil
}

// DeleteReportDefinition delete report definition.
// reference: https://docs.aws.amazon.com/zh_cn/aws-cost-management/latest/APIReference/
// API_cur_DeleteReportDefinition.html
func (a *Aws) DeleteReportDefinition(kt *kit.Kit, opt *typesBill.AwsBillDeleteReportDefinitionReq) error {
	client, err := a.clientSet.costAndUsageReportClient(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor delete report definition client failed, opt: %v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return err
	}

	req := &curservice.DeleteReportDefinitionInput{
		ReportName: converter.ValToPtr(opt.ReportName),
	}

	_, err = client.DeleteReportDefinitionWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor delete report definition error, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	return nil
}

// CreateStack create stack.
// reference: https://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/API_CreateStack.html
func (a *Aws) CreateStack(kt *kit.Kit, opt *typesBill.AwsCreateStackReq) (string, error) {
	client, err := a.clientSet.cloudFormationClient(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor formation client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return "", err
	}

	req := &cloudformation.CreateStackInput{
		StackName:    converter.ValToPtr(opt.StackName),
		TemplateURL:  converter.ValToPtr(opt.TemplateURL),
		Capabilities: opt.Capabilities,
	}
	resp, err := client.CreateStackWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor formation create stack failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return "", err
	}

	return converter.PtrToVal(resp.StackId), nil
}

// DescribeStack describe stack.
// reference: https://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/API_DescribeStacks.html
func (a *Aws) DescribeStack(kt *kit.Kit, opt *typesBill.AwsDeleteStackReq) ([]*cloudformation.Stack, error) {
	client, err := a.clientSet.cloudFormationClient(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor formation client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	req := &cloudformation.DescribeStacksInput{
		StackName: converter.ValToPtr(opt.StackID),
	}
	resp, err := client.DescribeStacksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor formation create stack failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	return resp.Stacks, nil
}

// DeleteStack delete stack.
// reference: https://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/API_DeleteStack.html
func (a *Aws) DeleteStack(kt *kit.Kit, opt *typesBill.AwsDeleteStackReq) error {
	client, err := a.clientSet.cloudFormationClient(opt.Region)
	if err != nil {
		logs.Errorf("aws adaptor formation client failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	req := &cloudformation.DeleteStackInput{
		StackName: converter.ValToPtr(opt.StackID),
	}
	_, err = client.DeleteStackWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("aws adaptor formation delete stack failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	return nil
}

// -------------- 新增账号账单管理部分 --------------

// GetMainAccountBillList get bill list for main account
func (a *Aws) GetMainAccountBillList(kt *kit.Kit, opt *typesBill.AwsMainBillListOption,
	billInfo *billcore.RootAccountBillConfig[billcore.AwsBillConfigExtension]) (int64, []map[string]string, error) {

	where, err := parseRootCondition(opt)
	if err != nil {
		return 0, nil, err
	}

	// 只有第一页时才返回数量
	var total = int64(0)
	if opt.Page != nil && opt.Page.Offset == 0 {
		// get bill total
		total, err = a.GetRootAccountBillTotal(kt, where, billInfo)
		if err != nil {
			return 0, nil, err
		}
		if total == 0 {
			return 0, nil, nil
		}
	}

	sql := fmt.Sprintf(QueryBillSQL, "*", billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	if opt.Page != nil {
		sql += fmt.Sprintf(" OFFSET %d LIMIT %d", opt.Page.Offset, opt.Page.Limit)
	}
	list, err := a.GetRootAccountAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		return 0, nil, err
	}
	for i, item := range list {
		bill_usage_account_id := item["line_item_usage_account_id"]
		if bill_usage_account_id != opt.CloudAccountID {
			logs.Errorf("@%d: line_item_usage_account_id mismatch, want: %s, bill_id: %s, sql: %s, rid: %s",
				i, billInfo.RootAccountID, bill_usage_account_id, sql, kt.Rid)
		}
	}
	return total, list, nil
}

func parseRootCondition(opt *typesBill.AwsMainBillListOption) (string, error) {
	var condition = fmt.Sprintf("WHERE line_item_usage_account_id = '%s' ", opt.CloudAccountID)
	if opt.BeginDate != "" && opt.EndDate != "" {
		searchDate, err := time.Parse(constant.DateLayout, opt.BeginDate)
		if err != nil {
			return "", fmt.Errorf("conv search date failed, err: %v", err)
		}
		condition += fmt.Sprintf("AND year = '%d' AND month = '%d' AND "+
			"date(line_item_usage_start_date) >= date '%s' AND date(line_item_usage_start_date) <= date '%s'",
			searchDate.Year(), searchDate.Month(), opt.BeginDate, opt.EndDate)
	}
	return condition, nil
}

// GetRootAccountBillTotal get bill list total for root account
func (a *Aws) GetRootAccountBillTotal(
	kt *kit.Kit, where string, billInfo *billcore.RootAccountBillConfig[billcore.AwsBillConfigExtension]) (
	int64, error) {

	sql := fmt.Sprintf(QueryBillTotalSQL, billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	cloudList, err := a.GetRootAccountAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		return 0, err
	}

	total, err := strconv.ParseInt(cloudList[0]["_col0"], 10, 64)
	if err != nil {
		return 0, errf.Newf(errf.InvalidParameter, "get bill total parse id %d failed, err: %v", total, err)
	}

	return total, nil
}

// GetRootAccountAwsAthenaQuery get aws athena query
func (a *Aws) GetRootAccountAwsAthenaQuery(kt *kit.Kit, query string,
	billInfo *billcore.RootAccountBillConfig[billcore.AwsBillConfigExtension]) ([]map[string]string, error) {

	client, err := a.clientSet.athenaClient(billInfo.Extension.Region)
	if err != nil {
		return nil, err
	}

	var s athena.StartQueryExecutionInput
	s.SetQueryString(query)

	var r athena.ResultConfiguration
	r.SetOutputLocation(billInfo.Extension.SavePath)
	s.SetResultConfiguration(&r)

	result, err := client.StartQueryExecution(&s)
	if err != nil {
		logs.Errorf("aws athena start query error, billInfo: %+v, err: %v, rid: %s", billInfo, err, kt.Rid)
		return nil, err
	}

	var qri athena.GetQueryExecutionInput
	qri.SetQueryExecutionId(*result.QueryExecutionId)

	var qrop *athena.GetQueryExecutionOutput
	duration := time.Duration(100) * time.Millisecond

	for {
		qrop, err = client.GetQueryExecution(&qri)
		if err != nil {
			logs.Errorf("aws cloud athena get query loop err, queryExecutionId: %s, err: %v, rid: %s",
				*result.QueryExecutionId, err, kt.Rid)
			return nil, err
		}

		if *qrop.QueryExecution.Status.State != "RUNNING" && *qrop.QueryExecution.Status.State != "QUEUED" {
			break
		}
		time.Sleep(duration)
	}

	if *qrop.QueryExecution.Status.State == "SUCCEEDED" {
		var ip athena.GetQueryResultsInput
		ip.SetQueryExecutionId(*result.QueryExecutionId)

		op, err := client.GetQueryResults(&ip)
		if err != nil {
			logs.Errorf("aws cloud athena get query result err, queryExecutionId: %s, err: %v, rid: %s",
				*result.QueryExecutionId, err, kt.Rid)
			return nil, err
		}

		list := make([]map[string]string, 0)
		resultMap := make([]string, 0)
		for index, row := range op.ResultSet.Rows {
			// parse table field
			if index == 0 {
				for _, column := range row.Data {
					tmpField := converter.PtrToVal(column.VarCharValue)
					resultMap = append(resultMap, tmpField)
				}
			} else {
				tmpMap := make(map[string]string, 0)
				for colKey, column := range row.Data {
					tmpValue := converter.PtrToVal(column.VarCharValue)
					if tmpValue == "" || strings.IndexAny(tmpValue, "Ee") == -1 {
						tmpMap[resultMap[colKey]] = tmpValue
						continue
					}

					decimalNum, err := math.NewDecimalFromString(tmpValue)
					if err != nil {
						tmpMap[resultMap[colKey]] = tmpValue
						continue
					}
					tmpMap[resultMap[colKey]] = decimalNum.ToString()
				}
				list = append(list, tmpMap)
			}
		}

		return list, nil
	}

	var errMsg = *qrop.QueryExecution.Status.State
	if qrop.QueryExecution.Status.StateChangeReason != nil {
		errMsg = *qrop.QueryExecution.Status.StateChangeReason
	}

	if strings.Contains(errMsg, fmt.Sprintf("%s does not exist", billInfo.CloudDatabaseName)) {
		return nil, errf.Newf(
			errf.RecordNotFound, "root accountID: %s bill record is not found", billInfo.RootAccountID)
	}

	return nil, errf.Newf(errf.DecodeRequestFailed, "Aws Athena Query Failed(%s)", errMsg)
}
