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

// Package usermgr defines esb client to request usermgr.
package usermgr

import (
	"fmt"
	"net/http"
	"strconv"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// Client is an ApiGateway client to request usermgr.
type Client interface {
	ListUserMgrDepartment(kt *kit.Kit, req *ListDepartmentParams) (*ListDeptResult, error)
	ListAllDepartment(kt *kit.Kit) (map[string]*DeptInfo, error)
}

// NewClient initialize a new usermgr client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}
	c := &client.Capability{
		Client: cli,
		Discover: &apigateway.Discovery{
			Name:    "userManagerApiGateWay",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/api/c/compapi/v2")

	return &usermgr{
		config: cfg,
		client: restCli,
	}, nil
}

// usermgr is an APIGateway client to request usermgr.
type usermgr struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
}

func (um *usermgr) header(kt *kit.Kit) http.Header {
	header := http.Header{}
	header.Set(constant.RidKey, kt.Rid)
	header.Set(constant.BKGWAuthKey, um.config.GetAuthValue())
	return header
}

// ListUserMgrDepartment list usermgr department.
func (um *usermgr) ListUserMgrDepartment(kt *kit.Kit, req *ListDepartmentParams) (*ListDeptResult, error) {
	resp := new(ListDepartmentResp)
	cli := um.client.Get().
		SubResourcef("/usermanage/list_departments").
		WithContext(kt.Ctx).
		WithHeaders(um.header(kt))
	if req.Page > 0 {
		cli.WithParam("page", strconv.FormatInt(req.Page, 10))
	}
	if req.PageSize > 0 {
		cli.WithParam("page_size", strconv.FormatInt(req.PageSize, 10))
	}
	if len(req.LookupField) > 0 {
		cli.WithParam("lookup_field", req.LookupField)
	}
	if len(req.ExactLookups) > 0 {
		cli.WithParam("exact_lookups", req.ExactLookups)
	}
	err := cli.Do().Into(resp)
	if err != nil {
		logs.Errorf("list usermgr department error, req: %+v, resp: %+v, err: %+v, rid: %s", req, resp, err, resp.Rid)
		return nil, err
	}

	if resp.IsFailed() {
		logs.Errorf("list usermgr department failed, req: %+v, resp: %+v, err: %+v, respRid: %s, rid: %s",
			req, resp, err, resp.Rid, kt.Rid)
		return nil, fmt.Errorf("list usermgr department failed, req: %+v, code: %d, msg: %s, respRid: %s",
			req, resp.Code, resp.Message, resp.Rid)
	}

	return resp.ListDeptResult, nil
}

// ListAllDepartment 获取所有部门信息
func (um *usermgr) ListAllDepartment(kt *kit.Kit) (map[string]*DeptInfo, error) {
	var deptMap = make(map[string]*DeptInfo)
	page := int64(1)
	pageSize := int64(constant.DeptQueryDBMaxNum)
	for {
		umReq := &ListDepartmentParams{
			Page:     page,
			PageSize: pageSize,
		}
		list, dErr := um.ListUserMgrDepartment(kt, umReq)
		if dErr != nil {
			logs.Errorf("list all dept by usermgr failed, err: %v, rid: %s", dErr, kt.Rid)
			return nil, dErr
		}

		for _, item := range list.Results {
			// 跳过 level == 0 && item.HasChildren == false 的item
			if item.Level == 0 && !item.HasChildren {
				continue
			}

			deptMap[strconv.FormatInt(item.ID, 10)] = item
		}

		if (page * pageSize) >= list.Count {
			break
		}
		page++
	}

	return deptMap, nil
}
