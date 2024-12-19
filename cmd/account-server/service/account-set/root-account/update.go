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

package rootaccount

import (
	"fmt"

	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/account-server/account-set"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// UpdateRootAccount update root account with options
func (s *service) UpdateRootAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.RootAccountUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的更新权限
	if err := s.checkPermission(cts, meta.RootAccount, meta.Find); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := s.client.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var (
		result interface{}
	)

	switch baseInfo.Vendor {
	case enumor.Aws:
		result, err = s.updateForAws(cts, req, accountID)
	case enumor.HuaWei:
		result, err = s.updateForHuaWei(cts, req, accountID)
	case enumor.Gcp:
		result, err = s.updateForGcp(cts, req, accountID)
	case enumor.Azure:
		result, err = s.updateForAzure(cts, req, accountID)
	case enumor.Zenlayer:
		result, err = s.updateForZenlayer(cts, req, accountID)
	case enumor.Kaopu:
		result, err = s.updateForKaopu(cts, req, accountID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err != nil {
		err := fmt.Errorf("update [%s] root account error: %s, rid: %s", baseInfo.Vendor, err.Error(), cts.Kit.Rid)
		logs.Errorf(err.Error())
		return nil, err
	}

	return result, nil
}

func (s *service) updateForAws(cts *rest.Contexts, req *proto.RootAccountUpdateReq, accountID string) (interface{}, error) {
	var (
		extension *proto.AwsRootAccountExtensionUpdateReq
	)
	if req.Extension != nil {
		// 解析Extension
		extension = new(proto.AwsRootAccountExtensionUpdateReq)
		if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		// 校验Extension
		err := extension.Validate()
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	var shouldUpdatedExtension *dataproto.AwsRootAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.AwsRootAccountExtensionUpdateReq{
			CloudIamUsername: extension.CloudIamUsername,
			CloudSecretID:    &extension.CloudSecretID,
			CloudSecretKey:   &extension.CloudSecretKey,
		}
	}

	// 更新
	_, err := s.client.DataService().Aws.RootAccount.Update(
		cts.Kit,
		accountID,
		&dataproto.RootAccountUpdateReq[dataproto.AwsRootAccountExtensionUpdateReq]{
			Name:        req.Name,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Memo:        req.Memo,
			DeptID:      req.DeptID,
			Extension:   shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (s *service) updateForGcp(cts *rest.Contexts, req *proto.RootAccountUpdateReq, accountID string) (interface{}, error) {
	var (
		extension *proto.GcpRootAccountExtensionUpdateReq
	)
	if req.Extension != nil {
		// 解析Extension
		extension = new(proto.GcpRootAccountExtensionUpdateReq)
		if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		// 校验Extension
		err := extension.Validate()
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	var shouldUpdatedExtension *dataproto.GcpRootAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.GcpRootAccountExtensionUpdateReq{
			CloudProjectName:        extension.CloudProjectName,
			CloudServiceAccountID:   &extension.CloudServiceAccountID,
			CloudServiceAccountName: &extension.CloudServiceAccountName,
			CloudServiceSecretID:    &extension.CloudServiceSecretID,
			CloudServiceSecretKey:   &extension.CloudServiceSecretKey,
		}
	}

	// 更新
	_, err := s.client.DataService().Gcp.RootAccount.Update(
		cts.Kit,
		accountID,
		&dataproto.RootAccountUpdateReq[dataproto.GcpRootAccountExtensionUpdateReq]{
			Name:        req.Name,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Memo:        req.Memo,
			DeptID:      req.DeptID,
			Extension:   shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (s *service) updateForAzure(cts *rest.Contexts, req *proto.RootAccountUpdateReq, accountID string) (interface{}, error) {
	var (
		extension *proto.AzureRootAccountExtensionUpdateReq
	)
	if req.Extension != nil {
		// 解析Extension
		extension = new(proto.AzureRootAccountExtensionUpdateReq)
		if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		// 校验Extension
		err := extension.Validate()
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	var shouldUpdatedExtension *dataproto.AzureRootAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.AzureRootAccountExtensionUpdateReq{
			CloudTenantID:         extension.CloudTenantID,
			CloudSubscriptionName: extension.CloudSubscriptionName,
			CloudApplicationID:    &extension.CloudApplicationID,
			CloudApplicationName:  &extension.CloudApplicationName,
			CloudClientSecretKey:  &extension.CloudClientSecretKey,
		}
	}

	// 更新
	_, err := s.client.DataService().Azure.RootAccount.Update(
		cts.Kit,
		accountID,
		&dataproto.RootAccountUpdateReq[dataproto.AzureRootAccountExtensionUpdateReq]{
			Name:        req.Name,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Memo:        req.Memo,
			DeptID:      req.DeptID,
			Extension:   shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (s *service) updateForHuaWei(cts *rest.Contexts, req *proto.RootAccountUpdateReq, accountID string) (interface{}, error) {
	var (
		extension *proto.HuaWeiRootAccountExtensionUpdateReq
	)
	if req.Extension != nil {
		// 解析Extension
		extension = new(proto.HuaWeiRootAccountExtensionUpdateReq)
		if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		// 校验Extension
		err := extension.Validate()
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	var shouldUpdatedExtension *dataproto.HuaWeiRootAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.HuaWeiRootAccountExtensionUpdateReq{
			CloudSubAccountName: extension.CloudSubAccountName,
			CloudIamUserID:      extension.CloudIamUserID,
			CloudIamUsername:    extension.CloudIamUsername,
			CloudSecretID:       &extension.CloudSecretID,
			CloudSecretKey:      &extension.CloudSecretKey,
		}
	}

	// 更新
	_, err := s.client.DataService().HuaWei.RootAccount.Update(
		cts.Kit,
		accountID,
		&dataproto.RootAccountUpdateReq[dataproto.HuaWeiRootAccountExtensionUpdateReq]{
			Name:        req.Name,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Memo:        req.Memo,
			DeptID:      req.DeptID,
			Extension:   shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (s *service) updateForZenlayer(cts *rest.Contexts, req *proto.RootAccountUpdateReq, accountID string) (interface{}, error) {
	var (
		extension *proto.ZenlayerRootAccountExtensionUpdateReq
	)
	if req.Extension != nil {
		// 解析Extension
		extension = new(proto.ZenlayerRootAccountExtensionUpdateReq)
		if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		// 校验Extension
		err := extension.Validate()
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	var shouldUpdatedExtension *dataproto.ZenlayerRootAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.ZenlayerRootAccountExtensionUpdateReq{}
	}

	// 更新
	_, err := s.client.DataService().Zenlayer.RootAccount.Update(
		cts.Kit,
		accountID,
		&dataproto.RootAccountUpdateReq[dataproto.ZenlayerRootAccountExtensionUpdateReq]{
			Name:        req.Name,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Memo:        req.Memo,
			DeptID:      req.DeptID,
			Extension:   shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

func (s *service) updateForKaopu(cts *rest.Contexts, req *proto.RootAccountUpdateReq, accountID string) (interface{}, error) {
	var (
		extension *proto.KaopuRootAccountExtensionUpdateReq
	)
	if req.Extension != nil {
		// 解析Extension
		extension = new(proto.KaopuRootAccountExtensionUpdateReq)
		if err := common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		// 校验Extension
		err := extension.Validate()
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}
	var shouldUpdatedExtension *dataproto.KaopuRootAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.KaopuRootAccountExtensionUpdateReq{}
	}

	// 更新
	_, err := s.client.DataService().Kaopu.RootAccount.Update(
		cts.Kit,
		accountID,
		&dataproto.RootAccountUpdateReq[dataproto.KaopuRootAccountExtensionUpdateReq]{
			Name:        req.Name,
			Managers:    req.Managers,
			BakManagers: req.BakManagers,
			Memo:        req.Memo,
			DeptID:      req.DeptID,
			Extension:   shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}
