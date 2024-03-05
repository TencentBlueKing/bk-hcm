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
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typeargstpl "hcm/pkg/adaptor/types/argument-template"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcargstpl "hcm/pkg/api/hc-service/argument-template"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// SyncArgsTplOption ...
type SyncArgsTplOption struct {
	BkBizID int64 `json:"bk_biz_id" validate:"omitempty"`
}

// Validate ...
func (opt SyncArgsTplOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// --------------- Sync Address ------------------

// ArgsTplAddress ...
func (cli *client) ArgsTplAddress(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	fromCloud, err := cli.listFromCloudAddress(kt, params)
	if err != nil {
		return nil, err
	}
	logs.Infof("[%s] hcservice argument template address listFromCloud success, params: %+v, cloud_count: %d, rid: %s",
		enumor.TCloud, params, len(fromCloud), kt.Rid)

	fromDB, err := cli.listFromDB(kt, params, enumor.AddressType)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice sync argument template address listFromDB success, db_count: %d, rid: %s",
		enumor.TCloud, len(fromDB), kt.Rid)

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeargstpl.TCloudArgsTplAddress,
		*coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]](fromCloud, fromDB, isChangeAddress)

	logs.Infof("[%s] hcservice sync argument template diff address success, addNum: %d, updateNum: %d, delNum: %d, "+
		"rid: %s", enumor.TCloud, len(addSlice), len(updateMap), len(delCloudIDs), kt.Rid)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteAddress(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createAddress(kt, params.AccountID, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateAddress(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) deleteAddress(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delFromCloud, err := cli.listFromCloudAddress(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delFromCloud) > 0 {
		logs.Errorf("[%s] validate argument template address not exist failed, before delete, opt: %v, "+
			"failed_count: %d, rid: %s", enumor.TCloud, checkParams, len(delFromCloud), kt.Rid)
		return fmt.Errorf("validate argument template address not exist failed, before delete")
	}

	deleteReq := &protocloud.ArgsTplBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.ArgsTpl.BatchDeleteArgsTpl(kt, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete address failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to delete address success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateAddress(kt *kit.Kit, accountID string,
	updateMap map[string]typeargstpl.TCloudArgsTplAddress) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, updateMap is <= 0, not update")
	}

	for id, one := range updateMap {
		tmpAddressSet := make([]hcargstpl.TemplateInfo, 0, len(one.AddressExtraSet))
		for _, cloudAddress := range one.AddressExtraSet {
			tmpAddressSet = append(tmpAddressSet, hcargstpl.TemplateInfo{
				Address:     cloudAddress.Address,
				Description: cloudAddress.Description,
			})
		}

		templateJson, err := types.NewJsonField(tmpAddressSet)
		if err != nil {
			return fmt.Errorf("json marshal template failed, err: %w", err)
		}

		var updateReq = protocloud.ArgsTplBatchUpdateExprReq{
			IDs:       []string{id},
			Name:      converter.PtrToVal(one.AddressTemplateName),
			Templates: templateJson,
		}

		if _, err = cli.dbCli.Global.ArgsTpl.BatchUpdateArgsTpl(kt, &updateReq); err != nil {
			logs.Errorf("[%s] request dataservice BatchUpdateArgsTpl address failed, err: %v, rid: %s",
				enumor.TCloud, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync argument template to update address success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createAddress(kt *kit.Kit, accountID string, opt *SyncArgsTplOption,
	addSlice []typeargstpl.TCloudArgsTplAddress) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, addSlice is <= 0, not create")
	}

	var createReq = new(protocloud.ArgsTplBatchCreateReq[coreargstpl.TCloudArgsTplExtension])

	for _, one := range addSlice {
		tmpAddressSet := make([]hcargstpl.TemplateInfo, 0, len(one.AddressExtraSet))
		for _, cloudAddress := range one.AddressExtraSet {
			tmpAddressSet = append(tmpAddressSet, hcargstpl.TemplateInfo{
				Address:     cloudAddress.Address,
				Description: cloudAddress.Description,
			})
		}

		templateJson, err := types.NewJsonField(tmpAddressSet)
		if err != nil {
			return fmt.Errorf("json marshal template failed, err: %w", err)
		}

		tmp := []protocloud.ArgsTplBatchCreate[coreargstpl.TCloudArgsTplExtension]{
			{
				CloudID:   one.GetCloudID(),
				Name:      converter.PtrToVal(one.AddressTemplateName),
				Vendor:    string(enumor.TCloud),
				AccountID: accountID,
				BkBizID:   opt.BkBizID,
				Type:      enumor.AddressType,
				Templates: templateJson,
			},
		}

		createReq.ArgumentTemplates = append(createReq.ArgumentTemplates, tmp...)
	}

	_, err := cli.dbCli.TCloud.BatchCreateArgsTpl(kt, createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud address failed, createReq: %+v, err: %v, rid: %s",
			enumor.TCloud, createReq, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to create address success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) listFromCloudAddress(kt *kit.Kit, params *SyncBaseParams) (
	[]typeargstpl.TCloudArgsTplAddress, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list := make([]typeargstpl.TCloudArgsTplAddress, 0)
	for _, tmpCloudID := range params.CloudIDs {
		opt := &typeargstpl.TCloudListOption{
			Page: &adcore.TCloudPage{Offset: 0, Limit: 1},
			Filters: []*vpc.Filter{
				{
					Name:   tcommon.StringPtr("address-template-id"),
					Values: []*string{tcommon.StringPtr(tmpCloudID)},
				},
			},
		}
		result, _, err := cli.cloudCli.ListArgsTplAddress(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list argstpl address from cloud failed, account: %s, opt: %v, err: %v, rid: %s",
				enumor.TCloud, params.AccountID, opt, err, kt.Rid)
			return nil, err
		}

		list = append(list, result...)
	}

	return list, nil
}

func (cli *client) listFromCloudAddressGroup(kt *kit.Kit, params *SyncBaseParams) (
	[]typeargstpl.TCloudArgsTplAddressGroup, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list := make([]typeargstpl.TCloudArgsTplAddressGroup, 0)
	for _, tmpCloudID := range params.CloudIDs {
		opt := &typeargstpl.TCloudListOption{
			Page: &adcore.TCloudPage{Offset: 0, Limit: 1},
			Filters: []*vpc.Filter{
				{
					Name:   tcommon.StringPtr("address-template-group-id"),
					Values: []*string{tcommon.StringPtr(tmpCloudID)},
				},
			},
		}
		result, _, err := cli.cloudCli.ListArgsTplAddressGroup(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list argstpl address group from cloud failed, account: %s, opt: %v, err: %v, rid: %s",
				enumor.TCloud, params.AccountID, opt, err, kt.Rid)
			return nil, err
		}

		list = append(list, result...)
	}

	return list, nil
}

func (cli *client) listFromCloudService(kt *kit.Kit, params *SyncBaseParams) (
	[]typeargstpl.TCloudArgsTplService, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list := make([]typeargstpl.TCloudArgsTplService, 0)
	for _, tmpCloudID := range params.CloudIDs {
		opt := &typeargstpl.TCloudListOption{
			Page: &adcore.TCloudPage{Offset: 0, Limit: 1},
			Filters: []*vpc.Filter{
				{
					Name:   tcommon.StringPtr("service-template-id"),
					Values: []*string{tcommon.StringPtr(tmpCloudID)},
				},
			},
		}
		result, _, err := cli.cloudCli.ListArgsTplService(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list argstpl service from cloud failed, account: %s, opt: %v, err: %v, rid: %s",
				enumor.TCloud, params.AccountID, opt, err, kt.Rid)
			return nil, err
		}

		list = append(list, result...)
	}

	return list, nil
}

func (cli *client) listFromCloudServiceGroup(kt *kit.Kit, params *SyncBaseParams) (
	[]typeargstpl.TCloudArgsTplServiceGroup, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list := make([]typeargstpl.TCloudArgsTplServiceGroup, 0)
	for _, tmpCloudID := range params.CloudIDs {
		opt := &typeargstpl.TCloudListOption{
			Page: &adcore.TCloudPage{Offset: 0, Limit: 1},
			Filters: []*vpc.Filter{
				{
					Name:   tcommon.StringPtr("service-template-group-id"),
					Values: []*string{tcommon.StringPtr(tmpCloudID)},
				},
			},
		}
		result, _, err := cli.cloudCli.ListArgsTplServiceGroup(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list argstpl service group from cloud failed, account: %s, opt: %v, err: %v, rid: %s",
				enumor.TCloud, params.AccountID, opt, err, kt.Rid)
			return nil, err
		}

		list = append(list, result...)
	}

	return list, nil
}

func (cli *client) listFromDB(kt *kit.Kit, params *SyncBaseParams, templateType enumor.TemplateType) (
	[]*coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "type",
					Op:    filter.Equal.Factory(),
					Value: templateType,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.TCloud.ListArgsTplExt(kt, req)
	if err != nil {
		logs.Errorf("[%s] list argument template from db failed, account: %s, templateType: %s, req: %v, "+
			"err: %v, rid: %s", enumor.TCloud, params.AccountID, templateType, req, err, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isChangeAddress(cloud typeargstpl.TCloudArgsTplAddress,
	db *coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]) bool {

	if converter.PtrToVal(cloud.AddressTemplateName) != db.Name {
		return true
	}

	dbTemplates := converter.PtrToVal(db.Templates)
	if len(cloud.AddressExtraSet) != len(dbTemplates) {
		return true
	}

	if len(cloud.AddressExtraSet) > 0 {
		for idx, item := range cloud.AddressExtraSet {
			if converter.PtrToVal(item.Address) != converter.PtrToVal(dbTemplates[idx].Address) ||
				converter.PtrToVal(item.Description) != converter.PtrToVal(dbTemplates[idx].Description) {
				return true
			}
		}
	}

	return false
}

// RemoveArgsTplAddressDeleteFromCloud ...
func (cli *client) RemoveArgsTplAddressDeleteFromCloud(kt *kit.Kit, accountID, region string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "type", Op: filter.Equal.Factory(), Value: enumor.AddressType},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ArgsTpl.ListArgsTpl(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list address failed, req: %v, err: %v, rid: %s",
				enumor.TCloud, req, err, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listFromCloudAddress(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.AddressTemplateId))
			}

			cloudIDs = converter.MapKeyToStringSlice(cloudIDMap)
			if err = cli.deleteAddress(kt, accountID, region, cloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

// --------------- Sync Address Group ------------------

// ArgsTplAddressGroup ...
func (cli *client) ArgsTplAddressGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (
	*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	fromCloud, err := cli.listFromCloudAddressGroup(kt, params)
	if err != nil {
		return nil, err
	}
	logs.Infof("[%s] hcservice argument template address group listFromCloud success, params: %+v, cloud_count: %d, "+
		"rid: %s", enumor.TCloud, params, len(fromCloud), kt.Rid)

	fromDB, err := cli.listFromDB(kt, params, enumor.AddressGroupType)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice sync argument template address group listFromDB success, db_count: %d, rid: %s",
		enumor.TCloud, len(fromDB), kt.Rid)

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeargstpl.TCloudArgsTplAddressGroup,
		*coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]](fromCloud, fromDB, isChangeAddressGroup)

	logs.Infof("[%s] hcservice sync argument template diff address group success, addNum: %d, updateNum: %d, "+
		"delNum: %d, rid: %s", enumor.TCloud, len(addSlice), len(updateMap), len(delCloudIDs), kt.Rid)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteAddressGroup(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createAddressGroup(kt, params.AccountID, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateAddressGroup(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) deleteAddressGroup(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delFromCloud, err := cli.listFromCloudAddressGroup(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delFromCloud) > 0 {
		logs.Errorf("[%s] validate argument template address group not exist failed, before delete, opt: %v, "+
			"failed_count: %d, rid: %s", enumor.TCloud, checkParams, len(delFromCloud), kt.Rid)
		return fmt.Errorf("validate argument template address group not exist failed, before delete")
	}

	deleteReq := &protocloud.ArgsTplBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.ArgsTpl.BatchDeleteArgsTpl(kt, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete address group failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to delete address group success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateAddressGroup(kt *kit.Kit, accountID string,
	updateMap map[string]typeargstpl.TCloudArgsTplAddressGroup) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, updateMap is <= 0, not update")
	}

	for id, one := range updateMap {
		groupTemplateJson, err := types.NewJsonField(one.AddressTemplateIdSet)
		if err != nil {
			return fmt.Errorf("json marshal group template failed, err: %w", err)
		}

		var updateReq = protocloud.ArgsTplBatchUpdateExprReq{
			IDs:            []string{id},
			Name:           converter.PtrToVal(one.AddressTemplateGroupName),
			GroupTemplates: groupTemplateJson,
		}

		if _, err = cli.dbCli.Global.ArgsTpl.BatchUpdateArgsTpl(kt, &updateReq); err != nil {
			logs.Errorf("[%s] request dataservice BatchUpdateArgsTpl address group failed, err: %v, rid: %s",
				enumor.TCloud, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync argument template to update address group success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createAddressGroup(kt *kit.Kit, accountID string, opt *SyncArgsTplOption,
	addSlice []typeargstpl.TCloudArgsTplAddressGroup) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, addSlice is <= 0, not create")
	}

	var createReq = new(protocloud.ArgsTplBatchCreateReq[coreargstpl.TCloudArgsTplExtension])

	for _, one := range addSlice {
		groupTemplateJson, err := types.NewJsonField(one.AddressTemplateIdSet)
		if err != nil {
			return fmt.Errorf("json marshal group templates failed, err: %w", err)
		}

		tmp := []protocloud.ArgsTplBatchCreate[coreargstpl.TCloudArgsTplExtension]{
			{
				CloudID:        one.GetCloudID(),
				Name:           converter.PtrToVal(one.AddressTemplateGroupName),
				Vendor:         string(enumor.TCloud),
				AccountID:      accountID,
				BkBizID:        opt.BkBizID,
				Type:           enumor.AddressGroupType,
				GroupTemplates: groupTemplateJson,
			},
		}

		createReq.ArgumentTemplates = append(createReq.ArgumentTemplates, tmp...)
	}

	_, err := cli.dbCli.TCloud.BatchCreateArgsTpl(kt, createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud address group failed, createReq: %+v, err: %v, rid: %s",
			enumor.TCloud, createReq, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to create address group success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(addSlice), kt.Rid)

	return nil
}

func isChangeAddressGroup(cloud typeargstpl.TCloudArgsTplAddressGroup,
	db *coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]) bool {

	if converter.PtrToVal(cloud.AddressTemplateGroupName) != db.Name {
		return true
	}

	dbGroupTemplates := converter.PtrToVal(db.GroupTemplates)
	if len(cloud.AddressTemplateIdSet) != len(dbGroupTemplates) {
		return true
	}

	if len(cloud.AddressTemplateIdSet) > 0 {
		for idx, itemValue := range cloud.AddressTemplateIdSet {
			if converter.PtrToVal(itemValue) != dbGroupTemplates[idx] {
				return true
			}
		}
	}

	return false
}

// RemoveArgsTplAddressGroupDeleteFromCloud ...
func (cli *client) RemoveArgsTplAddressGroupDeleteFromCloud(kt *kit.Kit, accountID, region string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "type", Op: filter.Equal.Factory(), Value: enumor.AddressGroupType},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ArgsTpl.ListArgsTpl(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list address group failed, req: %v, err: %v, rid: %s",
				enumor.TCloud, req, err, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listFromCloudAddressGroup(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.AddressTemplateGroupId))
			}

			cloudIDs = converter.MapKeyToStringSlice(cloudIDMap)
			if err = cli.deleteAddressGroup(kt, accountID, region, cloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

// --------------- Sync Service ------------------

// ArgsTplService ...
func (cli *client) ArgsTplService(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	fromCloud, err := cli.listFromCloudService(kt, params)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice argument template service listFromCloud success, params: %+v, cloud_count: %d, rid: %s",
		enumor.TCloud, params, len(fromCloud), kt.Rid)

	fromDB, err := cli.listFromDB(kt, params, enumor.ServiceType)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice sync argument template service listFromDB success, db_count: %d, rid: %s",
		enumor.TCloud, len(fromDB), kt.Rid)

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeargstpl.TCloudArgsTplService,
		*coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]](fromCloud, fromDB, isChangeService)

	logs.Infof("[%s] hcservice sync argument template diff service success, addNum: %d, updateNum: %d, delNum: %d, "+
		"rid: %s", enumor.TCloud, len(addSlice), len(updateMap), len(delCloudIDs), kt.Rid)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteService(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createService(kt, params.AccountID, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateService(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) deleteService(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delFromCloud, err := cli.listFromCloudService(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delFromCloud) > 0 {
		logs.Errorf("[%s] validate argument template service not exist failed, before delete, opt: %v, "+
			"failed_count: %d, rid: %s", enumor.TCloud, checkParams, len(delFromCloud), kt.Rid)
		return fmt.Errorf("validate argument template service not exist failed, before delete")
	}

	deleteReq := &protocloud.ArgsTplBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.ArgsTpl.BatchDeleteArgsTpl(kt, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete service failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to delete service success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateService(kt *kit.Kit, accountID string,
	updateMap map[string]typeargstpl.TCloudArgsTplService) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, updateMap is <= 0, not update")
	}

	for id, one := range updateMap {
		tmpServiceSet := make([]hcargstpl.TemplateInfo, 0, len(one.ServiceExtraSet))
		for _, cloudAddress := range one.ServiceExtraSet {
			tmpServiceSet = append(tmpServiceSet, hcargstpl.TemplateInfo{
				Address:     cloudAddress.Service,
				Description: cloudAddress.Description,
			})
		}

		templateJson, err := types.NewJsonField(tmpServiceSet)
		if err != nil {
			return fmt.Errorf("json marshal template failed, err: %w", err)
		}

		var updateReq = protocloud.ArgsTplBatchUpdateExprReq{
			IDs:       []string{id},
			Name:      converter.PtrToVal(one.ServiceTemplateName),
			Templates: templateJson,
		}

		if _, err = cli.dbCli.Global.ArgsTpl.BatchUpdateArgsTpl(kt, &updateReq); err != nil {
			logs.Errorf("[%s] request dataservice BatchUpdateArgsTpl service failed, err: %v, rid: %s",
				enumor.TCloud, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync argument template to update service success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createService(kt *kit.Kit, accountID string, opt *SyncArgsTplOption,
	addSlice []typeargstpl.TCloudArgsTplService) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, addSlice is <= 0, not create")
	}

	var createReq = new(protocloud.ArgsTplBatchCreateReq[coreargstpl.TCloudArgsTplExtension])

	for _, one := range addSlice {
		tmpServiceSet := make([]hcargstpl.TemplateInfo, 0, len(one.ServiceExtraSet))
		for _, cloudAddress := range one.ServiceExtraSet {
			tmpServiceSet = append(tmpServiceSet, hcargstpl.TemplateInfo{
				Address:     cloudAddress.Service,
				Description: cloudAddress.Description,
			})
		}

		templateJson, err := types.NewJsonField(tmpServiceSet)
		if err != nil {
			return fmt.Errorf("json marshal template failed, err: %w", err)
		}

		tmp := []protocloud.ArgsTplBatchCreate[coreargstpl.TCloudArgsTplExtension]{
			{
				CloudID:   one.GetCloudID(),
				Name:      converter.PtrToVal(one.ServiceTemplateName),
				Vendor:    string(enumor.TCloud),
				AccountID: accountID,
				BkBizID:   opt.BkBizID,
				Type:      enumor.ServiceType,
				Templates: templateJson,
			},
		}

		createReq.ArgumentTemplates = append(createReq.ArgumentTemplates, tmp...)
	}

	_, err := cli.dbCli.TCloud.BatchCreateArgsTpl(kt, createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud service failed, createReq: %+v, err: %v, rid: %s",
			enumor.TCloud, createReq, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to create service success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(addSlice), kt.Rid)

	return nil
}

func isChangeService(cloud typeargstpl.TCloudArgsTplService,
	db *coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]) bool {

	if converter.PtrToVal(cloud.ServiceTemplateName) != db.Name {
		return true
	}

	dbTemplates := converter.PtrToVal(db.Templates)
	if len(cloud.ServiceExtraSet) != len(dbTemplates) {
		return true
	}

	if len(cloud.ServiceExtraSet) > 0 {
		for idx, item := range cloud.ServiceExtraSet {
			if converter.PtrToVal(item.Service) != converter.PtrToVal(dbTemplates[idx].Address) ||
				converter.PtrToVal(item.Description) != converter.PtrToVal(dbTemplates[idx].Description) {
				return true
			}
		}
	}

	return false
}

// RemoveArgsTplServiceDeleteFromCloud ...
func (cli *client) RemoveArgsTplServiceDeleteFromCloud(kt *kit.Kit, accountID, region string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "type", Op: filter.Equal.Factory(), Value: enumor.ServiceType},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ArgsTpl.ListArgsTpl(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list service failed, req: %v, err: %v, rid: %s",
				enumor.TCloud, req, err, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listFromCloudService(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.ServiceTemplateId))
			}

			cloudIDs = converter.MapKeyToStringSlice(cloudIDMap)
			if err = cli.deleteService(kt, accountID, region, cloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

// --------------- Sync Service Group ------------------

// ArgsTplServiceGroup ...
func (cli *client) ArgsTplServiceGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (
	*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	fromCloud, err := cli.listFromCloudServiceGroup(kt, params)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice argument template service group listFromCloud success, params: %+v, cloud_count: %d, "+
		"rid: %s", enumor.TCloud, params, len(fromCloud), kt.Rid)

	fromDB, err := cli.listFromDB(kt, params, enumor.ServiceGroupType)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice sync argument template service group listFromDB success, db_count: %d, rid: %s",
		enumor.TCloud, len(fromDB), kt.Rid)

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeargstpl.TCloudArgsTplServiceGroup,
		*coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]](fromCloud, fromDB, isChangeServiceGroup)

	logs.Infof("[%s] hcservice sync argument template diff service group success, addNum: %d, updateNum: %d, "+
		"delNum: %d, rid: %s", enumor.TCloud, len(addSlice), len(updateMap), len(delCloudIDs), kt.Rid)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteServiceGroup(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createServiceGroup(kt, params.AccountID, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateServiceGroup(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) deleteServiceGroup(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delFromCloud, err := cli.listFromCloudServiceGroup(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delFromCloud) > 0 {
		logs.Errorf("[%s] validate argument template service group not exist failed, before delete, opt: %v, "+
			"failed_count: %d, rid: %s", enumor.TCloud, checkParams, len(delFromCloud), kt.Rid)
		return fmt.Errorf("validate argument template service group not exist failed, before delete")
	}

	deleteReq := &protocloud.ArgsTplBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.ArgsTpl.BatchDeleteArgsTpl(kt, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete service group failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to delete service group success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateServiceGroup(kt *kit.Kit, accountID string,
	updateMap map[string]typeargstpl.TCloudArgsTplServiceGroup) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, updateMap is <= 0, not update")
	}

	for id, one := range updateMap {
		groupTemplateJson, err := types.NewJsonField(one.ServiceTemplateIdSet)
		if err != nil {
			return fmt.Errorf("json marshal group template failed, err: %w", err)
		}

		var updateReq = protocloud.ArgsTplBatchUpdateExprReq{
			IDs:            []string{id},
			Name:           converter.PtrToVal(one.ServiceTemplateGroupName),
			GroupTemplates: groupTemplateJson,
		}

		if _, err = cli.dbCli.Global.ArgsTpl.BatchUpdateArgsTpl(kt, &updateReq); err != nil {
			logs.Errorf("[%s] request dataservice BatchUpdateArgsTpl service group failed, err: %v, rid: %s",
				enumor.TCloud, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync argument template to update service group success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createServiceGroup(kt *kit.Kit, accountID string, opt *SyncArgsTplOption,
	addSlice []typeargstpl.TCloudArgsTplServiceGroup) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, addSlice is <= 0, not create")
	}

	var createReq = new(protocloud.ArgsTplBatchCreateReq[coreargstpl.TCloudArgsTplExtension])

	for _, one := range addSlice {
		groupTemplateJson, err := types.NewJsonField(one.ServiceTemplateIdSet)
		if err != nil {
			return fmt.Errorf("json marshal group templates failed, err: %w", err)
		}

		tmp := []protocloud.ArgsTplBatchCreate[coreargstpl.TCloudArgsTplExtension]{
			{
				CloudID:        one.GetCloudID(),
				Name:           converter.PtrToVal(one.ServiceTemplateGroupName),
				Vendor:         string(enumor.TCloud),
				AccountID:      accountID,
				BkBizID:        opt.BkBizID,
				Type:           enumor.ServiceGroupType,
				GroupTemplates: groupTemplateJson,
			},
		}

		createReq.ArgumentTemplates = append(createReq.ArgumentTemplates, tmp...)
	}

	_, err := cli.dbCli.TCloud.BatchCreateArgsTpl(kt, createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud service group failed, createReq: %+v, err: %v, rid: %s",
			enumor.TCloud, createReq, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync argument template to create service group success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(addSlice), kt.Rid)

	return nil
}

func isChangeServiceGroup(cloud typeargstpl.TCloudArgsTplServiceGroup,
	db *coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension]) bool {

	if converter.PtrToVal(cloud.ServiceTemplateGroupName) != db.Name {
		return true
	}

	dbGroupTemplates := converter.PtrToVal(db.GroupTemplates)
	if len(cloud.ServiceTemplateIdSet) != len(dbGroupTemplates) {
		return true
	}

	if len(cloud.ServiceTemplateIdSet) > 0 {
		for idx, itemValue := range cloud.ServiceTemplateIdSet {
			if converter.PtrToVal(itemValue) != dbGroupTemplates[idx] {
				return true
			}
		}
	}

	return false
}

// RemoveArgsTplServiceGroupDeleteFromCloud ...
func (cli *client) RemoveArgsTplServiceGroupDeleteFromCloud(kt *kit.Kit, accountID, region string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "type", Op: filter.Equal.Factory(), Value: enumor.ServiceGroupType},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ArgsTpl.ListArgsTpl(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list service group failed, req: %v, err: %v, rid: %s",
				enumor.TCloud, req, err, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listFromCloudServiceGroup(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.ServiceTemplateGroupId))
			}

			cloudIDs = converter.MapKeyToStringSlice(cloudIDMap)
			if err = cli.deleteServiceGroup(kt, accountID, region, cloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}
