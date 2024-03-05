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
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/sync/handler"
	argstpl "hcm/pkg/adaptor/types/argument-template"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncArgsTpl ....
func (svc *service) SyncArgsTpl(cts *rest.Contexts) (interface{}, error) {
	argsTplHandler := &argsTplAddressHandler{cli: svc.syncCli}

	addressErr := handler.ResourceSync(cts, argsTplHandler)
	if addressErr != nil {
		return nil, addressErr
	}

	addressGroupErr := handler.ResourceSync(cts, &argsTplAddressGroupHandler{
		cli: svc.syncCli, request: argsTplHandler.request, syncCli: argsTplHandler.syncCli})
	if addressGroupErr != nil {
		return nil, addressGroupErr
	}

	serviceErr := handler.ResourceSync(cts, &argsTplServiceHandler{
		cli: svc.syncCli, request: argsTplHandler.request, syncCli: argsTplHandler.syncCli})
	if serviceErr != nil {
		return nil, serviceErr
	}

	serviceGroupErr := handler.ResourceSync(cts, &argsTplServiceGroupHandler{
		cli: svc.syncCli, request: argsTplHandler.request, syncCli: argsTplHandler.syncCli})
	if serviceGroupErr != nil {
		return nil, serviceGroupErr
	}

	return nil, nil
}

// ------------- Sync Address ---------

// argsTplAddressHandler sync handler.
type argsTplAddressHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.TCloudSyncReq
	syncCli tcloud.Interface
	offset  uint64
}

var _ handler.Handler = new(argsTplAddressHandler)

