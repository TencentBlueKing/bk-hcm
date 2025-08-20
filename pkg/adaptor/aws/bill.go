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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/math"
	"hcm/pkg/tools/times"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	curservice "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/shopspring/decimal"
)

const (
	// QueryBillSQL 查询云账单的SQL
	QueryBillSQL = "SELECT %s FROM %s.%s %s "
	// QueryBillTotalSQL 查询云账单总数量的SQL
	QueryBillTotalSQL = "SELECT COUNT(*) FROM %s.%s %s "
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
	// AthenaSavePath s3://{BucketName}/{CurPrefix}/{CurName}/QueryLog
	AthenaSavePath    = "s3://%s/%s/%s/QueryLog"
	CrawlerCfnFileKey = "/%s/%s/crawler-cfn.yml"
	// YmlURL https://{BucketName}.s3.amazonaws.com/{CurPrefix}/{CurName}/crawler-cfn.yml
	YmlURL = "https://%s.s3.amazonaws.com/%s/%s/crawler-cfn.yml"
)

const (
	// QueryRootBillSelectField select fields
	QueryRootBillSelectField = ` 
	bill_payer_account_id,
	identity_line_item_id,
	line_item_usage_account_id,
	bill_invoice_id,
	bill_billing_entity,
	line_item_product_code,
	product_product_family,
	product_product_name,
	line_item_usage_type,
	product_instance_type,
	product_region,
	product_location,
	line_item_resource_id,
	pricing_term,
	line_item_line_item_type,
	line_item_line_item_description,
	'' AS line_item_usage_start_date,
	'' AS line_item_usage_end_date,
	pricing_unit,
	line_item_currency_code,
	pricing_public_on_demand_rate,
	line_item_unblended_rate,
	savings_plan_savings_plan_rate,
	line_item_net_unblended_rate,
	line_item_operation,
	savings_plan_savings_plan_a_r_n,
	SUM(line_item_usage_amount) AS line_item_usage_amount,
	SUM(pricing_public_on_demand_cost) AS pricing_public_on_demand_cost,
	SUM(line_item_unblended_cost) AS line_item_unblended_cost,
	SUM(line_item_net_unblended_cost) AS line_item_net_unblended_cost,
	SUM(savings_plan_savings_plan_effective_cost) AS savings_plan_savings_plan_effective_cost,
	SUM(savings_plan_net_savings_plan_effective_cost) AS savings_plan_net_savings_plan_effective_cost,
	SUM(reservation_effective_cost) AS reservation_effective_cost,
	SUM(reservation_net_effective_cost) AS reservation_net_effective_cost 
`
	// QueryRootBillGroupBySQL group by sql fragment
	QueryRootBillGroupBySQL = ` GROUP BY 
	identity_line_item_id, bill_payer_account_id,
	line_item_usage_account_id,
	bill_invoice_id,
	bill_billing_entity,
	savings_plan_savings_plan_a_r_n,
	line_item_product_code,
	product_product_family,
	product_product_name,
	line_item_usage_type,
	product_instance_type,
	product_region,
	product_location,
	line_item_resource_id,
	pricing_term,
	line_item_line_item_type,
	line_item_line_item_description,
	pricing_unit,
	line_item_currency_code,
	pricing_public_on_demand_rate,
	line_item_unblended_rate,
	savings_plan_savings_plan_rate, 
	line_item_net_unblended_rate, 
	line_item_operation `

	// QueryRootBillOrderBySQL 没有指定字段的情况下得到账单顺序会乱，多页的时候会导致账单重复、遗漏
	QueryRootBillOrderBySQL = " ORDER BY identity_line_item_id,line_item_usage_start_date,line_item_usage_end_date "
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

// GetAwsAthenaQuery ...
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
		return getAwsAthenaQuerySuccessSituation(kt, client, result)
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

func getAwsAthenaQuerySuccessSituation(kt *kit.Kit, client *athena.Athena, result *athena.StartQueryExecutionOutput) (
	[]map[string]string, error) {

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
			tmpMap := make(map[string]string)
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

	sql := fmt.Sprintf(QueryBillSQL, QueryRootBillSelectField, billInfo.CloudDatabaseName, billInfo.CloudTableName,
		where)
	sql += QueryRootBillGroupBySQL
	sql += QueryRootBillOrderBySQL
	if opt.Page != nil {
		sql += fmt.Sprintf(" OFFSET %d LIMIT %d", opt.Page.Offset, opt.Page.Limit)
	}
	list, err := a.GetRootAccountAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		return 0, nil, err
	}
	// 暂不返回数量
	return 0, list, nil
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
func (a *Aws) GetRootAccountBillTotal(kt *kit.Kit, where string, billInfo *billcore.AwsRootBillConfig) (int64, error) {

	sql := fmt.Sprintf(QueryBillTotalSQL, billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	sql += QueryRootBillGroupBySQL
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
func (a *Aws) GetRootAccountAwsAthenaQuery(kt *kit.Kit, query string, billInfo *billcore.AwsRootBillConfig) (
	[]map[string]string, error) {

	logs.V(5).Infof("aws root account athena query sql: [%s], rid: %s", query, kt.Rid)
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
		return getRootAccountAwsAthenaQuerySuccessSituation(kt, result, client)
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

func getRootAccountAwsAthenaQuerySuccessSituation(kt *kit.Kit, result *athena.StartQueryExecutionOutput,
	client *athena.Athena) ([]map[string]string, error) {

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

const (
	// AwsSPQuerySQL ...
	AwsSPQuerySQL = `SELECT 
			count(DISTINCT line_item_usage_account_id) AS account_cnt,
			sum(line_item_unblended_cost) AS unblended_cost, 
			sum(savings_plan_savings_plan_effective_cost) AS sp_cost,
			sum(savings_plan_net_savings_plan_effective_cost) AS sp_net_cost
			FROM %s.%s  
			WHERE line_item_line_item_type = 'SavingsPlanCoveredUsage'
	`
)

// GetRootSpTotalUsage get sp total usage for root account
func (a *Aws) GetRootSpTotalUsage(kt *kit.Kit, billInfo *billcore.AwsRootBillConfig,
	opt *typesBill.AwsRootSpUsageOption) (*typesBill.AwsSpUsageTotalResult, error) {

	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "bill info is required")
	}
	if opt == nil {
		return nil, errf.Newf(errf.RecordNotFound, "opt for get sp usage is required")
	}

	sql := fmt.Sprintf(AwsSPQuerySQL, billInfo.CloudDatabaseName, billInfo.CloudTableName)
	sql += fmt.Sprintf(" AND bill_payer_account_id = '%s'", opt.PayerCloudID)
	sql += fmt.Sprintf(" AND year = '%d'", opt.Year)
	sql += fmt.Sprintf(" AND month = '%d'", opt.Month)
	sql += fmt.Sprintf(" AND date(line_item_usage_start_date) >= date '%d-%02d-%02d' ",
		opt.Year, opt.Month, opt.StartDay)
	sql += fmt.Sprintf(" AND date(line_item_usage_start_date) <= date '%d-%02d-%02d' ",
		opt.Year, opt.Month, opt.EndDay)
	if len(opt.UsageCloudIDs) > 0 {
		sql += fmt.Sprintf(" AND line_item_usage_account_id IN ('%s') ", strings.Join(opt.UsageCloudIDs, "','"))
	}
	if len(opt.SpArnPrefix) > 0 {
		sql += fmt.Sprintf(" AND savings_plan_savings_plan_a_r_n LIKE '%s%%'", opt.SpArnPrefix)
	}

	cloudList, err := a.GetRootAccountAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		logs.Errorf("fail to call aws athena query for get sp total usage, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(cloudList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "no sp usage found")
	}

	ret := &typesBill.AwsSpUsageTotalResult{}
	ret.AccountCount, err = strconv.ParseUint(cloudList[0]["account_cnt"], 10, 64)
	if err != nil {
		logs.Errorf("fail to parse account count, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	ret.Currency = enumor.CurrencyCode(cloudList[0]["pricing_currency"])
	ubCost, err := decimal.NewFromString(cloudList[0]["unblended_cost"])
	if err != nil {
		logs.Errorf("fail to parse unblended cost, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	ret.UnblendedCost = converter.ValToPtr(ubCost)
	spCost, err := decimal.NewFromString(cloudList[0]["sp_cost"])
	if err != nil {
		logs.Errorf("fail to parse sp cost, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	ret.SpCost = converter.ValToPtr(spCost)
	spNetCost, err := decimal.NewFromString(cloudList[0]["sp_net_cost"])
	if err != nil {
		logs.Errorf("fail to parse sp net cost, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	ret.SpNetCost = converter.ValToPtr(spNetCost)

	return ret, nil
}

// AwsListRootOutsideMonthBill get bill list outside given bill month for main account
func (a *Aws) AwsListRootOutsideMonthBill(kt *kit.Kit, opt *typesBill.AwsMainOutsideMonthBillLitOpt,
	billInfo *billcore.RootAccountBillConfig[billcore.AwsBillConfigExtension]) ([]map[string]string, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}
	day1 := time.Date(int(opt.Year), time.Month(opt.Month), 1, 0, 0, 0, 0, time.UTC)
	nextBillMonthYear, nextBillMonth := times.GetRelativeMonth(day1, 1)

	// 拼接查询条件
	var condition = fmt.Sprintf("WHERE bill_payer_account_id = '%s' ", opt.PayerAccountID)
	condition += fmt.Sprintf(" AND year = '%d'", opt.Year)
	condition += fmt.Sprintf(" AND month = '%d'", opt.Month)
	condition += fmt.Sprintf(`AND (date(line_item_usage_start_date) < date '%d-%02d-01' 
		OR date(line_item_usage_start_date) >= date '%d-%02d-01')`,
		opt.Year, opt.Month, nextBillMonthYear, nextBillMonth)

	if len(opt.UsageAccountIDs) > 0 {
		condition += fmt.Sprintf(" AND line_item_usage_account_id IN ('%s') ", strings.Join(opt.UsageAccountIDs, "','"))
	}
	sql := fmt.Sprintf(QueryBillSQL, QueryRootBillSelectField, billInfo.CloudDatabaseName, billInfo.CloudTableName,
		condition)
	sql += QueryRootBillGroupBySQL
	sql += QueryRootBillOrderBySQL
	if opt.Page != nil {
		sql += fmt.Sprintf(" OFFSET %d LIMIT %d", opt.Page.Offset, opt.Page.Limit)
	}
	list, err := a.GetRootAccountAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// AwsRootBillListByQueryFields get bill list by query fields for root account
func (a *Aws) AwsRootBillListByQueryFields(kt *kit.Kit, opt *typesBill.AwsRootDeductBillListOpt,
	billInfo *billcore.RootAccountBillConfig[billcore.AwsBillConfigExtension]) ([]map[string]string, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	// 拼接查询条件
	var condition = fmt.Sprintf("WHERE year = '%d' AND month = '%d'", opt.Year, opt.Month)
	if opt.PayerAccountID != "" {
		condition += fmt.Sprintf("AND bill_payer_account_id = '%s' ", opt.PayerAccountID)
	}
	if opt.BeginDate != "" && opt.EndDate != "" {
		condition += fmt.Sprintf(`AND (date(line_item_usage_start_date) >= date '%s' 
		AND date(line_item_usage_start_date) <= date '%s')`, opt.BeginDate, opt.EndDate)
	}
	if len(opt.FieldsMap) > 0 {
		for key, values := range opt.FieldsMap {
			if len(values) == 0 {
				continue
			}
			condition += fmt.Sprintf(" AND %s IN('%s') ", key, strings.Join(values, "','"))
		}
	}
	sql := fmt.Sprintf(QueryBillSQL, QueryRootBillSelectField, billInfo.CloudDatabaseName, billInfo.CloudTableName,
		condition)
	sql += QueryRootBillGroupBySQL
	sql += QueryRootBillOrderBySQL
	if opt.Page != nil {
		sql += fmt.Sprintf(" OFFSET %d LIMIT %d", opt.Page.Offset, opt.Page.Limit)
	}
	list, err := a.GetRootAccountAwsAthenaQuery(kt, sql, billInfo)
	if err != nil {
		return nil, err
	}
	return list, nil
}
