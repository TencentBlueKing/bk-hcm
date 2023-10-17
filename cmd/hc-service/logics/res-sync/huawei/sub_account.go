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

package huawei

import (
	"errors"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
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

	fromCloud, err := cli.listSubAccountFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	fromDB, err := cli.listSubAccountFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[account.HuaWeiAccount,
		coresubaccount.SubAccount[coresubaccount.HuaWeiExtension]](fromCloud, fromDB, isSubAccountChange)

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
	updateMap map[string]account.HuaWeiAccount) error {

	if len(updateMap) <= 0 {
		return errors.New("updateMap is required")
	}

	account, err := cli.dbCli.HuaWei.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	updateItems := make([]dssubaccount.UpdateField, 0, len(updateMap))

	for id, one := range updateMap {
		extension := &coresubaccount.HuaWeiExtension{
			CloudAccountID: one.DomainID,
			LastProjectID:  one.LastProjectID,
			Enabled:        one.Enabled,
		}

		ext, err := core.MarshalStruct(extension)
		if err != nil {
			return err
		}

		accountType := ""
		if account.Extension.CloudIamUserID != "" &&
			account.Extension.CloudIamUserID == one.ID {
			accountType = string(enumor.CurrentAccount)
		}

		tmpRes := dssubaccount.UpdateField{
			ID:          id,
			Name:        one.Name,
			Vendor:      enumor.HuaWei,
			Site:        account.Site,
			AccountID:   account.ID,
			AccountType: accountType,
			Extension:   ext,
			// Managers/BizIDs由用户设置不继承资源账号。
			Managers: nil,
			BkBizIDs: nil,
			Memo:     one.Description,
		}
		updateItems = append(updateItems, tmpRes)
	}

	updateReq := &dssubaccount.UpdateReq{
		Items: updateItems,
	}
	if err = cli.dbCli.Global.SubAccount.BatchUpdate(kt, updateReq); err != nil {
		logs.Errorf("[%s] update sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sub account to update sub account success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSubAccount(kt *kit.Kit, opt *SyncSubAccountOption, addSlice []account.HuaWeiAccount) error {

	if len(addSlice) <= 0 {
		return errors.New("addSlice is required")
	}

	account, err := cli.dbCli.HuaWei.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	createResources := make([]dssubaccount.CreateField, 0, len(addSlice))
	// 产品侧定义主账号数据较重要，定制化插入一条主账号数据
	mainAccount, err := cli.makeMainAccount(kt, account)
	if err != nil {
		return err
	}
	createResources = append(createResources, mainAccount...)

	for _, one := range addSlice {
		extension := &coresubaccount.HuaWeiExtension{
			CloudAccountID: one.DomainID,
			LastProjectID:  one.LastProjectID,
			Enabled:        one.Enabled,
		}

		accountType := ""
		if account.Extension.CloudIamUserID != "" &&
			account.Extension.CloudIamUserID == one.ID {
			accountType = string(enumor.CurrentAccount)
		}

		ext, err := core.MarshalStruct(extension)
		if err != nil {
			return err
		}

		tmpRes := dssubaccount.CreateField{
			CloudID:     one.ID,
			Name:        one.Name,
			Vendor:      enumor.HuaWei,
			Site:        account.Site,
			AccountID:   account.ID,
			AccountType: accountType,
			Extension:   ext,
			// Managers/BizIDs由用户设置不继承资源账号。
			Managers: nil,
			BkBizIDs: nil,
			Memo:     one.Description,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dssubaccount.CreateReq{
		Items: createResources,
	}
	if _, err = cli.dbCli.Global.SubAccount.BatchCreate(kt, createReq); err != nil {
		logs.Errorf("[%s] create sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sub account to create sub account success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) deleteSubAccount(kt *kit.Kit, opt *SyncSubAccountOption, delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return errors.New("delCloudIDs is required")
	}

	delFromCloud, err := cli.listSubAccountFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	// 主账号构造的数据云上一定没有，这里过滤掉
	delete(delCloudMap, string(enumor.MainAccount))
	for _, one := range delFromCloud {
		if _, exsit := delCloudMap[one.GetCloudID()]; exsit {
			logs.Errorf("[%s] validate account not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.HuaWei, opt, len(delFromCloud), kt.Rid)
			return errors.New("validate account not exist failed, before delete")
		}
	}

	delCloudIDs = converter.MapKeyToStringSlice(delCloudMap)
	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", parts),
		}
		if err = cli.dbCli.Global.SubAccount.BatchDelete(kt, deleteReq); err != nil {
			return err
		}
		if err != nil {
			logs.Errorf("[%s] delete sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync sub account to delete sub account success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func isSubAccountChange(cloud account.HuaWeiAccount, db coresubaccount.SubAccount[coresubaccount.HuaWeiExtension]) bool {

	if cloud.DomainID != db.Extension.CloudAccountID {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.LastProjectID, db.Extension.LastProjectID) {
		return true
	}

	if cloud.Enabled != db.Extension.Enabled {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Description, db.Memo) {
		return true
	}

	if cloud.Name != db.Name {
		return true
	}

	return false
}

func (cli *client) listSubAccountFromCloud(kt *kit.Kit, opt *SyncSubAccountOption) ([]account.HuaWeiAccount, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	results, err := cli.cloudCli.ListAccount(kt)
	if err != nil {
		logs.Errorf("[%s] list sub account from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	return results, nil
}

func (cli *client) listSubAccountFromDB(kt *kit.Kit, opt *SyncSubAccountOption) (
	[]coresubaccount.SubAccount[coresubaccount.HuaWeiExtension], error) {

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
					Value: enumor.HuaWei,
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
	results := make([]coresubaccount.SubAccount[coresubaccount.HuaWeiExtension], 0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.HuaWei.SubAccount.ListExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list sub account from db failed, err: %v, account: %s, req: %v, rid: %s",
				enumor.HuaWei, err, opt.AccountID, req, kt.Rid)
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

func (cli *client) makeMainAccount(kt *kit.Kit,
	account *protocloud.AccountGetResult[protocore.HuaWeiAccountExtension]) ([]dssubaccount.CreateField, error) {

	ret := make([]dssubaccount.CreateField, 0)

	isExsit, err := cli.isMainAccountInSubAccountDB(kt)
	if err != nil {
		return ret, err
	}
	if isExsit {
		return ret, nil
	}

	extension := &coresubaccount.HuaWeiExtension{
		CloudAccountID: account.Extension.CloudSubAccountID,
	}

	ext, err := core.MarshalStruct(extension)
	if err != nil {
		return ret, err
	}

	ret = append(ret, dssubaccount.CreateField{
		CloudID:     string(enumor.MainAccount),
		Name:        string(enumor.MainAccount),
		Vendor:      enumor.HuaWei,
		Site:        account.Site,
		AccountID:   account.ID,
		AccountType: string(enumor.MainAccount),
		Extension:   ext,
		// Managers/BizIDs由用户设置不继承资源账号。
		Managers: nil,
		BkBizIDs: nil,
		Memo:     nil,
	})

	return ret, nil
}

func (cli *client) isMainAccountInSubAccountDB(kt *kit.Kit) (bool, error) {
	ret := false

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.HuaWei,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.Equal.Factory(),
					Value: enumor.MainAccount,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	start := uint32(0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.HuaWei.SubAccount.ListExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list sub account from db failed, err: %v, req: %v, rid: %s",
				enumor.HuaWei, err, req, kt.Rid)
			return false, err
		}

		if len(resp.Details) == 1 {
			ret = true
			break
		}

		if len(resp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return ret, nil
}
