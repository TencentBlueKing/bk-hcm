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

package plan

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/thirdparty/api-gateway/finops"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

// runConfirmNotice 预测确认通知定时任务主函数
func (c *Controller) runConfirmNotice(ctx context.Context, loc *time.Location) {
	now := time.Now()

	// 上午10:00推送,计算下次推送的时间
	nextRunTime := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, loc)
	if now.After(nextRunTime) {
		nextRunTime = nextRunTime.Add(time.Hour * 24)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 等待到下一个检查时间
		time.Sleep(time.Until(nextRunTime))

		kt := core.NewBackendKit()
		// 判断今天是否为通知日（周一、周三、周五）
		if !isNoticeDay(nextRunTime) {
			logs.Infof("today is not notice day, skip. weekday: %v, rid: %s", nextRunTime.Weekday(), kt.Rid)
			nextRunTime = nextRunTime.Add(time.Hour * 24)
			continue
		}

		// 执行通知
		_, _, err := c.PushResPlanConfirmNotice(kt, []int64{})
		if err != nil {
			logs.Errorf("%s: failed to push confirm notice, err: %v, rid: %s",
				constant.ResPlanConfirmNotificationFailed, err, kt.Rid)
		}

		// 计算下一个检查时间
		nextRunTime = nextRunTime.Add(time.Hour * 24)
	}
}

// isNoticeDay 判断今天是否为通知日（周一、周三、周五）
func isNoticeDay(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Monday || weekday == time.Wednesday || weekday == time.Friday
}

