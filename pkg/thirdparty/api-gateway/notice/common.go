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

package notice

// GetCurAnnResp the response of the get current announcements
type GetCurAnnResp = []CurAnnData

// CurAnnData the data of the current announcements
type CurAnnData struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	ContentList  []Content `json:"content_list"`
	AnnounceType string    `json:"announce_type"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
}

// Content the content of the current announcements
type Content struct {
	Content  string `json:"content"`
	Language string `json:"language"`
}

// RegAppData the data of the register application
type RegAppData struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
