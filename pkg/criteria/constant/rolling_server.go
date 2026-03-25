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

package constant

const (
	// PlaceholderQuotaConfig defines the placeholder for quota config
	PlaceholderQuotaConfig = -1
	// FineProportion 罚金成本比例
	FineProportion = 0.5
	// SyncAllBiz sync rolling all business bill
	SyncAllBiz = -1
	// FirstDay 每月第一天
	FirstDay = 1
	// CalculateFineStartDay 滚服机器"超期未退还"开始计算罚金的启始天数， 由于存在时区问题，可能会导致罚金不准，所以这里少算业务一天的罚金，从31变成32
	CalculateFineStartDay = 32
	// RollingServerLatestReturnDay 滚服机器用户可使用的最晚归还天数
	RollingServerLatestReturnDay = 30
	// CalculateFineEndDay 滚服机器"超期未退还"开始计算罚金的最后天数
	CalculateFineEndDay = 121
	// RsFineExemptionDay 滚服机器"超期未退还"豁免期天数
	RsFineExemptionDay = 22

	// DefaultCityID 城市ID默认值
	DefaultCityID = 10000
	// DefaultBusinessModID 三级业务id默认值
	DefaultBusinessModID = 0
	// DefaultBusinessModName 三级业务名称默认值
	DefaultBusinessModName = ""
	// PlatformID 平台ID
	PlatformID = 496
	// ResClassID 资源规格ID
	ResClassID = 4911

	// IEGOperateDeptID IEG初始运维部门ID
	IEGOperateDeptID = 3

	// RsExpiredTerminationZonesMaxRunes 过期终止通知邮件中可用区列最大展示字符数
	RsExpiredTerminationZonesMaxRunes = 10
)

