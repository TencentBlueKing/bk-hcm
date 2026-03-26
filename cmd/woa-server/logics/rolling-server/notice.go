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

package rollingserver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	rstypes "hcm/cmd/woa-server/types/rolling-server"
	"hcm/pkg/api/core"
	config "hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

func (l *logics) pushReturnNotificationsPeriodically(loc *time.Location) {
	// 每周一上午10:00推送
	now := time.Now()
	localTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	monday := times.GetMondayOfWeek(localTime)
	nextRunTime := time.Date(monday.Year(), monday.Month(), monday.Day(), 10, 0, 0, 0, loc)
	if now.After(nextRunTime) {
		nextRunTime = nextRunTime.Add(7 * time.Hour * 24)
	}

	for {
		// 等待到下一个时间
		time.Sleep(time.Until(nextRunTime))

		kt := core.NewBackendKit()
		start := time.Now()
		if err := l.PushReturnNotifications(kt, []int64{}, []string{}); err != nil {
			logs.Errorf("%s: failed to push rolling server return notice, err: %v, start time: %v, rid: %s",
				constant.RollingServerReturnNotificationPushFailed, err, start, kt.Rid)
		} else {
			end := time.Now()
			logs.Infof("push rolling server return notice success, start time: %v, end time: %v, cost: %v, rid: %s",
				start, end, end.Sub(start), kt.Rid)
		}

		// 计算下一个时间
		nextRunTime = nextRunTime.Add(7 * time.Hour * 24)
	}
}