// PushResPlanConfirmNotice 筛选数据并发送资源计划确认通知
func (c *Controller) PushResPlanConfirmNotice(kt *kit.Kit, bkBizIDs []int64) (
	successIDs, failedIDs []int64, err error) {

	start := time.Now()
	logs.Infof("start to push confirm notice, bk_biz_ids: %v, rid: %s", bkBizIDs, kt.Rid)

	// 1. 筛选符合条件的预测数据
	demands, err := c.filterConfirmDemands(kt, bkBizIDs, start)
	if err != nil {
		logs.Errorf("failed to filter confirm demands, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	if len(demands) == 0 {
		logs.Infof("no confirm demands found, rid: %s", kt.Rid)
		return []int64{}, []int64{}, nil
	}

	// 2. 按业务ID聚合
	bkBizDemands := make(map[int64][]*rpd.ResPlanDemandTable)
	for _, demand := range demands {
		bkBizDemands[demand.BkBizID] = append(bkBizDemands[demand.BkBizID], demand)
	}

	// 3. 按业务发送邮件
	successIDs = make([]int64, 0)
	failedIDs = make([]int64, 0)
	for bkBizID, bizDemands := range bkBizDemands {
		err := c.generateAndSendConfirmEmail(kt, bkBizID, bizDemands)
		if err != nil {
			failedIDs = append(failedIDs, bkBizID)
			logs.Errorf("failed to send confirm notice to biz: %d, err: %v, rid: %s",
				bkBizID, err, kt.Rid)
			continue
		}
		successIDs = append(successIDs, bkBizID)
	}

	end := time.Now()
	logs.Infof("end to push confirm notice, success_ids: %v, failed_ids: %v, cost: %fs, rid: %s",
		successIDs, failedIDs, end.Sub(start).Seconds(), kt.Rid)

	return successIDs, failedIDs, nil
}

// filterConfirmDemands 筛选符合条件的预测数据
func (c *Controller) filterConfirmDemands(kt *kit.Kit, bkBizIDs []int64, now time.Time) (
	[]*rpd.ResPlanDemandTable, error) {

	// 排除最近3周内有变动的数据
	noChangeDeadline := now.AddDate(0, 0, -21)
	noChangeDeadlineStr := noChangeDeadline.Format(constant.TimeStdFormat)

	// 根据 weekday 提前计算需要通知的时间范围
	weekday := now.Weekday()
	var notifyTimeStart, notifyTimeEnd int
	switch weekday {
	case time.Monday:
		// 周一：通知第14、15、16周（13-16周后）
		notifyTimeStart = times.ConvTimeToCompactInt(now.AddDate(0, 0, 13*7))
		notifyTimeEnd = times.ConvTimeToCompactInt(now.AddDate(0, 0, 16*7))
	case time.Wednesday, time.Friday:
		// 周三、周五：只通知第14周（13-14周后）
		notifyTimeStart = times.ConvTimeToCompactInt(now.AddDate(0, 0, 13*7))
		notifyTimeEnd = times.ConvTimeToCompactInt(now.AddDate(0, 0, 14*7))
	default:
		// 其他日期不通知
		return []*rpd.ResPlanDemandTable{}, nil
	}

	// 构建查询条件
	rules := []*filter.AtomRule{
		// 条件1：未确认
		tools.RuleEqual("is_confirmed", false),
		// 条件2：到货日期在通知范围内
		tools.RuleGreaterThanEqual("expect_time", notifyTimeStart),
		tools.RuleLessThanEqual("expect_time", notifyTimeEnd),
		// 条件3：更新时间在21天前或更早（排除最近3周有变动的）
		tools.RuleLessThanEqual("updated_at", noChangeDeadlineStr),
	}

	// 业务ID列表
	if len(bkBizIDs) > 0 {
		rules = append(rules, tools.RuleIn("bk_biz_id", bkBizIDs))
	}

	filterExpr := tools.ExpressionAnd(rules...)

	// 查询预测数据
	demandDetails := make([]*rpd.ResPlanDemandTable, 0)
	page := core.NewDefaultBasePage()

	for {
		listReq := &rpproto.ResPlanDemandListReq{
			ListReq: core.ListReq{
				Filter: filterExpr,
				Page:   page,
			},
		}

		rst, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list res plan demand, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		// 所有查询结果都符合条件，直接添加
		for i := range rst.Details {
			demandDetails = append(demandDetails, &rst.Details[i])
		}

		if len(rst.Details) < int(page.Limit) {
			break
		}
		page.Start += uint32(page.Limit)
	}

	return demandDetails, nil
}

// generateAndSendConfirmEmail 生成并发送预测确认邮件
func (c *Controller) generateAndSendConfirmEmail(kt *kit.Kit, bkBizID int64, demands []*rpd.ResPlanDemandTable) error {
	if len(demands) == 0 {
		return nil
	}

	// 1. 生成邮件内容
	emailTitle, emailContent, receivers, ccReceivers, err := c.generateConfirmEmail(kt, bkBizID, demands)
	if err != nil {
		logs.Errorf("failed to generate confirm email, cannot send email, err: %v, bk_biz_id: %d, rid: %s", err,
			bkBizID, kt.Rid)
		return err
	}

	if len(receivers) == 0 {
		logs.Errorf("no receivers found for bk_biz_id: %d, rid: %s", bkBizID, kt.Rid)
		return errors.New("no receivers found")
	}

	// 3. 发送邮件
	err = c.sendConfirmEmail(kt, receivers, ccReceivers, emailTitle, emailContent)
	if err != nil {
		logs.Errorf("failed to send confirm email, err: %v, bk_biz_id: %d, rid: %s", err, bkBizID, kt.Rid)
		return err
	}

	return nil
}

// generateConfirmEmail 生成预测确认邮件内容，并返回收件人和抄送人
func (c *Controller) generateConfirmEmail(kt *kit.Kit, bkBizID int64, demands []*rpd.ResPlanDemandTable) (
	title string, content string, receivers []string, ccReceivers []string, err error) {

	if len(demands) == 0 {
		return "", "", nil, nil, nil
	}

	sendTime := time.Now()
	bizName := demands[0].BkBizName

	// 处理接收人和生成表格内容
	needQueryOpProductIDs, totalCPUCore, tableContent, creators, revisers, err := c.processDemandsForConfirm(
		kt, demands)
	if err != nil {
		logs.Errorf("failed to process demands for confirm, err: %v, rid: %s", err, kt.Rid)
		return "", "", nil, nil, fmt.Errorf("failed to process demands for confirm, err: %w", err)
	}

	// 查询预算提报人
	budgetOperators, err := c.queryBudgetOperators(kt, needQueryOpProductIDs, sendTime)
	if err != nil {
		logs.Errorf("failed to query budget operators, cannot generate email without receivers, err: %v, rid: %s",
			err, kt.Rid)
		return "", "", nil, nil, fmt.Errorf("failed to query budget operators, "+
			"cannot proceed without receivers, err: %w", err)
	}

	receiverSet := make(map[string]struct{})
	for _, creator := range slice.Filter(creators, func(s string) bool { return s != "" }) {
		receiverSet[creator] = struct{}{}
	}
	for _, reviser := range slice.Filter(revisers, func(s string) bool { return s != "" }) {
		receiverSet[reviser] = struct{}{}
	}
	for operator := range budgetOperators {
		if operator != "" {
			receiverSet[operator] = struct{}{}
		}
	}

	leaderQuerySet := make(map[string]struct{})
	for _, creator := range slice.Filter(creators, func(s string) bool { return s != "" }) {
		leaderQuerySet[creator] = struct{}{}
	}
	for operator := range budgetOperators {
		if operator != "" {
			leaderQuerySet[operator] = struct{}{}
		}
	}
	submitters := maps.Keys(leaderQuerySet)

	// 生成邮件内容
	title, content = c.buildConfirmEmailContent(bizName, sendTime, totalCPUCore, tableContent, bkBizID)

	receivers = maps.Keys(receiverSet)
	// 获取抄送人列表
	ccReceivers, err = c.getConfirmCcReceivers(kt, bkBizID, submitters)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to get cc receivers, err: %w", err)
	}

	return title, content, receivers, ccReceivers, nil
}

