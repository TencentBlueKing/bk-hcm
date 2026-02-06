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

// Package notifier handles recycle order notifications
package notifier

import (
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/pkg"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/tools/metadata"
)

// Notifier handles recycle order notifications
type Notifier struct {
	cmsiClient cmsi.Client
	thirdCli   *thirdparty.Client
}

// New creates a new Notifier instance
func New(cmsiClient cmsi.Client, thirdCli *thirdparty.Client) (*Notifier, error) {
	return &Notifier{
		cmsiClient: cmsiClient,
		thirdCli:   thirdCli,
	}, nil
}

// SendSuccessNotice 检查子单是否全部完成，则发送回收成功通知
func (r *Notifier) SendSuccessNotice(kt *kit.Kit, orderID uint64) (sent bool, err error) {
	if !cc.WoaServer().RecycleNotice.Recycle.Enable {
		logs.Infof("recycle notice is disabled, skip send success notice, orderID: %d, rid: %s", orderID, kt.Rid)
		return false, nil
	}
	// 查询该主单下所有子单
	recycleSubOrder, err := r.findManyRecycleOrder(kt, orderID)
	if err != nil {
		return false, err
	}

	if len(recycleSubOrder) == 0 {
		return false, nil
	}

	// 检查是否所有子单都完成
	for _, sub := range recycleSubOrder {
		if sub.Status != table.RecycleStatusDone {
			return false, nil
		}
	}

	// 发送通知
	if err = r.sendRecycleSuccessNotification(kt, recycleSubOrder); err != nil {
		logs.Errorf("send notification failed, orderID: %d, err: %v, rid: %s", orderID, err, kt.Rid)
		return false, err
	}

	return true, nil
}

// SendDetectFailedNotice 发送预检失败通知
func (r *Notifier) SendDetectFailedNotice(kt *kit.Kit, orderID uint64) (sent bool, err error) {
	return r.checkAndSendDetectFailed(kt, orderID, true)
}

// CheckAndSendDetectFailed 检查并通知预检失败
func (r *Notifier) CheckAndSendDetectFailed(kt *kit.Kit, orderID uint64) (sent bool, err error) {
	return r.checkAndSendDetectFailed(kt, orderID, false)
}

// checkAndSendDetectFailed 检查并通知预检失败
func (r *Notifier) checkAndSendDetectFailed(kt *kit.Kit, orderID uint64, forceSend bool) (sent bool, err error) {
	if !cc.WoaServer().RecycleNotice.Recycle.Enable {
		logs.Infof("recycle notice is disabled, skip send failed notice, orderID: %d, rid: %s", orderID, kt.Rid)
		return false, nil
	}

	recycleSubOrders, err := r.findManyRecycleOrder(kt, orderID)
	if err != nil {
		return false, err
	}
	if len(recycleSubOrders) == 0 {
		logs.Infof("no sub orders to notify, rid: %s", kt.Rid)
		return false, nil
	}

	if !forceSend {
		var recycleSubOrder *table.RecycleOrder
		for _, suborder := range recycleSubOrders {
			if suborder.Status == table.RecycleStatusDetectFailed {
				recycleSubOrder = suborder
				break
			}
		}

		if recycleSubOrder == nil {
			logs.Infof("no failed sub order for notification, orderID: %d, rid: %s", orderID, kt.Rid)
			return false, nil
		}

		if !r.hitFailedNoticeInterval(recycleSubOrder.UpdateAt) {
			logs.Infof("failed sub order not hit failed notice interval, orderID: %d, rid: %s", orderID, kt.Rid)
			return false, nil
		}
	}

	if err = r.sendRecycleFailedNotification(kt, recycleSubOrders); err != nil {
		logs.Errorf("send notification failed, orderID: %d, err: %v, rid: %s", orderID, err, kt.Rid)
		return false, err
	}

	return true, nil
}

