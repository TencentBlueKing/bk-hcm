/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package recycler

import (
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/metadata"
)

// StartFailedNoticeScanLoop 开始定时扫描需要发送失败通知的回收单
func (r *recycler) StartFailedNoticeScanLoop(kt *kit.Kit) {
	cfg := cc.WoaServer().RecycleNotice.Recycle
	if !cfg.Enable {
		logs.Infof("recycle notice scan loop is disabled, rid: %s", kt.Rid)
		return
	}

	logs.Infof("start failed notice scan loop, interval: %v, startUpDelay: %v, rid: %s",
		cfg.ScanInterval, cfg.StartUpDelay, kt.Rid)

	time.Sleep(cfg.StartUpDelay)

	if err := r.scanAndSendFailedNotice(kt); err != nil {
		logs.Errorf("scan and send failed notice (first run) failed, err: %v, rid: %s", err, kt.Rid)
	}

	// 定时扫描
	ticker := time.NewTicker(cfg.ScanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-kt.Ctx.Done():
			logs.Infof("failed notice scan loop exits, rid: %s", kt.Rid)
			return
		case <-ticker.C:
			subkit := kt.NewSubKit()
			if err := r.scanAndSendFailedNotice(subkit); err != nil {
				logs.Errorf("scan and send failed notice failed, err: %v, rid: %s", err, subkit.Rid)
			}
		}
	}
}

// scanAndSendFailedNotice 扫描有通知记录且需要后续通知的主单，发送失败通知
func (r *recycler) scanAndSendFailedNotice(kt *kit.Kit) error {
	// 查询所有处于预检失败状态的子单对应的主单
	filter := map[string]any{
		"status": table.RecycleStatusDetectFailed,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	orders, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("query orders failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("query orders failed: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	orderIDSet := make(map[uint64]struct{})
	for _, order := range orders {
		orderIDSet[order.OrderID] = struct{}{}
	}

	for orderID := range orderIDSet {
		_, err = r.notifier.CheckAndSendDetectFailed(kt, orderID)
		if err != nil {
			logs.Errorf("check and send detect failed, orderID: %d, err: %v, rid: %s", orderID, err, kt.Rid)
			continue
		}
	}

	return nil
}
