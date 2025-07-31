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

package azure

import (
	"errors"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesrg "hcm/pkg/adaptor/types/resource-group"
	"hcm/pkg/api/core"
	corerg "hcm/pkg/api/core/cloud/resource-group"
	datarg "hcm/pkg/api/data-service/cloud/resource-group"
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

// SyncRGOption ...
type SyncRGOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate ...
func (opt SyncRGOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ResourceGroup ...
func (cli *client) ResourceGroup(kt *kit.Kit, opt *SyncRGOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resourcegroupFromCloud, err := cli.listResourceGroupFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	resourcegroupFromDB, err := cli.listResourceGroupFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(resourcegroupFromCloud) == 0 && len(resourcegroupFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesrg.AzureResourceGroup, corerg.AzureRG](
		resourcegroupFromCloud, resourcegroupFromDB, isResourceGroupChange)

	if len(delCloudIDs) > 0 {
		if err := cli.deleteResourceGroup(kt, opt, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createResourceGroup(kt, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateResourceGroup(kt, opt, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) createResourceGroup(kt *kit.Kit, opt *SyncRGOption,
	addSlice []typesrg.AzureResourceGroup) error {

	if len(addSlice) <= 0 {
		return errors.New("resourcegroup addSlice is <= 0, not create")
	}

	list := make([]datarg.AzureRGBatchCreate, 0, len(addSlice))

	for _, one := range addSlice {
		rgOne := datarg.AzureRGBatchCreate{
			Name:      converter.PtrToVal(one.Name),
			Type:      string(converter.PtrToVal(one.Type)),
			Location:  converter.PtrToVal(one.Location),
			AccountID: opt.AccountID,
		}
		list = append(list, rgOne)
	}

	createReq := &datarg.AzureRGBatchCreateReq{
		ResourceGroups: list,
	}
	_, err := cli.dbCli.Azure.ResourceGroup.BatchCreateResourceGroup(kt.Ctx, kt.Header(),
		createReq)
	if err != nil {
		logs.Errorf("[%s] create resourcegroup failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync resourcegroup to create resourcegroup success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) updateResourceGroup(kt *kit.Kit, opt *SyncRGOption,
	updateMap map[string]typesrg.AzureResourceGroup) error {

	if len(updateMap) <= 0 {
		return errors.New("resourcegroup updateMap is <= 0, not update")
	}

	list := make([]datarg.AzureRGBatchUpdate, 0, len(updateMap))
	for id, one := range updateMap {
		rgOne := datarg.AzureRGBatchUpdate{
			ID:       id,
			Location: converter.PtrToVal(one.Location),
		}
		list = append(list, rgOne)
	}

	updateReq := &datarg.AzureRGBatchUpdateReq{
		ResourceGroups: list,
	}
	if err := cli.dbCli.Azure.ResourceGroup.BatchUpdateRG(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] update resourcegroup failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, opt.AccountID, opt, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync resourcegroup to update resourcegroup success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) deleteResourceGroup(kt *kit.Kit, opt *SyncRGOption, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return errors.New("resourcegroup delCloudIDs is <= 0, not delete")
	}

	delResourceGroupFromCloud, err := cli.listResourceGroupFromCloud(kt, opt)
	if err != nil {
		return err
	}

	delCloudMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range delResourceGroupFromCloud {
		if _, exsit := delCloudMap[converter.PtrToVal(one.Name)]; exsit {
			logs.Errorf("[%s] validate resourcegroup not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
				enumor.Azure, opt, len(delResourceGroupFromCloud), kt.Rid)
			return errors.New("validate resourcegroup not exist failed, before delete")
		}
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &datarg.AzureRGBatchDeleteReq{
			Filter: tools.ContainersExpression("name", parts),
		}
		err := cli.dbCli.Azure.ResourceGroup.BatchDeleteResourceGroup(kt.Ctx, kt.Header(), deleteReq)
		if err != nil {
			logs.Errorf("[%s] delete resourcegroup failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
				err, opt.AccountID, opt, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync resourcegroup to delete resourcegroup success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listResourceGroupFromCloud(kt *kit.Kit,
	opt *SyncRGOption) ([]typesrg.AzureResourceGroup, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resourcegroups, err := cli.cloudCli.ListResourceGroup(kt)
	if err != nil {
		logs.Errorf("[%s] list resourcegroup from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure,
			err, opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	results := make([]typesrg.AzureResourceGroup, 0)
	for _, one := range resourcegroups {
		results = append(results, converter.PtrToVal(one))
	}

	return results, nil
}

func (cli *client) listResourceGroupFromDB(kt *kit.Kit, opt *SyncRGOption) (
	[]corerg.AzureRG, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &datarg.AzureRGListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
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
	results := make([]corerg.AzureRG, 0)
	for {
		req.Page.Start = start
		resourcegroups, err := cli.dbCli.Azure.ResourceGroup.ListResourceGroup(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list resourcegroup from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
				opt.AccountID, req, kt.Rid)
			return nil, err
		}
		results = append(results, resourcegroups.Details...)

		if len(resourcegroups.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	return results, nil
}

func isResourceGroupChange(cloud typesrg.AzureResourceGroup, db corerg.AzureRG) bool {

	if converter.PtrToVal(cloud.Location) != db.Location {
		return true
	}

	return false
}