// processDemandsForConfirm 处理需求列表，收集接收人和需要查询预算提报人的需求
func (c *Controller) processDemandsForConfirm(kt *kit.Kit, demands []*rpd.ResPlanDemandTable) (
	[]int64, int64, string, []string, []string, error) {

	needQueryOpProductIDs := make([]int64, 0)
	var totalCPUCore int64
	tableContent := ""
	creators := make([]string, 0)
	revisers := make([]string, 0)

	for _, demand := range demands {
		// 生成表格行
		cpuCore, os := c.getDemandResourceInfo(demand)
		totalCPUCore += cpuCore
		rowContent, err := c.buildDemandTableRow(kt, demand, cpuCore, os)
		if err != nil {
			return nil, 0, "", nil, nil, fmt.Errorf("failed to build demand table row, err: %w", err)
		}
		tableContent += rowContent

		// reviser为admin时creator必定为admin，所以只需判断hasNormalReviser
		hasNormalReviser := demand.Reviser != "" && demand.Reviser != constant.BackendOperationUserKey
		if hasNormalReviser {
			// 有非admin的更新人，直接使用Reviser，不需要查询预算提报人
			revisers = append(revisers, demand.Reviser)
		}

		if demand.Creator != "" && demand.Creator != constant.BackendOperationUserKey {
			creators = append(creators, demand.Creator)
			continue
		}

		// 收集OpProductID用于查询预算提报人
		if demand.OpProductID > 0 {
			needQueryOpProductIDs = append(needQueryOpProductIDs, demand.OpProductID)
		}
	}

	needQueryOpProductIDs = slice.Unique(needQueryOpProductIDs)
	creators = slice.Unique(creators)
	revisers = slice.Unique(revisers)

	return needQueryOpProductIDs, totalCPUCore, tableContent, creators, revisers, nil
}

// getDemandResourceInfo 获取需求的资源信息
func (c *Controller) getDemandResourceInfo(demand *rpd.ResPlanDemandTable) (cpuCore int64, os int64) {
	cpuCore = converter.PtrToVal(demand.CpuCore)
	os = converter.PtrToVal(demand.OS).IntPart()
	return cpuCore, os
}

// buildDemandTableRow 构建需求表格行
func (c *Controller) buildDemandTableRow(kt *kit.Kit, demand *rpd.ResPlanDemandTable, cpuCore, os int64) (
	string, error) {

	expectTime := times.ConvCompactIntToTime(demand.ExpectTime)
	expectTimeStr := expectTime.Format(constant.DateLayout)
	demandRange, err := c.demandTime.GetDemandDateRangeInMonth(kt, expectTime)
	if err != nil {
		logs.Errorf("failed to get demand date range, err: %v, expect_time: %d, rid: %s", err,
			demand.ExpectTime, kt.Rid)
		return "", fmt.Errorf("failed to get demand date range, err: %w", err)
	}

	return fmt.Sprintf(ptypes.ConfirmEmailTableTemplate, demand.DeviceType, expectTimeStr,
		demandRange.Start, demandRange.End, demand.RegionName, os, cpuCore, demand.ObsProject), nil
}

