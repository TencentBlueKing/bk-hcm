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

package gcp

import (
	"fmt"
	"strings"
	"time"

	typesBill "hcm/pkg/adaptor/types/bill"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/math"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	// QueryBillFields 需要查询的云账单字段
	QueryBillFields = "billing_account_id," +
		"service.id as service_id," +
		"service.description as service_description," +
		"sku.id as sku_id," +
		"sku.description as sku_description," +
		"usage_start_time," +
		"usage_end_time," +
		"project.id as project_id," +
		"project.name as project_name," +
		"project.number as project_number," +
		"IFNULL(location.location,'') as location," +
		"IFNULL(location.country,'') as country," +
		"IFNULL(location.region,'') as region," +
		"IFNULL(location.zone,'') as zone," +
		"resource.name as resource_name," +
		"resource.global_name as resource_global_name," +
		"(CAST(cost * 1000000 AS int64)) / 1000000 as cost," +
		"currency," +
		"IFNULL(usage.amount,0) AS usage_amount," +
		"IFNULL(usage.unit, '') AS usage_unit," +
		"usage.amount_in_pricing_units as usage_amount_in_pricing_units," +
		"usage.pricing_unit as usage_pricing_unit," +
		"export_time," +
		"cost+IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0) AS total_cost," +
		"invoice.month as month," +
		"cost_type," +
		"ARRAY_TO_STRING(ARRAY(SELECT CONCAT(name, ':', CAST(amount AS STRING)) AS credit FROM UNNEST(credits)), ',') AS credits_amount," +
		"IFNULL((SELECT sum(CAST(amount*1000000 AS int64)) AS credit FROM UNNEST(credits)),0)/1000000 as return_cost," +
		"currency_conversion_rate"
	// QueryBillSQL 查询云账单的SQL
	QueryBillSQL = "SELECT %s FROM %s.%s %s"
	// QueryBillTotalSQL 查询云账单总数量的SQL
	QueryBillTotalSQL = "SELECT COUNT(*) FROM %s.%s %s"
)

// GetBillList demonstrates issuing a query and reading results.
func (g *Gcp) GetBillList(kt *kit.Kit, opt *typesBill.GcpBillListOption,
	billInfo *cloud.AccountBillConfig[cloud.GcpBillConfigExtension]) (interface{}, int64, error) {

	where, err := g.parseCondition(opt)
	if err != nil {
		logs.Errorf("gcp get bill list parse date failed, opt: %+v, err: %v", opt, err)
		return nil, 0, err
	}

	// 只有第一页时返回数量，降低查询费用
	total := int64(0)
	if opt.Page != nil && opt.Page.Offset == 0 {
		total, err = g.GetBillTotal(kt, where, billInfo)
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, 0, nil
		}
	}

	query := fmt.Sprintf(QueryBillSQL, QueryBillFields, billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	if opt.Page != nil {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", opt.Page.Limit, opt.Page.Offset)
	}

	list, _, err := g.GetBigQuery(kt, query)
	return list, total, err
}

