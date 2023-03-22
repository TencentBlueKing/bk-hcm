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
	typecore "hcm/pkg/adaptor/types/core"
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

// SyncGcpEipOption define sync gcp eip option.
type SyncGcpEipOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncGcpEipOption
func (opt SyncGcpEipOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncGcpEip sync eip self
func SyncGcpEip(kt *kit.Kit, req *SyncGcpEipOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := ad.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	cloudAllIDs := make(map[string]bool)
	for {
		opt := &eip.GcpEipListOption{
			Region: req.Region,
			Page: &typecore.GcpPage{
				PageToken: nextToken,
				PageSize:  int64(filter.DefaultMaxInLimit),
			},
		}
		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}

		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
			opt.Page = nil
		}

		datas, err := client.ListEip(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list gcp eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*GcpEipSync)
		cloudIDs := make([]string, 0, len(datas.Details))
		for _, data := range datas.Details {
			eipSync := new(GcpEipSync)
			eipSync.IsUpdate = false
			eipSync.Eip = data
			cloudMap[data.CloudID] = eipSync
			cloudIDs = append(cloudIDs, data.CloudID)
			cloudAllIDs[data.CloudID] = true
		}

		updateIDs, dsMap, err := getGcpEipDSSync(kt, cloudIDs, req, dataCli)
		if err != nil {
			logs.Errorf("request getGcpEipDSSync failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := syncGcpEipUpdate(kt, updateIDs, cloudMap, dsMap, dataCli)
			if err != nil {
				logs.Errorf("request syncGcpEipUpdate failed, err: %v, rid: %s", err, kt.Rid)
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
			err := syncGcpEipAdd(kt, addIDs, req, cloudMap, dataCli)
			if err != nil {
				logs.Errorf("request syncGcpEipAdd failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		if len(datas.NextPageToken) == 0 {
			break
		}
		nextToken = datas.NextPageToken
	}

	dsIDs, err := getGcpEipAllDS(kt, req, dataCli)
	if err != nil {
		logs.Errorf("request getGcpEipAllDS failed, err: %v, rid: %s", err, kt.Rid)
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

		nextToken := ""
		for {
			opt := &eip.GcpEipListOption{
				Region: req.Region,
				Page: &typecore.GcpPage{
					PageToken: nextToken,
					PageSize:  int64(filter.DefaultMaxInLimit),
				},
			}

			if nextToken != "" {
				opt.Page.PageToken = nextToken
			}

			if len(req.CloudIDs) > 0 {
				opt.CloudIDs = req.CloudIDs
			}

			datas, err := client.ListEip(kt, opt)
			if err != nil {
				logs.Errorf("request adaptor to list gcp eip failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
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

			if len(datas.NextPageToken) == 0 {
				break
			}
			nextToken = datas.NextPageToken
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

func syncGcpEipAdd(kt *kit.Kit, addIDs []string, req *SyncGcpEipOption,
	cloudMap map[string]*GcpEipSync, dataCli *dataservice.Client) error {

	var createReq dataproto.EipExtBatchCreateReq[dataproto.GcpEipExtensionCreateReq]

	for _, id := range addIDs {
		eip := &dataproto.EipExtCreateReq[dataproto.GcpEipExtensionCreateReq]{
			CloudID:   id,
			Region:    req.Region,
			AccountID: req.AccountID,
			Name:      cloudMap[id].Eip.Name,
			Status:    converter.PtrToVal(cloudMap[id].Eip.Status),
			PublicIp:  converter.PtrToVal(cloudMap[id].Eip.PublicIp),
			PrivateIp: converter.PtrToVal(cloudMap[id].Eip.PrivateIp),
			Extension: &dataproto.GcpEipExtensionCreateReq{
				AddressType:  cloudMap[id].Eip.AddressType,
				Description:  cloudMap[id].Eip.Description,
				IpVersion:    cloudMap[id].Eip.IpVersion,
				NetworkTier:  cloudMap[id].Eip.NetworkTier,
				PrefixLength: cloudMap[id].Eip.PrefixLength,
				Purpose:      cloudMap[id].Eip.Purpose,
				Network:      cloudMap[id].Eip.Network,
				Subnetwork:   cloudMap[id].Eip.Subnetwork,
				SelfLink:     cloudMap[id].Eip.SelfLink,
				Users:        cloudMap[id].Eip.Users,
			},
		}
		createReq = append(createReq, eip)
	}

	if len(createReq) > 0 {
		_, err := dataCli.Gcp.BatchCreateEip(kt.Ctx, kt.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create gcp eip failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func isGcpEipChange(db *GcpDSEipSync, cloud *GcpEipSync) bool {

	if converter.PtrToVal(cloud.Eip.Status) != db.Eip.Status {
		return true
	}

	if cloud.Eip.AddressType != db.Eip.Extension.AddressType {
		return true
	}

	if cloud.Eip.Description != db.Eip.Extension.Description {
		return true
	}

	if cloud.Eip.IpVersion != db.Eip.Extension.IpVersion {
		return true
	}

	if cloud.Eip.NetworkTier != db.Eip.Extension.NetworkTier {
		return true
	}

	if cloud.Eip.PrefixLength != db.Eip.Extension.PrefixLength {
		return true
	}

	if cloud.Eip.Purpose != db.Eip.Extension.Purpose {
		return true
	}

	if cloud.Eip.Network != db.Eip.Extension.Network {
		return true
	}

	if cloud.Eip.Subnetwork != db.Eip.Extension.Subnetwork {
		return true
	}

	if cloud.Eip.SelfLink != db.Eip.Extension.SelfLink {
		return true
	}

	if !assert.IsStringSliceEqual(db.Eip.Extension.Users, cloud.Eip.Users) {
		return true
	}

	return false
}

func syncGcpEipUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*GcpEipSync,
	dsMap map[string]*GcpDSEipSync, dataCli *dataservice.Client) error {

	var updateReq dataproto.EipExtBatchUpdateReq[dataproto.GcpEipExtensionUpdateReq]

	for _, id := range updateIDs {

		if !isGcpEipChange(dsMap[id], cloudMap[id]) {
			continue
		}

		eip := &dataproto.EipExtUpdateReq[dataproto.GcpEipExtensionUpdateReq]{
			ID:     dsMap[id].Eip.ID,
			Status: converter.PtrToVal(cloudMap[id].Eip.Status),
			Extension: &dataproto.GcpEipExtensionUpdateReq{
				AddressType:  cloudMap[id].Eip.AddressType,
				Description:  cloudMap[id].Eip.Description,
				IpVersion:    cloudMap[id].Eip.IpVersion,
				NetworkTier:  cloudMap[id].Eip.NetworkTier,
				PrefixLength: cloudMap[id].Eip.PrefixLength,
				Purpose:      cloudMap[id].Eip.Purpose,
				Network:      cloudMap[id].Eip.Network,
				Subnetwork:   cloudMap[id].Eip.Subnetwork,
				SelfLink:     cloudMap[id].Eip.SelfLink,
				Users:        cloudMap[id].Eip.Users,
			},
		}

		updateReq = append(updateReq, eip)
	}

	if len(updateReq) > 0 {
		if _, err := dataCli.Gcp.BatchUpdateEip(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateEip failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getGcpEipDSSync(kt *kit.Kit, cloudIDs []string, req *SyncGcpEipOption,
	dataCli *dataservice.Client) ([]string, map[string]*GcpDSEipSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*GcpDSEipSync)

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
						Value: enumor.Gcp,
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

		results, err := dataCli.Gcp.ListEip(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list eip failed, err: %v, rid: %s", err, kt.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(GcpDSEipSync)
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

func getGcpEipAllDS(kt *kit.Kit, req *SyncGcpEipOption, dataCli *dataservice.Client) ([]string, error) {
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
						Value: enumor.Gcp,
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

		results, err := dataCli.Gcp.ListEip(kt.Ctx, kt.Header(), dataReq)
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
