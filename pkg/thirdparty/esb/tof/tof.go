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

package tof

import (
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/types"
)

// Client  is an esb client to request tof.
type Client interface {
	// GetStaffInfo 根据员工英文名获取员工信息
	GetStaffInfo(kt *kit.Kit, loginName string) (*StaffInfo, error)
}

// NewClient initialize a new tof client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &tof{
		client: client,
		config: config,
	}
}

var _ Client = new(tof)

// tof is an esb client to request tof.
type tof struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

// GetStaffInfo 根据员工英文名获取员工信息
func (t *tof) GetStaffInfo(kt *kit.Kit, loginName string) (*StaffInfo, error) {
	if loginName == "" {
		return nil, fmt.Errorf("login_name is empty")
	}

	resp := new(GetStaffInfoResp)

	// 构建查询参数
	params := make(map[string]string)
	params["login_name"] = loginName

	// 构建请求头
	h := http.Header{}
	h.Set(constant.RidKey, kt.Rid)
	types.SetCommonHeader(&h, t.config)

	err := t.client.Get().
		SubResourcef("/tof/get_staff_direct_leader/").
		WithContext(kt.Ctx).
		WithHeaders(h).
		WithParams(params).
		Do().Into(resp)

	if err != nil {
		logs.Errorf("get staff direct leader failed, login_name: %s, err: %v, rid: %s", loginName, err, kt.Rid)
		return nil, err
	}

	if !resp.Result || resp.Code != "00" {
		logs.Errorf("tof returns error, login_name: %s, code: %s, msg: %s, rid: %s",
			loginName, resp.Code, resp.Message, kt.Rid)
		return nil, fmt.Errorf("tof api returns err, code: %s, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, nil
}
