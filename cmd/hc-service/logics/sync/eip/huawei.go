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

package eip

import (
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types/eip"
	apicore "hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncHuaWeiEipOption define sync huawei eip option.
type SyncHuaWeiEipOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncHuaWeiEipOption
func (opt SyncHuaWeiEipOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncHuaWeiEip sync eip self
func SyncHuaWeiEip(kt *kit.Kit, req *SyncHuaWeiEipOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := ad.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudAllIDs := make(map[string]bool)
	limit := int32(filter.DefaultMaxInLimit)
	marker := ""
	opt := &eip.HuaWeiEipListOption{
		Region: req.Region,
		Limit:  &limit,
	}
	for {
		if marker != "" {
			opt.Marker = &marker
		}
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		}

		datas, err := client.ListEip(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*HuaWeiEipSync)
		cloudIDs := make([]string, 0, len(datas.Details))
		for k, data := range datas.Details {
			eipSync := new(HuaWeiEipSync)
			eipSync.IsUpdate = false
			eipSync.Eip = data
			cloudMap[data.CloudID] = eipSync
			cloudIDs = append(cloudIDs, data.CloudID)
			cloudAllIDs[data.CloudID] = true
			if k == len(datas.Details)-1 {
				marker = data.CloudID
			}
		}

		updateIDs, dsMap, err := getHuaWeiEipDSSync(kt, cloudIDs, req, dataCli)
		if err != nil {
			logs.Errorf("request getHuaWeiEipDSSync failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := syncHuaWeiEipUpdate(kt, updateIDs, cloudMap, dsMap, dataCli)
			if err != nil {
				logs.Errorf("request syncHuaWeiEipUpdate failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		addIDs := make([]string, 0)
		for _, id := range updateIDs {
			if _, ok := cloudMap[id]; ok {
				cloudMap[id].IsUpdate = true
			}
		}

		for k, v := range cloudMap {
			if !v.IsUpdate {
				addIDs = append(addIDs, k)
			}
		}

		if len(addIDs) > 0 {
			err := syncHuaWeiEipAdd(kt, addIDs, req, cloudMap, dataCli)
			if err != nil {
				logs.Errorf("request syncHuaWeiEipAdd failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		if int32(len(datas.Details)) < limit {
			break
		}
	}

	dsIDs, err := getHuaWeiEipAllDS(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getHuaWeiEipAllDS failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for _, id := range dsIDs {
		if _, ok := cloudAllIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		for {
			if marker != "" {
				opt.Marker = &marker
			}
			if len(req.CloudIDs) > 0 {
				opt.CloudIDs = req.CloudIDs
			}

			datas, err := client.ListEip(kt, opt)
			if err != nil {
				logs.Errorf("request adaptor to list huawei eip failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			if len(datas.Details) > 0 {
				marker = datas.Details[len(datas.Details)-1].CloudID
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas.Details {
					if data.CloudID == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if int32(len(datas.Details)) < limit {
				break
			}
		}

		if len(realDeleteIDs) > 0 {
			err := syncEipDelete(kt, realDeleteIDs, dataCli)
			if err != nil {
				logs.Errorf("request syncEipDelete failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func syncHuaWeiEipAdd(kt *kit.Kit, addIDs []string, req *SyncHuaWeiEipOption,
	cloudMap map[string]*HuaWeiEipSync, dataCli *dataservice.Client) error {

	var createReq dataproto.EipExtBatchCreateReq[dataproto.HuaWeiEipExtensionCreateReq]

	for _, id := range addIDs {
		eip := &dataproto.EipExtCreateReq[dataproto.HuaWeiEipExtensionCreateReq]{
			CloudID:    id,
			Region:     req.Region,
			AccountID:  req.AccountID,
			Name:       cloudMap[id].Eip.Name,
			InstanceId: cloudMap[id].Eip.InstanceId,
			Status:     converter.PtrToVal(cloudMap[id].Eip.Status),
			PublicIp:   converter.PtrToVal(cloudMap[id].Eip.PublicIp),
			PrivateIp:  converter.PtrToVal(cloudMap[id].Eip.PrivateIp),
			Extension: &dataproto.HuaWeiEipExtensionCreateReq{
				PortID:              cloudMap[id].Eip.PortID,
				BandwidthId:         cloudMap[id].Eip.BandwidthId,
				BandwidthName:       cloudMap[id].Eip.BandwidthName,
				BandwidthSize:       cloudMap[id].Eip.BandwidthSize,
				EnterpriseProjectId: cloudMap[id].Eip.EnterpriseProjectId,
				Type:                cloudMap[id].Eip.Type,
				BandwidthShareType:  cloudMap[id].Eip.BandwidthShareType,
				ChargeMode:          cloudMap[id].Eip.ChargeMode,
			},
		}
		createReq = append(createReq, eip)
	}

	if len(createReq) > 0 {
		_, err := dataCli.HuaWei.BatchCreateEip(kt.Ctx, kt.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create huawei eip failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func isHuaWeiEipChange(db *HuaWeiDSEipSync, cloud *HuaWeiEipSync) bool {

	if converter.PtrToVal(cloud.Eip.Status) != db.Eip.Status {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Eip.InstanceId, db.Eip.InstanceID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Eip.PortID, db.Eip.Extension.PortID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Eip.BandwidthId, db.Eip.Extension.BandwidthId) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Eip.BandwidthName, db.Eip.Extension.BandwidthName) {
		return true
	}

	if !assert.IsPtrInt32Equal(cloud.Eip.BandwidthSize, db.Eip.Extension.BandwidthSize) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Eip.EnterpriseProjectId, db.Eip.Extension.EnterpriseProjectId) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.Eip.Type, db.Eip.Extension.Type) {
		return true
	}

	if cloud.Eip.BandwidthShareType != db.Eip.Extension.BandwidthShareType {
		return true
	}

	if cloud.Eip.ChargeMode != db.Eip.Extension.ChargeMode {
		return true
	}

	return false
}

func syncHuaWeiEipUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*HuaWeiEipSync,
	dsMap map[string]*HuaWeiDSEipSync, dataCli *dataservice.Client) error {

	var updateReq dataproto.EipExtBatchUpdateReq[dataproto.HuaWeiEipExtensionUpdateReq]

	for _, id := range updateIDs {

		if !isHuaWeiEipChange(dsMap[id], cloudMap[id]) {
			continue
		}

		eip := &dataproto.EipExtUpdateReq[dataproto.HuaWeiEipExtensionUpdateReq]{
			ID:         dsMap[id].Eip.ID,
			Status:     converter.PtrToVal(cloudMap[id].Eip.Status),
			InstanceId: cloudMap[id].Eip.InstanceId,
			Extension: &dataproto.HuaWeiEipExtensionUpdateReq{
				PortID:              cloudMap[id].Eip.PortID,
				BandwidthId:         cloudMap[id].Eip.BandwidthId,
				BandwidthName:       cloudMap[id].Eip.BandwidthName,
				BandwidthSize:       cloudMap[id].Eip.BandwidthSize,
				EnterpriseProjectId: cloudMap[id].Eip.EnterpriseProjectId,
				Type:                cloudMap[id].Eip.Type,
				BandwidthShareType:  cloudMap[id].Eip.BandwidthShareType,
				ChargeMode:          cloudMap[id].Eip.ChargeMode,
			},
		}

		nilStr := ""
		if cloudMap[id].Eip.InstanceId == nil {
			eip.InstanceId = converter.ValToPtr(nilStr)
		}

		updateReq = append(updateReq, eip)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.HuaWei.BatchUpdateEip(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateEip failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getHuaWeiEipDSSync(kt *kit.Kit, cloudIDs []string, req *SyncHuaWeiEipOption,
	dataCli *dataservice.Client) ([]string, map[string]*HuaWeiDSEipSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*HuaWeiDSEipSync)

	if len(cloudIDs) <= 0 {
		return updateIDs, dsMap, nil
	}

	start := 0
	for {
		dataReq := &dataproto.EipListReq{
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
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "cloud_id",
						Op:    filter.In.Factory(),
						Value: cloudIDs,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		results, err := dataCli.HuaWei.ListEip(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list eip failed, err: %v, rid: %s", err, kt.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(HuaWeiDSEipSync)
				dsImageSync.Eip = detail
				dsMap[detail.CloudID] = dsImageSync
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return updateIDs, dsMap, nil
}

func getHuaWeiEipAllDS(kt *kit.Kit, req *SyncHuaWeiEipOption,
	dataCli *dataservice.Client) ([]string, error) {

	start := 0
	dsIDs := make([]string, 0)
	for {
		dataReq := &dataproto.EipListReq{
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
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
				},
			},
			Page: &apicore.BasePage{
				Start: uint32(start),
				Limit: apicore.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.HuaWei.ListEip(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list eip failed, err: %v, rid: %s", err, kt.Rid)
			return dsIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsIDs = append(dsIDs, detail.CloudID)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsIDs, nil
}
