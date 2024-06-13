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

// Package bill ...
package bill

import (
	"time"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// AwsBillListResult defines aws bill list result.
type AwsBillListResult struct {
	Count   int64       `json:"count,omitempty"`
	Details interface{} `json:"details"`
}

// AwsBillListOption define aws bill list option.
type AwsBillListOption struct {
	AccountID string `json:"account_id" validate:"required"`
	// 起始日期，格式为yyyy-mm-dd，不支持跨月查询
	BeginDate string `json:"begin_date" validate:"required"`
	// 截止日期，格式为yyyy-mm-dd，不支持跨月查询
	EndDate string       `json:"end_date" validate:"required"`
	Page    *AwsBillPage `json:"page" validate:"omitempty"`
}

// Validate aws bill list option.
func (opt AwsBillListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	// aws的AthenaQuery属于sql查询，跟SDK接口的Limit限制不同
	if opt.Page != nil {
		if opt.Page.Limit == 0 {
			return errf.New(errf.InvalidParameter, "aws.limit is required")
		}
		if opt.Page.Limit > 100000 {
			return errf.New(errf.InvalidParameter, "aws.limit should <= 100000")
		}
	}

	if (opt.BeginDate == "" && opt.EndDate != "") || (opt.BeginDate != "" && opt.EndDate == "") {
		return errf.New(errf.InvalidParameter, "begin_date and end_date can not be empty")
	}

	beginDate, err := time.Parse(constant.DateLayout, opt.BeginDate)
	if err != nil {
		return err
	}

	endDate, err := time.Parse(constant.DateLayout, opt.EndDate)
	if err != nil {
		return err
	}

	if beginDate.Year() != endDate.Year() || beginDate.Month() != endDate.Month() {
		return errf.New(errf.InvalidParameter, "begin_date and end_date are not the same year and month.")
	}

	return nil
}

// AwsBillPage defines aws bill page option.
type AwsBillPage struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

// Validate AwsBillPage.
func (t AwsBillPage) Validate() error {
	if t.Limit == 0 {
		return errf.New(errf.InvalidParameter, "limit is required")
	}

	if t.Limit > core.AwsQueryLimit {
		return errf.New(errf.InvalidParameter, "aws.limit should <= 1000")
	}

	return nil
}

// -------------------------- CreateBucket --------------------------

// AwsBillBucketCreateReq define aws bill bucket create request.
type AwsBillBucketCreateReq struct {
	Bucket string `json:"bucket" validate:"required"`
	Region string `json:"region" validate:"omitempty"`
}

// Validate aws bill bucket create option.
func (opt AwsBillBucketCreateReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- DeleteBucket --------------------------

// AwsBillBucketDeleteReq define aws bill bucket delete request.
type AwsBillBucketDeleteReq struct {
	Bucket string `json:"bucket" validate:"required"`
	Region string `json:"region" validate:"required"`
}

// Validate aws bill bucket delete option.
func (opt AwsBillBucketDeleteReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- BucketPolicy --------------------------

// AwsBillBucketPolicyReq define aws bill bucket policy request.
type AwsBillBucketPolicyReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Bucket    string `json:"bucket" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate aws bill bucket policy option.
func (opt AwsBillBucketPolicyReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- PutReportDefinition --------------------------

// AwsBillPutReportDefinitionReq define aws bill put report definition request.
type AwsBillPutReportDefinitionReq struct {
	Bucket           string    `json:"bucket" validate:"required"`
	Region           string    `json:"region" validate:"required"`
	CurName          string    `json:"cur_name" validate:"required"`
	CurPrefix        string    `json:"cur_prefix" validate:"required"`
	Format           string    `json:"format" validate:"required"`
	TimeUnit         string    `json:"time_unit" validate:"required"`
	Compression      string    `json:"compression" validate:"required"`
	SchemaElements   []*string `json:"schema_elements" validate:"required"`
	Artifacts        []*string `json:"artifacts" validate:"required"`
	ReportVersioning string    `json:"report_versioning" validate:"required"`
}

// Validate aws bill put report definition option.
func (opt AwsBillPutReportDefinitionReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- DeleteReportDefinition --------------------------

// AwsBillDeleteReportDefinitionReq define aws bill delete report definition request.
type AwsBillDeleteReportDefinitionReq struct {
	Region     string `json:"region" validate:"required"`
	ReportName string `json:"report_name" validate:"required"`
}

// Validate aws bill delete report definition option.
func (opt AwsBillDeleteReportDefinitionReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- CreateStack --------------------------

// AwsCreateStackReq define aws create stack request.
type AwsCreateStackReq struct {
	Region       string    `json:"region" validate:"required"`
	StackName    string    `json:"stack_name" validate:"required"`
	TemplateURL  string    `json:"template_url" validate:"required"`
	Capabilities []*string `json:"capabilities" validate:"required"`
}

// Validate aws create stack option.
func (opt AwsCreateStackReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- DeleteStack --------------------------

// AwsDeleteStackReq define aws delete stack request.
type AwsDeleteStackReq struct {
	Region  string `json:"region" validate:"required"`
	StackID string `json:"stack_id" validate:"required"`
}

// Validate aws delete stack option.
func (opt AwsDeleteStackReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- GetObject --------------------------

// AwsBillGetObjectReq define aws bill get object request.
type AwsBillGetObjectReq struct {
	Bucket string `json:"bucket" validate:"required"`
	Region string `json:"region" validate:"required"`
	Key    string `json:"key" validate:"required"`
}

// Validate aws bill bucket get object option.
func (opt AwsBillGetObjectReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- List --------------------------

// TCloudBillListResult defines tcloud bill list result.
type TCloudBillListResult struct {
	Count   *uint64     `json:"count,omitempty"`
	Details interface{} `json:"details"`
}

// TCloudBillListOption define tcloud bill list option.
type TCloudBillListOption struct {
	AccountID string `json:"account_id" validate:"required"`
	// 月份，格式为yyyy-mm，不能早于开通账单2.0的月份，最多可拉取24个月内的数据,不支持跨月查询
	Month string `json:"month" validate:"omitempty"`
	// 起始日期，周期开始时间，格式为Y-m-d H:i:s，Month和BeginDate&EndDate必传一个，如果有该字段则Month字段无效
	BeginDate string `json:"begin_date" validate:"omitempty"`
	// 截止日期，周期结束时间，格式为Y-m-d H:i:s，Month和BeginDate&EndDate必传一个，如果有该字段则Month字段无效
	EndDate string `json:"end_date" validate:"omitempty"`
	// Limit: 最大值为100
	Page *core.TCloudPage `json:"page" validate:"omitempty"`
	// 本次请求的上下文信息，可用于下一次请求的请求参数中，加快查询速度
	// 注意：此字段可能返回 null，表示取不到有效值。
	Context *string `json:"Context" validate:"omitempty"`
}

// Validate tcloud bill list option.
func (opt TCloudBillListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if opt.Month == "" && opt.BeginDate == "" && opt.EndDate == "" {
		return errf.New(errf.InvalidParameter, "month and begin_date and end_date can not be empty")
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HuaWeiBillListResult defines huawei bill list result.
type HuaWeiBillListResult struct {
	Count   *uint64     `json:"count,omitempty"`
	Details interface{} `json:"details"`
}

// HuaWeiBillListOption defines huawei bill list options.
type HuaWeiBillListOption struct {
	AccountID string `json:"account_id" validate:"required"`
	// 查询的资源详单所在账期,东八区时间,格式为YYYY-MM。 示例:2019-01 说明: 不支持2019年1月份之前的资源详单。
	Month string `json:"month" validate:"required"`
	// Limit: 最大值为1000
	Page *HuaWeiBillPage `json:"page" validate:"omitempty"`
}

// Validate huawei bill list option.
func (opt HuaWeiBillListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HuaWeiFeeRecordListOption defines huawei fee record list options
type HuaWeiFeeRecordListOption struct {
	AccountID    string `json:"account_id" validate:"required"`
	SubAccountID string `json:"sub_account_id" validate:"required"`
	// 查询的资源详单所在账期,东八区时间,格式为YYYY-MM。 示例:2019-01 说明: 不支持2019年1月份之前的资源详单。
	Month string `json:"month" validate:"required"`
	// 查询的资源消费记录的开始日期,格式为YYYY-MM-DD
	// 说明: 必须和cycle(即资源的消费账期)在同一个月
	// bill_date_begin和bill_date_end两个参数必须同时出现,否则仅按照cycle(即资源的消费账期)进行查询。
	BillDateBegin string `json:"bill_date_begin" validate:"required"`
	BillDateEnd   string `json:"bill_date_end" validate:"required"`
	// Limit: 最大值为1000
	Page *HuaWeiBillPage `json:"page" validate:"omitempty"`
}

// Validate huawei huawei fee record list option.
func (opt HuaWeiFeeRecordListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HuaWeiBillPage define huawei bill page option.
type HuaWeiBillPage struct {
	Limit  *int32 `json:"limit,omitempty"`
	Offset *int32 `json:"offset,omitempty"`
}

const HuaWeiQueryLimit = 1000

// Validate huawei bill offset page extension.
func (h HuaWeiBillPage) Validate() error {
	if h.Limit == nil {
		return nil
	}

	if *h.Limit > HuaWeiQueryLimit {
		return errf.New(errf.InvalidParameter, "huawei.Limit should <= 1000")
	}

	return nil
}

// AzureBillListOption define azure bill list option.
type AzureBillListOption struct {
	// AccountID 账号ID
	AccountID string `json:"account_id" validate:"required"`
	// BeginDate 起始日期，格式为yyyy-mm-dd，不支持跨月查询
	BeginDate string `json:"begin_date" validate:"required"`
	// EndDate 截止日期，格式为yyyy-mm-dd，不支持跨月查询
	EndDate string `json:"end_date" validate:"required"`
	// Page 分页信息
	Page *AzureBillPage `json:"page" validate:"omitempty"`
}

const AzureQueryLimit = 1000

// Validate azure bill list option.
func (opt AzureBillListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
		if opt.Page.Limit > AzureQueryLimit {
			return errf.New(errf.InvalidParameter, "azure.page.Limit should be <= 1000")
		}
	}

	return nil
}

// AzureBillPage define azure bill page option.
type AzureBillPage struct {
	// Limit 每页查询数量
	Limit int32 `json:"limit,omitempty"`
	// NextLink 链接 (url) 到结果的下一页
	NextLink string `json:"next_link" validate:"omitempty"`
}

// Validate azure bill offset page extension.
func (a AzureBillPage) Validate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	return nil
}

// GcpBillListOption defines gcp bill list options.
// 时间戳docs: https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types?hl=zh-cn#timestamp_type
type GcpBillListOption struct {
	// BillAccountID bill账号ID
	BillAccountID string `json:"bill_account_id" validate:"required"`
	AccountID     string `json:"account_id" validate:"required"`
	// 包含费用专列项的账单的年份和月份，格式为YYYYMM 示例:201901，可以使用此字段获取账单上的总费用
	Month string `json:"month" validate:"omitempty"`
	// 起始时间戳，时间戳值表示绝对时间点，与任何时区或惯例（如夏令时）无关，可精确到微秒，
	// 格式：0001-01-01 00:00:00 至 9999-12-31 23:59:59.999999（世界协调时间 (UTC)）
	// 也可以使用UTC格式：2014-09-27T12:30:00.45Z
	BeginDate string `json:"begin_date" validate:"omitempty"`
	// 截止时间戳，时间戳值表示绝对时间点，与任何时区或惯例（如夏令时）无关，可精确到微秒
	EndDate string       `json:"end_date" validate:"omitempty"`
	Page    *GcpBillPage `json:"page" validate:"omitempty"`
	// ProjectID 账号所属的项目ID
	ProjectID string `json:"project_id" validate:"required"`
}

// Validate gcp bill list option.
func (opt GcpBillListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if opt.Month == "" && opt.BeginDate == "" && opt.EndDate == "" {
		return errf.New(errf.InvalidParameter, "month and begin_date and end_date can not be empty")
	}

	// gcp的BigQuery属于sql查询，跟SDK接口的Limit限制不同
	if opt.Page != nil {
		if opt.Page.Limit == 0 {
			return errf.New(errf.InvalidParameter, "page.limit is required")
		}
		if opt.Page.Limit > 100000 {
			return errf.New(errf.InvalidParameter, "page.limit should <= 100000")
		}
	}

	return nil
}

// GcpBillPage defines gcp bill list page.
type GcpBillPage struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

// Validate validate gcp bill list page.
func (opt GcpBillPage) Validate() error {
	if opt.Limit == 0 {
		return errf.New(errf.InvalidParameter, "limit is required")
	}

	if opt.Limit > core.GcpQueryLimit {
		return errf.New(errf.InvalidParameter, "gcp.limit should <= 1000")
	}

	return nil
}
