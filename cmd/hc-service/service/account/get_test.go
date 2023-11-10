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

package account

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/pkg/api/core/cloud"
	hsaccount "hcm/pkg/api/hc-service/account"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/emicklei/go-restful/v3"
)

func Test_TCloudGetResCountBySecret(t *testing.T) {

	sec := cloud.TCloudSecret{
		CloudSecretID:  os.Getenv("TCLOUD_SECRET_ID"),
		CloudSecretKey: os.Getenv("TCLOUD_SECRET_KEY"),
	}

	reqData, err := json.Marshal(sec)
	if err != nil {
		t.Errorf("fail to marshal req, sec:%v, err:%v", sec, err)
		return
	}

	request, err := http.NewRequest("POST", "url", bytes.NewReader(reqData))
	if err != nil {
		t.Errorf("constructing http request fail, err:%v", err)
		return
	}

	svc := &service{
		ad: cloudadaptor.NewCloudAdaptorClient(nil),
	}
	got, err := svc.TCloudGetResCountBySecret(&rest.Contexts{Kit: kit.New(), Request: restful.NewRequest(request)})
	if err != nil {
		t.Errorf("fail to get, err: %v", err)
		return
	}
	wantLen := 8
	if resCount, ok := got.([]hsaccount.ResCountItem); ok {
		if len(resCount) != wantLen {
			t.Errorf("result length mismatch: want: %d, got: %d,%v", wantLen, len(resCount), resCount)
		}
		for i, count := range resCount {
			t.Log(i, count.Type, count.Count)
		}
	} else {
		var want *[]hsaccount.ResCountItem
		t.Errorf("result type mismatch: want: %t, got: %t", any(want), got)
		return
	}

}

func Test_AwsGetResCountBySecret(t *testing.T) {

	sec := cloud.AwsSecret{
		CloudSecretID:  os.Getenv("AWS_SECRET_ID"),
		CloudSecretKey: os.Getenv("AWS_SECRET_KEY"),
	}

	reqData, err := json.Marshal(sec)
	if err != nil {
		t.Errorf("fail to marshal req, sec:%v, err:%v", sec, err)
		return
	}

	request, err := http.NewRequest("POST", "url", bytes.NewReader(reqData))
	if err != nil {
		t.Errorf("constructing http request fail, err:%v", err)
		return
	}

	svc := &service{
		ad: cloudadaptor.NewCloudAdaptorClient(nil),
	}
	got, err := svc.AwsGetResCountBySecret(&rest.Contexts{Kit: kit.New(), Request: restful.NewRequest(request)})
	if err != nil {
		t.Errorf("fail to get, err: %v", err)
		return
	}
	wantLen := 7
	if resCount, ok := got.([]hsaccount.ResCountItem); ok {
		if len(resCount) != wantLen {
			t.Fatalf("result length mismatch: want: %d, got: %d,%v", wantLen, len(resCount), resCount)
		}
		for i, count := range resCount {
			t.Log(i, count.Type, count.Count)
		}
	} else {
		var want *[]hsaccount.ResCountItem
		t.Errorf("result type mismatch: want: %t, got: %t", any(want), got)
		return
	}

}
