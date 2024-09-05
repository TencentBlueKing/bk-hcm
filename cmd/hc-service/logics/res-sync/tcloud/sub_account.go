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

package tcloud

import (
	"errors"
	"strconv"

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

	addSlice, updateMap, delCloudIDs := common.Diff[account.TCloudAccount,
		coresubaccount.SubAccount[coresubaccount.TCloudExtension]](fromCloud, fromDB, isSubAccountChange)

	account, err := cli.dbCli.TCloud.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubAccount(kt, opt, account.Extension.CloudMainAccountID, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createSubAccount(kt, account, addSlice); err != nil {
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
	updateMap map[string]account.TCloudAccount) error {

	if len(updateMap) <= 0 {
		return errors.New("updateMap is required")
	}

	account, err := cli.dbCli.TCloud.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	updateItems := make([]dssubaccount.UpdateField, 0, len(updateMap))

	for id, one := range updateMap {
		extension := &coresubaccount.TCloudExtension{
			CloudMainAccountID: account.Extension.CloudMainAccountID,
			Uin:                one.Uin,
			NickName:           one.NickName,
			CreateTime:         one.CreateTime,
		}

		ext, err := core.MarshalStruct(extension)
		if err != nil {
			return err
		}

		accountType := ""
		if account.Extension.CloudSubAccountID != "" &&
			account.Extension.CloudSubAccountID == strconv.FormatUint(converter.PtrToVal(one.Uin), 10) {
			accountType = string(enumor.CurrentAccount)
		}

		tmpRes := dssubaccount.UpdateField{
			ID:          id,
			Name:        converter.PtrToVal(one.Name),
			Vendor:      enumor.TCloud,
			Site:        account.Site,
			AccountID:   account.ID,
			AccountType: accountType,
			Extension:   ext,
			// Managers/BizIDs由用户设置不继承资源账号。
			Managers: nil,
			BkBizIDs: nil,
			Memo:     one.Remark,
		}
		updateItems = append(updateItems, tmpRes)
	}

	updateReq := &dssubaccount.UpdateReq{
		Items: updateItems,
	}
	if err = cli.dbCli.Global.SubAccount.BatchUpdate(kt, updateReq); err != nil {
		logs.Errorf("[%s] update sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sub account to update sub account success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSubAccount(kt *kit.Kit,
	mainAccount *protocloud.AccountGetResult[protocore.TCloudAccountExtension], addSlice []account.TCloudAccount) error {

	if len(addSlice) <= 0 {
		return errors.New("addSlice is required")
	}

	createResources := make([]dssubaccount.CreateField, 0)

	// 产品侧定义主账号数据较重要，定制化插入一条主账号数据
	mainAccountCreateRes, err := cli.makeMainAccount(kt, mainAccount)
	if err != nil {
		return err
	}
	createResources = append(createResources, mainAccountCreateRes...)

	for _, one := range addSlice {

		extension := &coresubaccount.TCloudExtension{
			CloudMainAccountID: mainAccount.Extension.CloudMainAccountID,
			Uin:                one.Uin,
			NickName:           one.NickName,
			CreateTime:         one.CreateTime,
		}

		ext, err := core.MarshalStruct(extension)
		if err != nil {
			return err
		}

		accountType := ""
		if mainAccount.Extension.CloudSubAccountID != "" &&
			mainAccount.Extension.CloudSubAccountID == strconv.FormatUint(converter.PtrToVal(one.Uin), 10) {
			accountType = string(enumor.CurrentAccount)
		}

		tmpRes := dssubaccount.CreateField{
			CloudID:     one.GetCloudID(),
			Name:        converter.PtrToVal(one.Name),
			Vendor:      enumor.TCloud,
			Site:        mainAccount.Site,
			AccountID:   mainAccount.ID,
			AccountType: accountType,
			Extension:   ext,
			// Managers/BizIDs由用户设置不继承资源账号。
			Managers: nil,
			BkBizIDs: nil,
			Memo:     one.Remark,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dssubaccount.CreateReq{
		Items: createResources,
	}
	if _, err = cli.dbCli.Global.SubAccount.BatchCreate(kt, createReq); err != nil {
		logs.Errorf("[%s] create sub account failed, err: %v, account: %s, rid: %s", enumor.TCloud,
			err, mainAccount.ID, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync sub account to create sub account success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		mainAccount.ID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) deleteSubAccount(
	kt *kit.Kit, opt *SyncSubAccountOption, mainAccountID string, delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return errors.New("delCloudIDs is required")
	}

	delFromCloud, err := cli.listSubAccountFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	// 主账号构造的数据云上一定没有，这里过滤掉
	delete(delCloudMap, mainAccountID)
	for _, one := range delFromCloud {
		if _, exsit := delCloudMap[one.GetCloudID()]; exsit {
			logs.Errorf("[%s] validate account not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.TCloud, opt, len(delFromCloud), kt.Rid)
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
			logs.Errorf("[%s] delete sub account failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync sub account to delete sub account success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func isSubAccountChange(cloud account.TCloudAccount, db coresubaccount.SubAccount[coresubaccount.TCloudExtension]) bool {

	if !assert.IsPtrUint64Equal(cloud.Uin, db.Extension.Uin) {
		return true
	}

	if converter.PtrToVal(cloud.Name) != db.Name {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Remark, db.Memo) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.NickName, db.Extension.NickName) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CreateTime, db.Extension.CreateTime) {
		return true
	}

	return false
}

func (cli *client) listSubAccountFromCloud(kt *kit.Kit, opt *SyncSubAccountOption) ([]account.TCloudAccount, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	results, err := cli.cloudCli.ListAccount(kt)
	if err != nil {
		logs.Errorf("[%s] list sub account from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	return results, nil
}

func (cli *client) listSubAccountFromDB(kt *kit.Kit, opt *SyncSubAccountOption) (
	[]coresubaccount.SubAccount[coresubaccount.TCloudExtension], error) {

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
					Value: enumor.TCloud,
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
	results := make([]coresubaccount.SubAccount[coresubaccount.TCloudExtension], 0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.TCloud.SubAccount.ListExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list sub account from db failed, err: %v, account: %s, req: %v, rid: %s",
				enumor.TCloud, err, opt.AccountID, req, kt.Rid)
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
	account *protocloud.AccountGetResult[protocore.TCloudAccountExtension]) ([]dssubaccount.CreateField, error) {

	ret := make([]dssubaccount.CreateField, 0)

	isExsit, err := cli.isMainAccountInSubAccountDB(kt, account.Extension.CloudMainAccountID)
	if err != nil {
		return ret, err
	}
	if isExsit {
		return ret, nil
	}

	uin, _ := strconv.Atoi(account.Extension.CloudMainAccountID)
	int64Uin := uint64(uin)
	extension := &coresubaccount.TCloudExtension{
		CloudMainAccountID: account.Extension.CloudMainAccountID,
		Uin:                converter.ValToPtr(int64Uin),
	}

	ext, err := core.MarshalStruct(extension)
	if err != nil {
		return ret, err
	}

	ret = append(ret, dssubaccount.CreateField{
		CloudID:     account.Extension.CloudMainAccountID,
		Name:        string(enumor.MainAccount),
		Vendor:      enumor.TCloud,
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

func (cli *client) isMainAccountInSubAccountDB(kt *kit.Kit, cloudID string) (bool, error) {
	ret := false

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.TCloud,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.Equal.Factory(),
					Value: cloudID,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	start := uint32(0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.TCloud.SubAccount.ListExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list sub account from db failed, err: %v, req: %v, rid: %s",
				enumor.TCloud, err, req, kt.Rid)
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
