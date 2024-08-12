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

	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// AwsBillListReq define aws bill list req.
type AwsBillListReq struct {
	AccountID string `json:"account_id" validate:"required"`
	// 起始日期，格式为yyyy-mm-dd，不支持跨月查询
	BeginDate string `json:"begin_date" validate:"required"`
	// 截止日期，格式为yyyy-mm-dd，不支持跨月查询
	EndDate string           `json:"end_date" validate:"required"`
	Page    *AwsBillListPage `json:"page" validate:"omitempty"`
}

// Validate aws bill list req.
func (opt AwsBillListReq) Validate() error {
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

// AwsBillListPage defines aws bill list page.
type AwsBillListPage struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

// Validate validate aws bill list page.
func (opt AwsBillListPage) Validate() error {
	if opt.Limit == 0 {
		return errf.New(errf.InvalidParameter, "limit is required")
	}

	if opt.Limit > core.AwsQueryLimit {
		return errf.New(errf.InvalidParameter, "aws.limit should <= 1000")
	}

	return nil
}

// -------------------------- BillPipeline --------------------------

// BillPipelineReq define bill pipeline request.
type BillPipelineReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate bill pipeline option.
func (opt BillPipelineReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- BocketPolicy --------------------------

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

// -------------------------- DeleteStack --------------------------

// AwsDeleteStackReq define aws delete stack request.
type AwsDeleteStackReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate aws delete stack option.
func (opt AwsDeleteStackReq) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	return nil
}

// -------------------------- List --------------------------

// TCloudBillListReq define tcloud bill list req.
type TCloudBillListReq struct {
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

// Validate tcloud bill list req.
func (opt TCloudBillListReq) Validate() error {
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

// HuaWeiBillListReq defines huawei bill list req.
type HuaWeiBillListReq struct {
	AccountID string `json:"account_id" validate:"required"`
	// 查询的资源详单所在账期,东八区时间,格式为YYYY-MM。 示例:2019-01 说明: 不支持2019年1月份之前的资源详单。
	Month string `json:"month" validate:"required"`
	// Limit: 最大值为1000
	Page *typesBill.HuaWeiBillPage `json:"page" validate:"omitempty"`
}

// Validate huawei bill list req.
func (opt HuaWeiBillListReq) Validate() error {
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

// HuaWeiFeeRecordListReq defines huawei fee record list request.
type HuaWeiFeeRecordListReq struct {
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
	Page *typesBill.HuaWeiBillPage `json:"page" validate:"omitempty"`
}

// Validate huawei fee record list list req.
func (opt HuaWeiFeeRecordListReq) Validate() error {
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

// AzureBillListReq define azure bill list req.
type AzureBillListReq struct {
	AccountID string `json:"account_id" validate:"required"`
	// 起始日期，格式为yyyy-mm-dd，不支持跨月查询
	BeginDate string `json:"begin_date" validate:"required"`
	// 截止日期，格式为yyyy-mm-dd，不支持跨月查询
	EndDate string                   `json:"end_date" validate:"required"`
	Page    *typesBill.AzureBillPage `json:"page" validate:"omitempty"`
}

// Validate azure bill list req.
func (opt AzureBillListReq) Validate() error {
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

// GcpBillListReq defines gcp bill list req.
type GcpBillListReq struct {
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
	EndDate string                 `json:"end_date" validate:"omitempty"`
	Page    *typesBill.GcpBillPage `json:"page" validate:"omitempty"`
}

// Validate gcp bill list req.
func (opt GcpBillListReq) Validate() error {
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

// GcpRootAccountBillListReq defines gcp bill list req.
type GcpRootAccountBillListReq struct {
	// RootAccountID root account id
	RootAccountID string `json:"root_account_id" validate:"required"`
	// MainAccountID main account id
	MainAccountID string `json:"main_account_id" validate:"omitempty"`
	// 包含费用专列项的账单的年份和月份，格式为YYYYMM 示例:201901，可以使用此字段获取账单上的总费用
	Month string `json:"month" validate:"omitempty"`
	// 起始时间戳，时间戳值表示绝对时间点，与任何时区或惯例（如夏令时）无关，可精确到微秒，
	// 格式：0001-01-01 00:00:00 至 9999-12-31 23:59:59.999999（世界协调时间 (UTC)）
	// 也可以使用UTC格式：2014-09-27T12:30:00.45Z
	BeginDate string `json:"begin_date" validate:"omitempty"`
	// ProjectID 项目ID
	ProjectID string `json:"project_id" validate:"omitempty"`
	// 截止时间戳，时间戳值表示绝对时间点，与任何时区或惯例（如夏令时）无关，可精确到微秒
	EndDate string                 `json:"end_date" validate:"omitempty"`
	Page    *typesBill.GcpBillPage `json:"page" validate:"omitempty"`
}

// Validate gcp bill list req.
func (opt GcpRootAccountBillListReq) Validate() error {
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

// AwsRootBillListReq defines aws bill record list request.
type AwsRootBillListReq struct {
	// 本地主账号
	RootAccountID string `json:"root_account_id" validate:"required"`
	// 云上子账号id
	MainAccountCloudID string `json:"main_account_cloud_id" validate:"required"`

	// 起始日期，格式为yyyy-mm-dd，不支持跨月查询
	BeginDate string `json:"begin_date" validate:"required"`
	// 截止日期，格式为yyyy-mm-dd，不支持跨月查询
	EndDate string           `json:"end_date" validate:"required"`
	Page    *AwsBillListPage `json:"page" validate:"omitempty"`
}

// Validate ...
func (r *AwsRootBillListReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureRootBillListReq azure root account bill list
type AzureRootBillListReq struct {
	// 本地主账号
	RootAccountID string `json:"root_account_id" validate:"required"`
	// 云上订阅id
	SubscriptionID string `json:"subscription_id" validate:"required"`

	// 起始日期，格式为yyyy-mm-dd，不支持跨月查询
	BeginDate string `json:"begin_date" validate:"required"`
	// 截止日期，格式为yyyy-mm-dd，不支持跨月查询
	EndDate string                   `json:"end_date" validate:"required"`
	Page    *typesBill.AzureBillPage `json:"page" validate:"omitempty"`
}

// Validate ...
func (r *AzureRootBillListReq) Validate() error {
	return validator.Validate.Struct(r)
}