// GetBillTotal get bill total num
func (g *Gcp) GetBillTotal(kt *kit.Kit, where string, billInfo *cloud.AccountBillConfig[cloud.GcpBillConfigExtension]) (
	int64, error) {

	sql := fmt.Sprintf(QueryBillTotalSQL, billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	_, total, err := g.GetBigQuery(kt, sql)
	if err != nil {
		return 0, err
	}
	return total, nil
}

const (
	// RootAccountQueryBillFields 需要查询的云账单字段
	RootAccountQueryBillFields = "ANY_VALUE(billing_account_id) as billing_account_id," +
		"ANY_VALUE(service.id) as service_id," +
		"ANY_VALUE(service.description) as service_description," +
		"sku.id as sku_id," +
		"ANY_VALUE(sku.description) as sku_description," +
		"project.id as project_id," +
		"ANY_VALUE(project.name) as project_name," +
		"ANY_VALUE(project.number) as project_number," +
		"ANY_VALUE(IFNULL(location.location,'')) as location," +
		"ANY_VALUE(IFNULL(location.country,'')) as country," +
		"ANY_VALUE(IFNULL(location.region,'')) as region," +
		"ANY_VALUE(IFNULL(location.zone,'')) as zone," +
		"ANY_VALUE(resource.name) as resource_name," +
		"ANY_VALUE(resource.global_name) as resource_global_name," +
		"(CAST(SUM(cost) * 1000000 AS int64)) / 1000000 as cost," +
		"ANY_VALUE(currency) as currency," +
		"SUM(IFNULL(usage.amount,0)) AS usage_amount," +
		"ANY_VALUE(IFNULL(usage.unit, '')) AS usage_unit," +
		"SUM(usage.amount_in_pricing_units) as usage_amount_in_pricing_units," +
		"ANY_VALUE(usage.pricing_unit) as usage_pricing_unit," +
		// 非promotion费用直接加到消耗金额中
		"SUM(cost)+SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c where c.type != 'PROMOTION'), 0)) AS total_cost," +
		"ANY_VALUE(invoice.month) as month," +
		"ANY_VALUE(cost_type) as cost_type," +
		"SUM(IFNULL((SELECT sum(CAST(amount*1000000 AS int64)) AS credit FROM UNNEST(credits)),0)/1000000) as return_cost," +
		// 将promotion类型保存起来
		`ARRAY_CONCAT_AGG(ARRAY(select AS STRUCT * from  UNNEST(credits) where type='PROMOTION')) as  credit_infos,` +
		"ANY_VALUE(currency_conversion_rate) as currency_conversion_rate"
	// RootAccountQueryBillSQL 查询云账单的SQL
	RootAccountQueryBillSQL = "SELECT %s FROM %s.%s %s GROUP BY sku.id, project.id"
	// RootAccountQueryBillTotalSQL 查询云账单总数量的SQL
	RootAccountQueryBillTotalSQL = "SELECT COUNT(*) FROM (SELECT DISTINCT sku.id, project.id FROM %s.%s %s " +
		"GROUP BY sku.id, project.id)"

	// RootCreditQuerySQL credit 查询sql
	RootCreditQuerySQL = `SELECT * EXCEPT(credits_agg),
ARRAY((
SELECT AS STRUCT id AS id, name, type, full_name, CAST( SUM (amount) AS string) AS amount
FROM UNNEST(credits_agg)
GROUP BY  id, name, type, full_name)) AS credits
FROM (
	SELECT
		ANY_VALUE(billing_account_id) AS billing_account_id,
		project.id AS project_id,
		ANY_VALUE(project.name) as project_name,
		ANY_VALUE(project.number) as project_number,
		ANY_VALUE(currency) as currency,
		ANY_VALUE(invoice.month) as month,
		ANY_VALUE(currency_conversion_rate) as currency_conversion_rate,
		CAST(SUM(IFNULL((
			SELECT SUM(CAST(amount AS DECIMAL)) AS credit
			FROM  UNNEST(credits)
			WHERE type = 'PROMOTION'),0)) AS STRING)  AS promotion_credit,
		ARRAY_CONCAT_AGG(ARRAY(
			SELECT AS STRUCT id,type,name,full_name,SUM(CAST(amount AS DECIMAL)) AS amount
			FROM UNNEST(credits)
			WHERE type='PROMOTION'
			GROUP BY id,type,name,full_name )) AS credits_agg,
	FROM
		%s.%s
		%s  
	GROUP BY project.id )
`
)

// GetRootAccountBillTotal get bill total num
func (g *Gcp) GetRootAccountBillTotal(
	kt *kit.Kit, where string, billInfo *billcore.RootAccountBillConfig[billcore.GcpBillConfigExtension]) (
	int64, error) {

	sql := fmt.Sprintf(RootAccountQueryBillTotalSQL, billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	_, total, err := g.GetBigQuery(kt, sql)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// QueryRootCreditList query credits list.
func (g *Gcp) QueryRootCreditList(kt *kit.Kit, opt *typesBill.GcpRootAccountBillListOption,
	billInfo *billcore.RootAccountBillConfig[billcore.GcpBillConfigExtension]) (interface{}, error) {

	conditionOpt := &typesBill.GcpBillListOption{
		BillAccountID: opt.RootAccountID,
		AccountID:     opt.MainAccountID,
		Month:         opt.Month,
		BeginDate:     opt.BeginDate,
		EndDate:       opt.EndDate,
		Page:          opt.Page,
		ProjectID:     opt.ProjectID,
	}
	where, err := g.parseRootCreditCondition(conditionOpt)
	if err != nil {
		logs.Errorf("gcp query bill credit list parse date failed, opt: %+v, err: %v", opt, err)
		return nil, err
	}

	query := fmt.Sprintf(RootCreditQuerySQL, billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	if opt.Page != nil {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", opt.Page.Limit, opt.Page.Offset)
	}

	list, _, err := g.GetBigQuery(kt, query)
	return list, err
}

// GetRootAccountBillList demonstrates issuing a query and reading results.
func (g *Gcp) GetRootAccountBillList(kt *kit.Kit, opt *typesBill.GcpRootAccountBillListOption,
	billInfo *billcore.RootAccountBillConfig[billcore.GcpBillConfigExtension]) (interface{}, int64, error) {

	conditionOpt := &typesBill.GcpBillListOption{
		BillAccountID: opt.RootAccountID,
		AccountID:     opt.MainAccountID,
		Month:         opt.Month,
		BeginDate:     opt.BeginDate,
		EndDate:       opt.EndDate,
		Page:          opt.Page,
		ProjectID:     opt.ProjectID,
	}
	where, err := g.parseRootAccountCondition(conditionOpt)
	if err != nil {
		logs.Errorf("gcp get bill list parse date failed, opt: %+v, err: %v", opt, err)
		return nil, 0, err
	}

	// 只有第一页时返回数量，降低查询费用
	total := int64(0)
	if opt.Page != nil && opt.Page.Offset == 0 {
		total, err = g.GetRootAccountBillTotal(kt, where, billInfo)
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, 0, nil
		}
	}

	query := fmt.Sprintf(RootAccountQueryBillSQL, RootAccountQueryBillFields,
		billInfo.CloudDatabaseName, billInfo.CloudTableName, where)
	if opt.Page != nil {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", opt.Page.Limit, opt.Page.Offset)
	}

	list, _, err := g.GetBigQuery(kt, query)
	return list, total, err
}

// GetBigQuery ...
func (g *Gcp) GetBigQuery(kt *kit.Kit, query string) ([]map[string]bigquery.Value, int64, error) {
	client, err := g.clientSet.bigQueryClient(kt)
	if err != nil {
		return nil, 0, fmt.Errorf("gcp.billquery.NewClient, err: %+v", err)
	}

	q := client.Query(query)
	it, err := q.Read(kt.Ctx)
	if err != nil {
		return nil, 0, err
	}

	var list []map[string]bigquery.Value
	var num int64
	for {
		var row map[string]bigquery.Value
		err = it.Next(&row)
		if err == iterator.Done {
			break
		}
		// 将第一个值转换为 int64 类型
		if intValue, ok := row["f0_"].(int64); ok {
			num = intValue
		}
		if err != nil {
			logs.Errorf("gcp get big query next failed, query: %s, err: %+v", query, err)
			return nil, 0, err
		}
		if totalCostStr, ok := row["total_cost"].(string); ok {
			if strings.IndexAny(totalCostStr, "Ee") != -1 {
				decimalNum, err := math.NewDecimalFromString(totalCostStr)
				if err == nil {
					row["total_cost"] = decimalNum.ToString()
				}
			}
		}

		list = append(list, row)
	}

	return list, num, nil
}

func (g *Gcp) parseCondition(opt *typesBill.GcpBillListOption) (string, error) {
	var condition []string
	if len(opt.ProjectID) != 0 {
		condition = []string{fmt.Sprintf("project.id = '%s'", opt.ProjectID)}
	}
	if opt.Month != "" {
		condition = append(condition, fmt.Sprintf("invoice.month = '%s'", opt.Month))
	} else if opt.BeginDate != "" && opt.EndDate != "" {
		beginDate, err := time.Parse(constant.TimeStdFormat, opt.BeginDate)
		if err != nil {
			return "", err
		}

		endDate, err := time.Parse(constant.TimeStdFormat, opt.EndDate)
		if err != nil {
			return "", err
		}
		condition = append(condition, fmt.Sprintf("TIMESTAMP_TRUNC(_PARTITIONTIME, DAY) BETWEEN TIMESTAMP(\"%s\") AND "+
			"TIMESTAMP(\"%s\")", beginDate.Format(constant.DateLayout), endDate.Format(constant.DateLayout)))
	}

	if len(condition) > 0 {
		return "WHERE " + strings.Join(condition, " AND "), nil
	}

	return "", nil
}

func (g *Gcp) parseRootAccountCondition(opt *typesBill.GcpBillListOption) (string, error) {
	var condition []string
	if len(opt.ProjectID) != 0 {
		condition = []string{fmt.Sprintf("project.id = '%s'", opt.ProjectID)}
	} else {
		condition = []string{"project.id IS NULL"}
	}

	if opt.Month != "" {
		condition = append(condition, fmt.Sprintf("invoice.month = '%s'", opt.Month))
	}
	if opt.BeginDate != "" && opt.EndDate != "" {
		beginDate, err := time.Parse(constant.TimeStdFormat, opt.BeginDate)
		if err != nil {
			return "", err
		}

		endDate, err := time.Parse(constant.TimeStdFormat, opt.EndDate)
		if err != nil {
			return "", err
		}
		condition = append(condition, fmt.Sprintf("TIMESTAMP_TRUNC(_PARTITIONTIME, DAY) BETWEEN TIMESTAMP(\"%s\") AND "+
			"TIMESTAMP(\"%s\")", beginDate.Format(constant.DateLayout), endDate.Format(constant.DateLayout)))
	}

	if len(condition) > 0 {
		return "WHERE " + strings.Join(condition, " AND "), nil
	}

	return "", nil
}
func (g *Gcp) parseRootCreditCondition(opt *typesBill.GcpBillListOption) (string, error) {
	var condition []string
	if len(opt.ProjectID) != 0 {
		if opt.ProjectID == "NULL" {
			condition = []string{"project.id IS NULL"}
		} else {
			condition = []string{fmt.Sprintf("project.id = '%s'", opt.ProjectID)}
		}
	}
	// 不传project id 会返回所有project的信息包括project 为NULL

	if opt.Month != "" {
		condition = append(condition, fmt.Sprintf("invoice.month = '%s'", opt.Month))
	}
	if opt.BeginDate != "" && opt.EndDate != "" {
		beginDate, err := time.Parse(constant.TimeStdFormat, opt.BeginDate)
		if err != nil {
			return "", err
		}

		endDate, err := time.Parse(constant.TimeStdFormat, opt.EndDate)
		if err != nil {
			return "", err
		}
		condition = append(condition, fmt.Sprintf("TIMESTAMP_TRUNC(_PARTITIONTIME, DAY) BETWEEN TIMESTAMP(\"%s\") AND "+
			"TIMESTAMP(\"%s\")", beginDate.Format(constant.DateLayout), endDate.Format(constant.DateLayout)))
	}

	if len(condition) > 0 {
		return "WHERE " + strings.Join(condition, " AND "), nil
	}

	return "", nil
}