// queryBudgetOperators 查询预算申报操作人
func (c *Controller) queryBudgetOperators(kt *kit.Kit, opProductIDs []int64, sendTime time.Time) (
	map[string]struct{}, error) {

	if len(opProductIDs) == 0 {
		return make(map[string]struct{}), nil
	}

	// 过滤并去重运营产品ID
	uniqueOpProductIDs := slice.Unique(slice.Filter(opProductIDs, func(id int64) bool {
		return id > 0
	}))
	if len(uniqueOpProductIDs) == 0 {
		return make(map[string]struct{}), nil
	}

	// 查询预算提报人
	currentYear := sendTime.Year()
	operatorsMap := make(map[string]struct{})

	for _, batchOpProductIDs := range slice.Split(uniqueOpProductIDs, finops.MaxOpProductIDsPerRequest) {
		budgetParams := &finops.GetBudgetDeclarationOperatorParam{
			Years:        []int{currentYear},
			OpProductIDs: batchOpProductIDs,
		}
		budgetResult, err := c.finOpsCli.GetBudgetDeclarationOperator(kt, budgetParams)
		if err != nil {
			logs.Errorf("failed to get budget declaration operator, batch_size: %d, err: %v, rid: %s",
				len(batchOpProductIDs), err, kt.Rid)
			return nil, fmt.Errorf("failed to get budget declaration operator, "+
				"cannot proceed without receivers, err: %w", err)
		}

		operatorsMap = mergeBudgetOperators(operatorsMap, budgetResult, currentYear)
	}

	return operatorsMap, nil
}

// mergeBudgetOperators 将预算申报结果合并到提报人映射中
func mergeBudgetOperators(operatorsMap map[string]struct{}, budgetResult *finops.GetBudgetDeclarationOperatorResult,
	currentYear int) map[string]struct{} {

	for _, item := range budgetResult.Items {
		if item.Year != currentYear {
			continue
		}
		for _, comp := range item.Composition {
			// 同时获取提交人列表和创建人列表，并去重
			for _, committer := range slice.Filter(comp.Committers, func(s string) bool { return s != "" }) {
				operatorsMap[committer] = struct{}{}
			}
			for _, creator := range slice.Filter(comp.Creators, func(s string) bool { return s != "" }) {
				operatorsMap[creator] = struct{}{}
			}
		}
	}
	return operatorsMap
}

// buildConfirmEmailContent 构建确认邮件内容
func (c *Controller) buildConfirmEmailContent(bizName string, sendTime time.Time, totalCPUCore int64,
	tableContent string, bkBizID int64) (title, content string) {

	listResPlanURL := fmt.Sprintf(ptypes.ConfirmListURL, c.bkHcmURL, bkBizID)
	title = fmt.Sprintf(ptypes.ConfirmEmailTitleTemplate, bizName, sendTime.Format(constant.DateLayout))
	data := fmt.Sprintf("%d年%d月", sendTime.Year(), int(sendTime.Month()))

	content = fmt.Sprintf(ptypes.ConfirmEmailContentTemplate,
		c.bkHcmURL, c.bkHcmURL, bizName, bizName, data, totalCPUCore, listResPlanURL, tableContent)

	return title, content
}

// getConfirmCcReceivers 汇总抄送人员
func (c *Controller) getConfirmCcReceivers(kt *kit.Kit, bkBizID int64, submitters []string) ([]string, error) {
	ccSet := make(map[string]struct{})

	// 1. 获取业务运维
	if bizMaintainerMap, err := c.bizLogics.GetBkBizMaintainer(kt, []int64{bkBizID}); err != nil {
		logs.Warnf("failed to get bk biz maintainer for cc, err: %v, bk_biz_id: %d, rid: %s", err, bkBizID,
			kt.Rid)
	} else if maintainers, ok := bizMaintainerMap[bkBizID]; ok {
		for _, m := range slice.Filter(maintainers, func(s string) bool { return s != "" }) {
			ccSet[m] = struct{}{}
		}
	}

	// 2. 获取所有提报人的leader
	esbCli := esb.EsbClient()
	if esbCli == nil {
		logs.Errorf("tof client is nil, cannot get leader for creators, bk_biz_id: %d, rid: %s", bkBizID, kt.Rid)
		return nil, fmt.Errorf("tof client is nil, cannot get leader for creators")
	}
	for _, submitter := range slice.Filter(submitters, func(s string) bool { return s != "" }) {
		staffInfo, err := esbCli.Tof().GetStaffInfo(kt, submitter)
		if err != nil {
			logs.Warnf("failed to get staff info for submitter: %s, err: %v, rid: %s", submitter, err, kt.Rid)
			continue
		}
		if staffInfo != nil && staffInfo.LoginName != "" {
			ccSet[staffInfo.LoginName] = struct{}{}
		}
	}

	for _, cc := range slice.Filter(c.resPlanCfg.ConfirmNotice.CcReceivers, func(s string) bool { return s != "" }) {
		ccSet[cc] = struct{}{}
	}

	return maps.Keys(ccSet), nil
}

