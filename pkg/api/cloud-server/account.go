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

// Package cloudserver defines cloud-server api call protocols.
package cloudserver

// type AccountExtension interface {
// 	Validate() error
// 	ConvertToDataServiceExtension() cloud.AccountExtension
// }
//
// type TCloudAccountExtension struct {
// 	MainAccount string `json:"main_account" validate:"required"`
// 	SubAccount  string `json:"sub_account" validate:"required"`
// 	SecretID    string `json:"secret_id" validate:"required"`
// 	SecretKey   string `json:"secret_key" validate:"required"`
// }
//
// func (e TCloudAccountExtension) Validate() error {
// 	return validator.Validate.Struct(e)
// }
//
// func (e TCloudAccountExtension) ConvertToDataServiceExtension() cloud.AccountExtension {
// 	return cloud.AccountExtension{
// 		TCloud: &cloud.TCloudAccountExtension{
// 			MainAccountCid: e.MainAccount,
// 			SubAccountCid:  e.SubAccount,
// 			Secret: &cloud.BaseSecret{
// 				Cid: e.SecretID,
// 				Key: e.SecretKey,
// 			},
// 		},
// 	}
// }
//
// type AwsAccountExtension struct {
// 	AccountID   string `json:"account_id" validate:"required"`
// 	IamUsername string `json:"iam_username" validate:"required"`
// 	SecretID    string `json:"secret_id" validate:"required"`
// 	SecretKey   string `json:"secret_key" validate:"required"`
// }
//
// func (e AwsAccountExtension) Validate() error {
// 	return validator.Validate.Struct(e)
// }
//
// func (e AwsAccountExtension) ConvertToDataServiceExtension() cloud.AccountExtension {
// 	return cloud.AccountExtension{
// 		Aws: &cloud.AwsAccountExtension{},
// 	}
// }
//
// // CreateAccountReq defines create cloud account http request.
// type CreateAccountReq struct {
// 	Vendor enumor.Vendor `json:"vendor" validate:"required"`
// 	// FIXME: 没法统一 Spec,Extension,Attachment里的数据做进一步校验
// 	Spec       *cloud.AccountSpec       `json:"spec" validate:"required"`
// 	Extension  json.RawMessage          `json:"extension" validate:"required"`
// 	Attachment *cloud.AccountAttachment `json:"attachment" validate:"required"`
// }
//
// // Validate create account request.
// func (req CreateAccountReq) Validate() error {
// 	return validator.Validate.Struct(req)
// }
//
// func (req CreateAccountReq) UnmarshalExtension() (AccountExtension, error) {
//
// }
//
// // ListAccountReq ...
// type ListAccountReq struct {
// 	Filter *filter.Expression `json:"filter" validate:"omitempty"`
// 	Page   *types.BasePage    `json:"page" validate:"required"`
// }
//
// // Validate ...
// func (l *ListAccountReq) Validate() error {
// 	return validator.Validate.Struct(l)
// }
//
// // ListAccountResult defines list instances for iam pull resource callback result.
// type ListAccountResult struct {
// 	Count   uint64           `json:"count,omitempty"`
// 	Details []*cloud.Account `json:"details,omitempty"`
// }
//
// type CheckAccountReq struct {
// 	Vendor    enumor.Vendor           `json:"vendor" validate:"required"`
// 	Extension *cloud.AccountExtension `json:"extension" validate:"required"`
// }
