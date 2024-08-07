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

package mainaccount

const (
	GcpLoginAddress      = "https://console.cloud.google.com/welcome?project=%s"
	AwsLoginAddress      = "https://signin.aws.amazon.com/"
	HuaweiLoginAddress   = "https://auth.huaweicloud.com/authui/login.html?service=https://console.huaweicloud.com"
	AzureLoginAddress    = "https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/Overview"
	ZenlayerLoginAddress = "https://console.zenlayer.com/auth/login"
	KaopuLoginAddress    = "https://console.kaopuyun.com/user/#/login"

	EmailTitleTemplate   = "【HCM】 %s账号创建成功通知"
	EmailContentTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>云账号创建成功通知</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #fff;
            padding: 20px;
            border-radius: 5px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
        }
        h1 {
            color: #007bff;
        }
        .account-info {
            background-color: #f8f9fa;
            padding: 10px;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>二级账号创建成功通知</h1>
        <p>您好，您已成功创建二级账号，以下是账号信息，请妥善保管：</p>
        <div class="account-info">
            <p>云厂商：<strong>%s</strong></p>
            <p>账号名：<strong>%s</strong></p>
            <p>账号ID：<strong>%s</strong></p>
        </div>
        <p>请使用以下地址登录云厂商：<a href="%s">%s</a></p>
        <p>如果您有任何疑问，请联系海垒平台管理员。</p>
        <p>该邮件由系统自动发出，请勿回复。</p>
    </div>
</body>
</html>`
)