// Prepare ...
func (hd *argsTplAddressHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		logs.Errorf("request tcloud argument template address prepare failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}

// Next ...
func (hd *argsTplAddressHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &argstpl.TCloudListOption{
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	result, _, err := hd.syncCli.CloudCli().ListArgsTplAddress(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud argument template address failed, opt: %v, err: %v, rid: %s",
			listOpt, err, kt.Rid)
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(result))
	for _, one := range result {
		cloudIDs = append(cloudIDs, converter.PtrToVal(one.AddressTemplateId))
	}

	hd.offset += typecore.TCloudQueryLimit
	return cloudIDs, nil
}

// Sync ...
func (hd *argsTplAddressHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.ArgsTplAddress(kt, params,
		&tcloud.SyncArgsTplOption{BkBizID: constant.UnassignedBiz}); err != nil {
		logs.Errorf("sync tcloud argument template address failed, opt: %v, err: %v, rid: %s", params, err, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *argsTplAddressHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveArgsTplAddressDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove argument template address delete from cloud failed, accountID: %s, region: %s, "+
			"err: %v, rid: %s", hd.request.AccountID, hd.request.Region, err, kt.Rid)
		return err
	}

	return nil
}

// Name get cloud resource type name
func (hd *argsTplAddressHandler) Name() enumor.CloudResourceType {
	return enumor.ArgumentTemplateResType
}

// ------------- Sync Address Group ---------

// argsTplAddressGroupHandler sync handler.
type argsTplAddressGroupHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.TCloudSyncReq
	syncCli tcloud.Interface
	offset  uint64
}

var _ handler.Handler = new(argsTplAddressGroupHandler)

// Prepare ...
func (hd *argsTplAddressGroupHandler) Prepare(_ *rest.Contexts) error {
	return nil
}

// Next ...
func (hd *argsTplAddressGroupHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &argstpl.TCloudListOption{
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	result, _, err := hd.syncCli.CloudCli().ListArgsTplAddressGroup(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud argument template address group failed, opt: %v, err: %v, rid: %s",
			listOpt, err, kt.Rid)
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(result))
	for _, one := range result {
		cloudIDs = append(cloudIDs, converter.PtrToVal(one.AddressTemplateGroupId))
	}

	hd.offset += typecore.TCloudQueryLimit
	return cloudIDs, nil
}

// Sync ...
func (hd *argsTplAddressGroupHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.ArgsTplAddressGroup(kt, params,
		&tcloud.SyncArgsTplOption{BkBizID: constant.UnassignedBiz}); err != nil {
		logs.Errorf("sync tcloud argument template address group failed, opt: %v, err: %v, rid: %s",
			params, err, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *argsTplAddressGroupHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveArgsTplAddressGroupDeleteFromCloud(
		kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove argument template address group delete from cloud failed, accountID: %s, region: %s, "+
			"err: %v, rid: %s", hd.request.AccountID, hd.request.Region, err, kt.Rid)
		return err
	}

	return nil
}

// Name get cloud resource type name
func (hd *argsTplAddressGroupHandler) Name() enumor.CloudResourceType {
	return enumor.ArgumentTemplateResType
}

// ------------- Sync Service ---------

// argsTplServiceHandler sync handler.
type argsTplServiceHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.TCloudSyncReq
	syncCli tcloud.Interface
	offset  uint64
}

var _ handler.Handler = new(argsTplServiceHandler)

// Prepare ...
func (hd *argsTplServiceHandler) Prepare(_ *rest.Contexts) error {
	return nil
}

// Next ...
func (hd *argsTplServiceHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &argstpl.TCloudListOption{
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	result, _, err := hd.syncCli.CloudCli().ListArgsTplService(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud argument template service failed, opt: %v, err: %v, rid: %s",
			listOpt, err, kt.Rid)
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(result))
	for _, one := range result {
		cloudIDs = append(cloudIDs, converter.PtrToVal(one.ServiceTemplateId))
	}

	hd.offset += typecore.TCloudQueryLimit
	return cloudIDs, nil
}

// Sync ...
func (hd *argsTplServiceHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.ArgsTplService(kt, params,
		&tcloud.SyncArgsTplOption{BkBizID: constant.UnassignedBiz}); err != nil {
		logs.Errorf("sync tcloud argument template service failed, opt: %v, err: %v, rid: %s", params, err, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *argsTplServiceHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveArgsTplServiceDeleteFromCloud(kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove argument template service delete from cloud failed, accountID: %s, region: %s, "+
			"err: %v, rid: %s", hd.request.AccountID, hd.request.Region, err, kt.Rid)
		return err
	}

	return nil
}

// Name get cloud resource type name
func (hd *argsTplServiceHandler) Name() enumor.CloudResourceType {
	return enumor.ArgumentTemplateResType
}

// ------------- Sync Service Group ---------

// argsTplServiceGroupHandler sync handler.
type argsTplServiceGroupHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.TCloudSyncReq
	syncCli tcloud.Interface
	offset  uint64
}

var _ handler.Handler = new(argsTplServiceGroupHandler)

// Prepare ...
func (hd *argsTplServiceGroupHandler) Prepare(_ *rest.Contexts) error {
	return nil
}

// Next ...
func (hd *argsTplServiceGroupHandler) Next(kt *kit.Kit) ([]string, error) {
	listOpt := &argstpl.TCloudListOption{
		Page: &typecore.TCloudPage{
			Offset: hd.offset,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	result, _, err := hd.syncCli.CloudCli().ListArgsTplServiceGroup(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor list tcloud argument template service group failed, opt: %v, err: %v, rid: %s",
			listOpt, err, kt.Rid)
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	cloudIDs := make([]string, 0, len(result))
	for _, one := range result {
		cloudIDs = append(cloudIDs, converter.PtrToVal(one.ServiceTemplateGroupId))
	}

	hd.offset += typecore.TCloudQueryLimit
	return cloudIDs, nil
}

// Sync ...
func (hd *argsTplServiceGroupHandler) Sync(kt *kit.Kit, cloudIDs []string) error {
	params := &tcloud.SyncBaseParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
		CloudIDs:  cloudIDs,
	}
	if _, err := hd.syncCli.ArgsTplServiceGroup(kt, params,
		&tcloud.SyncArgsTplOption{BkBizID: constant.UnassignedBiz}); err != nil {
		logs.Errorf("sync tcloud argument template service group failed, opt: %v, err: %v, rid: %s",
			params, err, kt.Rid)
		return err
	}

	return nil
}

// RemoveDeleteFromCloud ...
func (hd *argsTplServiceGroupHandler) RemoveDeleteFromCloud(kt *kit.Kit) error {
	if err := hd.syncCli.RemoveArgsTplServiceGroupDeleteFromCloud(
		kt, hd.request.AccountID, hd.request.Region); err != nil {
		logs.Errorf("remove argument template service group delete from cloud failed, accountID: %s, region: %s, "+
			"err: %v, rid: %s", hd.request.AccountID, hd.request.Region, err, kt.Rid)
		return err
	}

	return nil
}

// Name get cloud resource type name
func (hd *argsTplServiceGroupHandler) Name() enumor.CloudResourceType {
	return enumor.ArgumentTemplateResType
}