// PushReturnNotifications push return notifications
func (l *logics) PushReturnNotifications(kt *kit.Kit, bizIDs []int64, extraReceivers []string) error {
	// 如果没传bizID, 则表示推送给所有业务
	if len(bizIDs) == 0 {
		var err error
		bizIDs, err = l.listIEGBizIDs(kt)
		if err != nil {
			logs.Errorf("list ieg biz ids failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	resPoolBizMap, err := l.listResPoolBizIDs(kt)
	if err != nil {
		logs.Errorf("list rolling resource pool business failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	bizIDNameMap := make(map[int64]string)
	for _, subBizIDs := range slice.Split(bizIDs, constant.BatchOperationMaxLimit) {
		rules := []cmdb.Rule{&cmdb.AtomRule{Field: "bk_biz_id", Operator: cmdb.OperatorIn, Value: subBizIDs}}
		expression := &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: rules}}
		params := &cmdb.SearchBizParams{BizPropertyFilter: expression, Fields: []string{"bk_biz_id", "bk_biz_name"}}
		resp, err := l.cmdbClient.SearchBusiness(kt, params)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, bizIDs: %v, rid: %s", err, subBizIDs, kt.Rid)
			return err
		}
		if len(resp.Info) != len(subBizIDs) {
			logs.Errorf("search business failed, bizIDs: %v, resp: %v, rid: %s", subBizIDs, resp.Info, kt.Rid)
			return fmt.Errorf("search business failed, bizIDs: %v, resp: %v", subBizIDs, resp.Info)
		}
		for _, info := range resp.Info {
			bizIDNameMap[info.BizID] = info.BizName
		}
	}

	for _, bizID := range bizIDs {
		if _, ok := resPoolBizMap[bizID]; ok {
			logs.Infof("skip resource pool business return notifications, bizID: %d, rid: %s", bizID, kt.Rid)
			continue
		}

		bizName, ok := bizIDNameMap[bizID]
		if !ok {
			logs.Errorf("get biz name failed, bizID: %d, rid: %s", bizID, kt.Rid)
			return errors.New("get biz name failed")
		}

		err := l.pushBizReturnNotification(kt, bizID, bizName, extraReceivers)
		if err != nil {
			logs.Errorf("%s: push biz return notification failed, bizID: %d, err: %v, rid: %s",
				constant.RollingServerReturnNotificationPushFailed, bizID, err, kt.Rid)
		}
	}

	return nil
}

func (l *logics) pushBizReturnNotification(kt *kit.Kit, bizID int64, bizName string, extraReceivers []string) error {
	// 1.获取滚服未退还完成的子单信息
	now := time.Now()
	startYear, startMonth, startDay := subDay(now.Year(), int(now.Month()), now.Day(), constant.CalculateFineEndDay)
	startDate := rstypes.AppliedRecordDate{Year: startYear, Month: startMonth, Day: startDay}
	endDate := rstypes.AppliedRecordDate{Year: now.Year(), Month: int(now.Month()), Day: now.Day()}
	unReturnedSubOrderMsgs, err := l.findUnReturnedSubOrderMsg(kt, bizID, startDate, endDate, false)
	if err != nil {
		logs.Errorf("find unreturned sub order msg failed, err: %v, bizID: %d, startDate: %v, endDate: %v, rid: %s",
			err, bizID, startDate, endDate, kt.Rid)
		return err
	}
	if len(unReturnedSubOrderMsgs) == 0 {
		return nil
	}

	// 2.获取消息接收人列表、业务名称等，构建推送信息
	receivers, cc, err := l.getReceiverAndCC(kt, bizID, unReturnedSubOrderMsgs, extraReceivers)
	if err != nil {
		logs.Errorf("get receiver and cc failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}

	pushMsg := &rstypes.PushReturnNoticeMsg{
		BizID:                  bizID,
		BizName:                bizName,
		UnReturnedSubOrderMsgs: unReturnedSubOrderMsgs,
		Receivers:              receivers,
		CC:                     cc,
	}

	// 3.发送邮件通知
	sendMailErr := l.generateAndSendMail(kt, pushMsg)
	if sendMailErr != nil {
		logs.Errorf("generate and send mail failed, err: %v, bizID: %d, rid: %s", sendMailErr, bizID, kt.Rid)
	}

	// 4.发送企业微信通知
	weComErr := l.generateAndSendWeCom(kt, pushMsg)
	if weComErr != nil {
		logs.Errorf("generate and send wecom failed, err: %v, bizID: %d, rid: %s", weComErr, bizID, kt.Rid)
	}

	if sendMailErr != nil && weComErr != nil {
		return fmt.Errorf("send mail and wecom failed, bizID: %d, sendMailErr: %v, weComErr: %v", bizID, sendMailErr,
			weComErr)
	}

	return nil
}

func (l *logics) getReceiverAndCC(kt *kit.Kit, bizID int64, unReturnedSubOrderMsgs []rstypes.UnReturnedSubOrderMsg,
	extraReceivers []string) ([]string, []string, error) {

	receivers := make([]string, 0)
	cc := make([]string, 0)

	if config.WoaServer().RollingServer.ReturnNotification.SendToBusiness {
		for _, msg := range unReturnedSubOrderMsgs {
			receivers = append(receivers, msg.AppliedUser)
		}

		bizMaintainerMap, err := l.bizLogics.GetBkBizMaintainer(kt, []int64{bizID})
		if err != nil {
			logs.Errorf("failed to get bk biz maintainer, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
			return nil, nil, err
		}
		var ok bool
		cc, ok = bizMaintainerMap[bizID]
		if !ok {
			logs.Errorf("failed to get bk biz maintainer, bizID: %d, rid: %s", err, bizID, kt.Rid)
			return nil, nil, fmt.Errorf("failed to get bk biz maintainer, bizID: %d", bizID)
		}
	}

	if len(extraReceivers) > 0 {
		receivers = append(receivers, extraReceivers...)
	}

	if len(config.WoaServer().RollingServer.ReturnNotification.DefaultReceivers) > 0 {
		receivers = append(receivers, config.WoaServer().RollingServer.ReturnNotification.DefaultReceivers...)
	}

	if len(receivers) == 0 {
		logs.Errorf("no receivers found, bizID: %d, rid: %s", bizID, kt.Rid)
		return nil, nil, errors.New("no receivers found")
	}

	return slice.Unique(receivers), slice.Unique(cc), nil
}

func (l *logics) generateAndSendMail(kt *kit.Kit, pushMsg *rstypes.PushReturnNoticeMsg) error {
	title, content, err := l.generateMail(kt, pushMsg)
	if err != nil {
		logs.Errorf("generate mail failed, err: %v, bizID: %d, rid: %s", err, pushMsg.BizID, kt.Rid)
		return err
	}

	mail := &cmsi.CmsiMail{
		ReceiverUserName: strings.Join(pushMsg.Receivers, ","),
		Title:            title,
		Content:          content,
		CcUserName:       strings.Join(pushMsg.CC, ","),
	}

	return l.cmsiClient.SendMail(kt, mail)
}

func (l *logics) generateMail(kt *kit.Kit, pushMsg *rstypes.PushReturnNoticeMsg) (string, string, error) {
	// 1.构建邮件标题
	title := fmt.Sprintf(constant.RsReturnNoticeTitle, pushMsg.BizName)

	// 2.构建邮件每一行的内容
	tableContent := ""
	var allNeedReturnedCore int64
	for _, msg := range pushMsg.UnReturnedSubOrderMsgs {
		appliedDate := fmt.Sprintf("%d-%02d-%02d", msg.AppliedYear, msg.AppliedMonth, msg.AppliedDay)
		needReturnedDate := fmt.Sprintf("%d-%02d-%02d", msg.NeedReturnedYear, msg.NeedReturnedMonth,
			msg.NeedReturnedDay)
		needReturnedCore := msg.AppliedCore - msg.ReturnedCore - msg.ExemptedReturnedCore

		tableContent += fmt.Sprintf(constant.RsReturnNoticeEmailTableTemplate, appliedDate, needReturnedDate,
			renderRsReturnNoticeFineState(msg.FineState), msg.SubOrderID, msg.AppliedUser,
			strconv.FormatInt(msg.AppliedCore, 10), strconv.FormatInt(needReturnedCore, 10))
		allNeedReturnedCore += needReturnedCore
	}

	// 3.构建邮件整体内容
	bkHcmURL := config.WoaServer().BkHcmURL
	now := time.Now()
	date := fmt.Sprintf("%d年%d月%d日", now.Year(), now.Month(), now.Day())
	queryStartTime := now.AddDate(0, 0, -constant.CalculateFineEndDay)
	queryStartDate := fmt.Sprintf("%d-%02d-%02d", queryStartTime.Year(), queryStartTime.Month(), queryStartTime.Day())
	queryEndDate := fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())

	content := fmt.Sprintf(constant.RsReturnNoticeEmailContentTemplate,
		bkHcmURL, bkHcmURL, title, pushMsg.BizName, date, allNeedReturnedCore, tableContent, bkHcmURL, pushMsg.BizID,
		queryStartDate, queryEndDate)

	return title, content, nil
}

func renderRsReturnNoticeFineState(fineState enumor.RsUnReturnedSubOrderFineState) string {
	var renderTemplate string
	switch fineState {
	case enumor.RsFineExemptionState:
		renderTemplate = constant.RsFineExemptionStateTemplate
	case enumor.RsImpendingFineState:
		renderTemplate = constant.RsImpendingFineStateTemplate
	case enumor.RsHasFineState:
		renderTemplate = constant.RsHasFineStateTemplate
	default:
		renderTemplate = constant.RsFineExemptionStateTemplate
	}

	return fmt.Sprintf(renderTemplate, fineState)
}

func (l *logics) generateAndSendWeCom(kt *kit.Kit, pushMsg *rstypes.PushReturnNoticeMsg) error {
	content, err := l.generateWeComContent(kt, pushMsg)
	if err != nil {
		logs.Errorf("generate wecom content failed, err: %v, bizID: %d, rid: %s", err, pushMsg.BizID, kt.Rid)
		return err
	}

	users := make([]string, 0)
	users = append(users, pushMsg.Receivers...)
	users = append(users, pushMsg.CC...)
	users = slice.Unique(users)

	failedUsers := make([]string, 0)
	for _, user := range users {
		resp, err := l.thirdCli.BkChat.SendApplyDoneMsg(kt.Ctx, kt.Header(), user, content)
		if err != nil {
			logs.Warnf("failed to send bkchat message, err: %v, bizID: %d, user: %s, rid: %s", err, pushMsg.BizID,
				user, kt.Rid)
			failedUsers = append(failedUsers, user)
			continue
		}
		if resp.Code != 0 {
			logs.Errorf("failed to send bkchat message, code: %d, msg: %s, bizID: %d, user: %s, rid: %s", resp.Code,
				resp.Msg, pushMsg.BizID, user, kt.Rid)
			failedUsers = append(failedUsers, user)
		}
	}

	if len(failedUsers) > 0 {
		return fmt.Errorf("failed to send bkchat message to users: %v", failedUsers)
	}

	return nil
}

func (l *logics) generateWeComContent(kt *kit.Kit, pushMsg *rstypes.PushReturnNoticeMsg) (string, error) {
	title := fmt.Sprintf(constant.RsReturnNoticeTitle, pushMsg.BizName)
	now := time.Now()
	date := fmt.Sprintf("%d年%d月%d日", now.Year(), now.Month(), now.Day())
	var allNeedReturnedCore int64
	for _, msg := range pushMsg.UnReturnedSubOrderMsgs {
		allNeedReturnedCore += msg.AppliedCore - msg.ReturnedCore - msg.ExemptedReturnedCore
	}
	bkHcmURL := config.WoaServer().BkHcmURL
	queryStartTime := now.AddDate(0, 0, -constant.CalculateFineEndDay)
	queryStartDate := fmt.Sprintf("%d-%02d-%02d", queryStartTime.Year(), queryStartTime.Month(), queryStartTime.Day())
	queryEndDate := fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())

	content := fmt.Sprintf(constant.RsReturnNoticeWeComContentTemplate, title, pushMsg.BizName, date,
		allNeedReturnedCore, bkHcmURL, pushMsg.BizID, queryStartDate, queryEndDate)

	return content, nil
}
