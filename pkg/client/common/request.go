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

package common

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// Empty alias for struct{} type, use for request without request or response data. e.g.
//  1. without response body: Request[T,Empty](cli,method,kt,req,url,args...)
//  2. without request body: Request[Empty,R](cli,method,kt,NoData,req,url,args...)
//  3. without both request and response body: Request[Empty,Empty](cli,method,kt,NoData,req,url,args...)
//
// see also its instance pointer NoData
type Empty = struct{}

// NoData is a pointer to Empty type instance, use it as placeholder for request without no request body.
// e.g. Request[Empty,R](cli,method,kt,NoData,url,args...)
var NoData = &Empty{}

// RequestNoResp is a helper method to build request without response body,
// same as Request[T,Empty](cli,method,kt,req,url,args...) but ignore response
func RequestNoResp[T any](cli rest.ClientInterface, method rest.VerbType, kt *kit.Kit, req *T,
	url string, urlArgs ...any) error {
	_, err := Request[T, Empty](cli, method, kt, req, url, urlArgs...)
	return err
}

// Request is a general-purpose helper method to build request.
// Type parameter `T` is the type of request type, and `R` is the type of response.
func Request[T any, R any](cli rest.ClientInterface, method rest.VerbType, kt *kit.Kit, req *T,
	url string, urlArgs ...any) (*R, error) {

	resp := new(core.BaseResp[*R])

	err := cli.Verb(method).
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef(url, urlArgs...).
		WithHeaders(kt.Header()).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}
