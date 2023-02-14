// Package iam TODO
/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */
package iam

import (
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/thirdparty/esb/types"
)

type esbInstanceWithCreatorParams struct {
	*types.CommParams           `json:",inline"`
	*client.InstanceWithCreator `json:",inline"`
}

type esbIamCreatorActionResp struct {
	types.BaseResponse `json:",inline"`
	Data               []client.CreatorActionPolicy `json:"data"`
}

type esbIamGetApplyPermUrlParams struct {
	*types.CommParams   `json:",inline"`
	*meta.IamPermission `json:",inline"`
}

type esbIamGetApplyPermUrlResp struct {
	types.BaseResponse `json:",inline"`
	Data               GetApplyPermUrlResult `json:"data"`
}

type GetApplyPermUrlResult struct {
	Url string `json:"url"`
}
