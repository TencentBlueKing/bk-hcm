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

package rest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/emicklei/go-restful/v3"
)

// Contexts request context.
type Contexts struct {
	Kit            *kit.Kit
	Request        *restful.Request
	resp           *restful.Response
	respStatusCode int

	// request meta info
	bizID string
}

// RequestBody 返回拷贝的body内容
func (c *Contexts) RequestBody() ([]byte, error) {
	byt, err := ioutil.ReadAll(c.Request.Request.Body)
	if err != nil {
		return nil, err
	}

	c.Request.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byt))

	return byt, nil
}

// ReDecodeInto 解析Body到结构体中，且把body内容重新写入Body.
func (c *Contexts) ReDecodeInto(to interface{}) error {
	byt, err := ioutil.ReadAll(c.Request.Request.Body)
	if err != nil {
		return err
	}

	c.Request.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byt))

	err = json.Unmarshal(byt, to)
	if err != nil {
		logs.ErrorDepthf(1, "decode request body failed, err: %s, rid: %s", err.Error(), c.Kit.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil
}

// DecodeInto decode request body to a struct, if failed, then return the
// response with an error
func (c *Contexts) DecodeInto(to interface{}) error {
	err := json.NewDecoder(c.Request.Request.Body).Decode(to)
	if err != nil {
		logs.ErrorDepthf(1, "decode request body failed, err: %s, rid: %s", err.Error(), c.Kit.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil
}

// PathParameter get path parameter value by its name.
func (c *Contexts) PathParameter(name string) PathParam {
	return PathParam(c.Request.PathParameter(name))
}

// DecodePathParamInto decode path parameter value into its actual type.
func (c *Contexts) DecodePathParamInto(name string, to interface{}) error {
	param := c.PathParameter(name)
	err := json.Unmarshal([]byte(param), &to)
	if err != nil {
		logs.ErrorDepthf(1, "decode path parameter %s failed, err: %v, rid: %s", param, err, c.Kit.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return nil
}

// PathParam defines path parameter value type.
type PathParam string

// String convert path parameter to string.
func (p PathParam) String() string {
	return string(p)
}

// Uint64 convert path parameter to uint64.
func (p PathParam) Uint64() (uint64, error) {
	value, err := strconv.ParseUint(string(p), 10, 64)
	if err != nil {
		logs.ErrorDepthf(1, "decode path parameter %s failed, err: %v", p, err)
		return 0, errf.NewFromErr(errf.InvalidParameter, err)
	}
	return value, nil
}

// Int64 convert path parameter to int64.
func (p PathParam) Int64() (int64, error) {
	value, err := strconv.ParseInt(string(p), 10, 64)
	if err != nil {
		logs.ErrorDepthf(1, "decode path parameter %s failed, err: %v", p, err)
		return 0, errf.NewFromErr(errf.InvalidParameter, err)
	}
	return value, nil
}

// WithStatusCode set the response status header code
func (c *Contexts) WithStatusCode(statusCode int) *Contexts {
	c.respStatusCode = statusCode
	return c
}

// respEntity response request with a success response.
func (c *Contexts) respEntity(data interface{}) {
	if c.respStatusCode != 0 {
		c.resp.WriteHeader(c.respStatusCode)
	}

	c.resp.Header().Set(constant.RidKey, c.Kit.Rid)
	c.resp.AddHeader(restful.HEADER_ContentType, restful.MIME_JSON)

	resp := &Response{
		Code:    errf.OK,
		Message: "",
		Data:    data,
	}

	if err := json.NewEncoder(c.resp.ResponseWriter).Encode(resp); err != nil {
		logs.ErrorDepthf(1, "do response failed, err: %s, rid: %s", err.Error(), c.Kit.Rid)
		return
	}

	return
}

// respError response request with error response.
func (c *Contexts) respError(err error) {
	if c.respStatusCode > 0 {
		c.resp.WriteHeader(c.respStatusCode)
	}

	if c.Kit != nil {
		c.resp.Header().Set(constant.RidKey, c.Kit.Rid)
	}

	resp := errf.Error(err).Resp()

	encodeErr := json.NewEncoder(c.resp.ResponseWriter).Encode(resp)
	if encodeErr != nil {
		logs.ErrorDepthf(1, "response with error failed, err: %v, rid: %s", encodeErr, c.Kit.Rid)
		return
	}

	return
}

// respErrorWithEntity response request with error response.
func (c *Contexts) respErrorWithEntity(data interface{}, err error) {
	if c.respStatusCode != 0 {
		c.resp.WriteHeader(c.respStatusCode)
	}

	c.resp.Header().Set(constant.RidKey, c.Kit.Rid)
	c.resp.AddHeader(restful.HEADER_ContentType, restful.MIME_JSON)

	parsedErr := errf.Error(err)
	resp := &Response{
		Code:    parsedErr.Code,
		Message: parsedErr.Message,
		Data:    data,
	}

	if err := json.NewEncoder(c.resp.ResponseWriter).Encode(resp); err != nil {
		logs.ErrorDepthf(1, "do response failed, err: %s, rid: %s", err.Error(), c.Kit.Rid)
		return
	}

	return
}
