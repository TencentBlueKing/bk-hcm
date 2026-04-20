/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
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

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core"
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SubAccountSecret sync sub account secret.
func (cli *client) SubAccountSecret(kt *kit.Kit, opt *SyncSubAccountOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	fromCloud, err := cli.listSubAccountSecretFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	fromDB, err := cli.listSubAccountSecretFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[account.TCloudSubAccountSecret,
		coresass.SubAccountSecret[coresass.TCloudSubAccountSecretExtension]](
		fromCloud, fromDB, isSubAccountSecretChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubAccountSecret(kt, opt, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createSubAccountSecret(kt, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateSubAccountSecret(kt, opt, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// listSubAccountSecretFromCloud lists all access keys from cloud for all sub-accounts under the given account.
// It first fetches sub-accounts from DB to retrieve their UINs, then calls ListAccessKeys for each,
// and finally enriches the result with LastUsedTime via GetSecurityLastUsed.
func (cli *client) listSubAccountSecretFromCloud(kt *kit.Kit, opt *SyncSubAccountOption) (
	[]account.TCloudSubAccountSecret, error) {

	subAccounts, err := cli.listSubAccountFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	secrets := make([]account.TCloudSubAccountSecret, 0)
	for _, subAccount := range subAccounts {
		if subAccount.Extension == nil || converter.PtrToVal(subAccount.Extension.Uin) == 0 {
			logs.Errorf("[%s] sub account %s has no valid uin, skip listing access keys, account: %s, rid: %s",
				enumor.TCloud, subAccount.ID, opt.AccountID, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, errors.New("sub account has no valid uin"))
		}

		if subAccount.AccountType == string(enumor.MainAccount) {
			logs.Warnf("[%s] sub account %s is main account, skip listing access keys, account: %s, rid: %s",
				enumor.TCloud, subAccount.ID, opt.AccountID, kt.Rid)
			continue
		}

		uin := converter.PtrToVal(subAccount.Extension.Uin)
		keys, err := cli.cloudCli.ListAccessKeys(kt, &account.ListAccessKeysOption{TargetUin: uin})
		if err != nil {
			logs.Errorf("[%s] list access keys failed, id: %s, account: %s, err: %v, rid: %s",
				enumor.TCloud, subAccount.ID, opt.AccountID, err, kt.Rid)
			return nil, err
		}

		for _, key := range keys {
			secrets = append(secrets, account.TCloudSubAccountSecret{
				AccountID:          opt.AccountID,
				SubAccountID:       subAccount.ID,
				CloudMainAccountID: subAccount.Extension.CloudMainAccountID,
				CloudSubAccountID:  subAccount.CloudID,
				AccessKeyID:        key.AccessKeyID,
				Status:             key.Status,
				CreateTime:         key.CreateTime,
			})
		}
	}

	if err = cli.fetchLastUsedTime(kt, opt, secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

// fetchLastUsedTime fills in the LastUsedTime for each cloud secret by calling GetSecurityLastUsed in batches
// of at most GetSecurityLastUsedMaxKeys keys per request.
func (cli *client) fetchLastUsedTime(kt *kit.Kit, opt *SyncSubAccountOption,
	secrets []account.TCloudSubAccountSecret) error {

	if len(secrets) == 0 {
		return nil
	}

	accessKeyIDs := make([]string, 0, len(secrets))
	for _, s := range secrets {
		accessKeyIDs = append(accessKeyIDs, s.AccessKeyID)
	}

	lastUsedMap := make(map[string]*string, len(accessKeyIDs))
	batches := slice.Split(accessKeyIDs, account.GetSecurityLastUsedMaxKeys)
	for _, batch := range batches {
		results, err := cli.cloudCli.GetSecurityLastUsed(kt, &account.GetSecurityLastUsedOption{
			SecretIdList: batch,
		})
		if err != nil {
			logs.Errorf("[%s] get security last used failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return err
		}

		for _, r := range results {
			if r.LastUsedDate != nil {
				lastUsedMap[r.SecretId] = r.LastUsedDate
			}
		}
	}

	for i := range secrets {
		secrets[i].LastUsedTime = lastUsedMap[secrets[i].AccessKeyID]
	}

	return nil
}

// listSubAccountSecretFromDB lists all sub account secrets from DB for the given account, paginating as needed.
func (cli *client) listSubAccountSecretFromDB(kt *kit.Kit, opt *SyncSubAccountOption) (
	[]coresass.SubAccountSecret[coresass.TCloudSubAccountSecretExtension], error) {

	req := &protocloud.SubAccountSecretExtListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: enumor.TCloud},
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: opt.AccountID},
			},
		},
		Page: core.NewDefaultBasePage(),
	}

	results := make([]coresass.SubAccountSecret[coresass.TCloudSubAccountSecretExtension], 0)
	start := uint32(0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.TCloud.SubAccountSecret.ListSubAccountSecretWithExtension(kt, req)
		if err != nil {
			logs.Errorf("[%s] list sub account secret from db failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
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

// isSubAccountSecretChange reports whether the cloud secret differs from the DB record.
// DisabledTime is intentionally excluded as it is managed locally, not by cloud sync.
func isSubAccountSecretChange(cloud account.TCloudSubAccountSecret,
	db coresass.SubAccountSecret[coresass.TCloudSubAccountSecretExtension]) bool {

	if hssubaccount.TCloudAccessKeyStatusToSecretStatus(cloud.Status) != db.Status {
		return true
	}

	if !assert.IsPtrStringEqual(converter.ValToPtr(cloud.CreateTime), db.CloudCreatedAt) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.LastUsedTime, db.LastUsedTime) {
		return true
	}

	return false
}

// createSubAccountSecret batch creates new sub account secrets in DB.
func (cli *client) createSubAccountSecret(kt *kit.Kit, opt *SyncSubAccountOption,
	addSlice []account.TCloudSubAccountSecret) error {

	creates := make([]protocloud.SubAccountSecretCreate[coresass.TCloudSubAccountSecretExtension], 0, len(addSlice))
	for _, entry := range addSlice {
		status := hssubaccount.TCloudAccessKeyStatusToSecretStatus(entry.Status)
		var createTime *string
		if entry.CreateTime != "" {
			createTime = converter.ValToPtr(entry.CreateTime)
		}

		creates = append(creates, protocloud.SubAccountSecretCreate[coresass.TCloudSubAccountSecretExtension]{
			AccountID:      entry.AccountID,
			SubAccountID:   entry.SubAccountID,
			Status:         status,
			CloudCreatedAt: createTime,
			LastUsedTime:   entry.LastUsedTime,
			Extension: &coresass.TCloudSubAccountSecretExtension{
				CloudSecretID:      entry.AccessKeyID,
				CloudMainAccountID: entry.CloudMainAccountID,
				CloudSubAccountID:  entry.CloudSubAccountID,
			},
		})
	}

	batches := slice.Split(creates, constant.CloudResourceSyncMaxLimit)
	for _, batch := range batches {
		req := &protocloud.SubAccountSecretBatchCreateReq[coresass.TCloudSubAccountSecretExtension]{
			SubAccountSecrets: batch,
		}
		if _, err := cli.dbCli.TCloud.SubAccountSecret.BatchCreateSubAccountSecret(kt, req); err != nil {
			logs.Errorf("[%s] create sub account secret failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync sub account secret to create success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

// updateSubAccountSecret batch updates changed sub account secrets in DB.
// DisabledTime is not updated as it is managed locally.
func (cli *client) updateSubAccountSecret(kt *kit.Kit, opt *SyncSubAccountOption,
	updateMap map[string]account.TCloudSubAccountSecret) error {

	updates := make([]protocloud.SubAccountSecretUpdate[coresass.TCloudSubAccountSecretExtension], 0, len(updateMap))
	for dbID, entry := range updateMap {
		status := hssubaccount.TCloudAccessKeyStatusToSecretStatus(entry.Status)
		var createTime *string
		if entry.CreateTime != "" {
			createTime = converter.ValToPtr(entry.CreateTime)
		}

		updates = append(updates, protocloud.SubAccountSecretUpdate[coresass.TCloudSubAccountSecretExtension]{
			ID:             dbID,
			Status:         converter.ValToPtr(status),
			CloudCreatedAt: createTime,
			LastUsedTime:   entry.LastUsedTime,
			Extension: &coresass.TCloudSubAccountSecretExtension{
				CloudSecretID:      entry.AccessKeyID,
				CloudMainAccountID: entry.CloudMainAccountID,
				CloudSubAccountID:  entry.CloudSubAccountID,
			},
		})
	}

	batches := slice.Split(updates, constant.CloudResourceSyncMaxLimit)
	for _, batch := range batches {
		req := &protocloud.SubAccountSecretBatchUpdateReq[coresass.TCloudSubAccountSecretExtension]{
			SubAccountSecrets: batch,
		}
		if err := cli.dbCli.TCloud.SubAccountSecret.BatchUpdateSubAccountSecret(kt, req); err != nil {
			logs.Errorf("[%s] update sub account secret failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync sub account secret to update success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

// deleteSubAccountSecret batch deletes sub account secrets from DB by cloud IDs.
func (cli *client) deleteSubAccountSecret(kt *kit.Kit, opt *SyncSubAccountOption, delCloudIDs []string) error {
	batches := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, batch := range batches {
		req := &protocloud.SubAccountSecretBatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", batch),
		}
		if err := cli.dbCli.Global.SubAccountSecret.BatchDelete(kt, req); err != nil {
			logs.Errorf("[%s] delete sub account secret failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync sub account secret to delete success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}
