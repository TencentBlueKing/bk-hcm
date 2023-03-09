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

package sync

import (
	"net/http"
	"time"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/TencentBlueKing/gopkg/collection/set"
)

// InitSyncService initial the sync service
func InitSyncService(c *capability.Capability) {
	svc := &syncSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("SyncAll", http.MethodPost, "/vendors/accounts/all/sync", svc.SyncAll)

	h.Load(c.WebService)
}

type syncSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// SyncAll sync all resource.
func (svc syncSvc) SyncAll(cts *rest.Contexts) (interface{}, error) {

	err := SyncAllResource(svc.client, cts.Kit, cts.Kit.Header())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// SyncTiming timing sync all resource
func SyncTiming(c *client.ClientSet) {

	kit := kit.New()
	header := http.Header{
		constant.UserKey:    []string{constant.SyncTimingUserKey},
		constant.RidKey:     []string{kit.Rid},
		constant.AppCodeKey: []string{constant.SyncTimingAppCodeKey},
	}

	for {

		logs.Infof("syncTiming all begin, rid: %s", kit.Rid)

		err := SyncAllResource(c, kit, header)
		if err != nil {
			logs.Errorf("sync SyncAllResource failed, err: %v, rid: %s", err, kit.Rid)
		}

		logs.Infof("syncTiming all resource end , rid: %s", kit.Rid)

		time.Sleep(time.Second * 600)
	}
}

// SyncAllResource sync all resource
func SyncAllResource(c *client.ClientSet, kit *kit.Kit, header http.Header) error {

	result, err := c.DataService().Global.Account.List(
		kit.Ctx,
		header,
		&dataproto.AccountListReq{
			Filter: tools.EqualExpression("type", constant.SyncTimingAccountResource),
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	)
	if err != nil {
		logs.Errorf("sync list all account failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	err = SyncPublicResource(kit, c, result, header)
	if err != nil {
		logs.Errorf("sync public resource failed, err: %v, rid: %s", err, kit.Rid)
	}

	for _, account := range result.Details {
		switch account.Vendor {
		case enumor.TCloud:
			err = SyncTCloudAll(c, kit, header, account.ID)
		case enumor.Aws:
			err = SyncAwsAll(c, kit, header, account.ID)
		case enumor.HuaWei:
			err = SyncHuaWeiAll(c, kit, header, account.ID)
		case enumor.Azure:
			err = SyncAzureAll(c, kit, header, account.ID)
		case enumor.Gcp:
			err = SyncGcpAll(c, kit, header, account.ID)
		default:
			logs.Errorf("sync vendor %s not support, rid: %s", account.Vendor, kit.Rid)
		}
	}

	return err
}

// SyncPublicResource sync all public resource
func SyncPublicResource(kit *kit.Kit, c *client.ClientSet, accounts *dataproto.AccountListResult,
	header http.Header) error {

	accountSet := set.NewStringSet()

	for _, account := range accounts.Details {
		switch account.Vendor {
		case enumor.TCloud:
			if accountSet.Has(string(enumor.TCloud)) {
				continue
			}
			accountSet.Add(string(enumor.TCloud))
			err := SyncTCloudPublicResource(kit, c, header, account.ID)
			if err != nil {
				logs.Errorf("sync do tcloud sync public resource failed, err: %v, rid: %s", err, kit.Rid)
			}
		case enumor.HuaWei:
			if accountSet.Has(string(enumor.HuaWei)) {
				continue
			}
			accountSet.Add(string(enumor.HuaWei))
			err := SyncHuaWeiPublicResource(kit, c, header, account.ID)
			if err != nil {
				logs.Errorf("sync do huawei sync public resource failed, err: %v, rid: %s", err, kit.Rid)
			}
		case enumor.Aws:
			if accountSet.Has(string(enumor.Aws)) {
				continue
			}
			accountSet.Add(string(enumor.Aws))
			err := SyncAwsPublicResource(kit, c, header, account.ID)
			if err != nil {
				logs.Errorf("sync do aws sync public resource failed, err: %v, rid: %s", err, kit.Rid)
			}
		case enumor.Gcp:
			if accountSet.Has(string(enumor.Gcp)) {
				continue
			}
			accountSet.Add(string(enumor.Gcp))
			err := SyncGcpPublicResource(kit, c, header, account.ID)
			if err != nil {
				logs.Errorf("sync do gcp sync public resource failed, err: %v, rid: %s", err, kit.Rid)
			}
		case enumor.Azure:
			if accountSet.Has(string(enumor.Azure)) {
				continue
			}
			accountSet.Add(string(enumor.Azure))
			err := SyncAzurePublicResource(kit, c, header, account.ID)
			if err != nil {
				logs.Errorf("sync do azure sync public resource failed, err: %v, rid: %s", err, kit.Rid)
			}
		default:
			logs.Errorf("sync vendor %s not support, rid: %s", account.Vendor, kit.Rid)
		}
	}

	return nil
}
