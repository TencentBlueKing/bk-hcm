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

package gcp

import (
	"errors"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	dataservice "hcm/pkg/api/data-service"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncSubAccountOption define sync account option.
type SyncSubAccountOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate SyncSubAccountOption
func (opt SyncSubAccountOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SubAccount sync sub account.
func (cli *client) SubAccount(kt *kit.Kit, opt *SyncSubAccountOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	fromCloud, err := cli.listAccountFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	fromDB, err := cli.listAccountFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[account.GcpAccount,
		coresubaccount.SubAccount[coresubaccount.GcpExtension]](fromCloud, fromDB, isSubAccountChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubAccount(kt, opt, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createSubAccount(kt, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateSubAccount(kt, opt, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) updateSubAccount(kt *kit.Kit, opt *SyncSubAccountOption,
	updateMap map[string]account.GcpAccount) error {

	if len(updateMap) <= 0 {
		return errors.New("updateMap is required")
	}

	account, err := cli.dbCli.Gcp.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	updateItems := make([]dssubaccount.UpdateField, 0, len(updateMap))

	for id, one := range updateMap {
		extension := &coresubaccount.GcpExtension{
			CloudProjectID:   account.Extension.CloudProjectID,
			CloudProjectName: account.Extension.CloudProjectName,
		}

		ext, err := core.MarshalStruct(extension)
		if err != nil {
			return err
		}

		tmpRes := dssubaccount.UpdateField{
			ID:        id,
			Name:      one.Name,
			Vendor:    enumor.Gcp,
			Site:      account.Site,
			AccountID: account.ID,
			Extension: ext,
			// Managers/BizIDs由用户设置不继承资源账号。
			Managers: nil,
			BkBizIDs: nil,
			Memo:     nil,
		}
		updateItems = append(updateItems, tmpRes)
	}

	updateReq := &dssubaccount.UpdateReq{
		Items: updateItems,
	}
	if err = cli.dbCli.Global.SubAccount.BatchUpdate(kt, updateReq); err != nil {
		logs.Errorf("[%s] update sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sub account to update sub account success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSubAccount(kt *kit.Kit, opt *SyncSubAccountOption, addSlice []account.GcpAccount) error {

	if len(addSlice) <= 0 {
		return errors.New("addSlice is required")
	}

	account, err := cli.dbCli.Gcp.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	createResources := make([]dssubaccount.CreateField, 0, len(addSlice))

	for _, one := range addSlice {
		extension := &coresubaccount.GcpExtension{
			CloudProjectID:   account.Extension.CloudProjectID,
			CloudProjectName: account.Extension.CloudProjectName,
		}

		ext, err := core.MarshalStruct(extension)
		if err != nil {
			return err
		}

		tmpRes := dssubaccount.CreateField{
			CloudID:   one.Name,
			Name:      one.Name,
			Vendor:    enumor.Gcp,
			Site:      account.Site,
			AccountID: account.ID,
			Extension: ext,
			// Managers/BizIDs由用户设置不继承资源账号。
			Managers: nil,
			BkBizIDs: nil,
			Memo:     nil,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dssubaccount.CreateReq{
		Items: createResources,
	}
	if _, err = cli.dbCli.Global.SubAccount.BatchCreate(kt, createReq); err != nil {
		logs.Errorf("[%s] create sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sub account to create sub account success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) deleteSubAccount(kt *kit.Kit, opt *SyncSubAccountOption, delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return errors.New("delCloudIDs is required")
	}

	delFromCloud, err := cli.listAccountFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range delFromCloud {
		if _, exsit := delCloudMap[one.GetCloudID()]; exsit {
			logs.Errorf("[%s] validate account not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.Gcp, opt, len(delFromCloud), kt.Rid)
			return errors.New("validate account not exist failed, before delete")
		}
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", parts),
		}
		if err = cli.dbCli.Global.SubAccount.BatchDelete(kt, deleteReq); err != nil {
			return err
		}
		if err != nil {
			logs.Errorf("[%s] delete sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync sub account to delete sub account success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func isSubAccountChange(cloud account.GcpAccount, db coresubaccount.SubAccount[coresubaccount.GcpExtension]) bool {

	if cloud.Name != db.Name {
		return true
	}

	return false
}

func (cli *client) listAccountFromCloud(kt *kit.Kit, opt *SyncSubAccountOption) ([]account.GcpAccount, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	results, err := cli.cloudCli.ListAccount(kt)
	if err != nil {
		logs.Errorf("[%s] list sub account from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	return results, nil
}

func (cli *client) listAccountFromDB(kt *kit.Kit, opt *SyncSubAccountOption) (
	[]coresubaccount.SubAccount[coresubaccount.GcpExtension], error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.Gcp,
				},
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: opt.AccountID,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	start := uint32(0)
	results := make([]coresubaccount.SubAccount[coresubaccount.GcpExtension], 0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.Gcp.SubAccount.ListExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list sub account from db failed, err: %v, account: %s, req: %v, rid: %s",
				enumor.Gcp, err, opt.AccountID, req, kt.Rid)
			return nil, err
		}

		results = append(results, resp.Details...)

		if len(resp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return results, nil
}