// sendConfirmEmail 发送预测确认邮件
func (c *Controller) sendConfirmEmail(kt *kit.Kit, receivers, cc []string, emailTitle,
	emailContent string) error {

	// 如果不发送给业务，清空收件人和抄送
	if !c.resPlanCfg.ConfirmNotice.SendToBusiness {
		receivers = make([]string, 0)
		cc = make([]string, 0)
	}

	// 添加默认收件人
	if len(c.resPlanCfg.ConfirmNotice.DefaultReceivers) > 0 {
		receivers = append(receivers, c.resPlanCfg.ConfirmNotice.DefaultReceivers...)
	}

	if len(receivers) == 0 {
		logs.Errorf("no receivers found, rid: %s", kt.Rid)
		return errors.New("no receivers found")
	}

	mail := &cmsi.CmsiMail{
		ReceiverUserName: strings.Join(receivers, ","),
		Title:            emailTitle,
		Content:          emailContent,
		CcUserName:       strings.Join(cc, ","),
	}

	return c.CmsiClient.SendMail(kt, mail)
}

// ConfirmResPlanDemands 确认资源预测需求
func (c *Controller) ConfirmResPlanDemands(kt *kit.Kit, bkBizID int64, demandIDs []string) (
	successIDs, failedIDs []string, err error) {

	if len(demandIDs) == 0 {
		return nil, nil, errors.New("demand_ids cannot be empty")
	}

	logs.Infof("start to confirm demands, bk_biz_id: %d, demand_ids: %v, rid: %s", bkBizID, demandIDs, kt.Rid)

	needUpdate := make([]string, 0)
	failed := make(map[string]struct{}) // 记录未找到或已确认的ID

	// 批量查询并筛选需更新的ID
	for _, batch := range slice.Split(demandIDs, int(filter.DefaultMaxInLimit)) {
		listReq := &rpproto.ResPlanDemandListReq{
			ListReq: core.ListReq{
				Filter: tools.ExpressionAnd(tools.RuleIn("id", batch), tools.RuleEqual("bk_biz_id",
					bkBizID)),
				Page: core.NewDefaultBasePage(),
			},
		}
		rst, err := c.client.DataService().Global.ResourcePlan.ListResPlanDemand(kt, listReq)
		if err != nil {
			logs.Errorf("query demands failed, batch: %v, err: %v, rid: %s", batch, err, kt.Rid)
			return nil, demandIDs, err
		}

		// 记录本批次已查询到的ID
		found := make(map[string]bool)
		for _, d := range rst.Details {
			found[d.ID] = true
			if !d.IsConfirmed {
				needUpdate = append(needUpdate, d.ID)
			} else {
				failed[d.ID] = struct{}{} // 已确认的加入失败
			}
		}

		// 未找到的ID加入失败
		for _, id := range batch {
			if !found[id] {
				failed[id] = struct{}{}
			}
		}
	}

	// 批量更新
	if len(needUpdate) > 0 {
		for _, batch := range slice.Split(needUpdate, constant.BatchOperationMaxLimit) {
			updateReqs := make([]rpproto.ResPlanDemandUpdateReq, len(batch))
			for i, id := range batch {
				updateReqs[i] = rpproto.ResPlanDemandUpdateReq{ID: id, IsConfirmed: true}
			}
			if err = c.client.DataService().Global.ResourcePlan.BatchUpdateResPlanDemand(kt,
				&rpproto.ResPlanDemandBatchUpdateReq{Demands: updateReqs}); err != nil {
				logs.Errorf("batch update failed, batch: %v, err: %v, rid: %s", batch, err, kt.Rid)
				return nil, demandIDs, err
			}
		}
	}

	successIDs = needUpdate
	failedIDs = maps.Keys(failed)

	logs.Infof("end to confirm demands, success_ids: %v, failed_ids: %v, rid: %s", successIDs, failedIDs, kt.Rid)

	return successIDs, failedIDs, nil
}