// sendRecycleSuccessNotification 发送回收成功的邮件 + 企微通知
func (r *Notifier) sendRecycleSuccessNotification(kt *kit.Kit, recycleSubOrders []*table.RecycleOrder) error {
	if len(recycleSubOrders) == 0 {
		logs.Errorf("no sub orders to notify, rid: %s", kt.Rid)
		return fmt.Errorf("no sub orders to notify")
	}

	if r.cmsiClient == nil {
		logs.Errorf("cmsi client is nil, orderID: %d, rid: %s", recycleSubOrders[0].OrderID, kt.Rid)
		return fmt.Errorf("cmsi client is nil")
	}

	if recycleSubOrders[0].User == "" {
		logs.Errorf("no receiver, orderID: %d, rid: %s", recycleSubOrders[0].OrderID, kt.Rid)
		return fmt.Errorf("no receiver, orderID: %d", recycleSubOrders[0].OrderID)
	}

	bkHcmURL := cc.WoaServer().BkHcmURL
	loc, err := time.LoadLocation(cc.WoaServer().LocalTimezone)
	if err != nil {
		logs.Errorf("load location failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 邮件标题
	title := fmt.Sprintf(constant.RecycleSuccessNoticeTitle, recycleSubOrders[0].BizName, recycleSubOrders[0].OrderID)
	createAt := recycleSubOrders[0].CreateAt.In(loc).Format(constant.DateTimeLayout)
	receiver, orderID, bizID := recycleSubOrders[0].User, recycleSubOrders[0].OrderID, recycleSubOrders[0].BizID

	// 构建子单信息表格
	var rows strings.Builder
	// 记录最后一个子单完成时间
	var latestDone time.Time
	var successNum uint
	for _, suborder := range recycleSubOrders {
		doneTime := suborder.UpdateAt.In(loc).Format(constant.DateTimeLayout)
		successNum += suborder.SuccessNum
		if suborder.UpdateAt.After(latestDone) {
			latestDone = suborder.UpdateAt
		}

		rows.WriteString(fmt.Sprintf(constant.RecycleSuccessNoticeEmailTableTemplate, suborder.SuborderID,
			suborder.SuccessNum, doneTime, bkHcmURL, suborder.BizID, suborder.SuborderID, suborder.BizID))
	}

	completed := latestDone.In(loc).Format(constant.DateTimeLayout)
	// 邮件内容
	emailContent := fmt.Sprintf(constant.RecycleSuccessNoticeEmailContentTemplate, title, bkHcmURL, bkHcmURL, title,
		receiver, successNum, orderID, createAt, rows.String())

	if err = r.sendMail(kt, receiver, title, emailContent, orderID); err != nil {
		logs.Errorf("send email failed, orderID: %d, err: %v, rid: %s", orderID, err, kt.Rid)
		return err
	}

	// 企微
	orderLink := fmt.Sprintf(constant.RecycleOrderLinkTemplate, bkHcmURL, bizID, orderID)
	wecomContent := fmt.Sprintf(constant.RecycleSuccessNoticeWeComContentTemplate, orderID, orderID, successNum,
		receiver, recycleSubOrders[0].BizName, createAt, completed, orderLink)

	return r.sendWeCom(kt, receiver, wecomContent, orderID)
}

// findManyRecycleOrder 查询该主单下所有子单
func (r *Notifier) findManyRecycleOrder(kt *kit.Kit, orderID uint64) (recycleSubOrders []*table.RecycleOrder,
	err error) {

	filter := mapstr.MapStr{
		"order_id": orderID,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}
	recycleSubOrders, err = dao.Set().RecycleOrder().FindManyRecycleOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("query sub orders failed, orderID: %d, err: %v, rid: %s", orderID, err, kt.Rid)
		return nil, fmt.Errorf("list sub orders failed: %w", err)
	}
	return recycleSubOrders, nil
}

// sendMail 发送邮件
func (r *Notifier) sendMail(kt *kit.Kit, receiver, title, content string, orderID uint64) error {
	mail := &cmsi.CmsiMail{
		Title:            title,
		Content:          content,
		ReceiverUserName: receiver,
	}
	if err := r.cmsiClient.SendMail(kt, mail); err != nil {
		logs.Errorf("send email failed, orderID: %d, receiver: %s, err: %v, rid: %s", orderID, receiver,
			err, kt.Rid)
		return fmt.Errorf("send recycle email failed, orderID: %d, err: %w", orderID, err)
	}
	return nil
}

// sendWeCom 发送企微通知
func (r *Notifier) sendWeCom(kt *kit.Kit, receiver, content string, orderID uint64) error {
	if r.thirdCli == nil || r.thirdCli.BkChat == nil {
		logs.Errorf("wecom client is nil, skip send wecom, orderID: %d, receiver: %s, rid: %s", orderID,
			receiver, kt.Rid)
		return fmt.Errorf("wecom client is nil, skip send wecom, orderID: %d", orderID)
	}

	resp, err := r.thirdCli.BkChat.SendApplyDoneMsg(kt.Ctx, kt.Header(), receiver, content)
	if err != nil {
		logs.Errorf("send wecom failed, orderID: %d, user: %s, err: %v, rid: %s", orderID, receiver, err, kt.Rid)
		return err
	}
	if resp.Code != 0 {
		logs.Errorf("send wecom failed, orderID: %d, user: %s, code: %d, msg: %s, rid: %s", orderID, receiver,
			resp.Code, resp.Msg, kt.Rid)
		return fmt.Errorf("send recycle wecom failed, orderID: %d, err: %s", orderID, resp.Msg)
	}
	return nil
}

// hitFailedNoticeInterval 根据失败发生时间判断是否命中配置的通知间隔
func (r *Notifier) hitFailedNoticeInterval(updateAt time.Time) bool {
	if updateAt.IsZero() {
		return false
	}
	intervals := cc.WoaServer().RecycleNotice.Recycle.FailedNoticeIntervals
	if len(intervals) == 0 {
		return false
	}
	elapsedMinutes := int(time.Since(updateAt).Minutes())
	for _, interval := range intervals {
		if elapsedMinutes == interval {
			return true
		}
	}
	return false
}

// sendRecycleFailedNotification 发送回收失败的邮件 + 企微通知
func (r *Notifier) sendRecycleFailedNotification(kt *kit.Kit, recycleSubOrders []*table.RecycleOrder) error {
	if len(recycleSubOrders) == 0 {
		logs.Errorf("no sub orders to notify, rid: %s", kt.Rid)
		return fmt.Errorf("no sub orders to notify")
	}

	if r.cmsiClient == nil {
		logs.Errorf("cmsi client is nil, rid: %s", kt.Rid)
		return fmt.Errorf("cmsi client is nil")
	}

	if recycleSubOrders[0].User == "" {
		logs.Errorf("no receiver, orderID: %d, rid: %s", recycleSubOrders[0].OrderID, kt.Rid)
		return fmt.Errorf("no receiver, orderID: %d", recycleSubOrders[0].OrderID)
	}

	bkHcmURL := cc.WoaServer().BkHcmURL
	loc, err := time.LoadLocation(cc.WoaServer().LocalTimezone)
	if err != nil {
		logs.Errorf("load location failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	receiver, orderID, bizID := recycleSubOrders[0].User, recycleSubOrders[0].OrderID, recycleSubOrders[0].BizID

	// 计算总回收数量
	var totalNum uint
	unfinishedCount := 0
	for _, sub := range recycleSubOrders {
		totalNum += sub.TotalNum
		if sub.Status != table.RecycleStatusDone {
			unfinishedCount++
		}
	}

	createAt := recycleSubOrders[0].CreateAt.In(loc).Format(constant.DateTimeLayout)
	// 构建子单表格
	var rows strings.Builder
	for _, sub := range recycleSubOrders {
		statusDesc, suggestion := r.getStatusDescAndSuggestion(sub.Status)
		rows.WriteString(fmt.Sprintf(constant.RecycleFailedNoticeEmailTableTemplate, sub.SuborderID, sub.TotalNum,
			statusDesc, suggestion, bkHcmURL, sub.BizID, sub.SuborderID, sub.BizID))
	}

	title := fmt.Sprintf(constant.RecycleFailedNoticeTitle, recycleSubOrders[0].BizName, recycleSubOrders[0].OrderID)
	emailContent := fmt.Sprintf(constant.RecycleFailedNoticeEmailContentTemplate, title, bkHcmURL, bkHcmURL, title,
		recycleSubOrders[0].BizName, totalNum, recycleSubOrders[0].OrderID, createAt, rows.String())

	if err = r.sendMail(kt, receiver, title, emailContent, orderID); err != nil {
		logs.Errorf("send failed email failed, orderID: %d, err: %v, rid: %s", orderID, err, kt.Rid)
		return err
	}

	// 企微通知
	orderLink := fmt.Sprintf(constant.RecycleOrderLinkTemplate, bkHcmURL, bizID, orderID)
	wecomContent := fmt.Sprintf(constant.RecycleFailedNoticeWeComContentTemplate, orderID, orderID, unfinishedCount,
		len(recycleSubOrders), receiver, recycleSubOrders[0].BizName, createAt, orderLink)

	return r.sendWeCom(kt, receiver, wecomContent, orderID)
}

// getStatusDescAndSuggestion 根据状态获取状态描述和处理建议
func (r *Notifier) getStatusDescAndSuggestion(status table.RecycleStatus) (statusDesc, suggestion string) {
	info, exists := recycleStatusInfoMap[string(status)]
	if !exists {
		return string(status), "--"
	}

	statusDesc = fmt.Sprintf(info.template, info.text)
	return statusDesc, info.suggestion
}

// statusInfo 状态信息
type statusInfo struct {
	text       string // 状态文本
	template   string // 状态样式模板
	suggestion string // 处理建议
}

var (
	statusSuccess = statusInfo{
		text:       "回收成功",
		template:   constant.RecycleStatusSuccessTemplate,
		suggestion: "已回收成功",
	}
	statusDetectFailed = statusInfo{
		text:       "预检未通过",
		template:   constant.RecycleStatusFailedTemplate,
		suggestion: "按预检要求处理",
	}
	statusPending = statusInfo{
		text:       "待预检",
		template:   constant.RecycleStatusPendingTemplate,
		suggestion: "待单据流转",
	}
	statusTerminated = statusInfo{
		text:       "已终止单据",
		template:   constant.RecycleStatusPendingTemplate,
		suggestion: "--",
	}
	statusDestroyFailed = statusInfo{
		text:       "销毁环节失败",
		template:   constant.RecycleStatusFailedTemplate,
		suggestion: "联系HCM助手处理",
	}

	// recycleStatusInfoMap 回收状态信息映射
	recycleStatusInfoMap = map[string]statusInfo{
		"DONE":           statusSuccess,
		"DETECT_FAILED":  statusDetectFailed,
		"UNCOMMIT":       statusPending,
		"COMMITTED":      statusPending,
		"DETECTING":      statusPending,
		"FOR_AUDIT":      statusPending,
		"REJECTED":       statusTerminated,
		"TERMINATE":      statusTerminated,
		"TRANSITING":     statusDestroyFailed,
		"TRANSIT_FAILED": statusDestroyFailed,
		"RETURNING":      statusDestroyFailed,
		"RETURN_FAILED":  statusDestroyFailed,
	}
)
