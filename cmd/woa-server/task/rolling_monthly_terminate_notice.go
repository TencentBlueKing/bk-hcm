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

package task

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/logics/biz"
	"hcm/cmd/woa-server/logics/config"
	model "hcm/cmd/woa-server/model/task"
	configtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	croncore "hcm/pkg/cron/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/slice"
)

// RollingMonthlyTerminateNoticeTask is the task for rolling monthly terminate notice.
type RollingMonthlyTerminateNoticeTask struct {
	clientSet    *client.ClientSet
	bizLogic     biz.Logics
	cmsiClient   cmsi.Client
	configLogics config.Logics
}

// Name return the name of the task.Name
func (r *RollingMonthlyTerminateNoticeTask) Name() string {
	return string(enumor.CronTaskRollingMonthlyTerminateNotice)
}

// Next return the next time to run the task. The task runs at 9:00 on the first day of each month.
func (r *RollingMonthlyTerminateNoticeTask) Next() (time.Time, error) {
	now := time.Now()
	loc, err := time.LoadLocation(cc.WoaServer().LocalTimezone)
	if err != nil {
		logs.Errorf("load timezone failed: %v, use UTC", err)
		loc = time.UTC
	}
	// 计算下一个月1日的9:00
	nextRun := time.Date(now.Year(), now.Month(), 1, 9, 0, 0, 0, loc)
	if !nextRun.After(now) {
		nextRun = nextRun.AddDate(0, 1, 0)
	}
	return nextRun, nil
}

// NewRollingMonthlyTerminateNoticeTask create a new rolling monthly terminate notice task.
func NewRollingMonthlyTerminateNoticeTask(clientSet *client.ClientSet, bizLogic biz.Logics, cmsiClient cmsi.Client,
	configLogics config.Logics) (croncore.Task, error) {
	return &RollingMonthlyTerminateNoticeTask{
		clientSet:    clientSet,
		bizLogic:     bizLogic,
		cmsiClient:   cmsiClient,
		configLogics: configLogics,
	}, nil
}

// Do execute the task with default last month.
func (r *RollingMonthlyTerminateNoticeTask) Do(kt *kit.Kit) error {
	loc, err := time.LoadLocation(cc.WoaServer().LocalTimezone)
	if err != nil {
		logs.Errorf("load timezone failed: %v, use UTC", err)
		loc = time.UTC
	}
	now := time.Now().In(loc)
	nowMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	startMonth := nowMonth.AddDate(0, -1, 0) // 默认处理上月

	return r.DoWithMonth(kt, startMonth)
}

// DoWithMonth execute the task with specified month.
func (r *RollingMonthlyTerminateNoticeTask) DoWithMonth(kt *kit.Kit, targetMonth time.Time) error {
	loc, err := time.LoadLocation(cc.WoaServer().LocalTimezone)
	if err != nil {
		logs.Errorf("load timezone failed: %v, use UTC", err)
		loc = time.UTC
	}

	targetMonth = time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, loc)
	nextMonth := targetMonth.AddDate(0, 1, 0) // 下月1号

	filter := map[string]interface{}{
		"create_at": mapstr.MapStr{
			pkg.BKDBGTE: targetMonth,
			pkg.BKDBLT:  nextMonth,
		},
		"require_type": enumor.RequireTypeRollServer,
		"stage": mapstr.MapStr{
			pkg.BKDBIN: []types.TicketStage{types.TicketStageAudit, types.TicketStageSuspend,
				types.TicketStageUncommit, types.TicketStageConfirming},
		},
	}
	page := metadata.BasePage{Limit: pkg.BKMaxInstanceLimit, Start: 0}

	userOrders, bizIDMap, err := r.processOrdersBatch(kt, page, filter)
	if err != nil {
		return err
	}

	return r.sendTerminationNotifications(kt, userOrders, bizIDMap, targetMonth)
}

