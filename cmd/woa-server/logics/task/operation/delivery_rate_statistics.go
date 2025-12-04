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

package operation

import (
	"time"

	model "hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetDeliveryRateStatistics aggregates delivery rate statistics by month within a range
func (op *operation) GetDeliveryRateStatistics(kt *kit.Kit, param *types.DeliveryRateStatisticsReq) (
	[]types.DeliveryRateStatisticsItem, error) {

	start, err := param.GetStartTime()
	if err != nil {
		logs.Errorf("parse start time failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	end, err := param.GetEndTime()
	if err != nil {
		logs.Errorf("parse end time failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	pipeline := op.buildDeliveryRateStatisticsPipeline(start, end)

	rst := make([]types.DeliveryRateStatisticsItem, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate delivery rate statistics failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return rst, nil
}

// buildDeliveryRateStatisticsPipeline builds the aggregation pipeline for delivery rate statistics
func (op *operation) buildDeliveryRateStatisticsPipeline(start, end time.Time) []map[string]interface{} {
	return []map[string]interface{}{
		// 第一步：过滤时间范围，排除采购到资源池的订单
		{pkg.BKDBMatch: map[string]interface{}{
			"create_at": map[string]interface{}{
				pkg.BKDBGTE: start,
				pkg.BKDBLTE: end,
			},
			"source": map[string]interface{}{
				pkg.BKDBNE: enumor.ApplyTicketSrcPurchaseToResPool,
			},
		}},
		// 第二步：提取年月信息
		{pkg.BKDBAddFields: map[string]interface{}{
			"year_month": map[string]interface{}{
				"$dateToString": map[string]interface{}{
					"format": "%Y-%m",
					"date":   "$create_at",
				},
			},
		}},
		// 第三步：按月份分组统计
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": "$year_month",
			// 需求总数（所有申请单的total_num之和）
			"total_sum": map[string]interface{}{pkg.BKDBSum: "$total_num"},
			// 成功交付数（所有申请单的success_num之和）
			"total_success": map[string]interface{}{pkg.BKDBSum: "$success_num"},
		}},
		// 第四步：计算交付率
		{pkg.BKDBProject: map[string]interface{}{
			"_id":        0,
			"year_month": "$_id",
			"delivery_rate": map[string]interface{}{
				pkg.BKDBRound: []interface{}{
					map[string]interface{}{
						"$cond": []interface{}{
							map[string]interface{}{pkg.BKDBEQ: []interface{}{"$total_sum", 0}},
							0.0,
							map[string]interface{}{
								"$multiply": []interface{}{
									map[string]interface{}{
										pkg.BKDBDivide: []interface{}{"$total_success", "$total_sum"},
									},
									100,
								},
							},
						},
					}, 2,
				},
			},
		}},
		// 第五步：排序
		{pkg.BKDBSort: map[string]interface{}{"year_month": 1}},
	}
}