const (
	// RsReturnNoticeTitle 滚服资源退还到期提醒标题模版
	RsReturnNoticeTitle = "HCM-自研云-%s业务-滚服项目-主机退还通知"

	// RsFineExemptionStateTemplate rolling server fine exemption state template.
	RsFineExemptionStateTemplate = `<span
                      style="
                        padding: 0 10px;
                        display: inline-block;
                        border-color: #14a5684d;
						background-color: #e4faf0;
						color: #14a568;
                        border-radius: 2px;
                        font-size: 12px;
                        line-height: 22px;
                        box-sizing: border-box;
                      "
                    >
                      %s
                    </span>`

	// RsImpendingFineStateTemplate rolling server impending fine state template.
	RsImpendingFineStateTemplate = `<span
                      style="
                        padding: 0 10px;
                        display: inline-block;
                        border-color: #fea5004d;
						background-color: #fff1db;
						color: #fe9c00;
                        border-radius: 2px;
                        font-size: 12px;
                        line-height: 22px;
                        box-sizing: border-box;
                      "
                    >
                      %s
                    </span>`

	// RsHasFineStateTemplate rolling server has fine state template.
	RsHasFineStateTemplate = `<span
                      style="
                        padding: 0 10px;
                        display: inline-block;
						border-color: #ea35364d;
						background-color: #feebea;
						color: #ea3636;
                        border-radius: 2px;
                        font-size: 12px;
                        line-height: 22px;
                        box-sizing: border-box;
                      "
                    >
                      %s
                    </span>`

	// RsReturnNoticeEmailTableTemplate 滚服资源退还到期提醒邮件表格模版
	RsReturnNoticeEmailTableTemplate = `<tr>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
</tr>
`
	// RsReturnNoticeEmailContentTemplate 滚服资源退还到期提醒邮件内容模版
	RsReturnNoticeEmailContentTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>{{ title }}</title>
  </head>
  <body style="margin: 0; padding: 0">
    <!-- banner start -->
    <div
      style="
        padding: 30px 40px 0;
        height: 148px;
        background-color: #2e3959;
        box-sizing: border-box;
      "
    >
      <img
        style="
          margin-right: 8px;
          width: 48px;
          height: 48px;
          vertical-align: middle;
        "
        src="data:image/png;base64,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"
        alt="logo"
      />
      <div style="display: inline-block; vertical-align: middle">·
        <div style="font-size: 18px; color: #fff">海垒</div>
        <a
          style="font-size: 14px; color: #ffffff80; text-decoration: none"
          href="%s"
          target="_blank"
        >
          %s
        </a>
      </div>
    </div>
    <!-- banner-end -->

    <!-- email-start -->
    <div style="margin: -48px 40px 0; box-sizing: border-box">
      <!-- email-header-start -->
      <div
        style="
          padding: 20px;
          border-radius: 2px;
          background-color: #fff;
          box-shadow: 0 4px 10px 2px #979ba54d;
          box-sizing: border-box;
        "
      >
        <h1
          style="
            margin: 0 0 16px;
            font-size: 18px;
            font-weight: 700;
            color: #394567;
          "
        >
          %s
        </h1>
        <div style="margin: 0; padding: 0; width: 100%%; line-height: 22px">
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">业务：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700"
                >%s</span
              >
            </span>
          </div>
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">日期：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700"
                >%s</span
              >
            </span>
          </div>
          <div style="display: inline-block">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56"
                >CPU核数（未退还）：</span
              >
              <span style="font-size: 12px; color: #e71818; font-weight: 700"
                >%d</span
              >
              <span style="font-size: 12px; color: #4d4f56">核</span>
            </span>
          </div>
        </div>
      </div>
      <!-- email-header-end -->

      <!-- email-body-start -->
      <div
        style="
          margin-top: 16px;
          padding: 20px 26px;
          border-radius: 2px;
          background-color: #fff;
          box-shadow: 0 4px 10px 2px #979ba54d;
          box-sizing: border-box;
        "
      >
        <!-- table-start -->
        <div style="margin-top: 16px; width: 100%%">
          <div style="width: 100%%; overflow-x: auto">
            <table
              style="
                width: 100%%;
                border: 1px solid #dcdee5;
                border-collapse: collapse;
                border-spacing: 0;
                border-radius: 2px;
                text-align: left;
                line-height: 42px;
              "
            >
              <thead
                style="
                  border-bottom: 1px solid #dcdee5;
                  background-color: #fafbfd;
                  color: #313238;
                "
              >
                <tr>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    申请日期
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    应退还日期
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    涉及罚金
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    申请单号
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    申请人
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    申请核数
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    未退还核数
                  </th>
                </tr>
              </thead>
              <tbody>
                %s
              </tbody>
            </table>
          </div>
        </div>
        <!-- table-end -->
        <div
          style="
            margin-top: 8px;
            margin-bottom: 8px;
            font-size: 12px;
            color: #4d4f56;
          "
        >
          更多信息请点击
          <a
            style="
              color: #3a84ff;
              font-size: 12px;
              line-height: 20px;
              text-decoration: none;
            "
            href="%s/#/business/rolling-server?bizs=%d&filter=roll_date=%s,%s"
            target="_blank"
          >
            链接查看详情
          </a>
        </div>
        <div style="margin-top: 12px; font-size: 12px; color: #4d4f56">
          注意：超过30天未退还将产生罚金，详情参考<a
            style="
              color: #3a84ff;
              font-size: 12px;
              line-height: 20px;
              text-decoration: none;
            "
            href="https://iwiki.woa.com/p/4012608772"
            target="_blank"
            >https://iwiki.woa.com/p/4012608772</a
          >，罚金最长收取3个月(主机申请后的第31到121天内收取罚金)，如有任何疑问，可企微联系 ICR(IEG资源服务助手)
        </div>
      </div>
      <!-- email-body-end -->
    </div>
    <!-- email-end -->
  </body>
</html>
`
	// RsReturnNoticeWeComContentTemplate 滚服资源退还到期提醒企微内容模版
	RsReturnNoticeWeComContentTemplate = `
%s	

业务：%s
日期：%s
CPU核数(未退还)：%d核
点击链接查看详情：%s/#/business/rolling-server?bizs=%d&filter=roll_date=%s,%s
`

	// HostDeliveryNoticeTitle 主机申请交付通知标题模板
	HostDeliveryNoticeTitle = "HCM-%s业务-主机申领交付通知-单号%d"

	// HostDeliveryNoticeEmailTableTemplate 主机申请交付通知邮件表格模版
	HostDeliveryNoticeEmailTableTemplate = `<tr>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %d
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
    <td
            style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      font-size: 12px;
                      color: #4d4f56;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
    >
        %s
    </td>
</tr>
`

	// HostDeliveryNoticeWeComContentTemplate 主机申请交付通知企微内容模版
	HostDeliveryNoticeWeComContentTemplate = `『HCM』资源申请【主机申领交付通知】

标题：资源申请单据完成[order_id:%d]
资源申请单号：%d
提单人：%s
业务：%s
需求类型：%s
创建时间：%s
备注：%s

资源申请单链接：%s/#/business/applications/host-application/detail/%d?bizs=%d&bkBizId=%d&resource_type=QCLOUDCVM`

	// HostDeliveryNoticeEmailContentTemplate 主机申请交付通知邮件内容模版
	HostDeliveryNoticeEmailContentTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>%s</title>
  </head>
  <body style="margin: 0; padding: 0">
    <!-- banner start -->
    <div style="padding: 30px 40px 0; height: 148px; background-color: #2e3959; box-sizing: border-box">
      <img
        style="margin-right: 8px; width: 48px; height: 48px; vertical-align: middle"
        src="data:image/png;base64,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"
        alt="logo"
      />
      <div style="display: inline-block; vertical-align: middle">
        <div style="font-size: 18px; color: #fff">海垒</div>
        <a
          style="font-size: 14px; color: #ffffff80; text-decoration: none"
          href="%s"
          target="_blank"
        >
          %s
        </a>
      </div>
    </div>
    <!-- banner-end -->

    <!-- email-start -->
    <div style="margin: -48px 40px 0; box-sizing: border-box">
      <!-- email-header-start -->
      <div
        style="
          padding: 20px;
          border-radius: 2px;
          background-color: #fff;
          box-shadow: 0 4px 10px 2px #979ba54d;
          box-sizing: border-box;
        "
      >
        <h1 style="margin: 0 0 16px; font-size: 18px; font-weight: 700; color: #394567">
          %s
        </h1>
        <div style="margin: 0; padding: 0; width: 100%%; line-height: 22px">
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">业务：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700">%s业务</span>
            </span>
          </div>
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">申领日期：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700">%s</span>
            </span>
          </div>
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">申领数量：</span>
              <span style="font-size: 12px; color: #e71818; font-weight: 700">%d</span>
              <span style="font-size: 12px; color: #4d4f56">台</span>
            </span>
          </div>
          <div style="display: inline-block">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">申领单号：</span>
              <span style="font-size: 12px; color: #e71818; font-weight: 700">%d</span>
            </span>
          </div>
        </div>
      </div>
      <!-- email-header-end -->

      <!-- email-body-start -->
      <div
        style="
          margin-top: 16px;
          padding: 20px 26px;
          border-radius: 2px;
          background-color: #fff;
          box-shadow: 0 4px 10px 2px #979ba54d;
          box-sizing: border-box;
        "
      >
        <!-- table-start -->
        <div style="margin-top: 16px; width: 100%%">
          <div style="width: 100%%; overflow-x: auto">
            <table
              style="
                width: 100%%;
                border: 1px solid #dcdee5;
                border-collapse: collapse;
                border-spacing: 0;
                border-radius: 2px;
                text-align: left;
                line-height: 42px;
              "
            >
              <thead style="border-bottom: 1px solid #dcdee5; background-color: #fafbfd; color: #313238">
                <tr>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    机型
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    需求数量
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    地域
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    可用区
                  </th>
                  <th
                    style="
                      padding: 0 16px;
                      border-bottom: 1px solid #dcdee5;
                      background-color: #fafbfd;
                      font-size: 12px;
                      color: #313238;
                      font-weight: normal;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    "
                  >
                    所属母机数量
                  </th>
                </tr>
              </thead>
              <tbody>
                %s
              </tbody>
            </table>
          </div>
        </div>
        <!-- table-end -->

        <!-- more-info-start -->
        <div style="margin-top: 8px; margin-bottom: 8px; font-size: 12px; color: #4d4f56">
          更多交付信息，请
          <a
            style="color: #3a84ff; font-size: 12px; line-height: 20px; text-decoration: none"
            href="%s/#/business/ticket?bizs=%d&type=host_apply&scene_apply=device&host_apply_device_rules=%s"
            target="_blank"
          >
            点击链接查看详情
          </a>
          。如有任何疑问，可企微联系 ICR(IEG资源服务助手)
        </div>
        <!-- more-info-end -->
      </div>
      <!-- email-body-end -->
    </div>
    <!-- email-end -->
  </body>
</html>`

	// RsExpiredTerminationTitle 滚服项目过期终止通知标题模版
	RsExpiredTerminationTitle = "HCM-%s业务-主机申领单据-滚服项目-过期终止通知-%s月份"

	// RsExpiredTerminationEmailTableTemplate 滚服项目过期终止通知邮件表格模版
	RsExpiredTerminationEmailTableTemplate = `<tr>
	<td style="padding: 8px; border-bottom: 1px solid #dfe1e6; text-align: left">%s</td>		
	<td style="padding: 8px; border-bottom: 1px solid #dfe1e6; text-align: left">%s</td>
	<td style="padding: 8px; border-bottom: 1px solid #dfe1e6; text-align: left">%s</td>
	<td style="padding: 8px; border-bottom: 1px solid #dfe1e6; text-align: left">%s</td>
	<td style="padding: 8px; border-bottom: 1px solid #dfe1e6; text-align: left">%d</td>
	<td style="padding: 8px; border-bottom: 1px solid #dfe1e6; text-align: left">%s</td>
</tr>
`

	// RsExpiredTerminationEmailContentTemplate 滚服项目过期终止通知邮件内容模版
	RsExpiredTerminationEmailContentTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>%s</title>
  </head>
  <body style="margin: 0; padding: 0">
    <!-- banner start -->
    <div
      style="
        padding: 30px 40px 0;
        height: 148px;
        background-color: #2e3959;
        box-sizing: border-box;
      "
    >
      <img
        style="
          margin-right: 8px;
          width: 48px;
          height: 48px;
          vertical-align: middle;
        "
		src="data:image/png;base64,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"
        alt="logo"
      />
      <div style="display: inline-block; vertical-align: middle">
        <div style="font-size: 18px; color: #fff">海垒</div>
        <a
          style="font-size: 14px; color: #ffffff80; text-decoration: none"
          href="%s"
          target="_blank"
        >
          %s
        </a>
      </div>
    </div>
    <!-- banner-end -->

    <!-- email-start -->
    <div style="margin: -48px 40px 0; box-sizing: border-box">
      <!-- email-header-start -->
      <div
        style="
          padding: 20px;
          border-radius: 2px;
          background-color: #fff;
          box-shadow: 0 4px 10px 2px #979ba54d;
          box-sizing: border-box;
        "
      >
        <h1 style="margin: 0 0 16px; font-size: 18px; font-weight: 700; color: #394567">
          %s
        </h1>
        <div style="margin: 0; padding: 0; width: 100%%; line-height: 22px">
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">业务：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700">%s</span>
            </span>
          </div>
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">提单月份：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700">%s</span>
            </span>
          </div>
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">终止日期：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700">%s</span>
            </span>
          </div>
          <div style="display: inline-block; margin-right: 64px">
            <span
              style="
                display: inline-block;
                width: 3px;
                height: 12px;
                background-color: #699df4;
                border-radius: 2px;
                vertical-align: middle;
              "
            ></span>
            <span style="vertical-align: middle">
              <span style="font-size: 12px; color: #4d4f56">单据状态：</span>
              <span style="font-size: 12px; color: #313238; font-weight: 700">已终止</span>
            </span>
          </div>
        </div>
      </div>
      <!-- email-header-end -->

        <!-- table-start -->
        <div style="margin-top: 16px; width: 100%%">
          <div style="margin-bottom: 10px; font-size: 12px; color: #2e3959">申领详情</div>
          <div style="width: 100%%; overflow-x: auto">
          <table
            style="
              width: 100%%;
              border: 1px solid #dcdee5;
              border-collapse: collapse;
              border-spacing: 0;
              border-radius: 2px;
              text-align: left;
              line-height: 42px;
            "
          >
            <thead
              style="
                border-bottom: 1px solid #dcdee5;
                background-color: #fafbfd;
                color: #313238;
              "
            >
              <tr style="height: 42px">
                <th
                  style="
                    padding: 0 16px;
                    border-bottom: 1px solid #dcdee5;
                    background-color: #fafbfd;
                    font-size: 12px;
                    color: #313238;
                    font-weight: normal;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                  "
                >
                  所属子单号
                </th>
                <th
                  style="
                    padding: 0 16px;
                    border-bottom: 1px solid #dcdee5;
                    background-color: #fafbfd;
                    font-size: 12px;
                    color: #313238;
                    font-weight: normal;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                  "
                >
                  机型
                </th>
                <th
                  style="
                    padding: 0 16px;
                    border-bottom: 1px solid #dcdee5;
                    background-color: #fafbfd;
                    font-size: 12px;
                    color: #313238;
                    font-weight: normal;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                  "
                >
                  地域
                </th>
                <th
                  style="
                    padding: 0 16px;
                    border-bottom: 1px solid #dcdee5;
                    background-color: #fafbfd;
                    font-size: 12px;
                    color: #313238;
                    font-weight: normal;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                  "
                >
                  可用区
                </th>
                 <th
                  style="
                    padding: 0 16px;
                    border-bottom: 1px solid #dcdee5;
                    background-color: #fafbfd;
                    font-size: 12px;
                    color: #313238;
                    font-weight: normal;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                  "
                >
                  实例数量
                </th>
                <th
                  style="
                    padding: 0 16px;
                    border-bottom: 1px solid #dcdee5;
                    background-color: #fafbfd;
                    font-size: 12px;
                    color: #313238;
                    font-weight: normal;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                  "
                >
                  资源分布方式
                </th>
              </tr>
            </thead>
            <tbody>
              %s
            </tbody>
          </table>
        </div>
      </div>
        <!-- table-end -->

        <!-- more-info-start -->
        <div style="padding: 12px 0 32px; font-size: 12px; color: #4d4f56">
            <p style="margin: 0 0 4px;">滚服项目申领规则：</p>
            <p style="margin: 0 0 4px;"><strong>•月度生效原则：</strong>本项目可申领主机资源<strong>按月生效</strong>，
				申领操作<strong>仅限当月有效</strong>，申领单不跨月交付。</p>
            <p style="margin: 0 0 4px;"><strong>•自动清理机制：</strong>为保障额度准确，平台将于<strong>每月1日</strong>
				自动终止(关闭)所有上月遗留的未完成申请。</p>
            <p style="margin: 0 0 4px;"><strong>•后续操作说明：</strong>若您仍有需求，请直接<strong>重新提交当月的新申请单</strong>。
				如有任何疑问，可企微联系ICR(IEG资源服务助手)</p>
        </div>
        <!-- more-info-end -->
    </div>
    <!-- email-end -->
  </body>
</html>`
)
