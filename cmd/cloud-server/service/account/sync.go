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
	"fmt"

	"hcm/cmd/cloud-server/service/sync/aws"
	"hcm/cmd/cloud-server/service/sync/azure"
	"hcm/cmd/cloud-server/service/sync/gcp"
	"hcm/cmd/cloud-server/service/sync/huawei"
	"hcm/cmd/cloud-server/service/sync/tcloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Sync ...
func (a *accountSvc) Sync(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的更新权限
	if err := a.checkPermission(cts, meta.Update, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.AccountCloudResType,
		accountID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// TODO：添加审计

	switch baseInfo.Vendor {
	case enumor.TCloud:
		opt := &tcloud.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: true,
		}
		err = tcloud.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.Aws:
		opt := &aws.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: true,
		}
		err = aws.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.HuaWei:
		opt := &huawei.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: true,
		}
		err = huawei.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.Gcp:
		opt := &gcp.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: true,
		}
		err = gcp.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.Azure:
		opt := &azure.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: true,
		}
		err = azure.SyncAllResource(cts.Kit, a.client, opt)

	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}
	if err != nil {
		logs.Errorf("sync all resource failed, err: %v, vendor: %s, rid: %s", err, baseInfo.Vendor, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
