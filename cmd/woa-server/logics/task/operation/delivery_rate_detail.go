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
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// GetDeliveryRateDetail aggregates delivery rate detail by biz and month within a range
func (op *operation) GetDeliveryRateDetail(kt *kit.Kit, param *types.DeliveryRateDetailReq) (
	*types.DeliveryRateDetailResp, error) {

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

	pipeline := op.buildDeliveryRateDetailPipeline(start, end)

	rst := make([]types.DeliveryRateDetailItem, 0)
	if err := model.Operation().ApplyOrder().AggregateAll(kt.Ctx, pipeline, &rst); err != nil {
		logs.Errorf("aggregate delivery rate detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.DeliveryRateDetailResp{Details: rst}, nil
}

// buildDeliveryRateDetailPipeline builds the aggregation pipeline for delivery rate detail
func (op *operation) buildDeliveryRateDetailPipeline(start, end time.Time) []map[string]interface{} {
	return []map[string]interface{}{
		// 第一步：过滤时间范围
		{pkg.BKDBMatch: map[string]interface{}{
			"create_at": map[string]interface{}{
				pkg.BKDBGTE: start,
				pkg.BKDBLTE: end,
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
		// 第三步：按业务和月份分组统计
		{pkg.BKDBGroup: map[string]interface{}{
			"_id": map[string]interface{}{
				"bk_biz_id":  "$bk_biz_id",
				"year_month": "$year_month",
			},
			// 统计订单数（只统计stage=DONE且status=DONE的单据）
			"total_orders": map[string]interface{}{pkg.BKDBSum: 1},
			// 已完成单据数（stage=DONE）
			"done_orders": map[string]interface{}{
				"$sum": map[string]interface{}{
					"$cond": []interface{}{
						map[string]interface{}{pkg.BKDBEQ: []interface{}{"$stage", types.TicketStageDone}},
						1,
						0,
					},
				},
			},
			// 需求总数（所有申请单的total_num之和）
			"total_num_sum": map[string]interface{}{pkg.BKDBSum: "$total_num"},
			// 成功交付数（所有申请单的success_num之和）
			"success_num_sum": map[string]interface{}{pkg.BKDBSum: "$success_num"},
		}},
		// 第四步：计算交付率
		{pkg.BKDBAddFields: map[string]interface{}{
			"host_delivery_rate": map[string]interface{}{
				"$cond": []interface{}{
					map[string]interface{}{pkg.BKDBEQ: []interface{}{"$total_num_sum", 0}},
					0.0,
					map[string]interface{}{
						"$multiply": []interface{}{
							map[string]interface{}{
								pkg.BKDBDivide: []interface{}{"$success_num_sum", "$total_num_sum"},
							},
							100,
						},
					},
				},
			},
		}},
		// 第五步：格式化输出
		{pkg.BKDBProject: map[string]interface{}{
			"_id":             0,
			"bk_biz_id":       "$_id.bk_biz_id",
			"year_month":      "$_id.year_month",
			"total_orders":    1,
			"done_orders":     1,
			"total_num_sum":   1,
			"success_num_sum": 1,
			"host_delivery_rate": map[string]interface{}{
				pkg.BKDBRound: []interface{}{"$host_delivery_rate", 2},
			},
		}},
		// 第六步：排序（按主机交付率降序，交付率相同则按业务ID和月份升序）
		{pkg.BKDBSort: map[string]interface{}{
			"host_delivery_rate": pkg.BKDBDesc,
			"bk_biz_id":          pkg.BKDBAsc,
			"year_month":         pkg.BKDBAsc,
		}},
	}
}
