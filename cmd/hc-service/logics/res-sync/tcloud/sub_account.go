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
	"fmt"
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
	"hcm/pkg/dal/table/types"
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

// SubAccount sync subaccount.
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

	addSlice, updateMap, delCloudIDs := common.Diff[account.TCloudAccountWithExt,
		coresubaccount.SubAccount[coresubaccount.TCloudExtension]](fromCloud, fromDB, isSubAccountChange)

	// 获取三级账号的二级账号
	parentAccount, err := cli.dbCli.TCloud.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubAccount(kt, opt, parentAccount.Extension.CloudMainAccountID, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createSubAccount(kt, parentAccount, addSlice); err != nil {
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
	updateMap map[string]account.TCloudAccountWithExt) error {

	if len(updateMap) <= 0 {
		return errors.New("updateMap is required")
	}

	// 获取三级账号的二级账号
	parentAccount, err := cli.dbCli.TCloud.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to list account failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	subAccountIDs := make([]string, 0, len(updateMap))
	for id := range updateMap {
		subAccountIDs = append(subAccountIDs, id)
	}
	locSubAccountMap, err := cli.listSubAccountByID(kt, opt, subAccountIDs)
	if err != nil {
		return err
	}

	updateItems := make([]dssubaccount.UpdateField, 0, len(updateMap))
	for id, one := range updateMap {
		ext, err := core.MarshalStruct(one.Extension)
		if err != nil {
			return err
		}

		accountType := ""
		if parentAccount.Extension.CloudSubAccountID != "" &&
			parentAccount.Extension.CloudSubAccountID == strconv.FormatUint(converter.PtrToVal(one.Uin), 10) {
			accountType = string(enumor.CurrentAccount)
		}

		locSubAccount, exist := locSubAccountMap[id]
		if !exist {
			logs.Errorf("sync sub account failed, sub account %s not exist from DB", id)
			return fmt.Errorf("sync sub account failed, sub account %s not exist from DB", id)
		}

		tmpRes := dssubaccount.UpdateField{
			ID:             id,
			Name:           converter.PtrToVal(one.Name),
			Vendor:         enumor.TCloud,
			Site:           parentAccount.Site,
			AccountID:      parentAccount.ID,
			AccountType:    accountType,
			Extension:      &ext,
			Email:          one.Email,
			PhoneNum:       one.PhoneNum,
			CountryCode:    one.CountryCode,
			CloudCreatedAt: one.CreateTime,
			// Managers/由用户设置不继承资源账号。
			Managers: nil,
			Memo:     one.Remark,
		}

		// 如果DB中子账号没有业务ID，则需要继承主账号的业务ID
		if len(locSubAccount.BkBizIDs) == 0 {
			tmpRes.BkBizIDs = types.Int64Array{parentAccount.BkBizID}
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

// buildSubAccountExtension 构建子账号扩展信息
func (cli *client) buildSubAccountExtension(dbAccount *protocloud.AccountGetResult[protocore.TCloudAccountExtension],
	subAccount *account.TCloudAccount, safeAuthFlagMap map[uint64]account.SafeAuthFlagCollResult,
) *coresubaccount.TCloudExtension {

	// 获取安全认证标记
	var loginFlag *enumor.AccountProtectionFlag
	var actionFlag *enumor.AccountProtectionFlag
	if subAccount.Uin != nil {
		if flag, ok := safeAuthFlagMap[converter.PtrToVal(subAccount.Uin)]; ok {
			if flag.LoginFlag != nil {
				loginFlag = flag.LoginFlag.ToProtectionFlag()
			}
			if flag.ActionFlag != nil {
				actionFlag = flag.ActionFlag.ToProtectionFlag()
			}
		}
	}

	return &coresubaccount.TCloudExtension{
		CloudMainAccountID: dbAccount.Extension.CloudMainAccountID,
		Uin:                subAccount.Uin,
		NickName:           subAccount.NickName,
		CreateTime:         subAccount.CreateTime,
		ConsoleLogin:       enumor.GenfromConsoleLogin(subAccount.ConsoleLogin),
		LoginFlag:          loginFlag,
		ActionFlag:         actionFlag,
	}
}

func (cli *client) listSubAccountByID(kt *kit.Kit, opt *SyncSubAccountOption,
	subAccountIDs []string) (map[string]coresubaccount.SubAccount[coresubaccount.TCloudExtension], error) {

	if len(subAccountIDs) == 0 {
		return map[string]coresubaccount.SubAccount[coresubaccount.TCloudExtension]{}, nil
	}

	result := make(map[string]coresubaccount.SubAccount[coresubaccount.TCloudExtension], len(subAccountIDs))
	idChunks := slice.Split(subAccountIDs, int(core.DefaultMaxPageLimit))
	for _, ids := range idChunks {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", enumor.TCloud),
				tools.RuleEqual("account_id", opt.AccountID),
				tools.RuleIn("id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		resp, err := cli.dbCli.TCloud.SubAccount.ListExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list sub account by ids failed, err: %v, account: %s, req: %v, rid: %s",
				enumor.TCloud, err, opt.AccountID, req, kt.Rid)
			return nil, err
		}

		for _, one := range resp.Details {
			result[one.ID] = one
		}
	}

	return result, nil
}

func (cli *client) createSubAccount(kt *kit.Kit,
	mainAccount *protocloud.AccountGetResult[protocore.TCloudAccountExtension],
	addSlice []account.TCloudAccountWithExt) error {

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
		accountType := ""
		if mainAccount.Extension.CloudSubAccountID != "" &&
			mainAccount.Extension.CloudSubAccountID == strconv.FormatUint(converter.PtrToVal(one.Uin), 10) {
			accountType = string(enumor.CurrentAccount)
		}

		ext, err := core.MarshalStruct(one.Extension)
		if err != nil {
			return err
		}
		// Managers由用户设置不继承资源账号，业务id则继承二级账号的管理业务ID
		tmpRes := dssubaccount.CreateField{
			CloudID:        one.GetCloudID(),
			Name:           converter.PtrToVal(one.Name),
			Vendor:         enumor.TCloud,
			Site:           mainAccount.Site,
			AccountID:      mainAccount.ID,
			AccountType:    accountType,
			Extension:      ext,
			Email:          one.Email,
			PhoneNum:       one.PhoneNum,
			CountryCode:    one.CountryCode,
			CloudCreatedAt: one.CreateTime,
			Managers:       nil,
			BkBizIDs:       types.Int64Array{mainAccount.BkBizID},
			Memo:           one.Remark,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dssubaccount.CreateReq{Items: createResources}
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

func isSubAccountChange(cloud account.TCloudAccountWithExt,
	db coresubaccount.SubAccount[coresubaccount.TCloudExtension]) bool {

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

	if !assert.IsPtrStringEqual(cloud.Email, db.Email) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.PhoneNum, db.PhoneNum) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CountryCode, db.CountryCode) {
		return true
	}

	// 比较 Extension 中的字段
	if cloud.Extension != nil && db.Extension != nil {
		if !assert.IsPtrEqual(cloud.Extension.LoginFlag, db.Extension.LoginFlag) {
			return true
		}
		if !assert.IsPtrEqual(cloud.Extension.ActionFlag, db.Extension.ActionFlag) {
			return true
		}
		if !assert.IsPtrEqual(cloud.Extension.ConsoleLogin, db.Extension.ConsoleLogin) {
			return true
		}
	}
	if cloud.Extension != nil && db.Extension == nil || cloud.Extension == nil && db.Extension != nil {
		return true
	}

	return false
}

func (cli *client) listSubAccountFromCloud(kt *kit.Kit, opt *SyncSubAccountOption) (
	[]account.TCloudAccountWithExt, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	results, err := cli.cloudCli.ListAccount(kt)
	if err != nil {
		logs.Errorf("[%s] list sub account from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}
	if len(results) == 0 {
		return []account.TCloudAccountWithExt{}, nil
	}

	parAccount, err := cli.dbCli.TCloud.Account.Get(kt.Ctx, kt.Header(), opt.AccountID)
	if err != nil {
		logs.Errorf("request ds to get account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	uinList := make([]uint64, 0, len(results))
	for _, one := range results {
		if one.Uin != nil {
			uinList = append(uinList, converter.PtrToVal(one.Uin))
		}
	}
	safeAuthFlagMap, err := cli.batchDescribeSafeAuthFlagColl(kt, opt.AccountID, uinList)
	if err != nil {
		return nil, err
	}

	accountWithExtList := make([]account.TCloudAccountWithExt, 0, len(results))
	for _, one := range results {
		ext := cli.buildSubAccountExtension(parAccount, &one, safeAuthFlagMap)
		accountWithExtList = append(accountWithExtList, account.TCloudAccountWithExt{
			TCloudAccount: one,
			Extension:     ext,
		})
	}

	return accountWithExtList, nil
}

func (cli *client) listSubAccountFromDB(kt *kit.Kit, opt *SyncSubAccountOption) (
	[]coresubaccount.SubAccount[coresubaccount.TCloudExtension], error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloud),
			tools.RuleEqual("account_id", opt.AccountID),
		),
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

// batchDescribeSafeAuthFlagColl 分批次调用DescribeSafeAuthFlagColl获取子账号安全认证标记
func (cli *client) batchDescribeSafeAuthFlagColl(kt *kit.Kit, accountID string, uins []uint64) (
	map[uint64]account.SafeAuthFlagCollResult, error) {

	result := make(map[uint64]account.SafeAuthFlagCollResult)
	if len(uins) == 0 {
		return result, nil
	}

	// 分批次，每批次最多10个
	uinBatches := slice.Split(uins, account.DescribeSafeAuthFlagCollMaxUIN)
	for _, batch := range uinBatches {
		safeAuthFlags, err := cli.cloudCli.DescribeSafeAuthFlagColl(kt, &account.DescribeSafeAuthFlagCollOption{
			SubUins: batch,
		})
		if err != nil {
			logs.Errorf("describe safe auth flag coll failed, account: %s, sub_uin: %v, err: %v, rid: %s",
				accountID, batch, err, kt.Rid)
			return nil, fmt.Errorf("describe safe auth flag coll failed, err: %v", err)
		}

		for _, flag := range safeAuthFlags {
			result[flag.SubUin] = flag
		}
	}

	return result, nil
}