// processOrdersBatch 处理订单批次，终止订单、收集用户订单信息
func (r *RollingMonthlyTerminateNoticeTask) processOrdersBatch(kt *kit.Kit, page metadata.BasePage,
	filter map[string]interface{}) (map[string][]*types.ApplyOrder, map[int64]struct{}, error) {

	// 先查出所有待处理单据
	allOrders := make([]*types.ApplyOrder, 0)
	for {
		orders, err := model.Operation().ApplyOrder().FindManyApplyOrder(kt.Ctx, page, filter)
		if err != nil {
			logs.Errorf("find apply order failed, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}
		allOrders = append(allOrders, orders...)
		if len(orders) < page.Limit {
			break
		}
		page.Start += page.Limit
	}

	// 批量终止，仅将成功终止的单据加入 userOrders，失败批次不发送通知
	userOrders := make(map[string][]*types.ApplyOrder)
	bizIDMap := make(map[int64]struct{})
	for _, batch := range slice.Split(allOrders, pkg.BKMaxInstanceLimit) {
		subOrderIds := make([]string, 0, len(batch))
		for _, order := range batch {
			subOrderIds = append(subOrderIds, order.SubOrderId)
		}
		if err := r.batchTerminateOrders(kt, subOrderIds); err != nil {
			logs.Errorf("batch terminate orders failed, suborders: %v, err: %v, rid: %s", subOrderIds, err,
				kt.Rid)
			continue
		}
		for _, order := range batch {
			userOrders[order.User] = append(userOrders[order.User], order)
			bizIDMap[order.BkBizId] = struct{}{}
		}
	}

	if len(allOrders) > 0 && len(userOrders) == 0 {
		logs.Errorf("all %d orders failed to terminate", len(allOrders))
		return nil, nil, fmt.Errorf("all %d orders failed to terminate", len(allOrders))
	}

	return userOrders, bizIDMap, nil
}

// batchTerminateOrders 批量终止订单
func (r *RollingMonthlyTerminateNoticeTask) batchTerminateOrders(kt *kit.Kit, subOrderIds []string) error {
	update := &mapstr.MapStr{
		"stage":     types.TicketStageTerminate,
		"status":    types.ApplyStatusTerminate,
		"update_at": time.Now(),
	}
	updateFilter := &mapstr.MapStr{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: subOrderIds,
		},
	}

	err := model.Operation().ApplyOrder().UpdateApplyOrder(kt.Ctx, updateFilter, update)
	if err != nil {
		logs.Errorf("batch update apply order terminate failed, suborders: %v, err: %v, rid: %s", subOrderIds,
			err, kt.Rid)
		return err
	}

	return nil
}

// sendTerminationNotifications 发送终止通知邮件
func (r *RollingMonthlyTerminateNoticeTask) sendTerminationNotifications(kt *kit.Kit,
	userOrders map[string][]*types.ApplyOrder, bizIDMap map[int64]struct{}, startMonth time.Time) error {

	bizIDs := maps.Keys(bizIDMap)
	bizNames, err := r.bizLogic.GetBizNames(kt, bizIDs)
	if err != nil {
		logs.Errorf("get biz names failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	var errs []error
	for user, orders := range userOrders {
		if err := r.sendMailsByUserAndBiz(kt, user, orders, bizNames, startMonth); err != nil {
			logs.Errorf("send mails failed for user: %s, err: %v, rid: %s", user, err, kt.Rid)
			errs = append(errs, fmt.Errorf("user %s: %w", user, err))
			continue
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// sendMailsByUserAndBiz 按用户和业务发送邮件，单个业务失败不影响其他业务继续发送
func (r *RollingMonthlyTerminateNoticeTask) sendMailsByUserAndBiz(kt *kit.Kit, user string, orders []*types.ApplyOrder,
	bizNames map[int64]string, startMonth time.Time) error {

	ordersByBiz := make(map[int64][]*types.ApplyOrder)
	for _, order := range orders {
		ordersByBiz[order.BkBizId] = append(ordersByBiz[order.BkBizId], order)
	}
	var errs []error
	for bizID, bizOrders := range ordersByBiz {
		bizName := bizNames[bizID]
		if strings.TrimSpace(bizName) == "" {
			bizName = fmt.Sprintf("%d", bizID)
			logs.Warnf("biz name not found or empty for biz_id: %d, use fallback, rid: %s", bizID, kt.Rid)
		}
		if err := r.sendTerminateMail(kt, user, bizName, bizOrders, startMonth); err != nil {
			logs.Errorf("send terminate mail failed, user: %s, biz: %d, err: %v, rid: %s", user, bizID, err,
				kt.Rid)
			errs = append(errs, fmt.Errorf("biz %d: %w", bizID, err))
			continue
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// sendTerminateMail 发送邮件
func (r *RollingMonthlyTerminateNoticeTask) sendTerminateMail(kt *kit.Kit, user, bizName string,
	orders []*types.ApplyOrder, startMonth time.Time) error {

	if r.cmsiClient == nil {
		return fmt.Errorf("cmsi client is nil")
	}

	regionMap, zoneMap, err := r.getRegionZoneCnMap(kt, orders)
	if err != nil {
		logs.Errorf("get region zone map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	monthOnly := fmt.Sprintf("%d", startMonth.Month())
	yearMonth := fmt.Sprintf("%d年%02d月", startMonth.Year(), startMonth.Month())
	terminateStr := time.Now().Format(constant.DateLayout)
	title := fmt.Sprintf(constant.RsExpiredTerminationTitle, bizName, monthOnly)
	content := buildTerminateMailContent(kt, bizName, monthOnly, yearMonth, terminateStr, orders, regionMap, zoneMap)

	mail := &cmsi.CmsiMail{
		ReceiverUserName: user,
		Title:            title,
		Content:          content,
	}
	return r.cmsiClient.SendMail(kt, mail)
}

// getRegionZoneCnMap 地域和可用区映射
func (r *RollingMonthlyTerminateNoticeTask) getRegionZoneCnMap(kt *kit.Kit, orders []*types.ApplyOrder) (
	map[string]string, map[string]string, error) {

	regionMap := make(map[string]string)
	zoneMap := make(map[string]string)

	regions := make([]string, 0)
	zones := make([]string, 0)

	for _, order := range orders {
		if order.Spec == nil {
			continue
		}
		if order.Spec.Region != "" {
			regions = append(regions, order.Spec.Region)
		}
		if len(order.Spec.Zones) > 0 {
			zones = append(zones, order.Spec.Zones...)
		}
	}

	regions = slice.Unique(regions)
	zones = slice.Unique(zones)
	regionResult, err := r.configLogics.Region().GetRegion(kt)
	if err != nil {
		logs.Errorf("get region config failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	for _, region := range regionResult.Info {
		regionMap[region.Region] = region.RegionCn
	}

	if len(regions) > 0 {
		zoneReq := &configtypes.GetZoneParam{
			Region: regions,
			Zone:   zones,
		}
		zoneResult, err := r.configLogics.Zone().GetZone(kt, zoneReq)
		if err != nil {
			logs.Errorf("get zone config failed, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		for _, zone := range zoneResult.Info {
			zoneMap[zone.Zone] = zone.ZoneCn
		}
	}

	return regionMap, zoneMap, nil
}

// buildTerminateMailContent 构建邮件内容
func buildTerminateMailContent(kt *kit.Kit, bizName, monthOnly, yearMonth, terminateStr string,
	orders []*types.ApplyOrder, regionMap, zoneMap map[string]string) string {

	tableRows := buildTerminateMailTableRows(kt, orders, regionMap, zoneMap)
	bkHcmURL := cc.WoaServer().BkHcmURL

	emailTitle := fmt.Sprintf(constant.RsExpiredTerminationTitle, bizName, monthOnly)
	return fmt.Sprintf(constant.RsExpiredTerminationEmailContentTemplate, emailTitle, bkHcmURL, bkHcmURL, emailTitle,
		bizName, yearMonth, terminateStr, tableRows)
}

// buildTerminateMailTableRows 构建表格内容
func buildTerminateMailTableRows(kt *kit.Kit, orders []*types.ApplyOrder, regionMap, zoneMap map[string]string) string {
	var table strings.Builder
	for _, order := range orders {
		spec := order.Spec
		if spec == nil {
			logs.Warnf("order spec is nil, skip suborder: %s, rid: %s", order.SubOrderId, kt.Rid)
			continue
		}
		region := spec.Region
		if cnRegion, ok := regionMap[region]; ok {
			region = cnRegion
		}

		cnZones := make([]string, 0, len(spec.Zones))
		for _, zone := range spec.Zones {
			if zone == "all" {
				cnZones = append(cnZones, "全部可用区")
			} else if cnZone, ok := zoneMap[zone]; ok {
				cnZones = append(cnZones, cnZone)
			} else {
				cnZones = append(cnZones, zone)
			}
		}
		zonesStr := truncateZonesStr(strings.Join(cnZones, ","), len(cnZones),
			constant.RsExpiredTerminationZonesMaxRunes)
		assign := spec.ResAssign.GetName()

		table.WriteString(fmt.Sprintf(constant.RsExpiredTerminationEmailTableTemplate,
			order.SubOrderId, spec.DeviceType, region, zonesStr, order.TotalNum, assign))
	}
	return table.String()
}

// truncateZonesStr 当可用区名称拼接成的字符串过长时截断，超出部分用「...(共N个可用区)」代替
func truncateZonesStr(zonesCommaSep string, zoneCount int, maxRunes int) string {
	runes := []rune(zonesCommaSep)
	if len(runes) <= maxRunes {
		return zonesCommaSep
	}
	suffix := fmt.Sprintf("...(共%d个可用区)", zoneCount)
	truncRunes := maxRunes - len([]rune(suffix))
	if truncRunes <= 0 {
		return suffix
	}
	return string(runes[:truncRunes]) + suffix
}

// GetURL get the url of the task, require every task to have external api in service.
func (r *RollingMonthlyTerminateNoticeTask) GetURL() string {
	return "/cross_month/termination"
}
