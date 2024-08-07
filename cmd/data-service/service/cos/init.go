/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package cos

import (
	"net/http"
	"time"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/data-service/cos"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/objectstore"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitService initialize the raw bill service
func InitService(cap *capability.Capability) {
	svc := &service{
		ostore: cap.ObjectStore,
	}
	h := rest.NewHandler()
	h.Add("GenerateTemporalUrl", http.MethodPost, "/cos/temporal_urls/{action}/generate", svc.GenerateTemporalUrl)

	h.Load(cap.WebService)
}

type service struct {
	ostore objectstore.Storage
}

// GenerateTemporalUrl ...
func (s service) GenerateTemporalUrl(cts *rest.Contexts) (any, error) {
	req := new(cos.GenerateTemporalUrlReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	action := objectstore.OperateAction(cts.PathParameter("action"))
	cred, url, err := s.ostore.GetPreSignedURL(cts.Kit, action, time.Second*time.Duration(req.TTLSeconds), req.Filename)
	if err != nil {
		logs.Errorf("fail to get presigned download URL, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return &cos.GenerateTemporalUrlResult{
		AK:    cred.TmpSecretID,
		Token: cred.SessionToken,
		URL:   url,
	}, nil
}
