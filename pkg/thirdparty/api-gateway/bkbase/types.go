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

package bkbase

import (
	"encoding/json"
	"time"

	"hcm/pkg/tools/times"
)

const (
	// CodeSuccess is the response code which represents a success request.
	CodeSuccess string = "00"

	// DefaultQueryLimit is the bkbase default query batch size.
	DefaultQueryLimit int = 5000

	// DateLayout layout for bkbase date, '%Y%m%d'
	DateLayout = "20060102"
)

// Date is used to specify the date range in a BKBase SQL query.
type Date time.Time

// NewDate returns a new instance of Date from given time.Time instance
func NewDate(t time.Time) *Date {
	date := Date(t)
	return &date
}

// Layout return the layout of Date
func (d *Date) Layout() string {
	return DateLayout
}

// Format  as a string ('%Y%m%d')
func (d *Date) Format() string {
	return time.Time(*d).Format(DateLayout)
}

// String implements fmt.Stringer, used in fmt.Sprintf
func (d *Date) String() string {
	return d.Format()
}

// DateBefore returns a date that is prior to the current time.
// It is intended for use in comparisons with the `thedate` column for the purpose of filtering by a date range.
func DateBefore(nDays int) *Date {
	return NewDate(time.Now().Add(times.Day * time.Duration(-nDays)))
}

// QuerySyncReq query sync bkbase data request
type QuerySyncReq struct {
	AuthMethod    string `json:"bkdata_authentication_method"`
	DataToken     string `json:"bkdata_data_token"`
	AppCode       string `json:"bk_app_code"`
	AppSecret     string `json:"bk_app_secret"`
	PreferStorage string `json:"prefer_storage"`
	Sql           string `json:"sql"`
}

// QuerySyncResp query sync bkbase data response
type QuerySyncResp struct {
	Result  bool         `json:"result"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    QuerySyncRst `json:"data"`
}

// QuerySyncRst query sync bkbase data result
type QuerySyncRst struct {
	Cluster      string          `json:"cluster"`
	TotalRecords int             `json:"totalRecords"`
	TimeTaken    float64         `json:"timetaken"`
	Sql          string          `json:"sql"`
	List         json.RawMessage `json:"list"`
}
